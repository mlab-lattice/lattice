package service

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strconv"

	systemdefinition "github.com/mlab-lattice/core/pkg/system/definition"
	systemdefinitionblock "github.com/mlab-lattice/core/pkg/system/definition/block"
	systemtree "github.com/mlab-lattice/core/pkg/system/tree"

	crv1 "github.com/mlab-lattice/kubernetes-integration/pkg/api/customresource/v1"

	corev1 "k8s.io/api/core/v1"
	extensions "k8s.io/api/extensions/v1beta1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/uuid"

	"github.com/golang/glog"
)

const (
	envoyConfigDirectory           = "/etc/envoy"
	envoyConfigDirectoryVolumeName = "envoyconfig"
)

// getDeployment returns an *extensions.Deployment configured for a given Service
func (sc *ServiceController) getDeployment(svc *crv1.Service) (*extensions.Deployment, error) {
	// Need a consistent view of our config while generating the Job
	sc.configLock.RLock()
	defer sc.configLock.RUnlock()

	dName := getDeploymentName(svc)
	dLabels := map[string]string{
		crv1.LabelKeyServiceDeployment: svc.Name,
	}
	dAnnotations, err := sc.getDeploymentAnnotations(svc)
	if err != nil {
		return nil, err
	}

	dSpec, err := sc.getDeploymentSpec(svc)
	if err != nil {
		return nil, err
	}

	d := &extensions.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:            dName,
			Annotations:     dAnnotations,
			Labels:          dLabels,
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

func (sc *ServiceController) getDeploymentAnnotations(svc *crv1.Service) (map[string]string, error) {
	annotations := map[string]string{}
	svcDefinitionJsonBytes, err := json.Marshal(svc.Spec.Definition)
	if err != nil {
		return nil, err
	}

	annotations[crv1.AnnotationKeyDeploymentServiceDefinition] = string(svcDefinitionJsonBytes)

	// FIXME: remove this when local DNS is working
	sys, err := sc.getServiceSystem(svc)
	if err != nil {
		return nil, err
	}
	sysSvcSlice := getSystemServicesSlice(sys)
	sysSvcJsonBytes, err := json.Marshal(sysSvcSlice)
	if err != nil {
		return nil, err
	}

	annotations[crv1.AnnotationKeySystemServices] = string(sysSvcJsonBytes)

	return annotations, nil
}

// getDeploymentSpec returns an *extensions.DeploymentSpec configured for a given Service.
// N.B.: getDeploymentSpec assumes a RLock is held on sc.configLock when called.
func (sc *ServiceController) getDeploymentSpec(svc *crv1.Service) (*extensions.DeploymentSpec, error) {
	containers := []corev1.Container{}
	initContainers := []corev1.Container{}

	// Create a container for each Component in the Service
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
			Name:            component.Name,
			Image:           svc.Spec.ComponentBuildArtifacts[component.Name].DockerImageFqn,
			ImagePullPolicy: corev1.PullIfNotPresent,
			Command:         component.Exec.Command,
			Ports:           ports,
			Env:             envs,
			// TODO: maybe add Resources
			// TODO: add VolumeMounts
			LivenessProbe: getLivenessProbe(component.HealthCheck),
		}

		if component.Init {
			initContainers = append(initContainers, container)
		} else {
			containers = append(containers, container)
		}
	}

	// Add envoy containers
	envoyConfig := sc.config.Envoy
	initContainers = append(initContainers, corev1.Container{
		// add a UUID to deal with the small chance that a user names their
		// service component the same thing we name our envoy container
		Name:            fmt.Sprintf("lattice-prepare-envoy-%v", uuid.NewUUID()),
		Image:           sc.config.Envoy.PrepareImage,
		ImagePullPolicy: corev1.PullIfNotPresent,
		Command:         []string{"/usr/local/bin/prepare-envoy.sh"},
		Env: []corev1.EnvVar{
			{
				Name:  "ENVOY_EGRESS_PORT",
				Value: strconv.Itoa(int(svc.Spec.EnvoyEgressPort)),
			},
			{
				Name:  "REDIRECT_EGRESS_CIDR_BLOCK",
				Value: envoyConfig.RedirectCidrBlock,
			},
			{
				Name:  "ENVOY_CONFIG_DIR",
				Value: envoyConfigDirectory,
			},
			{
				Name:  "ENVOY_ADMIN_PORT",
				Value: strconv.Itoa(int(svc.Spec.EnvoyAdminPort)),
			},
			{
				Name: "ENVOY_XDS_API_HOST",
				ValueFrom: &corev1.EnvVarSource{
					FieldRef: &corev1.ObjectFieldSelector{
						FieldPath: "status.hostIP",
					},
				},
			},
			{
				Name:  "ENVOY_XDS_API_PORT",
				Value: fmt.Sprintf("%v", envoyConfig.XdsApiPort),
			},
		},
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      envoyConfigDirectoryVolumeName,
				MountPath: envoyConfigDirectory,
			},
		},
		// Need CAP_NET_ADMIN to manipulate iptables
		SecurityContext: &corev1.SecurityContext{
			Capabilities: &corev1.Capabilities{
				Add: []corev1.Capability{"NET_ADMIN"},
			},
		},
	})

	envoyPorts := []corev1.ContainerPort{}
	for component, ports := range svc.Spec.Ports {
		for _, port := range ports {
			envoyPorts = append(
				envoyPorts,
				corev1.ContainerPort{
					Name:          component + "-" + port.Name,
					ContainerPort: port.EnvoyPort,
				},
			)
		}
	}
	containers = append(containers, corev1.Container{
		// add a UUID to deal with the small chance that a user names their
		// service component the same thing we name our envoy container
		Name:            fmt.Sprintf("lattice-envoy-%v", uuid.NewUUID()),
		Image:           envoyConfig.Image,
		ImagePullPolicy: corev1.PullIfNotPresent,
		Command:         []string{"/usr/local/bin/envoy"},
		Args: []string{
			"-c",
			fmt.Sprintf("%v/config.json", envoyConfigDirectory),
			"--service-cluster",
			svc.Namespace,
			"--service-node",
			svc.Spec.Path.ToDomain(false),
		},
		Ports: envoyPorts,
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      envoyConfigDirectoryVolumeName,
				MountPath: envoyConfigDirectory,
				ReadOnly:  true,
			},
		},
	})

	var replicas int32
	if svc.Spec.Definition.Resources.NumInstances != nil {
		replicas = *svc.Spec.Definition.Resources.NumInstances
	} else {
		// Spin up the min instances here, then later let autoscaler scale up.
		// TODO: when doing blue-green deploys, consider looking instead at the current number
		// 		 of replicas in the existing deployment
		replicas = *svc.Spec.Definition.Resources.MinInstances
	}
	deploymentName := getDeploymentName(svc)
	ds := extensions.DeploymentSpec{
		Replicas: &replicas,
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Name: deploymentName,
				Labels: map[string]string{
					crv1.LabelKeyServiceDeployment: svc.Name,
				},
			},
			Spec: corev1.PodSpec{
				// TODO: add user Volumes
				Volumes: []corev1.Volume{
					// Volume for Envoy config
					{
						Name: envoyConfigDirectoryVolumeName,
						VolumeSource: corev1.VolumeSource{
							EmptyDir: &corev1.EmptyDirVolumeSource{},
						},
					},
				},
				InitContainers: initContainers,
				Containers:     containers,
				DNSPolicy:      corev1.DNSDefault,
				// TODO: add NodeSelector (for cloud)
				// TODO: add Tolerations (for cloud)
			},
		},
	}

	// FIXME: remove when local dns working
	sys, err := sc.getServiceSystem(svc)
	if err != nil {
		return nil, err
	}
	sysSvcSlice := getSystemServicesSlice(sys)
	svcDomains := []string{}
	for _, svcPath := range sysSvcSlice {
		path, err := systemtree.NewNodePath(svcPath)
		if err != nil {
			return nil, err
		}

		svcDomains = append(svcDomains, path.ToDomain(true))
	}
	ds.Template.Spec.HostAliases = []corev1.HostAlias{
		{
			IP:        "172.16.29.0",
			Hostnames: svcDomains,
		},
	}

	return &ds, nil
}

func getLivenessProbe(hc *systemdefinitionblock.ComponentHealthCheck) *corev1.Probe {
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

func (sc *ServiceController) getDeploymentForService(svc *crv1.Service) (*extensions.Deployment, error) {
	// List all Deployments to find in the Service's namespace to find the Deployment the Service manages.
	deployments, err := sc.deploymentLister.Deployments(svc.Namespace).List(labels.Everything())
	if err != nil {
		return nil, err
	}

	matchingDeployments := []*extensions.Deployment{}
	svcControllerRef := metav1.NewControllerRef(svc, controllerKind)

	for _, deployment := range deployments {
		dControllerRef := metav1.GetControllerOf(deployment)

		if reflect.DeepEqual(svcControllerRef, dControllerRef) {
			matchingDeployments = append(matchingDeployments, deployment)
		}
	}

	if len(matchingDeployments) == 0 {
		return nil, nil
	}

	if len(matchingDeployments) > 1 {
		// TODO: maybe handle this better. Could choose one to make the source of truth.
		return nil, fmt.Errorf("Service %v has multiple Deployments", svc.Name)
	}

	return matchingDeployments[0], nil
}

func (sc *ServiceController) createDeployment(svc *crv1.Service) (*extensions.Deployment, error) {
	d, err := sc.getDeployment(svc)
	if err != nil {
		return nil, err
	}

	dResp, err := sc.kubeClient.ExtensionsV1beta1().Deployments(svc.Namespace).Create(d)
	if err != nil {
		// FIXME: send warn event
		return nil, err
	}

	glog.V(4).Infof("Created Deployment %s", dResp.Name)
	// FIXME: send normal event
	return dResp, nil
}

func (sc *ServiceController) syncDeploymentSpec(svc *crv1.Service, d *extensions.Deployment) (*extensions.Deployment, error) {
	dSvcDefStr, ok := d.Annotations[crv1.AnnotationKeyDeploymentServiceDefinition]
	if !ok {
		return nil, fmt.Errorf("deployment did not have %v annotation", crv1.AnnotationKeyDeploymentServiceDefinition)
	}

	dSvcDef := systemdefinition.Service{}
	err := json.Unmarshal([]byte(dSvcDefStr), &dSvcDef)
	if err != nil {
		return nil, err
	}

	// FIXME: remove this when local DNS works
	sysServicesStr, ok := d.Annotations[crv1.AnnotationKeySystemServices]
	if !ok {
		return nil, fmt.Errorf("deployment did not have %v annotation", crv1.AnnotationKeySystemServices)
	}

	sysServices := []string{}
	err = json.Unmarshal([]byte(sysServicesStr), &sysServices)
	if err != nil {
		return nil, err
	}
	sys, err := sc.getServiceSystem(svc)
	if err != nil {
		return nil, err
	}

	// If the deployment is already updated for this Service definition, nothing to do.
	if reflect.DeepEqual(dSvcDef, svc.Spec.Definition) && reflect.DeepEqual(sysServices, getSystemServicesSlice(sys)) {
		glog.V(4).Infof("Service %q Deployment Spec already up to date", svc.Name)
		return d, nil
	}

	glog.V(4).Infof("Service %q Deployment Spec is not up to date, updating", svc.Name)
	// TODO: when scaling probably want to look at the current spec's num replicas
	newDSpec, err := sc.getDeploymentSpec(svc)
	if err != nil {
		return nil, err
	}

	return sc.updateDeploymentSpec(svc, d, newDSpec)
}

func (sc *ServiceController) updateDeploymentSpec(svc *crv1.Service, d *extensions.Deployment, dSpec *extensions.DeploymentSpec) (*extensions.Deployment, error) {
	dAnnotations, err := sc.getDeploymentAnnotations(svc)
	if err != nil {
		return nil, err
	}

	for k, v := range dAnnotations {
		d.Annotations[k] = v
	}

	d.Spec = *dSpec

	// TODO: should we Patch here instead?
	result, err := sc.kubeClient.ExtensionsV1beta1().Deployments(d.Namespace).Update(d)
	return result, err
}

func (sc *ServiceController) getServiceSystem(svc *crv1.Service) (*crv1.System, error) {
	sysKey := fmt.Sprintf("%v/%v", svc.Namespace, svc.Namespace)
	sysObj, exists, err := sc.systemStore.GetByKey(sysKey)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, fmt.Errorf("no System in namespace %v", svc.Namespace)
	}

	sys := sysObj.(*crv1.System)
	return sys, nil
}

// FIXME: remove this when local DNS works
func getSystemServicesSlice(sys *crv1.System) []string {
	svcPaths := []string{}
	for path := range sys.Spec.Services {
		svcPaths = append(svcPaths, string(path))
	}

	sort.Strings(svcPaths)
	return svcPaths
}
