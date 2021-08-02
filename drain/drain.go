package drain

import (
	"context"
	"fmt"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"

	tcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

// Drain cordons and drains a Kubernetes Node.
func Drain(client kubernetes.Interface, node string) error {
	_, err := getPodsForDeletion(context.TODO(), client, node)

	nodeClient := client.CoreV1().Nodes()
	if err := Cordon(context.TODO(), nodeClient, node); err != nil {
		return err
	}

	return err
}

func getPodsForDeletion(ctx context.Context, client kubernetes.Interface, node string) ([]v1.Pod, error) {
	pods := []v1.Pod{}

	podList, err := client.CoreV1().Pods(v1.NamespaceAll).List(ctx, metav1.ListOptions{
		FieldSelector: fields.SelectorFromSet(fields.Set{"spec.nodeName": node}).String(),
	})
	if err != nil {
		return podList.Items, err
	}

	return pods, nil
}

// Cordon marks the given Kubernetes Node unschedulable.
func Cordon(ctx context.Context, nodeClient tcorev1.NodeInterface, node string) error {
	return setUnschedulable(ctx, nodeClient, node, true)
}

// Uncordon marks the given Kubernetes Node schedulable.
func Uncordon(ctx context.Context, nodeClient tcorev1.NodeInterface, node string) error {
	return setUnschedulable(ctx, nodeClient, node, false)
}

// setUnschedulable updates a Node's spec to mark it unschedulable or not.
func setUnschedulable(ctx context.Context, nodeClient tcorev1.NodeInterface, node string, unschedule bool) error {
	patch := []byte(fmt.Sprintf("{\"spec\":{\"unschedulable\":%t}}", unschedule))
	_, err := nodeClient.Patch(ctx, node, types.StrategicMergePatchType, patch, metav1.PatchOptions{})
	return err
}
