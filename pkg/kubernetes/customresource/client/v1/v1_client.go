package v1

import (
	"github.com/mlab-lattice/system/pkg/kubernetes/customresource/apis/lattice/v1"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"

	"k8s.io/client-go/rest"
)

type Interface interface {
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

// Client is used to interact with features provided by the V1 group.
type Client struct {
	restClient rest.Interface
}

func (c *Client) ComponentBuilds(namespace string) ComponentBuildInterface {
	return newComponentBuilds(c, namespace)
}

func (c *Client) Configs(namespace string) ConfigInterface {
	return newConfigs(c, namespace)
}

func (c *Client) Services(namespace string) ServiceInterface {
	return newServices(c, namespace)
}

func (c *Client) ServiceBuilds(namespace string) ServiceBuildInterface {
	return newServiceBuilds(c, namespace)
}

func (c *Client) Systems(namespace string) SystemInterface {
	return newSystems(c, namespace)
}

func (c *Client) SystemBuilds(namespace string) SystemBuildInterface {
	return newSystemBuilds(c, namespace)
}

func (c *Client) SystemRollouts(namespace string) SystemRolloutInterface {
	return newSystemRollouts(c, namespace)
}

func (c *Client) SystemTeardowns(namespace string) SystemTeardownInterface {
	return newSystemTeardowns(c, namespace)
}

// NewForConfig creates a new Client for the given config.
func NewForConfig(c *rest.Config) (*Client, error) {
	config := *c
	if err := setConfigDefaults(&config); err != nil {
		return nil, err
	}
	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}
	return &Client{client}, nil
}

// NewForConfigOrDie creates a new Client for the given config and
// panics if there is an error in the config.
func NewForConfigOrDie(c *rest.Config) *Client {
	client, err := NewForConfig(c)
	if err != nil {
		panic(err)
	}
	return client
}

// New creates a new Client for the given RESTClient.
func New(c rest.Interface) *Client {
	return &Client{c}
}

func setConfigDefaults(config *rest.Config) error {
	gv := v1.SchemeGroupVersion
	config.GroupVersion = &gv
	config.APIPath = "/apis"
	config.ContentType = runtime.ContentTypeJSON
	config.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: serializer.NewCodecFactory(Scheme)}

	if config.UserAgent == "" {
		config.UserAgent = rest.DefaultKubernetesUserAgent()
	}

	return nil
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *Client) RESTClient() rest.Interface {
	if c == nil {
		return nil
	}
	return c.restClient
}

var Scheme = runtime.NewScheme()
var Codecs = serializer.NewCodecFactory(Scheme)
var ParameterCodec = runtime.NewParameterCodec(Scheme)

func init() {
	AddToScheme(Scheme)
}

func AddToScheme(scheme *runtime.Scheme) {
	v1.AddToScheme(scheme)
}
