package service

import (
	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	kubeutil "github.com/mlab-lattice/system/pkg/backend/kubernetes/util/kubernetes"

	corev1 "k8s.io/api/core/v1"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

func (c *Controller) getKubeServiceForService(svc *crv1.Service) (*corev1.Service, bool, error) {
	necessary := false
	for _, component := range svc.Spec.Definition.Components {
		for _, port := range component.Ports {
			if port.ExternalAccess != nil && port.ExternalAccess.Public {
				necessary = true
				break
			}
		}
	}
	if !necessary {
		return nil, false, nil
	}

	ksvcName := kubeutil.GetKubeServiceNameForService(svc)
	ksvc, err := c.kubeServiceLister.Services(svc.Namespace).Get(ksvcName)
	if err != nil && !apierrors.IsNotFound(err) {
		return nil, true, err
	}

	return ksvc, true, nil
}
