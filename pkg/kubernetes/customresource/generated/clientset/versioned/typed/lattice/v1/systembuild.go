package v1

import (
	v1 "github.com/mlab-lattice/system/pkg/kubernetes/customresource/apis/lattice/v1"
	scheme "github.com/mlab-lattice/system/pkg/kubernetes/customresource/generated/clientset/versioned/scheme"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
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
	Delete(name string, options *meta_v1.DeleteOptions) error
	DeleteCollection(options *meta_v1.DeleteOptions, listOptions meta_v1.ListOptions) error
	Get(name string, options meta_v1.GetOptions) (*v1.SystemBuild, error)
	List(opts meta_v1.ListOptions) (*v1.SystemBuildList, error)
	Watch(opts meta_v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.SystemBuild, err error)
	SystemBuildExpansion
}

// systemBuilds implements SystemBuildInterface
type systemBuilds struct {
	client rest.Interface
	ns     string
}

// newSystemBuilds returns a SystemBuilds
func newSystemBuilds(c *LatticeV1Client, namespace string) *systemBuilds {
	return &systemBuilds{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the systemBuild, and returns the corresponding systemBuild object, and an error if there is any.
func (c *systemBuilds) Get(name string, options meta_v1.GetOptions) (result *v1.SystemBuild, err error) {
	result = &v1.SystemBuild{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("systembuilds").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of SystemBuilds that match those selectors.
func (c *systemBuilds) List(opts meta_v1.ListOptions) (result *v1.SystemBuildList, err error) {
	result = &v1.SystemBuildList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("systembuilds").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested systemBuilds.
func (c *systemBuilds) Watch(opts meta_v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("systembuilds").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a systemBuild and creates it.  Returns the server's representation of the systemBuild, and an error, if there is any.
func (c *systemBuilds) Create(systemBuild *v1.SystemBuild) (result *v1.SystemBuild, err error) {
	result = &v1.SystemBuild{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("systembuilds").
		Body(systemBuild).
		Do().
		Into(result)
	return
}

// Update takes the representation of a systemBuild and updates it. Returns the server's representation of the systemBuild, and an error, if there is any.
func (c *systemBuilds) Update(systemBuild *v1.SystemBuild) (result *v1.SystemBuild, err error) {
	result = &v1.SystemBuild{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("systembuilds").
		Name(systemBuild.Name).
		Body(systemBuild).
		Do().
		Into(result)
	return
}

// Delete takes name of the systemBuild and deletes it. Returns an error if one occurs.
func (c *systemBuilds) Delete(name string, options *meta_v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("systembuilds").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *systemBuilds) DeleteCollection(options *meta_v1.DeleteOptions, listOptions meta_v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("systembuilds").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched systemBuild.
func (c *systemBuilds) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.SystemBuild, err error) {
	result = &v1.SystemBuild{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("systembuilds").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
