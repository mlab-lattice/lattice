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

// FakeDeploies implements DeployInterface
type FakeDeploies struct {
	Fake *FakeLatticeV1
	ns   string
}

var deploiesResource = schema.GroupVersionResource{Group: "lattice.mlab.com", Version: "v1", Resource: "deploies"}

var deploiesKind = schema.GroupVersionKind{Group: "lattice.mlab.com", Version: "v1", Kind: "Deploy"}

// Get takes name of the deploy, and returns the corresponding deploy object, and an error if there is any.
func (c *FakeDeploies) Get(name string, options v1.GetOptions) (result *lattice_v1.Deploy, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(deploiesResource, c.ns, name), &lattice_v1.Deploy{})

	if obj == nil {
		return nil, err
	}
	return obj.(*lattice_v1.Deploy), err
}

// List takes label and field selectors, and returns the list of Deploies that match those selectors.
func (c *FakeDeploies) List(opts v1.ListOptions) (result *lattice_v1.DeployList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(deploiesResource, deploiesKind, c.ns, opts), &lattice_v1.DeployList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &lattice_v1.DeployList{}
	for _, item := range obj.(*lattice_v1.DeployList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested deploies.
func (c *FakeDeploies) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(deploiesResource, c.ns, opts))

}

// Create takes the representation of a deploy and creates it.  Returns the server's representation of the deploy, and an error, if there is any.
func (c *FakeDeploies) Create(deploy *lattice_v1.Deploy) (result *lattice_v1.Deploy, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(deploiesResource, c.ns, deploy), &lattice_v1.Deploy{})

	if obj == nil {
		return nil, err
	}
	return obj.(*lattice_v1.Deploy), err
}

// Update takes the representation of a deploy and updates it. Returns the server's representation of the deploy, and an error, if there is any.
func (c *FakeDeploies) Update(deploy *lattice_v1.Deploy) (result *lattice_v1.Deploy, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(deploiesResource, c.ns, deploy), &lattice_v1.Deploy{})

	if obj == nil {
		return nil, err
	}
	return obj.(*lattice_v1.Deploy), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeDeploies) UpdateStatus(deploy *lattice_v1.Deploy) (*lattice_v1.Deploy, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(deploiesResource, "status", c.ns, deploy), &lattice_v1.Deploy{})

	if obj == nil {
		return nil, err
	}
	return obj.(*lattice_v1.Deploy), err
}

// Delete takes name of the deploy and deletes it. Returns an error if one occurs.
func (c *FakeDeploies) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(deploiesResource, c.ns, name), &lattice_v1.Deploy{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeDeploies) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(deploiesResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &lattice_v1.DeployList{})
	return err
}

// Patch applies the patch and returns the patched deploy.
func (c *FakeDeploies) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *lattice_v1.Deploy, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(deploiesResource, c.ns, name, data, subresources...), &lattice_v1.Deploy{})

	if obj == nil {
		return nil, err
	}
	return obj.(*lattice_v1.Deploy), err
}
