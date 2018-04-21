package v1

import (
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceSingularLoadBalancer = "loadbalancer"
	ResourcePluralLoadBalancer   = "loadbalancers"
	ResourceScopeLoadBalancer    = apiextensionsv1beta1.NamespaceScoped
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type LoadBalancer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              LoadBalancerSpec   `json:"spec"`
	Status            LoadBalancerStatus `json:"status,omitempty"`
}

type LoadBalancerSpec struct {
	NodePool string `json:"nodePool"`
}

type LoadBalancerStatus struct {
	State              LoadBalancerState          `json:"state"`
	ObservedGeneration int64                      `json:"observedGeneration"`
	Ports              map[int32]LoadBalancerPort `json:"ports"`
}

type LoadBalancerState string

const (
	LoadBalancerStatePending      LoadBalancerState = "pending"
	LoadBalancerStateProvisioning LoadBalancerState = "provisioning"
	LoadBalancerStateCreated      LoadBalancerState = "succeeded"
	LoadBalancerStateFailed       LoadBalancerState = "failed"
)

type LoadBalancerPort struct {
	Address string `json:"address"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type LoadBalancerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []LoadBalancer `json:"items"`
}
