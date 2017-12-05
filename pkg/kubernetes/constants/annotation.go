package constants

// Keys used in "k8s.io/apimachinery/pkg/apis/meta/v1".ObjectMeta annotations
const (
	AnnotationKeyComponentBuildDefinitionHash = "component.build.lattice.mlab.com/definition-hash"
	AnnotationKeyDeploymentServiceDefinition  = "service.lattice.mlab.com/definition"
	// FIXME: remove this when local DNS works
	AnnotationKeySystemServices = "system.lattice.mlab.com/services"
)
