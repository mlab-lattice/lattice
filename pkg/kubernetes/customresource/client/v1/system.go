package v1

import (
	"github.com/mlab-lattice/system/pkg/kubernetes/customresource/v1"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"

	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/rest"
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
}

// Systems implements SystemInterface
type Systems struct {
	client rest.Interface
	ns     string
}

// newSystems returns a Systems
func newSystems(c *V1Client, namespace string) *Systems {
	return &Systems{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the System, and returns the corresponding System object, and an error if there is any.
func (c *Systems) Get(name string, options meta_v1.GetOptions) (result *v1.System, err error) {
	result = &v1.System{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource(v1.ResourcePluralSystem).
		Name(name).
		VersionedParams(&options, ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of Systems that match those selectors.
func (c *Systems) List(opts meta_v1.ListOptions) (result *v1.SystemList, err error) {
	result = &v1.SystemList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource(v1.ResourcePluralSystem).
		VersionedParams(&opts, ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested Systems.
func (c *Systems) Watch(opts meta_v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource(v1.ResourcePluralSystem).
		VersionedParams(&opts, ParameterCodec).
		Watch()
}

// Create takes the representation of a System and creates it.  Returns the server's representation of the System, and an error, if there is any.
func (c *Systems) Create(System *v1.System) (result *v1.System, err error) {
	result = &v1.System{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource(v1.ResourcePluralSystem).
		Body(System).
		Do().
		Into(result)
	return
}

// Update takes the representation of a System and updates it. Returns the server's representation of the System, and an error, if there is any.
func (c *Systems) Update(System *v1.System) (result *v1.System, err error) {
	result = &v1.System{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource(v1.ResourcePluralSystem).
		Name(System.Name).
		Body(System).
		Do().
		Into(result)
	return
}

func (c *Systems) UpdateStatus(System *v1.System) (result *v1.System, err error) {
	result = &v1.System{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource(v1.ResourcePluralSystem).
		Name(System.Name).
		SubResource("status").
		Body(System).
		Do().
		Into(result)
	return
}

// Delete takes name of the System and deletes it. Returns an error if one occurs.
func (c *Systems) Delete(name string, options *meta_v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource(v1.ResourcePluralSystem).
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *Systems) DeleteCollection(options *meta_v1.DeleteOptions, listOptions meta_v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource(v1.ResourcePluralSystem).
		VersionedParams(&listOptions, ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched System.
func (c *Systems) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.System, err error) {
	result = &v1.System{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource(v1.ResourcePluralSystem).
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
