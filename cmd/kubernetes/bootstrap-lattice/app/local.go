package app

import (
	"fmt"

	"github.com/mlab-lattice/system/pkg/kubernetes/constants"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"
)

func seedLocalSpecific(kubeClientset *kubernetes.Clientset, systemId string) {
	nodeClient := kubeClientset.CoreV1().Nodes()
	node, err := nodeClient.Get(systemId, metav1.GetOptions{})
	if err != nil {
		panic(err)
	}

	if node == nil {
		panic(fmt.Errorf("could not find node %v", systemId))
	}

	node.Labels[constants.MasterNodeLabelID] = "0"
	_, err = nodeClient.Update(node)
	if err != nil {
		panic(err)
	}
}
