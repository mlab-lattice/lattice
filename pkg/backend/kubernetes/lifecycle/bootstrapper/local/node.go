package local

import (
	"fmt"

	"github.com/mlab-lattice/system/pkg/backend/kubernetes/constants"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (b *DefaultBootstrapper) bootstrapLocalNode() ([]interface{}, error) {
	if b.Options.DryRun {
		return []interface{}{}, nil
	}

	nodes, err := b.KubeClient.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		return []interface{}{}, err
	}

	if len(nodes.Items) != 1 {
		return []interface{}{}, fmt.Errorf("expected exactly 1 node, found %v", len(nodes.Items))
	}

	node := nodes.Items[0]
	node.Labels[constants.MasterNodeLabelID] = "0"
	node.Labels[constants.LabelKeyMasterNode] = "true"

	_, err = b.KubeClient.CoreV1().Nodes().Update(&node)
	return []interface{}{}, err
}
