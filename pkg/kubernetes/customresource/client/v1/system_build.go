package v1

import (
	"github.com/mlab-lattice/system/pkg/kubernetes/customresource/v1"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"

	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/rest"
)

// SystemBuildsGetter has a method to return a SystemBuildInterface.
// A group's client should implement this interface.
type SystemBuildsGetter interface {
	SystemBuilds(namespace string) SystemBuildInterface
}

// SystemBuildInterface has methods to work with SystemBuild resources.
type SystemBuildInterface interface {
	Create(*v1.SystemBuild) (*v1.SystemBuild, error)
	Update(*v1.SystemBuild) (*v1.SystemBuild, error)
	UpdateStatus(*v1.SystemBuild) (*v1.SystemBuild, error)
	Delete(name string, options *meta_v1.DeleteOptions) error
	DeleteCollection(options *meta_v1.DeleteOptions, listOptions meta_v1.ListOptions) error
	Get(name string, options meta_v1.GetOptions) (*v1.SystemBuild, error)
	List(opts meta_v1.ListOptions) (*v1.SystemBuildList, error)
	Watch(opts meta_v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.SystemBuild, err error)
}

// SystemBuilds implements SystemBuildInterface
type SystemBuilds struct {
	client rest.Interface
	ns     string
}

// newSystemBuilds returns a SystemBuilds
func newSystemBuilds(c *Client, namespace string) *SystemBuilds {
	return &SystemBuilds{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the SystemBuild, and returns the corresponding SystemBuild object, and an error if there is any.
func (c *SystemBuilds) Get(name string, options meta_v1.GetOptions) (result *v1.SystemBuild, err error) {
	result = &v1.SystemBuild{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource(v1.ResourcePluralSystemBuild).
		Name(name).
		VersionedParams(&options, ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of SystemBuilds that match those selectors.
func (c *SystemBuilds) List(opts meta_v1.ListOptions) (result *v1.SystemBuildList, err error) {
	result = &v1.SystemBuildList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource(v1.ResourcePluralSystemBuild).
		VersionedParams(&opts, ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested SystemBuilds.
func (c *SystemBuilds) Watch(opts meta_v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource(v1.ResourcePluralSystemBuild).
		VersionedParams(&opts, ParameterCodec).
		Watch()
}

// Create takes the representation of a SystemBuild and creates it.  Returns the server's representation of the SystemBuild, and an error, if there is any.
func (c *SystemBuilds) Create(SystemBuild *v1.SystemBuild) (result *v1.SystemBuild, err error) {
	result = &v1.SystemBuild{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource(v1.ResourcePluralSystemBuild).
		Body(SystemBuild).
		Do().
		Into(result)
	return
}

// Update takes the representation of a SystemBuild and updates it. Returns the server's representation of the SystemBuild, and an error, if there is any.
func (c *SystemBuilds) Update(SystemBuild *v1.SystemBuild) (result *v1.SystemBuild, err error) {
	result = &v1.SystemBuild{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource(v1.ResourcePluralSystemBuild).
		Name(SystemBuild.Name).
		Body(SystemBuild).
		Do().
		Into(result)
	return
}

func (c *SystemBuilds) UpdateStatus(SystemBuild *v1.SystemBuild) (result *v1.SystemBuild, err error) {
	result = &v1.SystemBuild{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource(v1.ResourcePluralSystemBuild).
		Name(SystemBuild.Name).
		SubResource("status").
		Body(SystemBuild).
		Do().
		Into(result)
	return
}

// Delete takes name of the SystemBuild and deletes it. Returns an error if one occurs.
func (c *SystemBuilds) Delete(name string, options *meta_v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource(v1.ResourcePluralSystemBuild).
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *SystemBuilds) DeleteCollection(options *meta_v1.DeleteOptions, listOptions meta_v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource(v1.ResourcePluralSystemBuild).
		VersionedParams(&listOptions, ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched SystemBuild.
func (c *SystemBuilds) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.SystemBuild, err error) {
	result = &v1.SystemBuild{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource(v1.ResourcePluralSystemBuild).
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
