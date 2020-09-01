package fleetlock

import (
	"fmt"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// newKubeClient creates a Kubernetes client using a kubeconfig at the given
// path or using the Pod service account (i.e. in-cluster).
func newKubeClient(kubePath string) (kubernetes.Interface, error) {
	// Kubernetes REST client config
	config, err := clientcmd.BuildConfigFromFlags("", kubePath)
	if err != nil {
		return nil, fmt.Errorf("fleetlock: error getting Kubernetes client config: %v", err)
	}

	// create Kubernetes client
	return kubernetes.NewForConfig(config)
}
