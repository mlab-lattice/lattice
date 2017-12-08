package service

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strconv"

	"github.com/mlab-lattice/system/pkg/definition"
	"github.com/mlab-lattice/system/pkg/definition/block"
	"github.com/mlab-lattice/system/pkg/definition/tree"
	"github.com/mlab-lattice/system/pkg/kubernetes/constants"

	crv1 "github.com/mlab-lattice/system/pkg/kubernetes/customresource/apis/lattice/v1"
	kubeutil "github.com/mlab-lattice/system/pkg/kubernetes/util/kubernetes"

	appsv1beta2 "k8s.io/api/apps/v1beta2"
	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/golang/glog"
	"github.com/satori/go.uuid"
)

const (
	envoyConfigDirectory           = "/etc/envoy"
	envoyConfigDirectoryVolumeName = "envoyconfig"
)

// getDeployment returns an *extensions.Deployment configured for a given Service
func (sc *Controller) getDeployment(svc *crv1.Service) (*appsv1beta2.Deployment, error) {
	// Need a consistent view of our config while generating the Job
	sc.configLock.RLock()
	defer sc.configLock.RUnlock()

	dName := getDeploymentName(svc)
	dLabels := getDeploymentLabels(svc)
	dAnnotations, err := sc.getDeploymentAnnotations(svc)
	if err != nil {
		return nil, err
	}

	dSpec, err := sc.getDeploymentSpec(svc)
	if err != nil {
		return nil, err
	}

	d := &appsv1beta2.Deployment{
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

func getDeploymentLabels(svc *crv1.Service) map[string]string {
	return map[string]string{
		constants.LabelKeyServiceDeployment: svc.Name,
	}
}

func (sc *Controller) getDeploymentAnnotations(svc *crv1.Service) (map[string]string, error) {
	annotations := map[string]string{}
	definitionJSONBytes, err := json.Marshal(svc.Spec.Definition)
	if err != nil {
		return nil, err
	}

	annotations[constants.AnnotationKeyDeploymentServiceDefinition] = string(definitionJSONBytes)

	// FIXME: remove this when local DNS is working
	sys, err := sc.getServiceSystem(svc)
	if err != nil {
		return nil, err
	}
	services := getSystemServicesSlice(sys)
	servicesJSONBytes, err := json.Marshal(services)
	if err != nil {
		return nil, err
	}

	annotations[constants.AnnotationKeySystemServices] = string(servicesJSONBytes)

	return annotations, nil
}

// getDeploymentSpec returns an *extensions.DeploymentSpec configured for a given Service.
// N.B.: getDeploymentSpec assumes a RLock is held on sc.configLock when called.
func (sc *Controller) getDeploymentSpec(svc *crv1.Service) (*appsv1beta2.DeploymentSpec, error) {
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
		Name:            fmt.Sprintf("lattice-prepare-envoy-%v", uuid.NewV4().String()),
		Image:           sc.config.Envoy.PrepareImage,
		ImagePullPolicy: corev1.PullIfNotPresent,
		Env: []corev1.EnvVar{
			{
				Name:  "EGRESS_PORT",
				Value: strconv.Itoa(int(svc.Spec.EnvoyEgressPort)),
			},
			{
				Name:  "REDIRECT_EGRESS_CIDR_BLOCK",
				Value: envoyConfig.RedirectCidrBlock,
			},
			{
				Name:  "CONFIG_DIR",
				Value: envoyConfigDirectory,
			},
			{
				Name:  "ADMIN_PORT",
				Value: strconv.Itoa(int(svc.Spec.EnvoyAdminPort)),
			},
			{
				Name: "XDS_API_HOST",
				ValueFrom: &corev1.EnvVarSource{
					FieldRef: &corev1.ObjectFieldSelector{
						FieldPath: "status.hostIP",
					},
				},
			},
			{
				Name:  "XDS_API_PORT",
				Value: fmt.Sprintf("%v", envoyConfig.XDSAPIPort),
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
		Name:            fmt.Sprintf("lattice-envoy-%v", uuid.NewV4().String()),
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
	dName := getDeploymentName(svc)
	dLabels := getDeploymentLabels(svc)
	ds := appsv1beta2.DeploymentSpec{
		Replicas: &replicas,
		Selector: &metav1.LabelSelector{
			MatchLabels: dLabels,
		},
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Name:   dName,
				Labels: dLabels,
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
				Tolerations: []corev1.Toleration{
					kubeutil.GetServiceTaintToleration(svc.Name),
				},
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
		path, err := tree.NewNodePath(svcPath)
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

func getLivenessProbe(hc *block.ComponentHealthCheck) *corev1.Probe {
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

	if hc.HTTP != nil {
		headers := []corev1.HTTPHeader{}
		for k, v := range hc.HTTP.Headers {
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
					Path:        hc.HTTP.Path,
					Port:        intstr.FromString(hc.HTTP.Port),
					HTTPHeaders: headers,
				},
			},
		}
	}

	return &corev1.Probe{
		Handler: corev1.Handler{
			TCPSocket: &corev1.TCPSocketAction{
				Port: intstr.FromString(hc.TCP.Port),
			},
		},
	}
}

func (sc *Controller) getDeploymentForService(svc *crv1.Service) (*appsv1beta2.Deployment, error) {
	// List all Deployments to find in the Service's namespace to find the Deployment the Service manages.
	deployments, err := sc.deploymentLister.Deployments(svc.Namespace).List(labels.Everything())
	if err != nil {
		return nil, err
	}

	matchingDeployments := []*appsv1beta2.Deployment{}
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

func (sc *Controller) createDeployment(svc *crv1.Service) (*appsv1beta2.Deployment, error) {
	d, err := sc.getDeployment(svc)
	if err != nil {
		return nil, err
	}

	dResp, err := sc.kubeClient.AppsV1beta2().Deployments(svc.Namespace).Create(d)
	if err != nil {
		// FIXME: send warn event
		return nil, err
	}

	glog.V(4).Infof("Created Deployment %s", dResp.Name)
	// FIXME: send normal event
	return dResp, nil
}

func (sc *Controller) syncDeploymentSpec(svc *crv1.Service, d *appsv1beta2.Deployment) (*appsv1beta2.Deployment, error) {
	dSvcDefStr, ok := d.Annotations[constants.AnnotationKeyDeploymentServiceDefinition]
	if !ok {
		return nil, fmt.Errorf("deployment did not have %v annotation", constants.AnnotationKeyDeploymentServiceDefinition)
	}

	dSvcDef := definition.Service{}
	err := json.Unmarshal([]byte(dSvcDefStr), &dSvcDef)
	if err != nil {
		return nil, err
	}

	// FIXME: remove this when local DNS works
	sysServicesStr, ok := d.Annotations[constants.AnnotationKeySystemServices]
	if !ok {
		return nil, fmt.Errorf("deployment did not have %v annotation", constants.AnnotationKeySystemServices)
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

func (sc *Controller) updateDeploymentSpec(svc *crv1.Service, d *appsv1beta2.Deployment, dSpec *appsv1beta2.DeploymentSpec) (*appsv1beta2.Deployment, error) {
	dAnnotations, err := sc.getDeploymentAnnotations(svc)
	if err != nil {
		return nil, err
	}

	for k, v := range dAnnotations {
		d.Annotations[k] = v
	}

	d.Spec = *dSpec

	// TODO: should we Patch here instead?
	result, err := sc.kubeClient.AppsV1beta2().Deployments(d.Namespace).Update(d)
	return result, err
}

func (sc *Controller) getServiceSystem(svc *crv1.Service) (*crv1.System, error) {
	return sc.systemLister.Systems(svc.Namespace).Get(svc.Namespace)
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
