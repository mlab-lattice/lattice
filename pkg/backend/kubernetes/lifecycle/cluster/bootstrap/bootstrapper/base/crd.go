package base

import (
	latticev1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/cluster/bootstrap/bootstrapper"
)

func (b *DefaultBootstrapper) crdResources(resources *bootstrapper.ClusterResources) {
	customResourceDefinitions := latticev1.GetCustomResourceDefinitions()
	resources.CustomResourceDefinitions = append(resources.CustomResourceDefinitions, customResourceDefinitions...)
}
