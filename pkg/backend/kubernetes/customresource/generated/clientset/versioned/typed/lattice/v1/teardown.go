package v1

import (
	v1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	scheme "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/generated/clientset/versioned/scheme"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// TeardownsGetter has a method to return a TeardownInterface.
// A group's client should implement this interface.
type TeardownsGetter interface {
	Teardowns(namespace string) TeardownInterface
}

// TeardownInterface has methods to work with Teardown resources.
type TeardownInterface interface {
	Create(*v1.Teardown) (*v1.Teardown, error)
	Update(*v1.Teardown) (*v1.Teardown, error)
	UpdateStatus(*v1.Teardown) (*v1.Teardown, error)
	Delete(name string, options *meta_v1.DeleteOptions) error
	DeleteCollection(options *meta_v1.DeleteOptions, listOptions meta_v1.ListOptions) error
	Get(name string, options meta_v1.GetOptions) (*v1.Teardown, error)
	List(opts meta_v1.ListOptions) (*v1.TeardownList, error)
	Watch(opts meta_v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.Teardown, err error)
	TeardownExpansion
}

// teardowns implements TeardownInterface
type teardowns struct {
	client rest.Interface
	ns     string
}

// newTeardowns returns a Teardowns
func newTeardowns(c *LatticeV1Client, namespace string) *teardowns {
	return &teardowns{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the teardown, and returns the corresponding teardown object, and an error if there is any.
func (c *teardowns) Get(name string, options meta_v1.GetOptions) (result *v1.Teardown, err error) {
	result = &v1.Teardown{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("teardowns").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of Teardowns that match those selectors.
func (c *teardowns) List(opts meta_v1.ListOptions) (result *v1.TeardownList, err error) {
	result = &v1.TeardownList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("teardowns").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested teardowns.
func (c *teardowns) Watch(opts meta_v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("teardowns").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a teardown and creates it.  Returns the server's representation of the teardown, and an error, if there is any.
func (c *teardowns) Create(teardown *v1.Teardown) (result *v1.Teardown, err error) {
	result = &v1.Teardown{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("teardowns").
		Body(teardown).
		Do().
		Into(result)
	return
}

// Update takes the representation of a teardown and updates it. Returns the server's representation of the teardown, and an error, if there is any.
func (c *teardowns) Update(teardown *v1.Teardown) (result *v1.Teardown, err error) {
	result = &v1.Teardown{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("teardowns").
		Name(teardown.Name).
		Body(teardown).
		Do().
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().

func (c *teardowns) UpdateStatus(teardown *v1.Teardown) (result *v1.Teardown, err error) {
	result = &v1.Teardown{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("teardowns").
		Name(teardown.Name).
		SubResource("status").
		Body(teardown).
		Do().
		Into(result)
	return
}

// Delete takes name of the teardown and deletes it. Returns an error if one occurs.
func (c *teardowns) Delete(name string, options *meta_v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("teardowns").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *teardowns) DeleteCollection(options *meta_v1.DeleteOptions, listOptions meta_v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("teardowns").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched teardown.
func (c *teardowns) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.Teardown, err error) {
	result = &v1.Teardown{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("teardowns").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
