package v1

import (
	"github.com/mlab-lattice/system/pkg/kubernetes/customresource/apis/lattice/v1"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"

	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/rest"
)

// ConfigsGetter has a method to return a ConfigInterface.
// A group's client should implement this interface.
type ConfigsGetter interface {
	Configs(namespace string) ConfigInterface
}

// ConfigInterface has methods to work with Config resources.
type ConfigInterface interface {
	Create(*v1.Config) (*v1.Config, error)
	Update(*v1.Config) (*v1.Config, error)
	UpdateStatus(*v1.Config) (*v1.Config, error)
	Delete(name string, options *meta_v1.DeleteOptions) error
	DeleteCollection(options *meta_v1.DeleteOptions, listOptions meta_v1.ListOptions) error
	Get(name string, options meta_v1.GetOptions) (*v1.Config, error)
	List(opts meta_v1.ListOptions) (*v1.ConfigList, error)
	Watch(opts meta_v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.Config, err error)
}

// Configs implements ConfigInterface
type Configs struct {
	client rest.Interface
	ns     string
}

// newConfigs returns a Configs
func newConfigs(c *Client, namespace string) *Configs {
	return &Configs{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the Config, and returns the corresponding Config object, and an error if there is any.
func (c *Configs) Get(name string, options meta_v1.GetOptions) (result *v1.Config, err error) {
	result = &v1.Config{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource(v1.ResourcePluralConfig).
		Name(name).
		VersionedParams(&options, ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of Configs that match those selectors.
func (c *Configs) List(opts meta_v1.ListOptions) (result *v1.ConfigList, err error) {
	result = &v1.ConfigList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource(v1.ResourcePluralConfig).
		VersionedParams(&opts, ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested Configs.
func (c *Configs) Watch(opts meta_v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource(v1.ResourcePluralConfig).
		VersionedParams(&opts, ParameterCodec).
		Watch()
}

// Create takes the representation of a Config and creates it.  Returns the server's representation of the Config, and an error, if there is any.
func (c *Configs) Create(Config *v1.Config) (result *v1.Config, err error) {
	result = &v1.Config{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource(v1.ResourcePluralConfig).
		Body(Config).
		Do().
		Into(result)
	return
}

// Update takes the representation of a Config and updates it. Returns the server's representation of the Config, and an error, if there is any.
func (c *Configs) Update(Config *v1.Config) (result *v1.Config, err error) {
	result = &v1.Config{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource(v1.ResourcePluralConfig).
		Name(Config.Name).
		Body(Config).
		Do().
		Into(result)
	return
}

func (c *Configs) UpdateStatus(Config *v1.Config) (result *v1.Config, err error) {
	result = &v1.Config{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource(v1.ResourcePluralConfig).
		Name(Config.Name).
		SubResource("status").
		Body(Config).
		Do().
		Into(result)
	return
}

// Delete takes name of the Config and deletes it. Returns an error if one occurs.
func (c *Configs) Delete(name string, options *meta_v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource(v1.ResourcePluralConfig).
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *Configs) DeleteCollection(options *meta_v1.DeleteOptions, listOptions meta_v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource(v1.ResourcePluralConfig).
		VersionedParams(&listOptions, ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched Config.
func (c *Configs) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.Config, err error) {
	result = &v1.Config{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource(v1.ResourcePluralConfig).
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
