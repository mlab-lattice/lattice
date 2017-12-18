package v1

import (
	v1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	scheme "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/generated/clientset/versioned/scheme"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// SystemRolloutsGetter has a method to return a SystemRolloutInterface.
// A group's client should implement this interface.
type SystemRolloutsGetter interface {
	SystemRollouts(namespace string) SystemRolloutInterface
}

// SystemRolloutInterface has methods to work with SystemRollout resources.
type SystemRolloutInterface interface {
	Create(*v1.SystemRollout) (*v1.SystemRollout, error)
	Update(*v1.SystemRollout) (*v1.SystemRollout, error)
	UpdateStatus(*v1.SystemRollout) (*v1.SystemRollout, error)
	Delete(name string, options *meta_v1.DeleteOptions) error
	DeleteCollection(options *meta_v1.DeleteOptions, listOptions meta_v1.ListOptions) error
	Get(name string, options meta_v1.GetOptions) (*v1.SystemRollout, error)
	List(opts meta_v1.ListOptions) (*v1.SystemRolloutList, error)
	Watch(opts meta_v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.SystemRollout, err error)
	SystemRolloutExpansion
}

// systemRollouts implements SystemRolloutInterface
type systemRollouts struct {
	client rest.Interface
	ns     string
}

// newSystemRollouts returns a SystemRollouts
func newSystemRollouts(c *LatticeV1Client, namespace string) *systemRollouts {
	return &systemRollouts{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the systemRollout, and returns the corresponding systemRollout object, and an error if there is any.
func (c *systemRollouts) Get(name string, options meta_v1.GetOptions) (result *v1.SystemRollout, err error) {
	result = &v1.SystemRollout{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("systemrollouts").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of SystemRollouts that match those selectors.
func (c *systemRollouts) List(opts meta_v1.ListOptions) (result *v1.SystemRolloutList, err error) {
	result = &v1.SystemRolloutList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("systemrollouts").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested systemRollouts.
func (c *systemRollouts) Watch(opts meta_v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("systemrollouts").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a systemRollout and creates it.  Returns the server's representation of the systemRollout, and an error, if there is any.
func (c *systemRollouts) Create(systemRollout *v1.SystemRollout) (result *v1.SystemRollout, err error) {
	result = &v1.SystemRollout{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("systemrollouts").
		Body(systemRollout).
		Do().
		Into(result)
	return
}

// Update takes the representation of a systemRollout and updates it. Returns the server's representation of the systemRollout, and an error, if there is any.
func (c *systemRollouts) Update(systemRollout *v1.SystemRollout) (result *v1.SystemRollout, err error) {
	result = &v1.SystemRollout{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("systemrollouts").
		Name(systemRollout.Name).
		Body(systemRollout).
		Do().
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().

func (c *systemRollouts) UpdateStatus(systemRollout *v1.SystemRollout) (result *v1.SystemRollout, err error) {
	result = &v1.SystemRollout{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("systemrollouts").
		Name(systemRollout.Name).
		SubResource("status").
		Body(systemRollout).
		Do().
		Into(result)
	return
}

// Delete takes name of the systemRollout and deletes it. Returns an error if one occurs.
func (c *systemRollouts) Delete(name string, options *meta_v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("systemrollouts").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *systemRollouts) DeleteCollection(options *meta_v1.DeleteOptions, listOptions meta_v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("systemrollouts").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched systemRollout.
func (c *systemRollouts) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.SystemRollout, err error) {
	result = &v1.SystemRollout{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("systemrollouts").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
