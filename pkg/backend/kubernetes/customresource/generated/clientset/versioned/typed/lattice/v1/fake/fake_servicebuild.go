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

// FakeServiceBuilds implements ServiceBuildInterface
type FakeServiceBuilds struct {
	Fake *FakeLatticeV1
	ns   string
}

var servicebuildsResource = schema.GroupVersionResource{Group: "lattice.mlab.com", Version: "v1", Resource: "servicebuilds"}

var servicebuildsKind = schema.GroupVersionKind{Group: "lattice.mlab.com", Version: "v1", Kind: "ServiceBuild"}

// Get takes name of the serviceBuild, and returns the corresponding serviceBuild object, and an error if there is any.
func (c *FakeServiceBuilds) Get(name string, options v1.GetOptions) (result *lattice_v1.ServiceBuild, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(servicebuildsResource, c.ns, name), &lattice_v1.ServiceBuild{})

	if obj == nil {
		return nil, err
	}
	return obj.(*lattice_v1.ServiceBuild), err
}

// List takes label and field selectors, and returns the list of ServiceBuilds that match those selectors.
func (c *FakeServiceBuilds) List(opts v1.ListOptions) (result *lattice_v1.ServiceBuildList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(servicebuildsResource, servicebuildsKind, c.ns, opts), &lattice_v1.ServiceBuildList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &lattice_v1.ServiceBuildList{}
	for _, item := range obj.(*lattice_v1.ServiceBuildList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested serviceBuilds.
func (c *FakeServiceBuilds) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(servicebuildsResource, c.ns, opts))

}

// Create takes the representation of a serviceBuild and creates it.  Returns the server's representation of the serviceBuild, and an error, if there is any.
func (c *FakeServiceBuilds) Create(serviceBuild *lattice_v1.ServiceBuild) (result *lattice_v1.ServiceBuild, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(servicebuildsResource, c.ns, serviceBuild), &lattice_v1.ServiceBuild{})

	if obj == nil {
		return nil, err
	}
	return obj.(*lattice_v1.ServiceBuild), err
}

// Update takes the representation of a serviceBuild and updates it. Returns the server's representation of the serviceBuild, and an error, if there is any.
func (c *FakeServiceBuilds) Update(serviceBuild *lattice_v1.ServiceBuild) (result *lattice_v1.ServiceBuild, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(servicebuildsResource, c.ns, serviceBuild), &lattice_v1.ServiceBuild{})

	if obj == nil {
		return nil, err
	}
	return obj.(*lattice_v1.ServiceBuild), err
}

// Delete takes name of the serviceBuild and deletes it. Returns an error if one occurs.
func (c *FakeServiceBuilds) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(servicebuildsResource, c.ns, name), &lattice_v1.ServiceBuild{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeServiceBuilds) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(servicebuildsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &lattice_v1.ServiceBuildList{})
	return err
}

// Patch applies the patch and returns the patched serviceBuild.
func (c *FakeServiceBuilds) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *lattice_v1.ServiceBuild, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(servicebuildsResource, c.ns, name, data, subresources...), &lattice_v1.ServiceBuild{})

	if obj == nil {
		return nil, err
	}
	return obj.(*lattice_v1.ServiceBuild), err
}
