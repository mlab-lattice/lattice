package local

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strconv"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"

	corev1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
)

func (cp *DefaultLocalCloudProvider) ServiceAddressLoadBalancerNeedsUpdate(
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

		// XXX: something else need to happen here?

		return false, nil
	}

	spec, err := cp.kubeServiceSpec(address, service, serviceMeshPorts)
	if err != nil {
		return false, err
	}

	return serviceAddressKubeServiceSpecNeedsUpdate(spec, kubeService.Spec), nil
}

func (cp *DefaultLocalCloudProvider) EnsureServiceAddressLoadBalancer(
	latticeID v1.LatticeID,
	address *latticev1.Address,
	service *latticev1.Service,
	serviceMeshPorts map[int32]int32,
) error {
	// Try to find the kube service in the cache
	kubeService, err := cp.getKubeService(address)
	if err != nil {
		return err
	}

	if kubeService == nil {
		// If it wasn't found, try to create it.
		spec, err := cp.kubeServiceSpec(address, service, serviceMeshPorts)
		if err != nil {
			return err
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
				return err
			}

			kubeService, err = cp.kubeClient.CoreV1().Services(address.Namespace).Get(kubeServiceName, metav1.GetOptions{})
			if err != nil {
				if !errors.IsNotFound(err) {
					return err
				}

				err := fmt.Errorf(
					"could not create kube service %v for %v because it already exists, but could not find it",
					kubeServiceName,
					address.Description(cp.namespacePrefix),
				)
				return err
			}
		}
	}

	spec, err := cp.kubeServiceSpec(address, service, serviceMeshPorts)
	if err != nil {
		return err
	}

	if !serviceAddressKubeServiceSpecNeedsUpdate(spec, kubeService.Spec) {
		return nil
	}

	strategicMergePatchBytes, err := serviceAddressKubeServiceStrategicMergePatchBytes(spec, kubeService.Spec)
	if err != nil {
		return fmt.Errorf("error creating json patches for kube service spec patch: %v", err)
	}

	_, err = cp.kubeClient.CoreV1().Services(address.Namespace).Patch(
		kubeService.Name,
		types.StrategicMergePatchType,
		strategicMergePatchBytes,
	)
	if err != nil {
		return fmt.Errorf("error patching kube service for %v: %v", address.Description(cp.namespacePrefix), err)
	}

	return nil
}

func (cp *DefaultLocalCloudProvider) DestroyServiceAddressLoadBalancer(
	latticeID v1.LatticeID,
	address *latticev1.Address,
) error {
	kubeServiceName := serviceAddressKubeServiceLoadBalancerName(address)

	err := cp.kubeClient.CoreV1().Services(address.Namespace).Delete(kubeServiceName, nil)
	if err != nil {
		// if the kube service is already deleted then it's not an error
		if !errors.IsNotFound(err) {
			return err
		}
	}

	return nil
}

func (cp *DefaultLocalCloudProvider) ServiceAddressLoadBalancerAddAnnotations(
	latticeID v1.LatticeID,
	address *latticev1.Address,
	service *latticev1.Service,
	serviceMeshPorts map[int32]int32,
	annotations map[string]string,
) error {
	return nil
}

func (cp *DefaultLocalCloudProvider) ServiceAddressLoadBalancerPorts(
	latticeID v1.LatticeID,
	address *latticev1.Address,
	service *latticev1.Service,
	serviceMeshPorts map[int32]int32,
) (map[int32]string, error) {
	kubeServiceName := serviceAddressKubeServiceLoadBalancerName(address)
	kubeService, err := cp.kubeServiceLister.Services(address.Namespace).Get(kubeServiceName)
	if err != nil {
		return nil, err
	}

	kubeServicePorts := make(map[int32]int32)
	for _, port := range kubeService.Spec.Ports {
		kubeServicePorts[port.Port] = port.NodePort
	}

	ports := make(map[int32]string)
	for _, port := range service.Spec.Definition.ContainerPorts() {
		if port.Public() {
			kubeServicePort := kubeServicePorts[serviceMeshPorts[port.Port]]
			ports[port.Port] = fmt.Sprintf("%v:%v", cp.IP(), kubeServicePort)
		}
	}

	return ports, nil
}

func (cp *DefaultLocalCloudProvider) kubeServiceSpec(
	address *latticev1.Address,
	service *latticev1.Service,
	serviceMeshPorts map[int32]int32,
) (corev1.ServiceSpec, error) {
	var ports []corev1.ServicePort

	for _, port := range service.Spec.Definition.ContainerPorts() {
		if port.Public() {
			targetPort, ok := serviceMeshPorts[port.Port]
			if !ok {
				err := fmt.Errorf(
					"container port %v not found in service mesh ports for %v",
					port.Port,
					service.Description(cp.namespacePrefix),
				)
				return corev1.ServiceSpec{}, err
			}

			ports = append(ports, corev1.ServicePort{
				Name:     strconv.Itoa(int(port.Port)),
				Protocol: corev1.ProtocolTCP,
				Port:     targetPort,
			})
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

func (cp *DefaultLocalCloudProvider) getKubeService(address *latticev1.Address) (*corev1.Service, error) {
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

func serviceNeedsAddressLoadBalancer(service *latticev1.Service) bool {
	for _, port := range service.Spec.Definition.ContainerPorts() {
		if port.Public() {
			return true
		}
	}

	return false
}
