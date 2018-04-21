package fake

import (
	lattice_v1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeAddresses implements AddressInterface
type FakeAddresses struct {
	Fake *FakeLatticeV1
	ns   string
}

var addressesResource = schema.GroupVersionResource{Group: "lattice.mlab.com", Version: "v1", Resource: "addresses"}

var addressesKind = schema.GroupVersionKind{Group: "lattice.mlab.com", Version: "v1", Kind: "Address"}

// Get takes name of the address, and returns the corresponding address object, and an error if there is any.
func (c *FakeAddresses) Get(name string, options v1.GetOptions) (result *lattice_v1.Address, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(addressesResource, c.ns, name), &lattice_v1.Address{})

	if obj == nil {
		return nil, err
	}
	return obj.(*lattice_v1.Address), err
}

// List takes label and field selectors, and returns the list of Addresses that match those selectors.
func (c *FakeAddresses) List(opts v1.ListOptions) (result *lattice_v1.AddressList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(addressesResource, addressesKind, c.ns, opts), &lattice_v1.AddressList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &lattice_v1.AddressList{}
	for _, item := range obj.(*lattice_v1.AddressList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested addresses.
func (c *FakeAddresses) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(addressesResource, c.ns, opts))

}

// Create takes the representation of a address and creates it.  Returns the server's representation of the address, and an error, if there is any.
func (c *FakeAddresses) Create(address *lattice_v1.Address) (result *lattice_v1.Address, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(addressesResource, c.ns, address), &lattice_v1.Address{})

	if obj == nil {
		return nil, err
	}
	return obj.(*lattice_v1.Address), err
}

// Update takes the representation of a address and updates it. Returns the server's representation of the address, and an error, if there is any.
func (c *FakeAddresses) Update(address *lattice_v1.Address) (result *lattice_v1.Address, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(addressesResource, c.ns, address), &lattice_v1.Address{})

	if obj == nil {
		return nil, err
	}
	return obj.(*lattice_v1.Address), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeAddresses) UpdateStatus(address *lattice_v1.Address) (*lattice_v1.Address, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(addressesResource, "status", c.ns, address), &lattice_v1.Address{})

	if obj == nil {
		return nil, err
	}
	return obj.(*lattice_v1.Address), err
}

// Delete takes name of the address and deletes it. Returns an error if one occurs.
func (c *FakeAddresses) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(addressesResource, c.ns, name), &lattice_v1.Address{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeAddresses) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(addressesResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &lattice_v1.AddressList{})
	return err
}

// Patch applies the patch and returns the patched address.
func (c *FakeAddresses) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *lattice_v1.Address, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(addressesResource, c.ns, name, data, subresources...), &lattice_v1.Address{})

	if obj == nil {
		return nil, err
	}
	return obj.(*lattice_v1.Address), err
}
