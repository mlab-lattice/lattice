package basecontrollers

import (
	controller "github.com/mlab-lattice/lattice/cmd/kubernetes/lattice-controller-manager/app/common"
)

func GetControllerInitializers() map[string]controller.Initializer {
	return map[string]controller.Initializer{
		"component-build": initializeComponentBuildController,
		"service-address": initializeServiceAddressController,
		"service-build":   initializeServiceBuildController,
		"service":         initializeServiceController,
		"system":          initializeSystemController,
		"system-build":    initializeSystemBuildController,
		"system-rollout":  initializeSystemRolloutController,
	}
}
