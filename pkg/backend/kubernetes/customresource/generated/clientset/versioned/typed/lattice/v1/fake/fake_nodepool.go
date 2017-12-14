package fake

import (
	lattice_v1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeNodePools implements NodePoolInterface
type FakeNodePools struct {
	Fake *FakeLatticeV1
	ns   string
}

var nodepoolsResource = schema.GroupVersionResource{Group: "lattice.mlab.com", Version: "v1", Resource: "nodepools"}

var nodepoolsKind = schema.GroupVersionKind{Group: "lattice.mlab.com", Version: "v1", Kind: "NodePool"}

// Get takes name of the nodePool, and returns the corresponding nodePool object, and an error if there is any.
func (c *FakeNodePools) Get(name string, options v1.GetOptions) (result *lattice_v1.NodePool, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(nodepoolsResource, c.ns, name), &lattice_v1.NodePool{})

	if obj == nil {
		return nil, err
	}
	return obj.(*lattice_v1.NodePool), err
}

// List takes label and field selectors, and returns the list of NodePools that match those selectors.
func (c *FakeNodePools) List(opts v1.ListOptions) (result *lattice_v1.NodePoolList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(nodepoolsResource, nodepoolsKind, c.ns, opts), &lattice_v1.NodePoolList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &lattice_v1.NodePoolList{}
	for _, item := range obj.(*lattice_v1.NodePoolList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested nodePools.
func (c *FakeNodePools) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(nodepoolsResource, c.ns, opts))

}

// Create takes the representation of a nodePool and creates it.  Returns the server's representation of the nodePool, and an error, if there is any.
func (c *FakeNodePools) Create(nodePool *lattice_v1.NodePool) (result *lattice_v1.NodePool, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(nodepoolsResource, c.ns, nodePool), &lattice_v1.NodePool{})

	if obj == nil {
		return nil, err
	}
	return obj.(*lattice_v1.NodePool), err
}

// Update takes the representation of a nodePool and updates it. Returns the server's representation of the nodePool, and an error, if there is any.
func (c *FakeNodePools) Update(nodePool *lattice_v1.NodePool) (result *lattice_v1.NodePool, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(nodepoolsResource, c.ns, nodePool), &lattice_v1.NodePool{})

	if obj == nil {
		return nil, err
	}
	return obj.(*lattice_v1.NodePool), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeNodePools) UpdateStatus(nodePool *lattice_v1.NodePool) (*lattice_v1.NodePool, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(nodepoolsResource, "status", c.ns, nodePool), &lattice_v1.NodePool{})

	if obj == nil {
		return nil, err
	}
	return obj.(*lattice_v1.NodePool), err
}

// Delete takes name of the nodePool and deletes it. Returns an error if one occurs.
func (c *FakeNodePools) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(nodepoolsResource, c.ns, name), &lattice_v1.NodePool{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeNodePools) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(nodepoolsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &lattice_v1.NodePoolList{})
	return err
}

// Patch applies the patch and returns the patched nodePool.
func (c *FakeNodePools) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *lattice_v1.NodePool, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(nodepoolsResource, c.ns, name, data, subresources...), &lattice_v1.NodePool{})

	if obj == nil {
		return nil, err
	}
	return obj.(*lattice_v1.NodePool), err
}
