package latticecontrollers

import (
	controller "github.com/mlab-lattice/kubernetes-integration/cmd/controller-manager/app/common"
)

func GetControllerInitializers() map[string]controller.Initializer {
	return map[string]controller.Initializer{
		"service-build": initializeServiceBuildController,
		"system":        initializeSystemController,
		"system-build":  initializeSystemBuildController,
	}
}
