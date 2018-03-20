package v1

import (
	v1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	scheme "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/generated/clientset/versioned/scheme"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// DeploiesGetter has a method to return a DeployInterface.
// A group's client should implement this interface.
type DeploiesGetter interface {
	Deploies(namespace string) DeployInterface
}

// DeployInterface has methods to work with Deploy resources.
type DeployInterface interface {
	Create(*v1.Deploy) (*v1.Deploy, error)
	Update(*v1.Deploy) (*v1.Deploy, error)
	UpdateStatus(*v1.Deploy) (*v1.Deploy, error)
	Delete(name string, options *meta_v1.DeleteOptions) error
	DeleteCollection(options *meta_v1.DeleteOptions, listOptions meta_v1.ListOptions) error
	Get(name string, options meta_v1.GetOptions) (*v1.Deploy, error)
	List(opts meta_v1.ListOptions) (*v1.DeployList, error)
	Watch(opts meta_v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.Deploy, err error)
	DeployExpansion
}

// deploies implements DeployInterface
type deploies struct {
	client rest.Interface
	ns     string
}

// newDeploies returns a Deploies
func newDeploies(c *LatticeV1Client, namespace string) *deploies {
	return &deploies{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the deploy, and returns the corresponding deploy object, and an error if there is any.
func (c *deploies) Get(name string, options meta_v1.GetOptions) (result *v1.Deploy, err error) {
	result = &v1.Deploy{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("deploies").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of Deploies that match those selectors.
func (c *deploies) List(opts meta_v1.ListOptions) (result *v1.DeployList, err error) {
	result = &v1.DeployList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("deploies").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested deploies.
func (c *deploies) Watch(opts meta_v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("deploies").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a deploy and creates it.  Returns the server's representation of the deploy, and an error, if there is any.
func (c *deploies) Create(deploy *v1.Deploy) (result *v1.Deploy, err error) {
	result = &v1.Deploy{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("deploies").
		Body(deploy).
		Do().
		Into(result)
	return
}

// Update takes the representation of a deploy and updates it. Returns the server's representation of the deploy, and an error, if there is any.
func (c *deploies) Update(deploy *v1.Deploy) (result *v1.Deploy, err error) {
	result = &v1.Deploy{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("deploies").
		Name(deploy.Name).
		Body(deploy).
		Do().
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().

func (c *deploies) UpdateStatus(deploy *v1.Deploy) (result *v1.Deploy, err error) {
	result = &v1.Deploy{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("deploies").
		Name(deploy.Name).
		SubResource("status").
		Body(deploy).
		Do().
		Into(result)
	return
}

// Delete takes name of the deploy and deletes it. Returns an error if one occurs.
func (c *deploies) Delete(name string, options *meta_v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("deploies").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *deploies) DeleteCollection(options *meta_v1.DeleteOptions, listOptions meta_v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("deploies").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched deploy.
func (c *deploies) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.Deploy, err error) {
	result = &v1.Deploy{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("deploies").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
