package drain

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

const (
	// Wait up to this amount of time for pods to be evicted from a node.
	drainMaxWait = 60 * time.Second
)

// Config configures a Drainer.
type Config struct {
	Client kubernetes.Interface
	Logger *logrus.Logger
}

// Drainer manages cordoning nodes and evicting Pods.
type Drainer interface {
	// Drain cordons a node and evicts its Pods.
	Drain(ctx context.Context, node string) error
	// Cordon marks a Kubernetes Node as unschedulable.
	Cordon(ctx context.Context, node string) error
	// Uncordon marks a Kubernetes Node as schedulable.
	Uncordon(ctx context.Context, node string) error
}

// New returns a new Drainer.
func New(config *Config) Drainer {
	return &drainer{
		client: config.Client,
		log:    config.Logger,
	}
}

// drain is a Kubernetes node cordon and drainer.
type drainer struct {
	client kubernetes.Interface
	log    *logrus.Logger
}

// Cordon marks a Kubernetes Node as unschedulable.
func (d *drainer) Cordon(ctx context.Context, node string) error {
	d.log.WithField("node", node).Info("drainer: cordoning node")
	return d.setUnschedulable(ctx, node, true)
}

// Uncordon marks a Kubernetes Node as schedulable.
func (d *drainer) Uncordon(ctx context.Context, node string) error {
	d.log.WithField("node", node).Info("drainer: uncordoning node")
	return d.setUnschedulable(ctx, node, false)
}

// Drain drains a Kubernetes Node.
func (d *drainer) Drain(ctx context.Context, node string) error {
	fields := logrus.Fields{
		"node": node,
	}

	if err := d.Cordon(ctx, node); err != nil {
		d.log.WithFields(fields).Errorf("drainer: error cordoning node: %v", err)
		return err
	}

	d.log.WithFields(fields).Info("drainer: draining node")

	pods, err := d.getPodsForDeletion(ctx, node)
	if err != nil {
		d.log.WithFields(fields).Errorf("drainer: error getting pods: %v", err)
		return err
	}

	evictedPods := []string{}
	for _, pod := range pods {
		fields["pod"] = pod.GetName()
		d.log.WithFields(fields).Info("drainer: evicting pod")

		err := d.evictPod(ctx, pod)
		if err != nil {
			d.log.WithFields(fields).Errorf("drainer: error evicting pod: %v", err)
			return err
		}
		evictedPods = append(evictedPods, pod.GetName())
	}

	start := time.Now()
	for len(evictedPods) > 0 {
		podsDescription := ""
		if len(evictedPods) > 5 {
			podsDescription = strings.Join(evictedPods[:5], ", ") + " ..."
		} else {
			podsDescription = strings.Join(evictedPods, ", ")
		}
		d.log.WithFields(fields).Infof("drainer: waiting for %d pods to be evicted: %s", len(evictedPods), podsDescription)

		if time.Since(start) > drainMaxWait {
			d.log.WithFields(fields).Infof("drainer: waited maximum amount of time for evictions, continuing")
			break
		}

		pods, err := d.getPodsForDeletion(ctx, node)
		if err != nil {
			d.log.WithFields(fields).Errorf("drainer: error getting pods: %v", err)
			return err
		}
		podsByName := make(map[string]struct{}, len(pods))
		for _, pod := range pods {
			podsByName[pod.GetName()] = struct{}{}
		}
		remainingPods := []string{}
		for _, pod := range evictedPods {
			if _, ok := podsByName[pod]; ok {
				remainingPods = append(remainingPods, pod)
			}
		}
		evictedPods = remainingPods
		if len(evictedPods) > 0 {
			time.Sleep(1 * time.Second)
		}
	}

	d.log.WithFields(fields).Info("drainer: drained node")
	return nil
}

// Lists pods on a node and filters our mirror and daemonset Pods.
func (d *drainer) getPodsForDeletion(ctx context.Context, node string) ([]v1.Pod, error) {
	pods := []v1.Pod{}
	logFields := logrus.Fields{
		"node": node,
	}

	podList, err := d.client.CoreV1().Pods(v1.NamespaceAll).List(ctx, metav1.ListOptions{
		FieldSelector: fields.SelectorFromSet(fields.Set{"spec.nodeName": node}).String(),
	})
	if err != nil {
		return podList.Items, err
	}

	for _, pod := range podList.Items {
		logFields["pod"] = pod.GetName()

		// skip mirror pods
		if isMirrorPod(pod) {
			d.log.WithFields(logFields).Debug("skip mirror pod")
			continue
		}

		// skip daemonset pods
		if isDaemonSetPod(pod) {
			d.log.WithFields(logFields).Debug("skip daemonset pod")
			continue
		}

		pods = append(pods, pod)
	}

	return pods, nil
}

// evictPod tries to create an Eviction of the given Pod.
func (d *drainer) evictPod(ctx context.Context, pod v1.Pod) error {

	// https://pkg.go.dev/k8s.io/api/policy/v1beta1#Eviction
	eviction := &policyv1beta1.Eviction{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "policy/v1beta1",
			Kind:       "Eviction",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      pod.GetName(),
			Namespace: pod.GetNamespace(),
		},
		DeleteOptions: &metav1.DeleteOptions{
			GracePeriodSeconds: pod.Spec.TerminationGracePeriodSeconds,
		},
	}

	return d.client.PolicyV1beta1().Evictions(pod.GetNamespace()).Evict(ctx, eviction)
}

// setUnschedulable updates a Node's spec to mark it unschedulable or not.
func (d *drainer) setUnschedulable(ctx context.Context, node string, unschedule bool) error {
	patch := []byte(fmt.Sprintf("{\"spec\":{\"unschedulable\":%t}}", unschedule))
	_, err := d.client.CoreV1().Nodes().Patch(ctx, node, types.StrategicMergePatchType, patch, metav1.PatchOptions{})
	return err
}

// isMirrorPod returns true if a Pod is a mirror Pod (i.e. annotated with
// `kubernetes.io/config.mirror`)
func isMirrorPod(pod v1.Pod) bool {
	if _, present := pod.ObjectMeta.Annotations[v1.MirrorPodAnnotationKey]; present {
		return true
	}
	return false
}

// isDaemonSetPod returns true if a Pod is owned by a DaemonSet controller
func isDaemonSetPod(pod v1.Pod) bool {
	controller := metav1.GetControllerOf(&pod)
	if controller == nil {
		return false
	}

	return controller.Kind == "DaemonSet"
}
