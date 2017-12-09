package v1

import (
	v1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/generated/clientset/versioned/scheme"
	serializer "k8s.io/apimachinery/pkg/runtime/serializer"
	rest "k8s.io/client-go/rest"
)

type LatticeV1Interface interface {
	RESTClient() rest.Interface
	ComponentBuildsGetter
	ConfigsGetter
	ServicesGetter
	ServiceBuildsGetter
	SystemsGetter
	SystemBuildsGetter
	SystemRolloutsGetter
	SystemTeardownsGetter
}

// LatticeV1Client is used to interact with features provided by the lattice.mlab.com group.
type LatticeV1Client struct {
	restClient rest.Interface
}

func (c *LatticeV1Client) ComponentBuilds(namespace string) ComponentBuildInterface {
	return newComponentBuilds(c, namespace)
}

func (c *LatticeV1Client) Configs(namespace string) ConfigInterface {
	return newConfigs(c, namespace)
}

func (c *LatticeV1Client) Services(namespace string) ServiceInterface {
	return newServices(c, namespace)
}

func (c *LatticeV1Client) ServiceBuilds(namespace string) ServiceBuildInterface {
	return newServiceBuilds(c, namespace)
}

func (c *LatticeV1Client) Systems(namespace string) SystemInterface {
	return newSystems(c, namespace)
}

func (c *LatticeV1Client) SystemBuilds(namespace string) SystemBuildInterface {
	return newSystemBuilds(c, namespace)
}

func (c *LatticeV1Client) SystemRollouts(namespace string) SystemRolloutInterface {
	return newSystemRollouts(c, namespace)
}

func (c *LatticeV1Client) SystemTeardowns(namespace string) SystemTeardownInterface {
	return newSystemTeardowns(c, namespace)
}

// NewForConfig creates a new LatticeV1Client for the given config.
func NewForConfig(c *rest.Config) (*LatticeV1Client, error) {
	config := *c
	if err := setConfigDefaults(&config); err != nil {
		return nil, err
	}
	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}
	return &LatticeV1Client{client}, nil
}

// NewForConfigOrDie creates a new LatticeV1Client for the given config and
// panics if there is an error in the config.
func NewForConfigOrDie(c *rest.Config) *LatticeV1Client {
	client, err := NewForConfig(c)
	if err != nil {
		panic(err)
	}
	return client
}

// New creates a new LatticeV1Client for the given RESTClient.
func New(c rest.Interface) *LatticeV1Client {
	return &LatticeV1Client{c}
}

func setConfigDefaults(config *rest.Config) error {
	gv := v1.SchemeGroupVersion
	config.GroupVersion = &gv
	config.APIPath = "/apis"
	config.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: scheme.Codecs}

	if config.UserAgent == "" {
		config.UserAgent = rest.DefaultKubernetesUserAgent()
	}

	return nil
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *LatticeV1Client) RESTClient() rest.Interface {
	if c == nil {
		return nil
	}
	return c.restClient
}
