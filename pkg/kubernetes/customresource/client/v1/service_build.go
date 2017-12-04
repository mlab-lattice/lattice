package v1

import (
	"github.com/mlab-lattice/system/pkg/kubernetes/customresource/v1"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"

	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/rest"
)

// ServiceBuildsGetter has a method to return a ServiceBuildInterface.
// A group's client should implement this interface.
type ServiceBuildsGetter interface {
	ServiceBuilds(namespace string) ServiceBuildInterface
}

// ServiceBuildInterface has methods to work with ServiceBuild resources.
type ServiceBuildInterface interface {
	Create(*v1.ServiceBuild) (*v1.ServiceBuild, error)
	Update(*v1.ServiceBuild) (*v1.ServiceBuild, error)
	UpdateStatus(*v1.ServiceBuild) (*v1.ServiceBuild, error)
	Delete(name string, options *meta_v1.DeleteOptions) error
	DeleteCollection(options *meta_v1.DeleteOptions, listOptions meta_v1.ListOptions) error
	Get(name string, options meta_v1.GetOptions) (*v1.ServiceBuild, error)
	List(opts meta_v1.ListOptions) (*v1.ServiceBuildList, error)
	Watch(opts meta_v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.ServiceBuild, err error)
}

// ServiceBuilds implements ServiceBuildInterface
type ServiceBuilds struct {
	client rest.Interface
	ns     string
}

// newServiceBuilds returns a ServiceBuilds
func newServiceBuilds(c *V1Client, namespace string) *ServiceBuilds {
	return &ServiceBuilds{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the ServiceBuild, and returns the corresponding ServiceBuild object, and an error if there is any.
func (c *ServiceBuilds) Get(name string, options meta_v1.GetOptions) (result *v1.ServiceBuild, err error) {
	result = &v1.ServiceBuild{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource(v1.ResourcePluralServiceBuild).
		Name(name).
		VersionedParams(&options, ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of ServiceBuilds that match those selectors.
func (c *ServiceBuilds) List(opts meta_v1.ListOptions) (result *v1.ServiceBuildList, err error) {
	result = &v1.ServiceBuildList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource(v1.ResourcePluralServiceBuild).
		VersionedParams(&opts, ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested ServiceBuilds.
func (c *ServiceBuilds) Watch(opts meta_v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource(v1.ResourcePluralServiceBuild).
		VersionedParams(&opts, ParameterCodec).
		Watch()
}

// Create takes the representation of a ServiceBuild and creates it.  Returns the server's representation of the ServiceBuild, and an error, if there is any.
func (c *ServiceBuilds) Create(ServiceBuild *v1.ServiceBuild) (result *v1.ServiceBuild, err error) {
	result = &v1.ServiceBuild{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource(v1.ResourcePluralServiceBuild).
		Body(ServiceBuild).
		Do().
		Into(result)
	return
}

// Update takes the representation of a ServiceBuild and updates it. Returns the server's representation of the ServiceBuild, and an error, if there is any.
func (c *ServiceBuilds) Update(ServiceBuild *v1.ServiceBuild) (result *v1.ServiceBuild, err error) {
	result = &v1.ServiceBuild{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource(v1.ResourcePluralServiceBuild).
		Name(ServiceBuild.Name).
		Body(ServiceBuild).
		Do().
		Into(result)
	return
}

func (c *ServiceBuilds) UpdateStatus(ServiceBuild *v1.ServiceBuild) (result *v1.ServiceBuild, err error) {
	result = &v1.ServiceBuild{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource(v1.ResourcePluralServiceBuild).
		Name(ServiceBuild.Name).
		SubResource("status").
		Body(ServiceBuild).
		Do().
		Into(result)
	return
}

// Delete takes name of the ServiceBuild and deletes it. Returns an error if one occurs.
func (c *ServiceBuilds) Delete(name string, options *meta_v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource(v1.ResourcePluralServiceBuild).
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *ServiceBuilds) DeleteCollection(options *meta_v1.DeleteOptions, listOptions meta_v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource(v1.ResourcePluralServiceBuild).
		VersionedParams(&listOptions, ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched ServiceBuild.
func (c *ServiceBuilds) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.ServiceBuild, err error) {
	result = &v1.ServiceBuild{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource(v1.ResourcePluralServiceBuild).
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
