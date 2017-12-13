package v1

import (
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceSingularServiceAddress  = "ServiceAddress"
	ResourcePluralServiceAddress    = "ServiceAddresses"
	ResourceShortNameServiceAddress = "laddr"
	ResourceScopeServiceAddress     = apiextensionsv1beta1.NamespaceScoped
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ServiceAddress struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              ServiceAddressSpec   `json:"spec"`
	Status            ServiceAddressStatus `json:"status"`
}

type ServiceAddressSpec struct {
	EndpointGroups map[string]ServiceAddressEndpointGroup `json:"endpoints"`
	Ports          map[int32]ServiceAddressPort           `json:"ports,omitempty"`
}

type ServiceAddressEndpointGroup struct {
	Service     *string  `json:"service,omitempty"`
	IP          []string `json:"ip,omitempty"`
	LoadBalance *ServiceAddressPortEndpointGroupLoadBalanceConfig
}

type ServiceAddressPortEndpointGroupLoadBalanceConfig struct {
	Strategy string `json:"strategy"`
}

type ServiceAddressPort struct {
	HTTP *ServiceAddressPortHTTPConfig `json:"http,omitempty"`
	TCP  *ServiceAddressPortTCPConfig  `json:"tcp"`
}

type ServiceAddressPortHTTPConfig struct {
	Targets []ServiceAddressPortHTTPTargetConfig
}

type ServiceAddressPortHTTPTargetConfig struct {
	EndpointGroup string                                   `json:"endpointGroup"`
	Port          int32                                    `json:"port"`
	Weight        int32                                    `json:"weight"`
	HealthCheck   *ServiceAddressPortHTTPHealthCheckConfig `json:"healthCheck, omitempty"`
}

type ServiceAddressPortHTTPHealthCheckConfig struct {
	ServiceAddressPortBaseHealthCheckConfig `json:",inline"`
	Path                                    string `json:"path"`
}

type ServiceAddressPortTCPConfig struct {
	EndpointGroup string                                  `json:"endpointGroup"`
	HealthCheck   *ServiceAddressPortTCPHealthCheckConfig `json:"healthCheck,omitempty"`
}

type ServiceAddressPortBaseHealthCheckConfig struct {
	Timeout            int32 `json:"timeout"`
	Interval           int32 `json:"interval"`
	UnhealthyThreshold int32 `json:"unhealthyThreshold"`
	HealthyThreshold   int32 `json:"healthyThreshold"`
}

type ServiceAddressPortTCPHealthCheckConfig struct {
	ServiceAddressPortBaseHealthCheckConfig
	Payload          *string `json:"payload"`
	ExpectedResponse *string `json:"expectedResponse"`
}

type ServiceAddressStatus struct {
	State ServiceAddressState `json:"state,omitempty"`
}

type ServiceAddressState string

const (
	ServiceAddressStatePending   ServiceAddressState = "pending"
	ServiceAddressStateSucceeded ServiceAddressState = "created"
	ServiceAddressStateFailed    ServiceAddressState = "failed"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ServiceAddressList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []SystemTeardown `json:"items"`
}
