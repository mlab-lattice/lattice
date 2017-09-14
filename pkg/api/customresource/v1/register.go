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
			Singular:   ComponentBuildResourceSingular,
			Plural:     ComponentBuildResourcePlural,
			ShortNames: []string{ComponentBuildResourceShortName},
			Scope:      ComponentBuildResourceScope,
			Kind:       "ComponentBuild",
			ListKind:   "ComponentBuildList",
			Type:       &ComponentBuild{},
			ListType:   &ComponentBuildList{},
		},
		{
			Singular:   ConfigResourceSingular,
			Plural:     ConfigResourcePlural,
			ShortNames: []string{},
			Scope:      ConfigResourceScope,
			Kind:       "Config",
			ListKind:   "ConfigList",
			Type:       &Config{},
			ListType:   &ConfigList{},
		},
		{
			Singular:   ServiceResourceSingular,
			Plural:     ServiceResourcePlural,
			ShortNames: []string{ServiceResourceShortName},
			Scope:      ServiceResourceScope,
			Kind:       "Service",
			ListKind:   "ServiceList",
			Type:       &Service{},
			ListType:   &ServiceList{},
		},
		{
			Singular:   ServiceBuildResourceSingular,
			Plural:     ServiceBuildResourcePlural,
			ShortNames: []string{ServiceBuildResourceShortName},
			Scope:      ServiceBuildResourceScope,
			Kind:       "ServiceBuild",
			ListKind:   "ServiceBuildList",
			Type:       &ServiceBuild{},
			ListType:   &ServiceBuildList{},
		},
		{
			Singular:   SystemResourceSingular,
			Plural:     SystemResourcePlural,
			ShortNames: []string{SystemResourceShortName},
			Scope:      SystemResourceScope,
			Kind:       "System",
			ListKind:   "SystemList",
			Type:       &System{},
			ListType:   &SystemList{},
		},
		{
			Singular:   SystemBuildResourceSingular,
			Plural:     SystemBuildResourcePlural,
			ShortNames: []string{SystemBuildResourceShortName},
			Scope:      SystemBuildResourceScope,
			Kind:       "SystemBuild",
			ListKind:   "SystemBuildList",
			Type:       &SystemBuild{},
			ListType:   &SystemBuildList{},
		},
	}
)

// Resource takes an unqualified resource and returns a Group-qualified GroupResource.
func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
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
