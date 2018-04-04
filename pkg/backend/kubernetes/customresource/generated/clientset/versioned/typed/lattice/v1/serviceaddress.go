package v1

import (
	v1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	scheme "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/generated/clientset/versioned/scheme"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// ServiceAddressesGetter has a method to return a ServiceAddressInterface.
// A group's client should implement this interface.
type ServiceAddressesGetter interface {
	ServiceAddresses(namespace string) ServiceAddressInterface
}

// ServiceAddressInterface has methods to work with ServiceAddress resources.
type ServiceAddressInterface interface {
	Create(*v1.ServiceAddress) (*v1.ServiceAddress, error)
	Update(*v1.ServiceAddress) (*v1.ServiceAddress, error)
	UpdateStatus(*v1.ServiceAddress) (*v1.ServiceAddress, error)
	Delete(name string, options *meta_v1.DeleteOptions) error
	DeleteCollection(options *meta_v1.DeleteOptions, listOptions meta_v1.ListOptions) error
	Get(name string, options meta_v1.GetOptions) (*v1.ServiceAddress, error)
	List(opts meta_v1.ListOptions) (*v1.ServiceAddressList, error)
	Watch(opts meta_v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.ServiceAddress, err error)
	ServiceAddressExpansion
}

// serviceAddresses implements ServiceAddressInterface
type serviceAddresses struct {
	client rest.Interface
	ns     string
}

// newServiceAddresses returns a ServiceAddresses
func newServiceAddresses(c *LatticeV1Client, namespace string) *serviceAddresses {
	return &serviceAddresses{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the serviceAddress, and returns the corresponding serviceAddress object, and an error if there is any.
func (c *serviceAddresses) Get(name string, options meta_v1.GetOptions) (result *v1.ServiceAddress, err error) {
	result = &v1.ServiceAddress{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("serviceaddresses").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of ServiceAddresses that match those selectors.
func (c *serviceAddresses) List(opts meta_v1.ListOptions) (result *v1.ServiceAddressList, err error) {
	result = &v1.ServiceAddressList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("serviceaddresses").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested serviceAddresses.
func (c *serviceAddresses) Watch(opts meta_v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("serviceaddresses").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a serviceAddress and creates it.  Returns the server's representation of the serviceAddress, and an error, if there is any.
func (c *serviceAddresses) Create(serviceAddress *v1.ServiceAddress) (result *v1.ServiceAddress, err error) {
	result = &v1.ServiceAddress{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("serviceaddresses").
		Body(serviceAddress).
		Do().
		Into(result)
	return
}

// Update takes the representation of a serviceAddress and updates it. Returns the server's representation of the serviceAddress, and an error, if there is any.
func (c *serviceAddresses) Update(serviceAddress *v1.ServiceAddress) (result *v1.ServiceAddress, err error) {
	result = &v1.ServiceAddress{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("serviceaddresses").
		Name(serviceAddress.Name).
		Body(serviceAddress).
		Do().
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().

func (c *serviceAddresses) UpdateStatus(serviceAddress *v1.ServiceAddress) (result *v1.ServiceAddress, err error) {
	result = &v1.ServiceAddress{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("serviceaddresses").
		Name(serviceAddress.Name).
		SubResource("status").
		Body(serviceAddress).
		Do().
		Into(result)
	return
}

// Delete takes name of the serviceAddress and deletes it. Returns an error if one occurs.
func (c *serviceAddresses) Delete(name string, options *meta_v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("serviceaddresses").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *serviceAddresses) DeleteCollection(options *meta_v1.DeleteOptions, listOptions meta_v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("serviceaddresses").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched serviceAddress.
func (c *serviceAddresses) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.ServiceAddress, err error) {
	result = &v1.ServiceAddress{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("serviceaddresses").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
