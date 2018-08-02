package v1

import (
	"fmt"
	"sort"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	kubeutil "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/util/kubernetes"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/util/latticeutil"
	"github.com/mlab-lattice/lattice/pkg/definition/component"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	definitionv1 "github.com/mlab-lattice/lattice/pkg/definition/v1"

	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func PodTemplateSpecForComponent(
	component component.Interface,
	path tree.Path,
	latticeID v1.LatticeID,
	internalDNSDomain string,
	namespacePrefix, namespace, name string,
	labels map[string]string,
	buildArtifacts map[string]ContainerBuildArtifacts,
	restartPolicy corev1.RestartPolicy,
	affinity *corev1.Affinity,
	tolerations []corev1.Toleration,
) (*corev1.PodTemplateSpec, error) {
	// convert lattice containers into kube containers
	containersComponent, err := newContainersComponent(component)
	if err != nil {
		return nil, err
	}

	var kubeContainers []corev1.Container
	mainContainerBuildArtifact, ok := buildArtifacts[kubeutil.UserMainContainerName]
	if !ok {
		return nil, fmt.Errorf("build artifacts did not include artifact for main container")
	}

	mainContainer, err := KubeContainerForContainer(
		containersComponent.MainContainer,
		kubeutil.UserMainContainerName,
		mainContainerBuildArtifact,
		path,
	)
	if err != nil {
		return nil, err
	}

	kubeContainers = append(kubeContainers, mainContainer)

	for name, sidecar := range containersComponent.Sidecars {
		buildArtifact, ok := buildArtifacts[kubeutil.UserSidecarContainerName(name)]
		if !ok {
			return nil, fmt.Errorf("build artifacts did not include artifact for sidecar %v", name)
		}

		container, err := KubeContainerForContainer(
			sidecar,
			kubeutil.UserMainContainerName,
			buildArtifact,
			path,
		)
		if err != nil {
			return nil, err
		}

		kubeContainers = append(kubeContainers, container)
	}

	// create the proper DNS options
	systemID, err := kubeutil.SystemID(namespacePrefix, namespace)
	if err != nil {
		err := fmt.Errorf("error getting system ID: %v", err)
		return nil, err
	}

	baseSearchPath := kubeutil.FullyQualifiedInternalSystemSubdomain(systemID, latticeID, internalDNSDomain)
	dnsSearches := []string{baseSearchPath}

	parentNode, err := path.Parent()
	if err != nil {
		return nil, fmt.Errorf("cannot get parent node path: %v", err)
	}
	parentDomain := kubeutil.FullyQualifiedInternalAddressSubdomain(parentNode.ToDomain(), systemID, latticeID, internalDNSDomain)
	if !parentNode.IsRoot() {
		dnsSearches = append(dnsSearches, parentDomain)
	}

	ndotsValue := "15"
	dnsConfig := &corev1.PodDNSConfig{
		Nameservers: []string{},
		Options: []corev1.PodDNSConfigOption{
			{
				Name:  "ndots",
				Value: &ndotsValue,
			},
		},
		Searches: dnsSearches,
	}

	// create the pod template spec
	podSpecTemplate := corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
		Spec: corev1.PodSpec{
			Containers:    kubeContainers,
			RestartPolicy: restartPolicy,
			DNSPolicy:     corev1.DNSDefault,
			DNSConfig:     dnsConfig,
			Affinity:      affinity,
			Tolerations:   tolerations,
		},
	}
	return &podSpecTemplate, nil
}

type containersComponent struct {
	MainContainer definitionv1.Container
	Sidecars      map[string]definitionv1.Container
}

func newContainersComponent(component component.Interface) (*containersComponent, error) {
	switch c := component.(type) {
	case *definitionv1.Job:
		cc := &containersComponent{
			MainContainer: c.Container,
			Sidecars:      c.Sidecars,
		}
		return cc, nil

	case *definitionv1.Service:
		cc := &containersComponent{
			MainContainer: c.Container,
			Sidecars:      c.Sidecars,
		}
		return cc, nil
	}

	return nil, fmt.Errorf("invalid container component: %v", component)
}

func KubeContainerForContainer(
	container definitionv1.Container,
	containerName string,
	buildArtifacts ContainerBuildArtifacts,
	componentPath tree.Path,
) (corev1.Container, error) {
	var ports []corev1.ContainerPort
	for portNum := range container.Ports {
		ports = append(
			ports,
			corev1.ContainerPort{
				Name:          fmt.Sprintf("port-%v", portNum),
				Protocol:      corev1.ProtocolTCP,
				ContainerPort: portNum,
			},
		)
	}

	// Sort the env var names so the array order is deterministic
	// so we can more easily check to see if the spec needs
	// to be updated.
	var envVarNames []string
	for name := range container.Exec.Environment {
		envVarNames = append(envVarNames, name)
	}

	sort.Strings(envVarNames)

	var envVars []corev1.EnvVar
	for _, name := range envVarNames {
		envVar := container.Exec.Environment[name]
		if envVar.Value != nil {
			envVars = append(
				envVars,
				corev1.EnvVar{
					Name:  name,
					Value: *envVar.Value,
				},
			)
		} else if envVar.SecretRef != nil {
			secretName, err := latticeutil.HashPath(envVar.SecretRef.Value.Path())
			if err != nil {
				return corev1.Container{}, err
			}

			s := &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: secretName,
				},
				Key: envVar.SecretRef.Value.Subcomponent(),
			}

			envVars = append(
				envVars,
				corev1.EnvVar{
					Name: name,
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: s,
					},
				},
			)
		}
	}

	var probe *corev1.Probe
	if container.HealthCheck != nil && container.HealthCheck.HTTP != nil {
		probe = &corev1.Probe{
			Handler: corev1.Handler{
				HTTPGet: &corev1.HTTPGetAction{
					Path: container.HealthCheck.HTTP.Path,
					Port: intstr.FromInt(int(container.HealthCheck.HTTP.Port)),
				},
			},
		}
	}

	var command []string
	if container.Exec != nil {
		command = container.Exec.Command
	}

	kubeContainer := corev1.Container{
		Name:            containerName,
		Image:           buildArtifacts.DockerImageFQN,
		ImagePullPolicy: corev1.PullIfNotPresent,
		Command:         command,
		Ports:           ports,
		Env:             envVars,
		LivenessProbe:   probe,
		ReadinessProbe:  probe,
	}
	return kubeContainer, nil
}
