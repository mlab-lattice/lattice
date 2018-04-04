package base

import (
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/lifecycle/lattice/bootstrap/bootstrapper"
)

func (b *DefaultBootstrapper) crdResources(resources *bootstrapper.Resources) {
	customResourceDefinitions := latticev1.GetCustomResourceDefinitions()
	resources.CustomResourceDefinitions = append(resources.CustomResourceDefinitions, customResourceDefinitions...)
}
