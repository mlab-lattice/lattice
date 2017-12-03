package v1

import (
	"github.com/mlab-lattice/system/pkg/kubernetes/customresource/v1"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"

	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

// ComponentBuildsGetter has a method to return a ComponentBuildInterface.
// A group's client should implement this interface.
type ComponentBuildsGetter interface {
	ComponentBuilds(namespace string) ComponentBuildInterface
}

// ComponentBuildInterface has methods to work with ComponentBuild resources.
type ComponentBuildInterface interface {
	Create(*v1.ComponentBuild) (*v1.ComponentBuild, error)
	Update(*v1.ComponentBuild) (*v1.ComponentBuild, error)
	UpdateStatus(*v1.ComponentBuild) (*v1.ComponentBuild, error)
	Delete(name string, options *meta_v1.DeleteOptions) error
	DeleteCollection(options *meta_v1.DeleteOptions, listOptions meta_v1.ListOptions) error
	Get(name string, options meta_v1.GetOptions) (*v1.ComponentBuild, error)
	List(opts meta_v1.ListOptions) (*v1.ComponentBuildList, error)
	Watch(opts meta_v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.ComponentBuild, err error)
}

// ComponentBuilds implements ComponentBuildInterface
type ComponentBuilds struct {
	client rest.Interface
	ns     string
}

// newComponentBuilds returns a ComponentBuilds
func newComponentBuilds(c *V1Client, namespace string) *ComponentBuilds {
	return &ComponentBuilds{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the ComponentBuild, and returns the corresponding ComponentBuild object, and an error if there is any.
func (c *ComponentBuilds) Get(name string, options meta_v1.GetOptions) (result *v1.ComponentBuild, err error) {
	result = &v1.ComponentBuild{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource(v1.ResourcePluralComponentBuild).
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of ComponentBuilds that match those selectors.
func (c *ComponentBuilds) List(opts meta_v1.ListOptions) (result *v1.ComponentBuildList, err error) {
	result = &v1.ComponentBuildList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource(v1.ResourcePluralComponentBuild).
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested ComponentBuilds.
func (c *ComponentBuilds) Watch(opts meta_v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource(v1.ResourcePluralComponentBuild).
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a ComponentBuild and creates it.  Returns the server's representation of the ComponentBuild, and an error, if there is any.
func (c *ComponentBuilds) Create(ComponentBuild *v1.ComponentBuild) (result *v1.ComponentBuild, err error) {
	result = &v1.ComponentBuild{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource(v1.ResourcePluralComponentBuild).
		Body(ComponentBuild).
		Do().
		Into(result)
	return
}

// Update takes the representation of a ComponentBuild and updates it. Returns the server's representation of the ComponentBuild, and an error, if there is any.
func (c *ComponentBuilds) Update(ComponentBuild *v1.ComponentBuild) (result *v1.ComponentBuild, err error) {
	result = &v1.ComponentBuild{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource(v1.ResourcePluralComponentBuild).
		Name(ComponentBuild.Name).
		Body(ComponentBuild).
		Do().
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
func (c *ComponentBuilds) UpdateStatus(ComponentBuild *v1.ComponentBuild) (result *v1.ComponentBuild, err error) {
	result = &v1.ComponentBuild{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource(v1.ResourcePluralComponentBuild).
		Name(ComponentBuild.Name).
		SubResource("status").
		Body(ComponentBuild).
		Do().
		Into(result)
	return
}

// Delete takes name of the ComponentBuild and deletes it. Returns an error if one occurs.
func (c *ComponentBuilds) Delete(name string, options *meta_v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource(v1.ResourcePluralComponentBuild).
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *ComponentBuilds) DeleteCollection(options *meta_v1.DeleteOptions, listOptions meta_v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource(v1.ResourcePluralComponentBuild).
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched ComponentBuild.
func (c *ComponentBuilds) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.ComponentBuild, err error) {
	result = &v1.ComponentBuild{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource(v1.ResourcePluralComponentBuild).
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
