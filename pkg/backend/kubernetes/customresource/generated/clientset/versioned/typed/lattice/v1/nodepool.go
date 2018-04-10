package v1

import (
	v1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	scheme "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/generated/clientset/versioned/scheme"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// NodePoolsGetter has a method to return a NodePoolInterface.
// A group's client should implement this interface.
type NodePoolsGetter interface {
	NodePools(namespace string) NodePoolInterface
}

// NodePoolInterface has methods to work with NodePool resources.
type NodePoolInterface interface {
	Create(*v1.NodePool) (*v1.NodePool, error)
	Update(*v1.NodePool) (*v1.NodePool, error)
	UpdateStatus(*v1.NodePool) (*v1.NodePool, error)
	Delete(name string, options *meta_v1.DeleteOptions) error
	DeleteCollection(options *meta_v1.DeleteOptions, listOptions meta_v1.ListOptions) error
	Get(name string, options meta_v1.GetOptions) (*v1.NodePool, error)
	List(opts meta_v1.ListOptions) (*v1.NodePoolList, error)
	Watch(opts meta_v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.NodePool, err error)
	NodePoolExpansion
}

// nodePools implements NodePoolInterface
type nodePools struct {
	client rest.Interface
	ns     string
}

// newNodePools returns a NodePools
func newNodePools(c *LatticeV1Client, namespace string) *nodePools {
	return &nodePools{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the nodePool, and returns the corresponding nodePool object, and an error if there is any.
func (c *nodePools) Get(name string, options meta_v1.GetOptions) (result *v1.NodePool, err error) {
	result = &v1.NodePool{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("nodepools").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of NodePools that match those selectors.
func (c *nodePools) List(opts meta_v1.ListOptions) (result *v1.NodePoolList, err error) {
	result = &v1.NodePoolList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("nodepools").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested nodePools.
func (c *nodePools) Watch(opts meta_v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("nodepools").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a nodePool and creates it.  Returns the server's representation of the nodePool, and an error, if there is any.
func (c *nodePools) Create(nodePool *v1.NodePool) (result *v1.NodePool, err error) {
	result = &v1.NodePool{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("nodepools").
		Body(nodePool).
		Do().
		Into(result)
	return
}

// Update takes the representation of a nodePool and updates it. Returns the server's representation of the nodePool, and an error, if there is any.
func (c *nodePools) Update(nodePool *v1.NodePool) (result *v1.NodePool, err error) {
	result = &v1.NodePool{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("nodepools").
		Name(nodePool.Name).
		Body(nodePool).
		Do().
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().

func (c *nodePools) UpdateStatus(nodePool *v1.NodePool) (result *v1.NodePool, err error) {
	result = &v1.NodePool{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("nodepools").
		Name(nodePool.Name).
		SubResource("status").
		Body(nodePool).
		Do().
		Into(result)
	return
}

// Delete takes name of the nodePool and deletes it. Returns an error if one occurs.
func (c *nodePools) Delete(name string, options *meta_v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("nodepools").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *nodePools) DeleteCollection(options *meta_v1.DeleteOptions, listOptions meta_v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("nodepools").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched nodePool.
func (c *nodePools) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.NodePool, err error) {
	result = &v1.NodePool{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("nodepools").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
