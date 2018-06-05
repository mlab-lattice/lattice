package aws

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	kubetf "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/cloudprovider/aws/terraform"
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/util/kubernetes"
	"github.com/mlab-lattice/lattice/pkg/util/terraform"
	awstfprovider "github.com/mlab-lattice/lattice/pkg/util/terraform/provider/aws"

	corev1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
)

const (
	AnnotationKeyAddressServiceLoadBalancerDNSName = "service-load-balancer.address.aws.cloud-provider.lattice.mlab.com/dns-name"

	terraformOutputServiceAddressLoadBalancerDNSName = "dns_name"
)

func (cp *DefaultAWSCloudProvider) ServiceAddressLoadBalancerNeedsUpdate(
	latticeID v1.LatticeID,
	address *latticev1.Address,
	service *latticev1.Service,
	serviceMeshPorts map[int32]int32,
) (bool, error) {
	loadBalancerNeeded := serviceNeedsAddressLoadBalancer(service)

	kubeService, err := cp.getKubeService(address)
	if err != nil {
		return false, err
	}

	if kubeService == nil && loadBalancerNeeded {
		return true, nil
	}

	if !loadBalancerNeeded {
		if kubeService != nil {
			return true, nil
		}

		config, err := cp.serviceAddressLoadBalancerTerraformConfig(latticeID, address, nil)
		if err != nil {
			return false, err
		}

		result, _, err := terraform.Plan(serviceAddressLoadBalancerWorkDirectory(address.Name), config, true)
		if err != nil {
			return false, err
		}

		switch result {
		case terraform.PlanResultError:
			return false, fmt.Errorf("unknown error")

		case terraform.PlanResultEmpty:
			return false, nil

		case terraform.PlanResultNotEmpty:
			return true, nil

		default:
			return false, fmt.Errorf("unexpected terraform plan result: %v", result)
		}
	}

	spec, err := cp.kubeServiceSpec(address, service, serviceMeshPorts)
	if err != nil {
		return false, err
	}

	if serviceAddressKubeServiceSpecNeedsUpdate(spec, kubeService.Spec) {
		return true, nil
	}

	module, err := cp.serviceAddressLoadBalancerTerraformModule(latticeID, address, service, serviceMeshPorts, kubeService)
	if err != nil {
		return false, err
	}

	config, err := cp.serviceAddressLoadBalancerTerraformConfig(latticeID, address, module)
	if err != nil {
		return false, err
	}

	result, _, err := terraform.Plan(serviceAddressLoadBalancerWorkDirectory(address.Name), config, false)
	if err != nil {
		return false, err
	}

	switch result {
	case terraform.PlanResultError:
		return false, fmt.Errorf("unknown error")

	case terraform.PlanResultEmpty:
		return false, nil

	case terraform.PlanResultNotEmpty:
		return true, nil

	default:
		return false, fmt.Errorf("unexpected terraform plan result: %v", result)
	}
}

func (cp *DefaultAWSCloudProvider) EnsureServiceAddressLoadBalancer(
	latticeID v1.LatticeID,
	address *latticev1.Address,
	service *latticev1.Service,
	serviceMeshPorts map[int32]int32,
) error {
	if !serviceNeedsAddressLoadBalancer(service) {
		return cp.DestroyServiceAddressLoadBalancer(latticeID, address)
	}

	kubeService, err := cp.ensureKubeService(address, service, serviceMeshPorts)
	if err != nil {
		return err
	}

	module, err := cp.serviceAddressLoadBalancerTerraformModule(latticeID, address, service, serviceMeshPorts, kubeService)
	if err != nil {
		return err
	}

	config, err := cp.serviceAddressLoadBalancerTerraformConfig(latticeID, address, module)
	if err != nil {
		return err
	}

	_, err = terraform.Apply(serviceAddressLoadBalancerWorkDirectory(address.Name), config)
	if err != nil {
		return fmt.Errorf(
			"error applying terraform for %v service load balancer: %v",
			address.Description(cp.namespacePrefix),
			err,
		)
	}

	return nil
}

func (cp *DefaultAWSCloudProvider) DestroyServiceAddressLoadBalancer(
	latticeID v1.LatticeID,
	address *latticev1.Address,
) error {
	config, err := cp.serviceAddressLoadBalancerTerraformConfig(latticeID, address, nil)
	if err != nil {
		return err
	}

	_, err = terraform.Destroy(serviceAddressLoadBalancerWorkDirectory(address.Name), config)
	if err != nil {
		return fmt.Errorf(
			"error destroying terraform for %v service load balancer: %v",
			address.Description(cp.namespacePrefix),
			err,
		)
	}

	kubeService, err := cp.getKubeService(address)
	if err != nil {
		return err
	}

	if kubeService == nil {
		return nil
	}

	return cp.kubeClient.CoreV1().Services(kubeService.Namespace).Delete(kubeService.Name, nil)
}

func (cp *DefaultAWSCloudProvider) ServiceAddressLoadBalancerAddAnnotations(
	latticeID v1.LatticeID,
	address *latticev1.Address,
	service *latticev1.Service,
	serviceMeshPorts map[int32]int32,
	annotations map[string]string,
) error {
	info, err := cp.serviceAddressLoadBalancerInfo(latticeID, address, service, serviceMeshPorts)
	if err != nil {
		return err
	}

	annotations[AnnotationKeyAddressServiceLoadBalancerDNSName] = info.DNSName
	return nil
}

func (cp *DefaultAWSCloudProvider) ServiceAddressLoadBalancerPorts(
	latticeID v1.LatticeID,
	address *latticev1.Address,
	service *latticev1.Service,
	serviceMeshPorts map[int32]int32,
) (map[int32]string, error) {
	dnsName, ok := address.Annotations[AnnotationKeyAddressServiceLoadBalancerDNSName]
	if !ok {
		err := fmt.Errorf(
			"%v does not have annotation %v",
			address.Description(cp.namespacePrefix),
			AnnotationKeyAddressServiceLoadBalancerDNSName,
		)
		return nil, err
	}

	ports := make(map[int32]string)
	for _, componentPorts := range service.Spec.Ports {
		for _, componentPort := range componentPorts {
			if componentPort.Public {
				ports[componentPort.Port] = fmt.Sprintf("http://%v:%v", dnsName, componentPort.Port)
			}
		}
	}

	return ports, nil
}

func (cp *DefaultAWSCloudProvider) serviceAddressLoadBalancerTerraformConfig(
	latticeID v1.LatticeID,
	address *latticev1.Address,
	module *kubetf.ApplicationLoadBalancer,
) (*terraform.Config, error) {
	config := &terraform.Config{
		Provider: awstfprovider.Provider{
			Region: cp.region,
		},
		Backend: terraform.S3BackendConfig{
			Region:  cp.region,
			Bucket:  cp.terraformBackendOptions.S3.Bucket,
			Key:     kubetf.GetS3BackendServiceAddressLoadBalancerPathRoot(latticeID, address.Namespace, address.Name),
			Encrypt: true,
		},
	}

	if module != nil {
		config.Modules = map[string]interface{}{
			"load-balancer": module,
		}
		config.Output = map[string]terraform.ConfigOutput{
			terraformOutputServiceAddressLoadBalancerDNSName: {
				Value: fmt.Sprintf("${module.load-balancer.%v}", terraformOutputServiceAddressLoadBalancerDNSName),
			},
		}
	}

	return config, nil
}

func (cp *DefaultAWSCloudProvider) serviceAddressLoadBalancerTerraformModule(
	latticeID v1.LatticeID,
	address *latticev1.Address,
	service *latticev1.Service,
	serviceMeshPorts map[int32]int32,
	kubeService *corev1.Service,
) (*kubetf.ApplicationLoadBalancer, error) {
	systemID, err := kubernetes.SystemID(cp.namespacePrefix, address.Namespace)
	if err != nil {
		return nil, fmt.Errorf("error getting system ID for %v: %v", address.Description(cp.namespacePrefix), err)
	}

	nodePoolAnnotation, err := service.NodePoolAnnotation()
	if err != nil {
		return nil, err
	}

	serviceMeshPortsLookup := make(map[int32]int32)
	for k, v := range serviceMeshPorts {
		serviceMeshPortsLookup[v] = k
	}

	targetPorts := make(map[int32]int32)
	for _, kubeServicePort := range kubeService.Spec.Ports {
		servicePort, ok := serviceMeshPortsLookup[kubeServicePort.Port]
		if !ok {
			err := fmt.Errorf(
				"service mesh ports does not contain kube service port %v for %v",
				kubeServicePort.Port,
				service.Description(cp.namespacePrefix),
			)
			return nil, err
		}

		targetPorts[servicePort] = kubeServicePort.NodePort
	}

	autoscalingGroupSecurityGroupIDS := make(map[string]string)
	for ns := range nodePoolAnnotation {
		for nodePoolName := range nodePoolAnnotation[ns] {
			nodePool, err := cp.nodePoolLister.NodePools(ns).Get(nodePoolName)
			if err != nil {
				return nil, fmt.Errorf("error getting node pool %v/%v: %v", ns, nodePoolName, err)
			}

			autoscalingGroupName, ok := nodePool.Annotations[AnnotationKeyNodePoolAutoscalingGroupName]
			if !ok {
				err := fmt.Errorf(
					"%v does not have %v annotation",
					nodePool.Description(cp.namespacePrefix),
					AnnotationKeyNodePoolAutoscalingGroupName,
				)
				return nil, err
			}

			securityGroupID, ok := nodePool.Annotations[AnnotationKeyNodePoolSecurityGroupID]
			if !ok {
				err := fmt.Errorf(
					"%v does not have %v annotation",
					nodePool.Description(cp.namespacePrefix),
					AnnotationKeyNodePoolSecurityGroupID,
				)
				return nil, err
			}

			autoscalingGroupSecurityGroupIDS[autoscalingGroupName] = securityGroupID
		}
	}

	module := &kubetf.ApplicationLoadBalancer{
		Source: cp.terraformModulePath + kubetf.ModulePathApplicationLoadBalancer,

		Region: cp.region,

		LatticeID: latticeID,
		SystemID:  systemID,
		VPCID:     cp.vpcID,
		SubnetIDs: cp.subnetIDs,

		Name: address.Name,
		AutoscalingGroupSecurityGroupIDs: autoscalingGroupSecurityGroupIDS,
		Ports: targetPorts,
	}
	return module, nil
}

func (cp *DefaultAWSCloudProvider) ensureKubeService(
	address *latticev1.Address,
	service *latticev1.Service,
	serviceMeshPorts map[int32]int32,
) (*corev1.Service, error) {
	// Try to find the kube service in the cache
	kubeService, err := cp.getKubeService(address)
	if err != nil {
		return nil, err
	}

	if kubeService == nil {
		// If it wasn't found, try to create it.
		spec, err := cp.kubeServiceSpec(address, service, serviceMeshPorts)
		if err != nil {
			return nil, err
		}

		kubeServiceName := serviceAddressKubeServiceLoadBalancerName(address)
		kubeService = &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:            kubeServiceName,
				OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(address, latticev1.AddressKind)},
			},
			Spec: spec,
		}

		// If it wasn't found in our cache but we couldn't create it due to it already existing, we lost a race and
		// should retrieve it from the API.
		kubeService, err = cp.kubeClient.CoreV1().Services(address.Namespace).Create(kubeService)
		if err != nil {
			if !errors.IsAlreadyExists(err) {
				return nil, err
			}

			kubeService, err = cp.kubeClient.CoreV1().Services(address.Namespace).Get(kubeServiceName, metav1.GetOptions{})
			if err != nil {
				if !errors.IsNotFound(err) {
					return nil, err
				}

				err := fmt.Errorf(
					"could not create kube service %v for %v because it already exists, but could not find it",
					kubeServiceName,
					address.Description(cp.namespacePrefix),
				)
				return nil, err
			}
		}
	}

	spec, err := cp.kubeServiceSpec(address, service, serviceMeshPorts)
	if err != nil {
		return nil, err
	}

	if !serviceAddressKubeServiceSpecNeedsUpdate(spec, kubeService.Spec) {
		return kubeService, nil
	}

	strategicMergePatchBytes, err := serviceAddressKubeServiceStrategicMergePatchBytes(spec, kubeService.Spec)
	if err != nil {
		return nil, fmt.Errorf("error creating json patches for kube service spec patch: %v", err)
	}

	kubeService, err = cp.kubeClient.CoreV1().Services(address.Namespace).Patch(
		kubeService.Name,
		types.StrategicMergePatchType,
		strategicMergePatchBytes,
	)
	if err != nil {
		return nil, fmt.Errorf("error patching kube service for %v: %v", address.Description(cp.namespacePrefix), err)
	}

	return kubeService, nil
}

func (cp *DefaultAWSCloudProvider) getKubeService(address *latticev1.Address) (*corev1.Service, error) {
	// Try to find the kube service in the cache
	kubeServiceName := serviceAddressKubeServiceLoadBalancerName(address)
	kubeService, err := cp.kubeServiceLister.Services(address.Namespace).Get(kubeServiceName)
	if err != nil {
		if !errors.IsNotFound(err) {
			return nil, err
		}

		return nil, nil
	}

	return kubeService, nil
}

func (cp *DefaultAWSCloudProvider) kubeServiceSpec(
	address *latticev1.Address,
	service *latticev1.Service,
	serviceMeshPorts map[int32]int32,
) (corev1.ServiceSpec, error) {
	var ports []corev1.ServicePort

	for component, componentPorts := range service.Spec.Ports {
		for _, componentPort := range componentPorts {
			if componentPort.Public {
				targetPort, ok := serviceMeshPorts[componentPort.Port]
				if !ok {
					err := fmt.Errorf(
						"component port %v not found in service mesh ports for %v",
						componentPort.Port,
						service.Description(cp.namespacePrefix),
					)
					return corev1.ServiceSpec{}, err
				}

				ports = append(ports, corev1.ServicePort{
					// FIXME: need a better naming scheme
					Name:     fmt.Sprintf("%v-%v", component, componentPort.Name),
					Protocol: corev1.ProtocolTCP,
					Port:     targetPort,
				})
			}
		}
	}

	// sort the ports so we have a deterministic spec
	sort.Slice(ports, func(i, j int) bool {
		return ports[i].Port < ports[j].Port
	})

	labels := map[string]string{
		latticev1.ServiceIDLabelKey: service.Name,
	}

	// Note: if you add or remove any fields here,
	// update serviceAddressKubeServiceStrategicMergePatchBytes as well
	spec := corev1.ServiceSpec{
		Selector: labels,
		Type:     corev1.ServiceTypeNodePort,
		Ports:    ports,
	}
	return spec, nil
}

func serviceAddressKubeServiceSpecNeedsUpdate(desired, current corev1.ServiceSpec) bool {
	if desired.Type != current.Type {
		return true
	}

	if !reflect.DeepEqual(desired.Selector, current.Selector) {
		return true
	}

	if serviceAddressKubeServiceSpecPortsNeedUpdate(desired.Ports, current.Ports) {
		return true
	}

	return false
}

func serviceAddressKubeServiceSpecPortsNeedUpdate(desired, current []corev1.ServicePort) bool {
	currentPorts := make(map[int32]corev1.ServicePort)
	for _, p := range current {
		currentPorts[p.Port] = p
	}

	for _, p := range desired {
		current, ok := currentPorts[p.Port]
		if !ok {
			return true
		}

		if p.Protocol != current.Protocol {
			return true
		}

		if p.Name != current.Name {
			return true
		}
	}

	return false
}

func serviceAddressKubeServiceStrategicMergePatchBytes(desired, current corev1.ServiceSpec) ([]byte, error) {
	// there's about 0 documentation on how to do a merge patch with the kubernetes go client
	// the below was eventually divined from: https://github.com/kubernetes/kubernetes/blob/v1.10.1/pkg/kubectl/cmd/patch.go#L260-L284
	spec := current.DeepCopy()
	spec.Selector = desired.Selector
	spec.Type = desired.Type
	spec.Ports = desired.Ports

	currentJSON, err := json.Marshal(&current)
	if err != nil {
		return nil, fmt.Errorf("error marshalling current kube service spec: %v", err)
	}

	var currentJSONMap strategicpatch.JSONMap
	err = json.Unmarshal(currentJSON, &currentJSONMap)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling current kube service spec: %v", err)
	}

	desiredJSON, err := json.Marshal(&spec)
	if err != nil {
		return nil, fmt.Errorf("error marshalling desired kube service spec: %v", err)
	}

	var desiredJSONMap strategicpatch.JSONMap
	err = json.Unmarshal(desiredJSON, &desiredJSONMap)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling desired kube service spec: %v", err)
	}

	obj := corev1.ServiceSpec{}
	patchMap, err := strategicpatch.StrategicMergeMapPatch(currentJSONMap, desiredJSONMap, obj)
	if err != nil {
		return nil, fmt.Errorf("error getting strategic merge patch: %v", err)
	}

	patchMapBytes, err := json.Marshal(&patchMap)
	if err != nil {
		return nil, fmt.Errorf("error marshalling strategic merge patch: %v", err)
	}

	return patchMapBytes, nil
}

func serviceAddressKubeServiceLoadBalancerName(address *latticev1.Address) string {
	return fmt.Sprintf("load-balancer-address-%v", address.Name)
}

func serviceAddressLoadBalancerWorkDirectory(addressID string) string {
	return workDirectory("address/service-load-balancer", addressID)
}

type serviceAddressLoadBalancerInfo struct {
	DNSName string
}

func (cp *DefaultAWSCloudProvider) serviceAddressLoadBalancerInfo(
	latticeID v1.LatticeID,
	address *latticev1.Address,
	service *latticev1.Service,
	serviceMeshPorts map[int32]int32,
) (serviceAddressLoadBalancerInfo, error) {
	kubeService, err := cp.getKubeService(address)
	if err != nil {
		return serviceAddressLoadBalancerInfo{}, err
	}

	if kubeService == nil {
		return serviceAddressLoadBalancerInfo{}, fmt.Errorf("could not get load balancer kube service")
	}

	module, err := cp.serviceAddressLoadBalancerTerraformModule(latticeID, address, service, serviceMeshPorts, kubeService)
	if err != nil {
		return serviceAddressLoadBalancerInfo{}, err
	}

	config, err := cp.serviceAddressLoadBalancerTerraformConfig(latticeID, address, module)
	if err != nil {
		return serviceAddressLoadBalancerInfo{}, err
	}

	outputVars := []string{terraformOutputServiceAddressLoadBalancerDNSName}
	values, err := terraform.Output(serviceAddressLoadBalancerWorkDirectory(address.Name), config, outputVars)
	if err != nil {
		err := fmt.Errorf(
			"error getting terraform output for %v service load balancer: %v",
			address.Description(cp.namespacePrefix),
			err,
		)
		return serviceAddressLoadBalancerInfo{}, err
	}

	info := serviceAddressLoadBalancerInfo{
		DNSName: values[terraformOutputServiceAddressLoadBalancerDNSName],
	}
	return info, nil
}

func serviceNeedsAddressLoadBalancer(service *latticev1.Service) bool {
	for _, componentPorts := range service.Spec.Ports {
		for _, componentPort := range componentPorts {
			if componentPort.Public {
				return true
			}
		}
	}

	return false
}
