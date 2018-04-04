package v1

import (
	v1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	scheme "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/generated/clientset/versioned/scheme"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// SystemsGetter has a method to return a SystemInterface.
// A group's client should implement this interface.
type SystemsGetter interface {
	Systems(namespace string) SystemInterface
}

// SystemInterface has methods to work with System resources.
type SystemInterface interface {
	Create(*v1.System) (*v1.System, error)
	Update(*v1.System) (*v1.System, error)
	UpdateStatus(*v1.System) (*v1.System, error)
	Delete(name string, options *meta_v1.DeleteOptions) error
	DeleteCollection(options *meta_v1.DeleteOptions, listOptions meta_v1.ListOptions) error
	Get(name string, options meta_v1.GetOptions) (*v1.System, error)
	List(opts meta_v1.ListOptions) (*v1.SystemList, error)
	Watch(opts meta_v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.System, err error)
	SystemExpansion
}

// systems implements SystemInterface
type systems struct {
	client rest.Interface
	ns     string
}

// newSystems returns a Systems
func newSystems(c *LatticeV1Client, namespace string) *systems {
	return &systems{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the system, and returns the corresponding system object, and an error if there is any.
func (c *systems) Get(name string, options meta_v1.GetOptions) (result *v1.System, err error) {
	result = &v1.System{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("systems").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of Systems that match those selectors.
func (c *systems) List(opts meta_v1.ListOptions) (result *v1.SystemList, err error) {
	result = &v1.SystemList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("systems").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested systems.
func (c *systems) Watch(opts meta_v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("systems").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a system and creates it.  Returns the server's representation of the system, and an error, if there is any.
func (c *systems) Create(system *v1.System) (result *v1.System, err error) {
	result = &v1.System{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("systems").
		Body(system).
		Do().
		Into(result)
	return
}

// Update takes the representation of a system and updates it. Returns the server's representation of the system, and an error, if there is any.
func (c *systems) Update(system *v1.System) (result *v1.System, err error) {
	result = &v1.System{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("systems").
		Name(system.Name).
		Body(system).
		Do().
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().

func (c *systems) UpdateStatus(system *v1.System) (result *v1.System, err error) {
	result = &v1.System{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("systems").
		Name(system.Name).
		SubResource("status").
		Body(system).
		Do().
		Into(result)
	return
}

// Delete takes name of the system and deletes it. Returns an error if one occurs.
func (c *systems) Delete(name string, options *meta_v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("systems").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *systems) DeleteCollection(options *meta_v1.DeleteOptions, listOptions meta_v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("systems").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched system.
func (c *systems) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.System, err error) {
	result = &v1.System{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("systems").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
