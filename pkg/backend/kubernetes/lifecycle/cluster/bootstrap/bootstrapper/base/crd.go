package base

import (
	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/cluster/bootstrap/bootstrapper"
)

func (b *DefaultBootstrapper) crdResources(resources *bootstrapper.Resources) {
	customResourceDefinitions := crv1.GetCustomResourceDefinitions()
	resources.CustomResourceDefinitions = append(resources.CustomResourceDefinitions, customResourceDefinitions...)
}
