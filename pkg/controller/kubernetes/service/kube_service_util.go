package service

import (
	"fmt"

	systemdefinitionblock "github.com/mlab-lattice/core/pkg/system/definition/block"

	crv1 "github.com/mlab-lattice/kubernetes-integration/pkg/api/customresource/v1"

	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func (sc *ServiceController) getKubeService(svc *crv1.Service) (*corev1.Service, error) {
	ports := []corev1.ServicePort{}
	for componentName, cPorts := range svc.Spec.Ports {
		for _, port := range cPorts {
			protocol, err := getProtocol(port.Protocol)
			if err != nil {
				return nil, err
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
			Name:            svc.Name,
			Namespace:       svc.Namespace,
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(svc, controllerKind)},
		},
		Spec: corev1.ServiceSpec{
			Ports: ports,
			Selector: map[string]string{
				crv1.ServiceDeploymentLabelKey: svc.Name,
			},
			ClusterIP: "None",
			Type:      corev1.ServiceTypeClusterIP,
		},
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
