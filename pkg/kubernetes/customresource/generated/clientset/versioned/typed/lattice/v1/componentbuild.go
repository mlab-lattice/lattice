package v1

import (
	v1 "github.com/mlab-lattice/system/pkg/kubernetes/customresource/apis/lattice/v1"
	scheme "github.com/mlab-lattice/system/pkg/kubernetes/customresource/generated/clientset/versioned/scheme"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// ComponentBuildsGetter has a method to return a ComponentBuildInterface.
// A group's client should implement this interface.
type ComponentBuildsGetter interface {
	ComponentBuilds(namespace string) ComponentBuildInterface
}

// ComponentBuildInterface has methods to work with ComponentBuild resources.
type ComponentBuildInterface interface {
	Create(*v1.ComponentBuild) (*v1.ComponentBuild, error)
	Update(*v1.ComponentBuild) (*v1.ComponentBuild, error)
	Delete(name string, options *meta_v1.DeleteOptions) error
	DeleteCollection(options *meta_v1.DeleteOptions, listOptions meta_v1.ListOptions) error
	Get(name string, options meta_v1.GetOptions) (*v1.ComponentBuild, error)
	List(opts meta_v1.ListOptions) (*v1.ComponentBuildList, error)
	Watch(opts meta_v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.ComponentBuild, err error)
	ComponentBuildExpansion
}

// componentBuilds implements ComponentBuildInterface
type componentBuilds struct {
	client rest.Interface
	ns     string
}

// newComponentBuilds returns a ComponentBuilds
func newComponentBuilds(c *LatticeV1Client, namespace string) *componentBuilds {
	return &componentBuilds{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the componentBuild, and returns the corresponding componentBuild object, and an error if there is any.
func (c *componentBuilds) Get(name string, options meta_v1.GetOptions) (result *v1.ComponentBuild, err error) {
	result = &v1.ComponentBuild{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("componentbuilds").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of ComponentBuilds that match those selectors.
func (c *componentBuilds) List(opts meta_v1.ListOptions) (result *v1.ComponentBuildList, err error) {
	result = &v1.ComponentBuildList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("componentbuilds").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested componentBuilds.
func (c *componentBuilds) Watch(opts meta_v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("componentbuilds").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a componentBuild and creates it.  Returns the server's representation of the componentBuild, and an error, if there is any.
func (c *componentBuilds) Create(componentBuild *v1.ComponentBuild) (result *v1.ComponentBuild, err error) {
	result = &v1.ComponentBuild{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("componentbuilds").
		Body(componentBuild).
		Do().
		Into(result)
	return
}

// Update takes the representation of a componentBuild and updates it. Returns the server's representation of the componentBuild, and an error, if there is any.
func (c *componentBuilds) Update(componentBuild *v1.ComponentBuild) (result *v1.ComponentBuild, err error) {
	result = &v1.ComponentBuild{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("componentbuilds").
		Name(componentBuild.Name).
		Body(componentBuild).
		Do().
		Into(result)
	return
}

// Delete takes name of the componentBuild and deletes it. Returns an error if one occurs.
func (c *componentBuilds) Delete(name string, options *meta_v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("componentbuilds").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *componentBuilds) DeleteCollection(options *meta_v1.DeleteOptions, listOptions meta_v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("componentbuilds").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched componentBuild.
func (c *componentBuilds) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.ComponentBuild, err error) {
	result = &v1.ComponentBuild{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("componentbuilds").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
