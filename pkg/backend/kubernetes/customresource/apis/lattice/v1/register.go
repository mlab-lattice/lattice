package v1

import (
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const GroupName = "lattice.mlab.com"

var (
	SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)
	AddToScheme   = SchemeBuilder.AddToScheme

	SchemeGroupVersion = schema.GroupVersion{Group: GroupName, Version: "v1"}

	Resources = []struct {
		Singular   string
		Plural     string
		ShortNames []string
		Scope      apiextensionsv1beta1.ResourceScope
		Kind       string
		ListKind   string
		Type       runtime.Object
		ListType   runtime.Object
	}{
		{
			Singular:   ResourceSingularComponentBuild,
			Plural:     ResourcePluralComponentBuild,
			ShortNames: []string{ResourceShortNameComponentBuild},
			Scope:      ResourceScopeComponentBuild,
			Kind:       "ComponentBuild",
			ListKind:   "ComponentBuildList",
			Type:       &ComponentBuild{},
			ListType:   &ComponentBuildList{},
		},
		{
			Singular:   ResourceSingularConfig,
			Plural:     ResourcePluralConfig,
			ShortNames: []string{},
			Scope:      ResourceScopeConfig,
			Kind:       "Config",
			ListKind:   "ConfigList",
			Type:       &Config{},
			ListType:   &ConfigList{},
		},
		{
			Singular:   ResourceSingularEndpoint,
			Plural:     ResourcePluralEndpoint,
			ShortNames: []string{ResourceShortNameEndpoint},
			Scope:      ResourceScopeEndpoint,
			Kind:       "Endpoint",
			ListKind:   "EndpointList",
			Type:       &Endpoint{},
			ListType:   &EndpointList{},
		},
		{
			Singular:   ResourceSingularLoadBalancer,
			Plural:     ResourcePluralLoadBalancer,
			ShortNames: []string{ResourceShortNameLoadBalancer},
			Scope:      ResourceScopeLoadBalancer,
			Kind:       "LoadBalancer",
			ListKind:   "LoadBalancerList",
			Type:       &LoadBalancer{},
			ListType:   &LoadBalancerList{},
		},
		{
			Singular:   ResourceSingularNodePool,
			Plural:     ResourcePluralNodePool,
			ShortNames: []string{ResourceShortNameNodePool},
			Scope:      ResourceScopeNodePool,
			Kind:       "NodePool",
			ListKind:   "NodePoolList",
			Type:       &NodePool{},
			ListType:   &NodePoolList{},
		},
		{
			Singular:   ResourceSingularService,
			Plural:     ResourcePluralService,
			ShortNames: []string{ResourceShortNameService},
			Scope:      ResourceScopeService,
			Kind:       "Service",
			ListKind:   "ServiceList",
			Type:       &Service{},
			ListType:   &ServiceList{},
		},
		{
			Singular:   ResourceSingularServiceAddress,
			Plural:     ResourcePluralServiceAddress,
			ShortNames: []string{ResourceShortNameServiceAddress},
			Scope:      ResourceScopeServiceAddress,
			Kind:       "ServiceAddress",
			ListKind:   "ServiceAddressList",
			Type:       &ServiceAddress{},
			ListType:   &ServiceAddressList{},
		},
		{
			Singular:   ResourceSingularServiceBuild,
			Plural:     ResourcePluralServiceBuild,
			ShortNames: []string{ResourceShortNameServiceBuild},
			Scope:      ResourceScopeServiceBuild,
			Kind:       "ServiceBuild",
			ListKind:   "ServiceBuildList",
			Type:       &ServiceBuild{},
			ListType:   &ServiceBuildList{},
		},
		{
			Singular:   ResourceSingularSystem,
			Plural:     ResourcePluralSystem,
			ShortNames: []string{ResourceShortNameSystem},
			Scope:      ResourceScopeSystem,
			Kind:       "System",
			ListKind:   "SystemList",
			Type:       &System{},
			ListType:   &SystemList{},
		},
		{
			Singular:   ResourceSingularSystemBuild,
			Plural:     ResourcePluralSystemBuild,
			ShortNames: []string{ResourceShortNameSystemBuild},
			Scope:      ResourceScopeSystemBuild,
			Kind:       "SystemBuild",
			ListKind:   "SystemBuildList",
			Type:       &SystemBuild{},
			ListType:   &SystemBuildList{},
		},
		{
			Singular:   ResourceSingularSystemRollout,
			Plural:     ResourcePluralSystemRollout,
			ShortNames: []string{ResourceShortNameSystemRollout},
			Scope:      ResourceScopeSystemRollout,
			Kind:       "SystemRollout",
			ListKind:   "SystemRolloutList",
			Type:       &SystemRollout{},
			ListType:   &SystemRolloutList{},
		},
		{
			Singular:   ResourceSingularSystemTeardown,
			Plural:     ResourcePluralSystemTeardown,
			ShortNames: []string{ResourceShortNameSystemTeardown},
			Scope:      ResourceScopeSystemTeardown,
			Kind:       "SystemTeardown",
			ListKind:   "SystemTeardownList",
			Type:       &SystemTeardown{},
			ListType:   &SystemTeardownList{},
		},
	}
)

// Resource takes an unqualified resource and returns a Group-qualified GroupResource.
func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}

func GroupVersionResource(resource string) schema.GroupVersionResource {
	return SchemeGroupVersion.WithResource(resource)
}

// addKnownTypes adds the set of types defined in this package to the supplied scheme.
func addKnownTypes(scheme *runtime.Scheme) error {
	for _, resource := range Resources {
		scheme.AddKnownTypes(
			SchemeGroupVersion,
			resource.Type.(runtime.Object),
			resource.ListType.(runtime.Object),
		)
	}
	metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}

func GetCustomResourceDefinitions() []*apiextensionsv1beta1.CustomResourceDefinition {
	var definitions []*apiextensionsv1beta1.CustomResourceDefinition
	for _, resource := range Resources {
		name := resource.Plural + "." + GroupName

		definition := &apiextensionsv1beta1.CustomResourceDefinition{
			// Include TypeMeta so if this is a dry run it will be printed out
			TypeMeta: metav1.TypeMeta{
				Kind:       "CustomResourceDefinition",
				APIVersion: apiextensionsv1beta1.GroupName + "/v1beta1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
			Spec: apiextensionsv1beta1.CustomResourceDefinitionSpec{
				Group:   GroupName,
				Version: SchemeGroupVersion.Version,
				Scope:   resource.Scope,
				Names: apiextensionsv1beta1.CustomResourceDefinitionNames{
					Singular:   resource.Singular,
					Plural:     resource.Plural,
					ShortNames: resource.ShortNames,
					Kind:       resource.Kind,
					ListKind:   resource.ListKind,
				},
			},
		}

		definitions = append(definitions, definition)
	}
	return definitions
}
