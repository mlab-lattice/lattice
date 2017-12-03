package v1

import (
	"github.com/mlab-lattice/system/pkg/kubernetes/customresource/v1"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"

	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

// SystemTeardownsGetter has a method to return a SystemTeardownInterface.
// A group's client should implement this interface.
type SystemTeardownsGetter interface {
	SystemTeardowns(namespace string) SystemTeardownInterface
}

// SystemTeardownInterface has methods to work with SystemTeardown resources.
type SystemTeardownInterface interface {
	Create(*v1.SystemTeardown) (*v1.SystemTeardown, error)
	Update(*v1.SystemTeardown) (*v1.SystemTeardown, error)
	UpdateStatus(*v1.SystemTeardown) (*v1.SystemTeardown, error)
	Delete(name string, options *meta_v1.DeleteOptions) error
	DeleteCollection(options *meta_v1.DeleteOptions, listOptions meta_v1.ListOptions) error
	Get(name string, options meta_v1.GetOptions) (*v1.SystemTeardown, error)
	List(opts meta_v1.ListOptions) (*v1.SystemTeardownList, error)
	Watch(opts meta_v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.SystemTeardown, err error)
}

// SystemTeardowns implements SystemTeardownInterface
type SystemTeardowns struct {
	client rest.Interface
	ns     string
}

// newSystemTeardowns returns a SystemTeardowns
func newSystemTeardowns(c *V1Client, namespace string) *SystemTeardowns {
	return &SystemTeardowns{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the SystemTeardown, and returns the corresponding SystemTeardown object, and an error if there is any.
func (c *SystemTeardowns) Get(name string, options meta_v1.GetOptions) (result *v1.SystemTeardown, err error) {
	result = &v1.SystemTeardown{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource(v1.ResourcePluralSystemTeardown).
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of SystemTeardowns that match those selectors.
func (c *SystemTeardowns) List(opts meta_v1.ListOptions) (result *v1.SystemTeardownList, err error) {
	result = &v1.SystemTeardownList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource(v1.ResourcePluralSystemTeardown).
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested SystemTeardowns.
func (c *SystemTeardowns) Watch(opts meta_v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource(v1.ResourcePluralSystemTeardown).
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a SystemTeardown and creates it.  Returns the server's representation of the SystemTeardown, and an error, if there is any.
func (c *SystemTeardowns) Create(SystemTeardown *v1.SystemTeardown) (result *v1.SystemTeardown, err error) {
	result = &v1.SystemTeardown{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource(v1.ResourcePluralSystemTeardown).
		Body(SystemTeardown).
		Do().
		Into(result)
	return
}

// Update takes the representation of a SystemTeardown and updates it. Returns the server's representation of the SystemTeardown, and an error, if there is any.
func (c *SystemTeardowns) Update(SystemTeardown *v1.SystemTeardown) (result *v1.SystemTeardown, err error) {
	result = &v1.SystemTeardown{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource(v1.ResourcePluralSystemTeardown).
		Name(SystemTeardown.Name).
		Body(SystemTeardown).
		Do().
		Into(result)
	return
}

func (c *SystemTeardowns) UpdateStatus(SystemTeardown *v1.SystemTeardown) (result *v1.SystemTeardown, err error) {
	result = &v1.SystemTeardown{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource(v1.ResourcePluralSystemTeardown).
		Name(SystemTeardown.Name).
		SubResource("status").
		Body(SystemTeardown).
		Do().
		Into(result)
	return
}

// Delete takes name of the SystemTeardown and deletes it. Returns an error if one occurs.
func (c *SystemTeardowns) Delete(name string, options *meta_v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource(v1.ResourcePluralSystemTeardown).
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *SystemTeardowns) DeleteCollection(options *meta_v1.DeleteOptions, listOptions meta_v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource(v1.ResourcePluralSystemTeardown).
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched SystemTeardown.
func (c *SystemTeardowns) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.SystemTeardown, err error) {
	result = &v1.SystemTeardown{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource(v1.ResourcePluralSystemTeardown).
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
