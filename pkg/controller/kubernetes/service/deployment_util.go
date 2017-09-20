package service

import (
	"fmt"

	systemdefinitionblock "github.com/mlab-lattice/core/pkg/system/definition/block"

	crv1 "github.com/mlab-lattice/kubernetes-integration/pkg/api/customresource/v1"

	corev1 "k8s.io/api/core/v1"
	extensions "k8s.io/api/extensions/v1beta1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func (sc *ServiceController) getDeployment(svc *crv1.Service, svcBuild *crv1.ServiceBuild) (*extensions.Deployment, error) {
	name := getDeploymentName(svc)
	labels := map[string]string{
		crv1.ServiceDeploymentLabelKey: svc.Name,
	}

	dSpec, err := sc.getDeploymentSpec(svc, svcBuild)
	if err != nil {
		return nil, err
	}

	d := &extensions.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			Labels:          labels,
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(svc, controllerKind)},
		},
		Spec: *dSpec,
	}

	return d, nil
}

func getDeploymentName(svc *crv1.Service) string {
	// TODO: May change this to UUID when a Service can have multiple Deployments (e.g. Blue/Green & Canary)
	return fmt.Sprintf("lattice-service-%s", svc.Name)
}

func (sc *ServiceController) getDeploymentSpec(svc *crv1.Service, svcBuild *crv1.ServiceBuild) (*extensions.DeploymentSpec, error) {
	containers := []corev1.Container{}
	componentDockerImageFqns, err := sc.getComponentDockerImageFqns(svcBuild)
	if err != nil {
		return nil, err
	}

	for _, component := range svc.Spec.Definition.Components {
		ports := []corev1.ContainerPort{}
		for _, port := range component.Ports {
			ports = append(
				ports,
				corev1.ContainerPort{
					Name:          port.Name,
					ContainerPort: int32(port.Port),
				},
			)
		}

		envs := []corev1.EnvVar{}
		for k, v := range component.Exec.Environment {
			envs = append(
				envs,
				corev1.EnvVar{
					Name:  k,
					Value: v,
				},
			)
		}

		container := corev1.Container{
			Name:    component.Name,
			Image:   componentDockerImageFqns[component.Name],
			Command: component.Exec.Command,
			Ports:   ports,
			Env:     envs,
			// TODO: maybe add Resources
			// TODO: add VolumeMounts
			LivenessProbe: getLivenessProbe(component.HealthCheck),
		}

		containers = append(containers, container)
	}

	// Spin up the min instances here, then later let autoscaler scale up.
	replicas := int32(svc.Spec.Definition.Resources.MinInstances)
	ds := extensions.DeploymentSpec{
		Replicas: &replicas,
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Name: getDeploymentName(svc),
				Labels: map[string]string{
					crv1.ServiceDeploymentLabelKey: svc.Name,
				},
			},
			Spec: corev1.PodSpec{
				// TODO: add Volumes
				Containers: containers,
				DNSPolicy:  corev1.DNSDefault,
				// TODO: add NodeSelector (for cloud)
				// TODO: add Tolerations (for cloud)
				// TODO: add HostAliases (for local)
			},
		},
	}

	return &ds, nil
}

func (sc *ServiceController) getComponentDockerImageFqns(svcBuild *crv1.ServiceBuild) (map[string]string, error) {
	componentDockerImageFqns := map[string]string{}

	for cName, cBuildInfo := range svcBuild.Spec.Components {
		if cBuildInfo.BuildName == nil {
			return nil, fmt.Errorf("svcBuild %v Component %v does not have a ComponentBuildName", svcBuild.Name, cName)
		}

		cBuildName := *cBuildInfo.BuildName
		cBuildKey := svcBuild.Namespace + "/" + cBuildName
		cBuildObj, exists, err := sc.componentBuildStore.GetByKey(cBuildKey)

		if err != nil {
			return nil, err
		}

		if !exists {
			return nil, fmt.Errorf("cBuild %v not in cBuild Store", cBuildKey)
		}

		cBuild := cBuildObj.(*crv1.ComponentBuild)

		if cBuild.Spec.Artifacts == nil {
			return nil, fmt.Errorf("cBuild %v does not have Artifacts", cBuildKey)
		}

		componentDockerImageFqns[cName] = cBuild.Spec.Artifacts.DockerImageFqn
	}

	return componentDockerImageFqns, nil
}

func getLivenessProbe(hc *systemdefinitionblock.HealthCheck) *corev1.Probe {
	if hc == nil {
		return nil
	}

	if hc.Exec != nil {
		return &corev1.Probe{
			Handler: corev1.Handler{
				Exec: &corev1.ExecAction{
					Command: hc.Exec.Command,
				},
			},
		}
	}

	if hc.Http != nil {
		headers := []corev1.HTTPHeader{}
		for k, v := range hc.Http.Headers {
			headers = append(
				headers,
				corev1.HTTPHeader{
					Name:  k,
					Value: v,
				},
			)
		}

		return &corev1.Probe{
			Handler: corev1.Handler{
				HTTPGet: &corev1.HTTPGetAction{
					Path:        hc.Http.Path,
					Port:        intstr.FromString(hc.Http.Port),
					HTTPHeaders: headers,
				},
			},
		}
	}

	return &corev1.Probe{
		Handler: corev1.Handler{
			TCPSocket: &corev1.TCPSocketAction{
				Port: intstr.FromString(hc.Tcp.Port),
			},
		},
	}
}
