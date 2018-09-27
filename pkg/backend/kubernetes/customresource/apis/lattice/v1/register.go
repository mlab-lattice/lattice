package v1

import (
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
		Type     runtime.Object
		ListType runtime.Object
	}{
		{
			Type:     &Address{},
			ListType: &AddressList{},
		},
		{
			Type:     &Build{},
			ListType: &BuildList{},
		},
		{
			Type:     &ContainerBuild{},
			ListType: &ContainerBuildList{},
		},
		{
			Type:     &Config{},
			ListType: &ConfigList{},
		},
		{
			Type:     &Deploy{},
			ListType: &DeployList{},
		},
		{
			Type:     &GitTemplate{},
			ListType: &GitTemplateList{},
		},
		{
			Type:     &Job{},
			ListType: &JobList{},
		},
		{
			Type:     &JobRun{},
			ListType: &JobRunList{},
		},
		{
			Type:     &NodePool{},
			ListType: &NodePoolList{},
		},
		{
			Type:     &Service{},
			ListType: &ServiceList{},
		},
		{
			Type:     &System{},
			ListType: &SystemList{},
		},
		{
			Type:     &Teardown{},
			ListType: &TeardownList{},
		},
		{
			Type:     &Template{},
			ListType: &TemplateList{},
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
