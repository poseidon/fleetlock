package fleetlock

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/poseidon/fleetlock/drain"
)

// DrainNode cordons and drains a Kubernetes Node that matches the Zincati
// request ID.
func (s *Server) DrainNode(ctx context.Context, id string) error {
	// match Zincati ID to Kubernetes Node
	node, err := s.MatchNode(ctx, id)
	if err != nil {
		return err
	}

	fields := logrus.Fields{
		"id":   id,
		"node": node.GetName(),
	}

	s.log.WithFields(fields).Info("fleetlock: draining node")
	err = drain.Drain(s.kubeClient, node.GetName())
	if err != nil {
		s.log.WithFields(fields).Errorf("fleetlock: drain error: %v", err)
	}
	s.log.WithFields(fields).Info("fleetlock: drained node")

	return err
}

// UncordonNode uncordons a Kubernetes Node that matches the Zincati request ID.
func (s *Server) UncordonNode(ctx context.Context, id string) error {
	// match Zincati ID to Kubernetes Node
	node, err := s.MatchNode(ctx, id)
	if err != nil {
		return err
	}

	fields := logrus.Fields{
		"id":   id,
		"node": node.GetName(),
	}

	s.log.WithFields(fields).Info("fleetlock: uncordoning node")
	err = drain.Uncordon(ctx, s.kubeClient.CoreV1().Nodes(), node.GetName())
	if err != nil {
		s.log.WithFields(fields).Errorf("fleetlock: uncordon error: %v", err)
	}
	s.log.WithFields(fields).Info("fleetlock: uncordoned node")

	return err
}

// MatchNode matches a Zincati request ID to a Kubernetes Node.
// See ZincatiID for how Zincati and systemd compute IDs.
func (s *Server) MatchNode(ctx context.Context, id string) (*v1.Node, error) {
	fields := logrus.Fields{
		"id": id,
	}

	nodes, err := s.kubeClient.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
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
