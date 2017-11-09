package kubernetescontrollers

import (
	controller "github.com/mlab-lattice/system/cmd/kubernetes/lattice-controller-manager/app/common"
)

func GetControllerInitializers() map[string]controller.Initializer {
	return map[string]controller.Initializer{
		"component-build": initializeComponentBuildController,
		"service":         initializeServiceController,
	}
}
