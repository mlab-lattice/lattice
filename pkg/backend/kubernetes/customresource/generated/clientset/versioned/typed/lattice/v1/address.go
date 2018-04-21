package v1

import (
	v1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	scheme "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/generated/clientset/versioned/scheme"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// AddressesGetter has a method to return a AddressInterface.
// A group's client should implement this interface.
type AddressesGetter interface {
	Addresses(namespace string) AddressInterface
}

// AddressInterface has methods to work with Address resources.
type AddressInterface interface {
	Create(*v1.Address) (*v1.Address, error)
	Update(*v1.Address) (*v1.Address, error)
	UpdateStatus(*v1.Address) (*v1.Address, error)
	Delete(name string, options *meta_v1.DeleteOptions) error
	DeleteCollection(options *meta_v1.DeleteOptions, listOptions meta_v1.ListOptions) error
	Get(name string, options meta_v1.GetOptions) (*v1.Address, error)
	List(opts meta_v1.ListOptions) (*v1.AddressList, error)
	Watch(opts meta_v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.Address, err error)
	AddressExpansion
}

// addresses implements AddressInterface
type addresses struct {
	client rest.Interface
	ns     string
}

// newAddresses returns a Addresses
func newAddresses(c *LatticeV1Client, namespace string) *addresses {
	return &addresses{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the address, and returns the corresponding address object, and an error if there is any.
func (c *addresses) Get(name string, options meta_v1.GetOptions) (result *v1.Address, err error) {
	result = &v1.Address{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("addresses").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of Addresses that match those selectors.
func (c *addresses) List(opts meta_v1.ListOptions) (result *v1.AddressList, err error) {
	result = &v1.AddressList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("addresses").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested addresses.
func (c *addresses) Watch(opts meta_v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("addresses").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a address and creates it.  Returns the server's representation of the address, and an error, if there is any.
func (c *addresses) Create(address *v1.Address) (result *v1.Address, err error) {
	result = &v1.Address{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("addresses").
		Body(address).
		Do().
		Into(result)
	return
}

// Update takes the representation of a address and updates it. Returns the server's representation of the address, and an error, if there is any.
func (c *addresses) Update(address *v1.Address) (result *v1.Address, err error) {
	result = &v1.Address{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("addresses").
		Name(address.Name).
		Body(address).
		Do().
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().

func (c *addresses) UpdateStatus(address *v1.Address) (result *v1.Address, err error) {
	result = &v1.Address{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("addresses").
		Name(address.Name).
		SubResource("status").
		Body(address).
		Do().
		Into(result)
	return
}

// Delete takes name of the address and deletes it. Returns an error if one occurs.
func (c *addresses) Delete(name string, options *meta_v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("addresses").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *addresses) DeleteCollection(options *meta_v1.DeleteOptions, listOptions meta_v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("addresses").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched address.
func (c *addresses) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.Address, err error) {
	result = &v1.Address{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("addresses").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
