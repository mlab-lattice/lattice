package kubernetes

import (
	"fmt"

	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
)

func NodePoolIDLabelValue(nodePool *crv1.NodePool) string {
	return fmt.Sprintf("%v/%v", nodePool.Namespace, nodePool.Name)
}
