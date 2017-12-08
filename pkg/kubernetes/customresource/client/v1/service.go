package v1

import (
	"github.com/mlab-lattice/system/pkg/kubernetes/customresource/apis/lattice/v1"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"

	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/rest"
)

// ServicesGetter has a method to return a ServiceInterface.
// A group's client should implement this interface.
type ServicesGetter interface {
	Services(namespace string) ServiceInterface
}

// ServiceInterface has methods to work with Service resources.
type ServiceInterface interface {
	Create(*v1.Service) (*v1.Service, error)
	Update(*v1.Service) (*v1.Service, error)
	UpdateStatus(*v1.Service) (*v1.Service, error)
	Delete(name string, options *meta_v1.DeleteOptions) error
	DeleteCollection(options *meta_v1.DeleteOptions, listOptions meta_v1.ListOptions) error
	Get(name string, options meta_v1.GetOptions) (*v1.Service, error)
	List(opts meta_v1.ListOptions) (*v1.ServiceList, error)
	Watch(opts meta_v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.Service, err error)
}

// Services implements ServiceInterface
type Services struct {
	client rest.Interface
	ns     string
}

// newServices returns a Services
func newServices(c *Client, namespace string) *Services {
	return &Services{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the Service, and returns the corresponding Service object, and an error if there is any.
func (c *Services) Get(name string, options meta_v1.GetOptions) (result *v1.Service, err error) {
	result = &v1.Service{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource(v1.ResourcePluralService).
		Name(name).
		VersionedParams(&options, ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of Services that match those selectors.
func (c *Services) List(opts meta_v1.ListOptions) (result *v1.ServiceList, err error) {
	result = &v1.ServiceList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource(v1.ResourcePluralService).
		VersionedParams(&opts, ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested Services.
func (c *Services) Watch(opts meta_v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource(v1.ResourcePluralService).
		VersionedParams(&opts, ParameterCodec).
		Watch()
}

// Create takes the representation of a Service and creates it.  Returns the server's representation of the Service, and an error, if there is any.
func (c *Services) Create(Service *v1.Service) (result *v1.Service, err error) {
	result = &v1.Service{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource(v1.ResourcePluralService).
		Body(Service).
		Do().
		Into(result)
	return
}

// Update takes the representation of a Service and updates it. Returns the server's representation of the Service, and an error, if there is any.
func (c *Services) Update(Service *v1.Service) (result *v1.Service, err error) {
	result = &v1.Service{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource(v1.ResourcePluralService).
		Name(Service.Name).
		Body(Service).
		Do().
		Into(result)
	return
}

func (c *Services) UpdateStatus(Service *v1.Service) (result *v1.Service, err error) {
	result = &v1.Service{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource(v1.ResourcePluralService).
		Name(Service.Name).
		SubResource("status").
		Body(Service).
		Do().
		Into(result)
	return
}

// Delete takes name of the Service and deletes it. Returns an error if one occurs.
func (c *Services) Delete(name string, options *meta_v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource(v1.ResourcePluralService).
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *Services) DeleteCollection(options *meta_v1.DeleteOptions, listOptions meta_v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource(v1.ResourcePluralService).
		VersionedParams(&listOptions, ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched Service.
func (c *Services) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.Service, err error) {
	result = &v1.Service{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource(v1.ResourcePluralService).
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
