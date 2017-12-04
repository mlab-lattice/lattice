package system

import (
	"github.com/mlab-lattice/system/pkg/kubernetes/constants"
	crv1 "github.com/mlab-lattice/system/pkg/kubernetes/customresource/v1"
)

func (sc *Controller) removeFinalizer(sys *crv1.System) error {
	// Build up a list of all the finalizers except the system controller finalizer.
	finalizers := []string{}
	found := false
	for _, finalizer := range sys.Finalizers {
		if finalizer == constants.KubeFinalizerSystemController {
			found = true
			continue
		}
		finalizers = append(finalizers, finalizer)
	}

	// If the finalizer wasn't part of the list, nothing to do.
	if !found {
		return nil
	}

	// The finalizer was in the list, so we should remove it.
	sys.Finalizers = finalizers
	result, err := sc.latticeClient.V1().Systems(sys.Namespace).Update(sys)
	if err != nil {
		return err
	}
	*sys = *result
	return nil
}
