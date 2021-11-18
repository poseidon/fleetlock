package fleetlock

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Config configures a Fleetlock server.
type Config struct {
	// logger
	Logger *logrus.Logger
}

// Server implements the FleetLock protocol.
type Server struct {
	// logger
	log *logrus.Logger
	// metrics
	metrics *metrics

	// Kubernetes
	namespace  string
	kubeClient kubernetes.Interface
}

// NewServer returns a new fleetlock Server handler
func NewServer(config *Config) (http.Handler, error) {
	if config.Logger == nil {
		return nil, fmt.Errorf("fleetlock: logger must not be nil")
	}

	// set via downward API
	namespace := os.Getenv("NAMESPACE")
	if namespace == "" {
		namespace = "default"
	}

	// set for development
	kubeconfigPath := os.Getenv("KUBECONFIG")

	// Kubernetes client from kubeconfig or service account (in-cluster)
	kubeClient, err := newKubeClient(kubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("fleetlock: error creating Kubernetes client: %v", err)
	}

	// create prometheus registry
	registry := prometheus.NewRegistry()
	err = registerAll(registry,
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)
	if err != nil {
		return nil, fmt.Errorf("fleetlock: register collectors error: %v", err)
	}

	metrics := newMetrics()
	err = metrics.Register(registry)
	if err != nil {
		return nil, fmt.Errorf("fleetlock: register metrics error: %v", err)
	}

	s := &Server{
		log:        config.Logger,
		metrics:    metrics,
		namespace:  namespace,
		kubeClient: kubeClient,
	}

	mux := http.NewServeMux()
	chain := func(next http.Handler) http.Handler {
		return POSTHandler(HeaderHandler(fleetLockHeaderKey, "true", next))
	}
	mux.Handle("/v1/pre-reboot", chain(http.HandlerFunc(s.lock)))
	mux.Handle("/v1/steady-state", chain(http.HandlerFunc(s.unlock)))
	mux.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	mux.Handle("/-/healthy", healthHandler())
	return mux, nil
}

// newRebootLease creates a reboot lease.
func (s *Server) newRebootLease(group string) *RebootLease {
	return &RebootLease{
		Meta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("fleetlock-%s", group),
			Namespace: s.namespace,
		},
		Client: s.kubeClient.CoordinationV1(),
	}
}

// lock attempts to obtain a reboot lease lock.
func (s *Server) lock(w http.ResponseWriter, req *http.Request) {
	// decode Message from request
	msg, err := decodeMessage(req)
	if err != nil {
		s.log.Errorf("fleetlock: error decoding message: %v", err)
		encodeReply(w, NewReply(KindDecodeError, "error decoding message"))
		return
	}
	id := msg.ClientParmas.ID
	group := msg.ClientParmas.Group
	rebootLease := s.newRebootLease(group)

	fields := logrus.Fields{
		"id":    id,
		"group": group,
	}

	s.log.WithFields(fields).Info("fleetlock: attempt reboot lease lock")
	s.metrics.lockRequests.Inc()

	// get or create a reboot lease
	ctx := context.Background()
	lock, err := rebootLease.Get(ctx)
	if err != nil {
		if !errors.IsNotFound(err) {
			s.log.Errorf("fleetlock: error getting reboot lease %s: %v", rebootLease.Name(), err)
			encodeReply(w, NewReply(KindInternalError, "error getting reboot lease"))
			return
		}
	}

	fields["holder"] = lock.Holder

	// reboot lease already owned by node
	if lock.Holder == id {
		s.log.WithFields(fields).Info("fleetlock: retained reboot lease")
		s.metrics.lockState.With(prometheus.Labels{"group": group}).Set(1)
		fmt.Fprint(w, "retained reboot lease")

		// best effort, do not gate on drain succeeding
		_ = s.DrainNode(ctx, id)
		return
	}

	// reboot lease available
	if lock.Holder == "" {
		// obtain the reboot lease lock
		s.log.WithFields(fields).Info("fleetlock: reboot lease available, attempt")
		update := &RebootLock{
			Holder:           id,
			LeaseTransitions: lock.LeaseTransitions + 1,
		}
		err := rebootLease.Update(ctx, update)
		if err == nil {
			s.log.WithFields(fields).Info("fleetlock: obtained reboot lease")
			s.metrics.lockState.With(prometheus.Labels{"group": group}).Set(1)
			fmt.Fprintf(w, "obtained reboot lease")

			// best effort, do not gate on drain succeeding
			_ = s.DrainNode(ctx, id)
			return
		}
		s.log.WithFields(fields).Errorf("fleetlock: error obtaining reboot lease: %v", err)
	}

	// reboot lease held by different node
	s.log.WithFields(fields).Info("fleetlock: reboot lease lock unavailable")
	s.metrics.lockState.With(prometheus.Labels{"group": group}).Set(1)
	encodeReply(w, NewReply(KindLockHeld, "reboot lease lock unavailable, held by %s", lock.Holder))
}

// unlock attempts to release a reboot lease lock.
func (s *Server) unlock(w http.ResponseWriter, req *http.Request) {
	// decode Message from request
	msg, err := decodeMessage(req)
	if err != nil {
		s.log.Errorf("fleetlock: error decoding message: %v", err)
		encodeReply(w, NewReply(KindDecodeError, "error decoding message"))
		return
	}
	id := msg.ClientParmas.ID
	group := msg.ClientParmas.Group
	rebootLease := s.newRebootLease(group)

	fields := logrus.Fields{
		"id":    id,
		"group": group,
	}

	s.log.WithFields(fields).Info("fleetlock: attempt reboot lease unlock")
	s.metrics.unlockRequests.Inc()

	// get or create a reboot lease
	ctx := context.Background()
	lock, err := rebootLease.Get(ctx)
	if err != nil {
		if !errors.IsNotFound(err) {
			s.log.Errorf("fleetlock: error getting reboot lease %s: %v", rebootLease.Name(), err)
			encodeReply(w, NewReply(KindInternalError, "error getting reboot lease"))
			return
		}
	}

	// reboot lease is owned by node
	if lock.Holder == id {
		err := s.UncordonNode(ctx, id)
		if err != nil {
			s.log.Errorf("fleetlock: error uncordoning node: %v", err)
			encodeReply(w, NewReply(KindInternalError, "error uncordoning node"))
			return
		}

		// release reboot lease lock
		s.log.WithFields(fields).Info("fleetlock: unlock reboot lease")
		update := &RebootLock{
			Holder:           "",
			LeaseTransitions: lock.LeaseTransitions,
		}
		err = rebootLease.Update(ctx, update)
		if err != nil {
			s.log.WithFields(fields).Errorf("fleetlock: error unlocking reboot lease: %v", err)
			encodeReply(w, NewReply(KindInternalError, "error unlocking reboot lease"))
			return
		}

		s.metrics.lockState.With(prometheus.Labels{"group": group}).Set(0)
		s.metrics.lockTransitions.With(prometheus.Labels{"group": group}).Inc()
		s.log.WithFields(fields).Info("fleetlock: unlocked reboot lease")
		fmt.Fprintf(w, "unlocked reboot lease for %s", lock.Holder)
		return
	}

	// reboot lease available
	if lock.Holder == "" {
		s.metrics.lockState.With(prometheus.Labels{"group": group}).Set(0)
		fmt.Fprint(w, "reboot lease already unlocked")
		return
	}

	// reboot lease held by different node
	s.log.WithFields(fields).Info("fleetlock: reboot lease unlock unavailable")
	s.metrics.lockState.With(prometheus.Labels{"group": group}).Set(1)
	encodeReply(w, NewReply(KindLockHeld, "reboot lease unlock unavailable, held by %s", lock.Holder))
}

// healthHandler handles liveness checks with an ok status response.
func healthHandler() http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, "ok")
	}
	return http.HandlerFunc(fn)
}
