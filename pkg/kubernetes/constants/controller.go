package constants

import (
	crv1 "github.com/mlab-lattice/system/pkg/kubernetes/customresource/v1"
)

var (
	ControllerKindKubernetesComponentBuild = crv1.SchemeGroupVersion.WithKind("ComponentBuild")
)
