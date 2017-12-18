package service

import (
	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	kubeutil "github.com/mlab-lattice/system/pkg/backend/kubernetes/util/kubernetes"

	corev1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/api/errors"
)

func (c *Controller) syncServiceKubeService(service *crv1.Service) (*corev1.Service, error) {
	name := kubeutil.GetKubeServiceNameForService(service)
	kubeService, err := c.kubeServiceLister.Services(service.Namespace).Get(name)
	if err != nil && errors.IsNotFound(err) {
		return nil, nil
	}

	return kubeService, err
}
