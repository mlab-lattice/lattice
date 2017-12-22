package fake

import (
	v1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/generated/clientset/versioned/typed/lattice/v1"
	rest "k8s.io/client-go/rest"
	testing "k8s.io/client-go/testing"
)

type FakeLatticeV1 struct {
	*testing.Fake
}

func (c *FakeLatticeV1) ComponentBuilds(namespace string) v1.ComponentBuildInterface {
	return &FakeComponentBuilds{c, namespace}
}

func (c *FakeLatticeV1) Configs(namespace string) v1.ConfigInterface {
	return &FakeConfigs{c, namespace}
}

func (c *FakeLatticeV1) Endpoints(namespace string) v1.EndpointInterface {
	return &FakeEndpoints{c, namespace}
}

func (c *FakeLatticeV1) NodePools(namespace string) v1.NodePoolInterface {
	return &FakeNodePools{c, namespace}
}

func (c *FakeLatticeV1) Services(namespace string) v1.ServiceInterface {
	return &FakeServices{c, namespace}
}

func (c *FakeLatticeV1) ServiceAddresses(namespace string) v1.ServiceAddressInterface {
	return &FakeServiceAddresses{c, namespace}
}

func (c *FakeLatticeV1) ServiceBuilds(namespace string) v1.ServiceBuildInterface {
	return &FakeServiceBuilds{c, namespace}
}

func (c *FakeLatticeV1) Systems(namespace string) v1.SystemInterface {
	return &FakeSystems{c, namespace}
}

func (c *FakeLatticeV1) SystemBuilds(namespace string) v1.SystemBuildInterface {
	return &FakeSystemBuilds{c, namespace}
}

func (c *FakeLatticeV1) SystemRollouts(namespace string) v1.SystemRolloutInterface {
	return &FakeSystemRollouts{c, namespace}
}

func (c *FakeLatticeV1) SystemTeardowns(namespace string) v1.SystemTeardownInterface {
	return &FakeSystemTeardowns{c, namespace}
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *FakeLatticeV1) RESTClient() rest.Interface {
	var ret *rest.RESTClient
	return ret
}
