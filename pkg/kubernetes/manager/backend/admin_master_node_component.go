package backend

import (
	"fmt"
	"io"

	"github.com/mlab-lattice/system/pkg/kubernetes/constants"

	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (kb *KubernetesBackend) GetMasterNodeComponents(nodeId string) ([]string, error) {
	_, err := kb.getMasterNode(nodeId)
	if err != nil {
		return nil, err
	}

	// TODO: at some point we may want to check to see what components are actually running on the node
	// For now we'll assume that all master nodes run all of the components.
	components := []string{
		constants.MasterNodeComponentLatticeControllerMaster,
		constants.MasterNodeComponentManagerApi,
	}
	return components, nil
}

func (kb *KubernetesBackend) GetMasterNodeComponentLog(nodeId, componentName string, follow bool) (io.ReadCloser, error) {
	componentPod, err := kb.getMasterNodeComponentPod(nodeId, componentName)
	if err != nil {
		return nil, err
	}

	req := kb.KubeClientset.CoreV1().
		Pods(componentPod.Namespace).
		GetLogs(componentPod.Name, &corev1.PodLogOptions{Follow: follow})

	return req.Stream()
}

func (kb *KubernetesBackend) RestartMasterNodeComponent(nodeId, componentName string) error {
	componentPod, err := kb.getMasterNodeComponentPod(nodeId, componentName)
	if err != nil {
		return err
	}

	return kb.KubeClientset.CoreV1().
		Pods(componentPod.Namespace).
		Delete(componentPod.Name, &metav1.DeleteOptions{})
}

func (kb *KubernetesBackend) getMasterNode(nodeId string) (*corev1.Node, error) {
	masterNodeLabel := constants.MasterNodeLabelID + "=" + nodeId
	nodes, err := kb.KubeClientset.CoreV1().Nodes().List(metav1.ListOptions{
		LabelSelector: masterNodeLabel,
	})
	if err != nil {
		return nil, err
	}

	if len(nodes.Items) == 0 {
		return nil, fmt.Errorf("invalid node ID %v", nodeId)
	}

	if len(nodes.Items) > 1 {
		return nil, fmt.Errorf("more than one node tagged with %v", masterNodeLabel)
	}

	return &nodes.Items[0], nil
}

func (kb *KubernetesBackend) getMasterNodeComponentPod(nodeId, componentName string) (*corev1.Pod, error) {
	podsClient := kb.KubeClientset.CoreV1().Pods(constants.NamespaceLatticeInternal)
	pods, err := podsClient.List(metav1.ListOptions{
		LabelSelector: constants.MasterNodeLabelComponent,
	})
	if err != nil {
		return nil, err
	}

	masterNode, err := kb.getMasterNode(nodeId)
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

	return nil, fmt.Errorf("component %v is not running on node %v", componentName, nodeId)
}
