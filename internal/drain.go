package fleetlock

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/poseidon/fleetlock/internal/drainer"
)

// DrainNode matches a Zincati request to a node, cordons the node, and evicts
// its pods.
func (s *Server) DrainNode(ctx context.Context, id string) error {
	// match Zincati ID to Kubernetes Node
	node, err := s.matchNode(ctx, id)
	if err != nil {
		return err
	}

	drainer := drain.New(&drain.Config{
		Client: s.kubeClient,
		Logger: s.log,
	})
	return drainer.Drain(ctx, node.GetName())
}

// UncordonNode uncordons a Kubernetes Node that matches the Zincati request ID.
func (s *Server) UncordonNode(ctx context.Context, id string) error {
	// match Zincati ID to Kubernetes Node
	node, err := s.matchNode(ctx, id)
	if err != nil {
		return err
	}

	drainer := drain.New(&drain.Config{
		Client: s.kubeClient,
		Logger: s.log,
	})
	return drainer.Uncordon(ctx, node.GetName())
}

// MatchNode matches a Zincati request ID to a Kubernetes Node.
// See ZincatiID for how Zincati and systemd compute IDs.
func (s *Server) matchNode(ctx context.Context, id string) (*v1.Node, error) {
	fields := logrus.Fields{
		"id": id,
	}

	s.log.WithFields(fields).Info("fleetlock: match Zincati request to Kubernetes node")

	nodes, err := s.kubeClient.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		s.log.WithFields(fields).Infof("fleetlock: nodes list error: %v", err)
		return nil, err
	}

	for _, node := range nodes.Items {
		zincatiID, err := ZincatiID(node.Status.NodeInfo.SystemUUID)
		if err == nil && id == zincatiID {
			fields["node"] = node.GetName()
			s.log.WithFields(fields).Info("fleetlock: Zincati request matches Kubernetes node")
			return &node, nil
		}
	}

	s.log.WithFields(fields).Info("fleetlock: Zincati request matches no Kubernetes Nodes")
	return nil, fmt.Errorf("fleetlock: Zincati request matches no Kubernetes Nodes")
}
