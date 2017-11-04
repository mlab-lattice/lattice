package backend

import (
	"fmt"

	systemtree "github.com/mlab-lattice/core/pkg/system/tree"
	coretypes "github.com/mlab-lattice/core/pkg/types"

	crv1 "github.com/mlab-lattice/kubernetes-integration/pkg/api/customresource/v1"
	kubeutil "github.com/mlab-lattice/kubernetes-integration/pkg/util/kubernetes"

	corev1 "k8s.io/api/core/v1"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (kb *KubernetesBackend) ListSystemServices(ln coretypes.LatticeNamespace) ([]coretypes.Service, error) {
	result := &crv1.ServiceList{}
	err := kb.LatticeResourceClient.Get().
		Namespace(string(ln)).
		Resource(crv1.ServiceResourcePlural).
		Do().
		Into(result)

	if err != nil {
		return nil, err
	}

	svcs := []coretypes.Service{}
	for _, r := range result.Items {
		coreSvc, err := kb.transformService(&r)
		if err != nil {
			return nil, err
		}

		svcs = append(svcs, *coreSvc)
	}

	return svcs, nil
}

func (kb *KubernetesBackend) GetSystemService(ln coretypes.LatticeNamespace, path systemtree.NodePath) (*coretypes.Service, error) {
	// FIXME: find a way to query this
	svcs, err := kb.ListSystemServices(ln)
	if err != nil {
		return nil, err
	}

	for _, svc := range svcs {
		if svc.Path == path {
			return &svc, nil
		}
	}

	return nil, nil
}

func (kb *KubernetesBackend) transformService(svc *crv1.Service) (*coretypes.Service, error) {
	// FIXME: this only works for local systems with a single port
	kubeSvcName := kubeutil.GetKubeServiceNameForService(svc)
	kubeSvc, err := kb.KubeClientset.CoreV1().Services(svc.Namespace).Get(kubeSvcName, metav1.GetOptions{})

	var addr *string
	if err != nil {
		// If there was a genuine error, return it, otherwise keep addr set to nil
		if !apierrors.IsNotFound(err) {
			return nil, err
		}
	} else {
		if kubeSvc.Spec.Type == corev1.ServiceTypeNodePort {
			// Otherwise we found a kube Service
			sysIP, err := kb.getSystemIP()
			if err != nil {
				return nil, err
			}

			addrt := fmt.Sprintf("%v:%v", *sysIP, kubeSvc.Spec.Ports[0].NodePort)
			addr = &addrt
		}
	}

	coreSvc := &coretypes.Service{
		Id:      svc.Name,
		Path:    svc.Spec.Path,
		Address: addr,
	}
	return coreSvc, nil
}
