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

// FakeSystemRollouts implements SystemRolloutInterface
type FakeSystemRollouts struct {
	Fake *FakeLatticeV1
	ns   string
}

var systemrolloutsResource = schema.GroupVersionResource{Group: "lattice.mlab.com", Version: "v1", Resource: "systemrollouts"}

var systemrolloutsKind = schema.GroupVersionKind{Group: "lattice.mlab.com", Version: "v1", Kind: "SystemRollout"}

// Get takes name of the systemRollout, and returns the corresponding systemRollout object, and an error if there is any.
func (c *FakeSystemRollouts) Get(name string, options v1.GetOptions) (result *lattice_v1.SystemRollout, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(systemrolloutsResource, c.ns, name), &lattice_v1.SystemRollout{})

	if obj == nil {
		return nil, err
	}
	return obj.(*lattice_v1.SystemRollout), err
}

// List takes label and field selectors, and returns the list of SystemRollouts that match those selectors.
func (c *FakeSystemRollouts) List(opts v1.ListOptions) (result *lattice_v1.SystemRolloutList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(systemrolloutsResource, systemrolloutsKind, c.ns, opts), &lattice_v1.SystemRolloutList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &lattice_v1.SystemRolloutList{}
	for _, item := range obj.(*lattice_v1.SystemRolloutList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested systemRollouts.
func (c *FakeSystemRollouts) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(systemrolloutsResource, c.ns, opts))

}

// Create takes the representation of a systemRollout and creates it.  Returns the server's representation of the systemRollout, and an error, if there is any.
func (c *FakeSystemRollouts) Create(systemRollout *lattice_v1.SystemRollout) (result *lattice_v1.SystemRollout, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(systemrolloutsResource, c.ns, systemRollout), &lattice_v1.SystemRollout{})

	if obj == nil {
		return nil, err
	}
	return obj.(*lattice_v1.SystemRollout), err
}

// Update takes the representation of a systemRollout and updates it. Returns the server's representation of the systemRollout, and an error, if there is any.
func (c *FakeSystemRollouts) Update(systemRollout *lattice_v1.SystemRollout) (result *lattice_v1.SystemRollout, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(systemrolloutsResource, c.ns, systemRollout), &lattice_v1.SystemRollout{})

	if obj == nil {
		return nil, err
	}
	return obj.(*lattice_v1.SystemRollout), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeSystemRollouts) UpdateStatus(systemRollout *lattice_v1.SystemRollout) (*lattice_v1.SystemRollout, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(systemrolloutsResource, "status", c.ns, systemRollout), &lattice_v1.SystemRollout{})

	if obj == nil {
		return nil, err
	}
	return obj.(*lattice_v1.SystemRollout), err
}

// Delete takes name of the systemRollout and deletes it. Returns an error if one occurs.
func (c *FakeSystemRollouts) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(systemrolloutsResource, c.ns, name), &lattice_v1.SystemRollout{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeSystemRollouts) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(systemrolloutsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &lattice_v1.SystemRolloutList{})
	return err
}

// Patch applies the patch and returns the patched systemRollout.
func (c *FakeSystemRollouts) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *lattice_v1.SystemRollout, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(systemrolloutsResource, c.ns, name, data, subresources...), &lattice_v1.SystemRollout{})

	if obj == nil {
		return nil, err
	}
	return obj.(*lattice_v1.SystemRollout), err
}
