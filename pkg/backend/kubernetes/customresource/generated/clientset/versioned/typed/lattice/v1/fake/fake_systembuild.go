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

// FakeSystemBuilds implements SystemBuildInterface
type FakeSystemBuilds struct {
	Fake *FakeLatticeV1
	ns   string
}

var systembuildsResource = schema.GroupVersionResource{Group: "lattice.mlab.com", Version: "v1", Resource: "systembuilds"}

var systembuildsKind = schema.GroupVersionKind{Group: "lattice.mlab.com", Version: "v1", Kind: "SystemBuild"}

// Get takes name of the systemBuild, and returns the corresponding systemBuild object, and an error if there is any.
func (c *FakeSystemBuilds) Get(name string, options v1.GetOptions) (result *lattice_v1.SystemBuild, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(systembuildsResource, c.ns, name), &lattice_v1.SystemBuild{})

	if obj == nil {
		return nil, err
	}
	return obj.(*lattice_v1.SystemBuild), err
}

// List takes label and field selectors, and returns the list of SystemBuilds that match those selectors.
func (c *FakeSystemBuilds) List(opts v1.ListOptions) (result *lattice_v1.SystemBuildList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(systembuildsResource, systembuildsKind, c.ns, opts), &lattice_v1.SystemBuildList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &lattice_v1.SystemBuildList{}
	for _, item := range obj.(*lattice_v1.SystemBuildList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested systemBuilds.
func (c *FakeSystemBuilds) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(systembuildsResource, c.ns, opts))

}

// Create takes the representation of a systemBuild and creates it.  Returns the server's representation of the systemBuild, and an error, if there is any.
func (c *FakeSystemBuilds) Create(systemBuild *lattice_v1.SystemBuild) (result *lattice_v1.SystemBuild, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(systembuildsResource, c.ns, systemBuild), &lattice_v1.SystemBuild{})

	if obj == nil {
		return nil, err
	}
	return obj.(*lattice_v1.SystemBuild), err
}

// Update takes the representation of a systemBuild and updates it. Returns the server's representation of the systemBuild, and an error, if there is any.
func (c *FakeSystemBuilds) Update(systemBuild *lattice_v1.SystemBuild) (result *lattice_v1.SystemBuild, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(systembuildsResource, c.ns, systemBuild), &lattice_v1.SystemBuild{})

	if obj == nil {
		return nil, err
	}
	return obj.(*lattice_v1.SystemBuild), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeSystemBuilds) UpdateStatus(systemBuild *lattice_v1.SystemBuild) (*lattice_v1.SystemBuild, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(systembuildsResource, "status", c.ns, systemBuild), &lattice_v1.SystemBuild{})

	if obj == nil {
		return nil, err
	}
	return obj.(*lattice_v1.SystemBuild), err
}

// Delete takes name of the systemBuild and deletes it. Returns an error if one occurs.
func (c *FakeSystemBuilds) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(systembuildsResource, c.ns, name), &lattice_v1.SystemBuild{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeSystemBuilds) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(systembuildsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &lattice_v1.SystemBuildList{})
	return err
}

// Patch applies the patch and returns the patched systemBuild.
func (c *FakeSystemBuilds) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *lattice_v1.SystemBuild, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(systembuildsResource, c.ns, name, data, subresources...), &lattice_v1.SystemBuild{})

	if obj == nil {
		return nil, err
	}
	return obj.(*lattice_v1.SystemBuild), err
}
