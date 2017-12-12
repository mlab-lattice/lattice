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

// FakeEndpoints implements EndpointInterface
type FakeEndpoints struct {
	Fake *FakeLatticeV1
	ns   string
}

var endpointsResource = schema.GroupVersionResource{Group: "lattice.mlab.com", Version: "v1", Resource: "endpoints"}

var endpointsKind = schema.GroupVersionKind{Group: "lattice.mlab.com", Version: "v1", Kind: "Endpoint"}

// Get takes name of the endpoint, and returns the corresponding endpoint object, and an error if there is any.
func (c *FakeEndpoints) Get(name string, options v1.GetOptions) (result *lattice_v1.Endpoint, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(endpointsResource, c.ns, name), &lattice_v1.Endpoint{})

	if obj == nil {
		return nil, err
	}
	return obj.(*lattice_v1.Endpoint), err
}

// List takes label and field selectors, and returns the list of Endpoints that match those selectors.
func (c *FakeEndpoints) List(opts v1.ListOptions) (result *lattice_v1.EndpointList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(endpointsResource, endpointsKind, c.ns, opts), &lattice_v1.EndpointList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &lattice_v1.EndpointList{}
	for _, item := range obj.(*lattice_v1.EndpointList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested endpoints.
func (c *FakeEndpoints) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(endpointsResource, c.ns, opts))

}

// Create takes the representation of a endpoint and creates it.  Returns the server's representation of the endpoint, and an error, if there is any.
func (c *FakeEndpoints) Create(endpoint *lattice_v1.Endpoint) (result *lattice_v1.Endpoint, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(endpointsResource, c.ns, endpoint), &lattice_v1.Endpoint{})

	if obj == nil {
		return nil, err
	}
	return obj.(*lattice_v1.Endpoint), err
}

// Update takes the representation of a endpoint and updates it. Returns the server's representation of the endpoint, and an error, if there is any.
func (c *FakeEndpoints) Update(endpoint *lattice_v1.Endpoint) (result *lattice_v1.Endpoint, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(endpointsResource, c.ns, endpoint), &lattice_v1.Endpoint{})

	if obj == nil {
		return nil, err
	}
	return obj.(*lattice_v1.Endpoint), err
}

// Delete takes name of the endpoint and deletes it. Returns an error if one occurs.
func (c *FakeEndpoints) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(endpointsResource, c.ns, name), &lattice_v1.Endpoint{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeEndpoints) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(endpointsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &lattice_v1.EndpointList{})
	return err
}

// Patch applies the patch and returns the patched endpoint.
func (c *FakeEndpoints) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *lattice_v1.Endpoint, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(endpointsResource, c.ns, name, data, subresources...), &lattice_v1.Endpoint{})

	if obj == nil {
		return nil, err
	}
	return obj.(*lattice_v1.Endpoint), err
}
