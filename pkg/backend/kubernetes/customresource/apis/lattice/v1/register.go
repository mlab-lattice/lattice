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
		Singular          string
		Plural            string
		Scope             apiextensionsv1beta1.ResourceScope
		Kind              string
		ListKind          string
		Type              runtime.Object
		ListType          runtime.Object
		StatusSubresource bool
	}{
		{
			Singular:          ResourceSingularAddress,
			Plural:            ResourcePluralAddress,
			Scope:             ResourceScopeAddress,
			Kind:              AddressKind.Kind,
			ListKind:          AddressListKind.Kind,
			Type:              &Address{},
			ListType:          &AddressList{},
			StatusSubresource: true,
		},
		{
			Singular:          ResourceSingularBuild,
			Plural:            ResourcePluralBuild,
			Scope:             ResourceScopeBuild,
			Kind:              BuildKind.Kind,
			ListKind:          BuildListKind.Kind,
			Type:              &Build{},
			ListType:          &BuildList{},
			StatusSubresource: true,
		},
		{
			Singular:          ResourceSingularContainerBuild,
			Plural:            ResourcePluralContainerBuild,
			Scope:             ResourceScopeContainerBuild,
			Kind:              ContainerBuildKind.Kind,
			ListKind:          ContainerBuildListKind.Kind,
			Type:              &ContainerBuild{},
			ListType:          &ComponentBuildList{},
			StatusSubresource: true,
		},
		{
			Singular:          ResourceSingularConfig,
			Plural:            ResourcePluralConfig,
			Scope:             ResourceScopeConfig,
			Kind:              ConfigKind.Kind,
			ListKind:          ConfigListKind.Kind,
			Type:              &Config{},
			ListType:          &ConfigList{},
			StatusSubresource: false,
		},
		{
			Singular:          ResourceSingularDeploy,
			Plural:            ResourcePluralDeploy,
			Scope:             ResourceScopeDeploy,
			Kind:              DeployKind.Kind,
			ListKind:          DeployListKind.Kind,
			Type:              &Deploy{},
			ListType:          &DeployList{},
			StatusSubresource: true,
		},
		{
			Singular:          ResourceSingularNodePool,
			Plural:            ResourcePluralNodePool,
			Scope:             ResourceScopeNodePool,
			Kind:              NodePoolKind.Kind,
			ListKind:          NodePoolListKind.Kind,
			Type:              &NodePool{},
			ListType:          &NodePoolList{},
			StatusSubresource: true,
		},
		{
			Singular:          ResourceSingularService,
			Plural:            ResourcePluralService,
			Scope:             ResourceScopeService,
			Kind:              ServiceKind.Kind,
			ListKind:          ServiceListKind.Kind,
			Type:              &Service{},
			ListType:          &ServiceList{},
			StatusSubresource: true,
		},
		{
			Singular:          ResourceSingularServiceBuild,
			Plural:            ResourcePluralServiceBuild,
			Scope:             ResourceScopeServiceBuild,
			Kind:              ServiceBuildKind.Kind,
			ListKind:          ServiceBuildListKind.Kind,
			Type:              &ServiceBuild{},
			ListType:          &ServiceBuildList{},
			StatusSubresource: true,
		},
		{
			Singular:          ResourceSingularSystem,
			Plural:            ResourcePluralSystem,
			Scope:             ResourceScopeSystem,
			Kind:              SystemKind.Kind,
			ListKind:          SystemListKind.Kind,
			Type:              &System{},
			ListType:          &SystemList{},
			StatusSubresource: true,
		},
		{
			Singular:          ResourceSingularTeardown,
			Plural:            ResourcePluralTeardown,
			Scope:             ResourceScopeTeardown,
			Kind:              TeardownKind.Kind,
			ListKind:          TeardownListKind.Kind,
			Type:              &Teardown{},
			ListType:          &TeardownList{},
			StatusSubresource: true,
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
					Kind:       resource.Kind,
					ListKind:   resource.ListKind,
					Categories: []string{"all", "lattice"},
				},
			},
		}

		if resource.StatusSubresource {
			definition.Spec.Subresources = &apiextensionsv1beta1.CustomResourceSubresources{
				Status: &apiextensionsv1beta1.CustomResourceSubresourceStatus{},
			}
		}

		definitions = append(definitions, definition)
	}
	return definitions
}
