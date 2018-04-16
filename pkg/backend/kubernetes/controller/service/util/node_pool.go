package util

import (
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/constants"
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	latticeclientset "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/generated/clientset/versioned"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
)

func ServicesForNodePool(latticeClient latticeclientset.Interface, nodePool *latticev1.NodePool) ([]latticev1.Service, error) {
	// TODO(kevinrosendahl): will have to change query's namespace when supporting cluster-level node pools
	selector := labels.NewSelector()
	requirement, err := labels.NewRequirement(constants.LabelKeyNodeRoleLatticeNodePool, selection.Equals, []string{nodePool.IDLabelValue()})
	if err != nil {
		return nil, err
	}
	selector = selector.Add(*requirement)

	services, err := latticeClient.LatticeV1().Services(nodePool.Namespace).List(metav1.ListOptions{LabelSelector: selector.String()})
	if err != nil {
		return nil, err
	}

	return services.Items, nil
}
