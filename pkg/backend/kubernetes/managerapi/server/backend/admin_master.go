package backend

import (
	"fmt"
	"io"

	"github.com/mlab-lattice/system/pkg/backend/kubernetes/constants"

	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (kb *KubernetesBackend) GetMasterComponents() ([]string, error) {
	// TODO: at some point we may want to check to see what components are actually running on a node
	// For now we'll assume that all master nodes run all of the components.
	components := []string{
		constants.MasterNodeComponentLatticeControllerManager,
		constants.MasterNodeComponentManagerAPI,
	}
	return components, nil
}

func (kb *KubernetesBackend) GetMasterComponentLog(nodeID, componentName string, follow bool) (io.ReadCloser, bool, error) {
	componentPod, err := kb.getMasterNodeComponentPod(nodeID, componentName)
	if err != nil {
		return nil, false, err
	}
	if componentPod == nil {
		return nil, false, nil
	}

	req := kb.KubeClient.CoreV1().
		Pods(componentPod.Namespace).
		GetLogs(componentPod.Name, &corev1.PodLogOptions{Follow: follow})

	log, err := req.Stream()
	return log, true, err
}

func (kb *KubernetesBackend) RestartMasterComponent(nodeID, componentName string) (bool, error) {
	componentPod, err := kb.getMasterNodeComponentPod(nodeID, componentName)
	if err != nil {
		return false, err
	}

	if componentPod == nil {
		return false, nil
	}

	err = kb.KubeClient.CoreV1().
		Pods(componentPod.Namespace).
		Delete(componentPod.Name, &metav1.DeleteOptions{})
	return true, err
}

func (kb *KubernetesBackend) getMasterNode(nodeID string) (*corev1.Node, error) {
	masterNodeLabel := constants.MasterNodeLabelID + "=" + nodeID
	nodes, err := kb.KubeClient.CoreV1().Nodes().List(metav1.ListOptions{
		LabelSelector: masterNodeLabel,
	})
	if err != nil {
		return nil, err
	}

	if len(nodes.Items) == 0 {
		return nil, fmt.Errorf("invalid node ID %v", nodeID)
	}

	if len(nodes.Items) > 1 {
		return nil, fmt.Errorf("more than one node tagged with %v", masterNodeLabel)
	}

	return &nodes.Items[0], nil
}

func (kb *KubernetesBackend) getMasterNodeComponentPod(nodeID, componentName string) (*corev1.Pod, error) {
	podsClient := kb.KubeClient.CoreV1().Pods(constants.NamespaceLatticeInternal)
	pods, err := podsClient.List(metav1.ListOptions{
		LabelSelector: constants.MasterNodeLabelComponent,
	})
	if err != nil {
		return nil, err
	}

	masterNode, err := kb.getMasterNode(nodeID)
	if err != nil {
		return nil, err
	}

	for _, pod := range pods.Items {
		if pod.Labels[constants.MasterNodeLabelComponent] != componentName {
			continue
		}

		if pod.Spec.NodeName == masterNode.Name {
			return &pod, nil
		}
	}

	return nil, nil
}
