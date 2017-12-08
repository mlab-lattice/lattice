package fake

import (
	lattice_v1 "github.com/mlab-lattice/system/pkg/kubernetes/customresource/apis/lattice/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeSystemTeardowns implements SystemTeardownInterface
type FakeSystemTeardowns struct {
	Fake *FakeLatticeV1
	ns   string
}

var systemteardownsResource = schema.GroupVersionResource{Group: "lattice.mlab.com", Version: "v1", Resource: "systemteardowns"}

var systemteardownsKind = schema.GroupVersionKind{Group: "lattice.mlab.com", Version: "v1", Kind: "SystemTeardown"}

// Get takes name of the systemTeardown, and returns the corresponding systemTeardown object, and an error if there is any.
func (c *FakeSystemTeardowns) Get(name string, options v1.GetOptions) (result *lattice_v1.SystemTeardown, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(systemteardownsResource, c.ns, name), &lattice_v1.SystemTeardown{})

	if obj == nil {
		return nil, err
	}
	return obj.(*lattice_v1.SystemTeardown), err
}

// List takes label and field selectors, and returns the list of SystemTeardowns that match those selectors.
func (c *FakeSystemTeardowns) List(opts v1.ListOptions) (result *lattice_v1.SystemTeardownList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(systemteardownsResource, systemteardownsKind, c.ns, opts), &lattice_v1.SystemTeardownList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &lattice_v1.SystemTeardownList{}
	for _, item := range obj.(*lattice_v1.SystemTeardownList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested systemTeardowns.
func (c *FakeSystemTeardowns) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(systemteardownsResource, c.ns, opts))

}

// Create takes the representation of a systemTeardown and creates it.  Returns the server's representation of the systemTeardown, and an error, if there is any.
func (c *FakeSystemTeardowns) Create(systemTeardown *lattice_v1.SystemTeardown) (result *lattice_v1.SystemTeardown, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(systemteardownsResource, c.ns, systemTeardown), &lattice_v1.SystemTeardown{})

	if obj == nil {
		return nil, err
	}
	return obj.(*lattice_v1.SystemTeardown), err
}

// Update takes the representation of a systemTeardown and updates it. Returns the server's representation of the systemTeardown, and an error, if there is any.
func (c *FakeSystemTeardowns) Update(systemTeardown *lattice_v1.SystemTeardown) (result *lattice_v1.SystemTeardown, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(systemteardownsResource, c.ns, systemTeardown), &lattice_v1.SystemTeardown{})

	if obj == nil {
		return nil, err
	}
	return obj.(*lattice_v1.SystemTeardown), err
}

// Delete takes name of the systemTeardown and deletes it. Returns an error if one occurs.
func (c *FakeSystemTeardowns) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(systemteardownsResource, c.ns, name), &lattice_v1.SystemTeardown{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeSystemTeardowns) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(systemteardownsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &lattice_v1.SystemTeardownList{})
	return err
}

// Patch applies the patch and returns the patched systemTeardown.
func (c *FakeSystemTeardowns) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *lattice_v1.SystemTeardown, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(systemteardownsResource, c.ns, name, data, subresources...), &lattice_v1.SystemTeardown{})

	if obj == nil {
		return nil, err
	}
	return obj.(*lattice_v1.SystemTeardown), err
}
