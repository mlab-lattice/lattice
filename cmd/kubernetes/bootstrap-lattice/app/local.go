package app

import (
	"fmt"

	"github.com/mlab-lattice/system/pkg/backend/kubernetes/constants"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func seedLocalSpecific(systemID string) {
	nodeClient := kubeClient.CoreV1().Nodes()
	node, err := nodeClient.Get(systemID, metav1.GetOptions{})
	if err != nil {
		panic(err)
	}

	if node == nil {
		panic(fmt.Errorf("could not find node %v", systemID))
	}

	node.Labels[constants.MasterNodeLabelID] = "0"
	_, err = nodeClient.Update(node)
	if err != nil {
		panic(err)
	}
}
