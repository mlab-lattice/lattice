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

// FakeServiceAddresses implements ServiceAddressInterface
type FakeServiceAddresses struct {
	Fake *FakeLatticeV1
	ns   string
}

var serviceaddressesResource = schema.GroupVersionResource{Group: "lattice.mlab.com", Version: "v1", Resource: "serviceaddresses"}

var serviceaddressesKind = schema.GroupVersionKind{Group: "lattice.mlab.com", Version: "v1", Kind: "ServiceAddress"}

// Get takes name of the serviceAddress, and returns the corresponding serviceAddress object, and an error if there is any.
func (c *FakeServiceAddresses) Get(name string, options v1.GetOptions) (result *lattice_v1.ServiceAddress, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(serviceaddressesResource, c.ns, name), &lattice_v1.ServiceAddress{})

	if obj == nil {
		return nil, err
	}
	return obj.(*lattice_v1.ServiceAddress), err
}

// List takes label and field selectors, and returns the list of ServiceAddresses that match those selectors.
func (c *FakeServiceAddresses) List(opts v1.ListOptions) (result *lattice_v1.ServiceAddressList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(serviceaddressesResource, serviceaddressesKind, c.ns, opts), &lattice_v1.ServiceAddressList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &lattice_v1.ServiceAddressList{}
	for _, item := range obj.(*lattice_v1.ServiceAddressList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested serviceAddresses.
func (c *FakeServiceAddresses) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(serviceaddressesResource, c.ns, opts))

}

// Create takes the representation of a serviceAddress and creates it.  Returns the server's representation of the serviceAddress, and an error, if there is any.
func (c *FakeServiceAddresses) Create(serviceAddress *lattice_v1.ServiceAddress) (result *lattice_v1.ServiceAddress, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(serviceaddressesResource, c.ns, serviceAddress), &lattice_v1.ServiceAddress{})

	if obj == nil {
		return nil, err
	}
	return obj.(*lattice_v1.ServiceAddress), err
}

// Update takes the representation of a serviceAddress and updates it. Returns the server's representation of the serviceAddress, and an error, if there is any.
func (c *FakeServiceAddresses) Update(serviceAddress *lattice_v1.ServiceAddress) (result *lattice_v1.ServiceAddress, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(serviceaddressesResource, c.ns, serviceAddress), &lattice_v1.ServiceAddress{})

	if obj == nil {
		return nil, err
	}
	return obj.(*lattice_v1.ServiceAddress), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeServiceAddresses) UpdateStatus(serviceAddress *lattice_v1.ServiceAddress) (*lattice_v1.ServiceAddress, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(serviceaddressesResource, "status", c.ns, serviceAddress), &lattice_v1.ServiceAddress{})

	if obj == nil {
		return nil, err
	}
	return obj.(*lattice_v1.ServiceAddress), err
}

// Delete takes name of the serviceAddress and deletes it. Returns an error if one occurs.
func (c *FakeServiceAddresses) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(serviceaddressesResource, c.ns, name), &lattice_v1.ServiceAddress{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeServiceAddresses) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(serviceaddressesResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &lattice_v1.ServiceAddressList{})
	return err
}

// Patch applies the patch and returns the patched serviceAddress.
func (c *FakeServiceAddresses) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *lattice_v1.ServiceAddress, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(serviceaddressesResource, c.ns, name, data, subresources...), &lattice_v1.ServiceAddress{})

	if obj == nil {
		return nil, err
	}
	return obj.(*lattice_v1.ServiceAddress), err
}
