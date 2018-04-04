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

// FakeLoadBalancers implements LoadBalancerInterface
type FakeLoadBalancers struct {
	Fake *FakeLatticeV1
	ns   string
}

var loadbalancersResource = schema.GroupVersionResource{Group: "lattice.mlab.com", Version: "v1", Resource: "loadbalancers"}

var loadbalancersKind = schema.GroupVersionKind{Group: "lattice.mlab.com", Version: "v1", Kind: "LoadBalancer"}

// Get takes name of the loadBalancer, and returns the corresponding loadBalancer object, and an error if there is any.
func (c *FakeLoadBalancers) Get(name string, options v1.GetOptions) (result *lattice_v1.LoadBalancer, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(loadbalancersResource, c.ns, name), &lattice_v1.LoadBalancer{})

	if obj == nil {
		return nil, err
	}
	return obj.(*lattice_v1.LoadBalancer), err
}

// List takes label and field selectors, and returns the list of LoadBalancers that match those selectors.
func (c *FakeLoadBalancers) List(opts v1.ListOptions) (result *lattice_v1.LoadBalancerList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(loadbalancersResource, loadbalancersKind, c.ns, opts), &lattice_v1.LoadBalancerList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &lattice_v1.LoadBalancerList{}
	for _, item := range obj.(*lattice_v1.LoadBalancerList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested loadBalancers.
func (c *FakeLoadBalancers) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(loadbalancersResource, c.ns, opts))

}

// Create takes the representation of a loadBalancer and creates it.  Returns the server's representation of the loadBalancer, and an error, if there is any.
func (c *FakeLoadBalancers) Create(loadBalancer *lattice_v1.LoadBalancer) (result *lattice_v1.LoadBalancer, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(loadbalancersResource, c.ns, loadBalancer), &lattice_v1.LoadBalancer{})

	if obj == nil {
		return nil, err
	}
	return obj.(*lattice_v1.LoadBalancer), err
}

// Update takes the representation of a loadBalancer and updates it. Returns the server's representation of the loadBalancer, and an error, if there is any.
func (c *FakeLoadBalancers) Update(loadBalancer *lattice_v1.LoadBalancer) (result *lattice_v1.LoadBalancer, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(loadbalancersResource, c.ns, loadBalancer), &lattice_v1.LoadBalancer{})

	if obj == nil {
		return nil, err
	}
	return obj.(*lattice_v1.LoadBalancer), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeLoadBalancers) UpdateStatus(loadBalancer *lattice_v1.LoadBalancer) (*lattice_v1.LoadBalancer, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(loadbalancersResource, "status", c.ns, loadBalancer), &lattice_v1.LoadBalancer{})

	if obj == nil {
		return nil, err
	}
	return obj.(*lattice_v1.LoadBalancer), err
}

// Delete takes name of the loadBalancer and deletes it. Returns an error if one occurs.
func (c *FakeLoadBalancers) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(loadbalancersResource, c.ns, name), &lattice_v1.LoadBalancer{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeLoadBalancers) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(loadbalancersResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &lattice_v1.LoadBalancerList{})
	return err
}

// Patch applies the patch and returns the patched loadBalancer.
func (c *FakeLoadBalancers) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *lattice_v1.LoadBalancer, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(loadbalancersResource, c.ns, name, data, subresources...), &lattice_v1.LoadBalancer{})

	if obj == nil {
		return nil, err
	}
	return obj.(*lattice_v1.LoadBalancer), err
}
