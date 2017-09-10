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
		Singular string
		Plural   string
		Scope    apiextensionsv1beta1.ResourceScope
		Kind     string
		ListKind string
		Type     runtime.Object
		ListType runtime.Object
	}{
		{
			Singular: BuildResourceSingular,
			Plural:   BuildResourcePlural,
			Scope:    BuildResourceScope,
			Kind:     "Build",
			ListKind: "BuildList",
			Type:     &Build{},
			ListType: &BuildList{},
		},
		{
			Singular: ConfigResourceSingular,
			Plural:   ConfigResourcePlural,
			Scope:    ConfigResourceScope,
			Kind:     "Config",
			ListKind: "ConfigList",
			Type:     &Config{},
			ListType: &ConfigList{},
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
		scheme.AddKnownTypes(SchemeGroupVersion,
			resource.Type.(runtime.Object),
			resource.ListType.(runtime.Object),
		)
	}
	metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}
