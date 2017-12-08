package v1

import (
	v1 "github.com/mlab-lattice/system/pkg/kubernetes/customresource/apis/lattice/v1"
	scheme "github.com/mlab-lattice/system/pkg/kubernetes/customresource/generated/clientset/versioned/scheme"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
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
	Delete(name string, options *meta_v1.DeleteOptions) error
	DeleteCollection(options *meta_v1.DeleteOptions, listOptions meta_v1.ListOptions) error
	Get(name string, options meta_v1.GetOptions) (*v1.SystemTeardown, error)
	List(opts meta_v1.ListOptions) (*v1.SystemTeardownList, error)
	Watch(opts meta_v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.SystemTeardown, err error)
	SystemTeardownExpansion
}

// systemTeardowns implements SystemTeardownInterface
type systemTeardowns struct {
	client rest.Interface
	ns     string
}

// newSystemTeardowns returns a SystemTeardowns
func newSystemTeardowns(c *LatticeV1Client, namespace string) *systemTeardowns {
	return &systemTeardowns{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the systemTeardown, and returns the corresponding systemTeardown object, and an error if there is any.
func (c *systemTeardowns) Get(name string, options meta_v1.GetOptions) (result *v1.SystemTeardown, err error) {
	result = &v1.SystemTeardown{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("systemteardowns").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of SystemTeardowns that match those selectors.
func (c *systemTeardowns) List(opts meta_v1.ListOptions) (result *v1.SystemTeardownList, err error) {
	result = &v1.SystemTeardownList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("systemteardowns").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested systemTeardowns.
func (c *systemTeardowns) Watch(opts meta_v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("systemteardowns").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a systemTeardown and creates it.  Returns the server's representation of the systemTeardown, and an error, if there is any.
func (c *systemTeardowns) Create(systemTeardown *v1.SystemTeardown) (result *v1.SystemTeardown, err error) {
	result = &v1.SystemTeardown{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("systemteardowns").
		Body(systemTeardown).
		Do().
		Into(result)
	return
}

// Update takes the representation of a systemTeardown and updates it. Returns the server's representation of the systemTeardown, and an error, if there is any.
func (c *systemTeardowns) Update(systemTeardown *v1.SystemTeardown) (result *v1.SystemTeardown, err error) {
	result = &v1.SystemTeardown{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("systemteardowns").
		Name(systemTeardown.Name).
		Body(systemTeardown).
		Do().
		Into(result)
	return
}

// Delete takes name of the systemTeardown and deletes it. Returns an error if one occurs.
func (c *systemTeardowns) Delete(name string, options *meta_v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("systemteardowns").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *systemTeardowns) DeleteCollection(options *meta_v1.DeleteOptions, listOptions meta_v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("systemteardowns").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched systemTeardown.
func (c *systemTeardowns) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.SystemTeardown, err error) {
	result = &v1.SystemTeardown{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("systemteardowns").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
