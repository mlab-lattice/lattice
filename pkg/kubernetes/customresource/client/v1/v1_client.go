package v1

import (
	"github.com/mlab-lattice/system/pkg/kubernetes/customresource/v1"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"

	"k8s.io/client-go/rest"
)

type V1Interface interface {
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

// V1Client is used to interact with features provided by the V1 group.
type V1Client struct {
	restClient rest.Interface
}

func (c *V1Client) ComponentBuilds(namespace string) ComponentBuildInterface {
	return newComponentBuilds(c, namespace)
}

func (c *V1Client) Configs(namespace string) ConfigInterface {
	return newConfigs(c, namespace)
}

func (c *V1Client) Services(namespace string) ServiceInterface {
	return newServices(c, namespace)
}

func (c *V1Client) ServiceBuilds(namespace string) ServiceBuildInterface {
	return newServiceBuilds(c, namespace)
}

func (c *V1Client) Systems(namespace string) SystemInterface {
	return newSystems(c, namespace)
}

func (c *V1Client) SystemBuilds(namespace string) SystemBuildInterface {
	return newSystemBuilds(c, namespace)
}

func (c *V1Client) SystemRollouts(namespace string) SystemRolloutInterface {
	return newSystemRollouts(c, namespace)
}

func (c *V1Client) SystemTeardowns(namespace string) SystemTeardownInterface {
	return newSystemTeardowns(c, namespace)
}

// NewForConfig creates a new V1Client for the given config.
func NewForConfig(c *rest.Config) (*V1Client, error) {
	config := *c
	if err := setConfigDefaults(&config); err != nil {
		return nil, err
	}
	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}
	return &V1Client{client}, nil
}

// NewForConfigOrDie creates a new V1Client for the given config and
// panics if there is an error in the config.
func NewForConfigOrDie(c *rest.Config) *V1Client {
	client, err := NewForConfig(c)
	if err != nil {
		panic(err)
	}
	return client
}

// New creates a new V1Client for the given RESTClient.
func New(c rest.Interface) *V1Client {
	return &V1Client{c}
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
func (c *V1Client) RESTClient() rest.Interface {
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
