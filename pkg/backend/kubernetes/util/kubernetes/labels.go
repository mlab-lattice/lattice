package kubernetes

import (
	"fmt"

	latticev1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
)

func NodePoolIDLabelValue(nodePool *latticev1.NodePool) string {
	return fmt.Sprintf("%v.%v", nodePool.Namespace, nodePool.Name)
}
