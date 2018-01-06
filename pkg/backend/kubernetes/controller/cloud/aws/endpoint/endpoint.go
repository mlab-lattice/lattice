package endpoint

import (
	"fmt"

	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	kubetf "github.com/mlab-lattice/system/pkg/backend/kubernetes/terraform/aws"
	kubeutil "github.com/mlab-lattice/system/pkg/backend/kubernetes/util/kubernetes"
	tf "github.com/mlab-lattice/system/pkg/terraform"
	awstfprovider "github.com/mlab-lattice/system/pkg/terraform/provider/aws"
)

func (c *Controller) provisionEndpoint(endpoint *crv1.Endpoint) error {
	if endpoint.Spec.ExternalName != nil {
		return c.provisionExternalNameEndpoint(endpoint)
	}

	if endpoint.Spec.IP != nil {
		return c.provisionIPEndpoint(endpoint)
	}

	return fmt.Errorf("endpoint must have either ExternalName or IP")
}

func (c *Controller) provisionExternalNameEndpoint(endpoint *crv1.Endpoint) error {
	config, err := c.endpointConfig(endpoint, c.externalNameEndpointModule(endpoint))
	if err != nil {
		return err
	}

	return tf.Apply(workDirectory(endpoint), config)
}

func (c *Controller) provisionIPEndpoint(endpoint *crv1.Endpoint) error {
	config, err := c.endpointConfig(endpoint, c.ipEndpointModule(endpoint))
	if err != nil {
		return err
	}

	return tf.Apply(workDirectory(endpoint), config)
}

func (c *Controller) deprovisionEndpoint(endpoint *crv1.Endpoint) error {
	if endpoint.Spec.ExternalName != nil {
		return c.deprovisionExternalNameEndpoint(endpoint)
	}

	if endpoint.Spec.IP != nil {
		return c.deprovisionIPEndpoint(endpoint)
	}

	return fmt.Errorf("endpoint must have either ExternalName or IP")
}

func (c *Controller) deprovisionExternalNameEndpoint(endpoint *crv1.Endpoint) error {
	config, err := c.endpointConfig(endpoint, c.externalNameEndpointModule(endpoint))
	if err != nil {
		return err
	}

	return tf.Destroy(workDirectory(endpoint), config)
}

func (c *Controller) deprovisionIPEndpoint(endpoint *crv1.Endpoint) error {
	config, err := c.endpointConfig(endpoint, c.ipEndpointModule(endpoint))
	if err != nil {
		return err
	}

	return tf.Destroy(workDirectory(endpoint), config)
}

func (c *Controller) endpointConfig(endpoint *crv1.Endpoint, endpointModule interface{}) (*tf.Config, error) {
	systemID, err := kubeutil.SystemID(endpoint.Namespace)
	if err != nil {
		return nil, err
	}

	config := &tf.Config{
		Provider: awstfprovider.Provider{
			Region: c.awsCloudProvider.Region(),
		},
		Backend: tf.S3BackendConfig{
			Region: c.awsCloudProvider.Region(),
			Bucket: c.terraformBackendOptions.S3.Bucket,
			Key: fmt.Sprintf(
				"%v/%v/%v",
				kubetf.GetS3BackendSystemStatePathRoot(c.clusterID, systemID),
				"endpoints",
				endpoint.Name,
			),
			Encrypt: true,
		},
		Modules: map[string]interface{}{
			"endpoint": endpointModule,
		},
	}

	return config, nil
}

func (c *Controller) ipEndpointModule(endpoint *crv1.Endpoint) *kubetf.IPEndpoint {
	return kubetf.NewIPEndpointModule(
		c.terraformModuleRoot,
		c.awsCloudProvider.Region(),
		c.awsCloudProvider.Route53PrivateZoneID(),
		endpoint.Spec.Path.ToDomain(true),
		*endpoint.Spec.IP,
	)
}

func (c *Controller) externalNameEndpointModule(endpoint *crv1.Endpoint) *kubetf.ExternalNameEndpoint {
	return kubetf.NewExternalNameEndpointModule(
		c.terraformModuleRoot,
		c.awsCloudProvider.Region(),
		c.awsCloudProvider.Route53PrivateZoneID(),
		endpoint.Spec.Path.ToDomain(true),
		*endpoint.Spec.ExternalName,
	)
}

func workDirectory(endpoint *crv1.Endpoint) string {
	return "/tmp/lattice-controller-manager/controllers/cloud/aws/endpoint/terraform/" + endpoint.Name
}
