package terraform

import (
	"fmt"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
)

func GetS3BackendStatePathRoot(latticeID v1.LatticeID) string {
	return fmt.Sprintf("lattice/%v", latticeID)
}

func GetS3BackendNamespaceStatePathRoot(latticeID v1.LatticeID, namespace string) string {
	return fmt.Sprintf("%v/system/%v/aws/terraform/state", GetS3BackendStatePathRoot(latticeID), namespace)
}

func GetS3BackendNodePoolPathRoot(latticeID v1.LatticeID, namespace, nodePoolID string) string {
	return fmt.Sprintf("%v/node-pool/%v", GetS3BackendNamespaceStatePathRoot(latticeID, namespace), nodePoolID)
}

func GetS3BackendAddressPathRoot(latticeID v1.LatticeID, namespace, addressID string) string {
	return fmt.Sprintf("%v/address/%v", GetS3BackendNamespaceStatePathRoot(latticeID, namespace), addressID)
}

func GetS3BackendServiceAddressLoadBalancerPathRoot(latticeID v1.LatticeID, namespace, addressID string) string {
	return fmt.Sprintf("%v/service-load-balancer", GetS3BackendAddressPathRoot(latticeID, namespace, addressID))
}

func GetS3BackendRoute53PathRoot(latticeID v1.LatticeID, zoneID string) string {
	return fmt.Sprintf("%v/route53/%v", GetS3BackendStatePathRoot(latticeID), zoneID)
}
