package v1

import (
	"github.com/mlab-lattice/system/pkg/kubernetes/customresource/v1"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"

	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
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
}

// SystemRollouts implements SystemRolloutInterface
type SystemRollouts struct {
	client rest.Interface
	ns     string
}

// newSystemRollouts returns a SystemRollouts
func newSystemRollouts(c *V1Client, namespace string) *SystemRollouts {
	return &SystemRollouts{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the SystemRollout, and returns the corresponding SystemRollout object, and an error if there is any.
func (c *SystemRollouts) Get(name string, options meta_v1.GetOptions) (result *v1.SystemRollout, err error) {
	result = &v1.SystemRollout{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource(v1.ResourcePluralSystemRollout).
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of SystemRollouts that match those selectors.
func (c *SystemRollouts) List(opts meta_v1.ListOptions) (result *v1.SystemRolloutList, err error) {
	result = &v1.SystemRolloutList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource(v1.ResourcePluralSystemRollout).
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested SystemRollouts.
func (c *SystemRollouts) Watch(opts meta_v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource(v1.ResourcePluralSystemRollout).
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a SystemRollout and creates it.  Returns the server's representation of the SystemRollout, and an error, if there is any.
func (c *SystemRollouts) Create(SystemRollout *v1.SystemRollout) (result *v1.SystemRollout, err error) {
	result = &v1.SystemRollout{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource(v1.ResourcePluralSystemRollout).
		Body(SystemRollout).
		Do().
		Into(result)
	return
}

// Update takes the representation of a SystemRollout and updates it. Returns the server's representation of the SystemRollout, and an error, if there is any.
func (c *SystemRollouts) Update(SystemRollout *v1.SystemRollout) (result *v1.SystemRollout, err error) {
	result = &v1.SystemRollout{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource(v1.ResourcePluralSystemRollout).
		Name(SystemRollout.Name).
		Body(SystemRollout).
		Do().
		Into(result)
	return
}

func (c *SystemRollouts) UpdateStatus(SystemRollout *v1.SystemRollout) (result *v1.SystemRollout, err error) {
	result = &v1.SystemRollout{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource(v1.ResourcePluralSystemRollout).
		Name(SystemRollout.Name).
		SubResource("status").
		Body(SystemRollout).
		Do().
		Into(result)
	return
}

// Delete takes name of the SystemRollout and deletes it. Returns an error if one occurs.
func (c *SystemRollouts) Delete(name string, options *meta_v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource(v1.ResourcePluralSystemRollout).
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *SystemRollouts) DeleteCollection(options *meta_v1.DeleteOptions, listOptions meta_v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource(v1.ResourcePluralSystemRollout).
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched SystemRollout.
func (c *SystemRollouts) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.SystemRollout, err error) {
	result = &v1.SystemRollout{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource(v1.ResourcePluralSystemRollout).
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
