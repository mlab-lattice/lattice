package util

import (
	"fmt"

	kubeconstants "github.com/mlab-lattice/system/pkg/backend/kubernetes/constants"
	latticev1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	latticeclientset "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/generated/clientset/versioned"
	kubeutil "github.com/mlab-lattice/system/pkg/backend/kubernetes/util/kubernetes"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ServicesForNodePool(latticeClient latticeclientset.Interface, nodePool *latticev1.NodePool) ([]latticev1.Service, error) {
	// TODO(kevinrosendahl): will have to change query's namespace when supporting cluster-level node pools
	nodePoolLabelValue := kubeutil.NodePoolIDLabelValue(nodePool)
	listOptions := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%v == %v", kubeconstants.LabelKeyNodeRoleNodePool, nodePoolLabelValue),
	}

	services, err := latticeClient.LatticeV1().Services(nodePool.Namespace).List(listOptions)
	if err != nil {
		return nil, err
	}

	return services.Items, nil
}
