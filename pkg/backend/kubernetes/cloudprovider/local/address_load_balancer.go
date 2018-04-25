package local

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"

	corev1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
)

func (cp *DefaultLocalCloudProvider) EnsureServiceAddressLoadBalancer(
	latticeID v1.LatticeID,
	address *latticev1.Address,
	service *latticev1.Service,
) error {
	// Try to find the kube service in the cache
	kubeServiceName := serviceAddressKubeServiceLoadBalancerName(address)
	kubeService, err := cp.kubeServiceLister.Services(address.Namespace).Get(kubeServiceName)
	if err != nil {
		if !errors.IsNotFound(err) {
			return err
		}

		// If it wasn't found, try to create it.
		spec := cp.kubeServiceSpec(address, service)
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

				return fmt.Errorf(
					"could not create kube service %v for %v because it already exists, but could not find it",
					kubeServiceName,
					address.Description(cp.namespacePrefix),
				)
			}
		}
	}

	spec := cp.kubeServiceSpec(address, service)

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
		fmt.Errorf("error patching kube service for %v: %v", address.Description(cp.namespacePrefix), err)
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
	annotations map[string]string,
) error {
	return nil
}

func (cp *DefaultLocalCloudProvider) ServiceAddressLoadBalancerPorts(
	latticeID v1.LatticeID,
	address *latticev1.Address,
	service *latticev1.Service,
) (map[int32]string, error) {
	kubeServiceName := serviceAddressKubeServiceLoadBalancerName(address)
	kubeService, err := cp.kubeServiceLister.Services(address.Namespace).Get(kubeServiceName)
	if err != nil {
		return nil, err
	}

	ports := make(map[int32]string)
	for _, port := range kubeService.Spec.Ports {
		ports[port.Port] = fmt.Sprintf("%v:%v", cp.IP(), port.NodePort)
	}

	return ports, nil
}

func (cp *DefaultLocalCloudProvider) kubeServiceSpec(address *latticev1.Address, service *latticev1.Service) corev1.ServiceSpec {
	var ports []corev1.ServicePort

	for component, componentPorts := range service.Spec.Ports {
		for _, componentPort := range componentPorts {
			if componentPort.Public {
				ports = append(ports, corev1.ServicePort{
					// FIXME: need a better naming scheme
					Name: fmt.Sprintf("%v-%v", component, componentPort.Name),
					Port: componentPort.Port,
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
	return corev1.ServiceSpec{
		Selector: labels,
		Type:     corev1.ServiceTypeNodePort,
		Ports:    ports,
	}
}

func serviceAddressKubeServiceSpecNeedsUpdate(desired, current corev1.ServiceSpec) bool {
	if desired.Type != current.Type {
		return true
	}

	if !reflect.DeepEqual(desired.Selector, current.Selector) {
		return true
	}

	if !reflect.DeepEqual(desired.Ports, current.Ports) {
		return true
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
