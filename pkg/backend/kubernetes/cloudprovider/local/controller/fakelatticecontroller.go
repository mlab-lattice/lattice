package dnscontroller

import(
    latticev1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/generated/clientset/versioned/typed/lattice/v1"
    discovery "k8s.io/client-go/discovery"

    "k8s.io/apimachinery/pkg/runtime"
    "k8s.io/client-go/kubernetes/fake"
    rest "k8s.io/client-go/rest"
)

type fakeLatticeClient struct {
    fakeClient  fake.Clientset
    latticeV1   *latticev1.LatticeV1Client
}

// LatticeV1 retrieves the LatticeV1Client
func (c *fakeLatticeClient) LatticeV1() latticev1.LatticeV1Interface {
    return c.LatticeV1()
}

func (c *fakeLatticeClient) Lattice() latticev1.LatticeV1Interface {
    return c.latticeV1
}

// Discovery retrieves a fake discovery
func (c *fakeLatticeClient) Discovery() discovery.DiscoveryInterface {
    if c == nil {
        return nil
    }
    return c.fakeClient.Discovery()
}

func newLatticeFakeClient(c *rest.Config, fakeClientObj []runtime.Object) (*fakeLatticeClient, error) {
    cs, err := latticev1.NewForConfig(c)
    if err != nil {
        return nil, err
    }

    fakeClient := fake.NewSimpleClientset(fakeClientObj...)
    if err != nil {
        return nil, err
    }

    return &fakeLatticeClient{
        fakeClient: *fakeClient,
        latticeV1: cs,
    }, err
}

