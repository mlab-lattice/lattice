package service

import (
	"fmt"

	systemdefinitionblock "github.com/mlab-lattice/core/pkg/system/definition/block"

	crv1 "github.com/mlab-lattice/kubernetes-integration/pkg/api/customresource/v1"

	corev1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/golang/glog"
)

func getKubeServiceNameForService(svc *crv1.Service) string {
	// This ensures that the kube Service name is a DNS-1035 label:
	// "a DNS-1035 label must consist of lower case alphanumeric characters or '-',
	//  and must start and end with an alphanumeric character (e.g. 'my-name',
	//  or 'abc-123', regex used for validation is '[a-z]([-a-z0-9]*[a-z0-9])?')"
	return fmt.Sprintf("svc-%v-lattice", svc.Name)
}

func (sc *ServiceController) getKubeServiceForService(svc *crv1.Service) (*corev1.Service, error) {
	ksvc, err := sc.kubeServiceLister.Services(svc.Namespace).Get(getKubeServiceNameForService(svc))
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	return ksvc, nil
}

func (sc *ServiceController) getKubeService(svc *crv1.Service) (*corev1.Service, error) {
	ports := []corev1.ServicePort{}
	public := false
	for componentName, cPorts := range svc.Spec.Ports {
		for _, port := range cPorts {
			protocol, err := getProtocol(port.Protocol)
			if err != nil {
				return nil, err
			}

			if port.Public {
				public = true
			}

			ports = append(ports, corev1.ServicePort{
				Name:       fmt.Sprintf("%v-%v", componentName, port.Name),
				Protocol:   protocol,
				Port:       port.Port,
				TargetPort: intstr.FromInt(int(port.Port)),
			})
		}
	}

	ksvc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            getKubeServiceNameForService(svc),
			Namespace:       svc.Namespace,
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(svc, controllerKind)},
		},
		Spec: corev1.ServiceSpec{
			Ports: ports,
			Selector: map[string]string{
				crv1.LabelKeyServiceDeployment: svc.Name,
			},
			ClusterIP: "None",
			Type:      corev1.ServiceTypeClusterIP,
		},
	}

	if public {
		ksvc.Spec.ClusterIP = ""
		ksvc.Spec.Type = corev1.ServiceTypeNodePort
	}

	return ksvc, nil
}

func getProtocol(protocolString string) (corev1.Protocol, error) {
	switch protocolString {
	case systemdefinitionblock.HttpProtocol, systemdefinitionblock.TcpProtocol:
		return corev1.ProtocolTCP, nil
	default:
		return corev1.ProtocolTCP, fmt.Errorf("invalid protocol %v", protocolString)
	}
}

func (sc *ServiceController) createKubeService(svc *crv1.Service) (*corev1.Service, error) {
	ksvc, err := sc.getKubeService(svc)
	if err != nil {
		return nil, err
	}

	ksvcResp, err := sc.kubeClient.CoreV1().Services(svc.Namespace).Create(ksvc)
	if err != nil {
		// FIXME: send warn event
		return nil, err
	}

	glog.V(4).Infof("Created Service %s", ksvcResp.Name)
	// FIXME: send normal event
	return ksvcResp, nil
}
