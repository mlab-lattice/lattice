package endpoint

import (
	"fmt"

	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	kubetf "github.com/mlab-lattice/system/pkg/backend/kubernetes/terraform/aws"
	kubeutil "github.com/mlab-lattice/system/pkg/backend/kubernetes/util/kubernetes"
	tf "github.com/mlab-lattice/system/pkg/terraform"
	awstfprovider "github.com/mlab-lattice/system/pkg/terraform/provider/aws"

	"github.com/mlab-lattice/system/pkg/types"
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
	systemName, err := kubeutil.SystemID(endpoint.Namespace)
	if err != nil {
		return err
	}

	externalNameEndpointModule := kubetf.NewExternalNameEndpointModule(
		c.terraformModuleRoot,
		c.awsCloudProvider.Region(),
		c.awsCloudProvider.Route53PrivateZoneID(),
		endpoint.Spec.Path.ToDomain(true),
		*endpoint.Spec.ExternalName,
	)

	config := c.endpointConfig(endpoint.Name, systemName, externalNameEndpointModule)
	return tf.Apply(workDirectory(endpoint), config)
}

func (c *Controller) provisionIPEndpoint(endpoint *crv1.Endpoint) error {
	systemName, err := kubeutil.SystemID(endpoint.Namespace)
	if err != nil {
		return err
	}

	ipEndpointModule := kubetf.NewIPEndpointModule(
		c.terraformModuleRoot,
		c.awsCloudProvider.Region(),
		c.awsCloudProvider.Route53PrivateZoneID(),
		endpoint.Spec.Path.ToDomain(true),
		*endpoint.Spec.IP,
	)

	config := c.endpointConfig(endpoint.Name, systemName, ipEndpointModule)
	return tf.Apply(workDirectory(endpoint), config)
}

func (c *Controller) deprovisionIPEndpoint(endpoint *crv1.Endpoint) error {
	systemName, err := kubeutil.SystemID(endpoint.Namespace)
	if err != nil {
		return err
	}

	config := c.endpointConfig(endpoint.Name, systemName, nil)
	return tf.Destroy(workDirectory(endpoint), config)
}

func (c *Controller) endpointConfig(endpointName string, systemID types.SystemID, endpointModule interface{}) *tf.Config {
	return &tf.Config{
		Provider: awstfprovider.Provider{
			Region: c.awsCloudProvider.Region(),
		},
		Backend: tf.S3BackendConfig{
			Region: c.awsCloudProvider.Region(),
			Bucket: c.terraformBackendOptions.S3.Bucket,
			Key: fmt.Sprintf(
				"%v/%v/%v",
				kubetf.GetS3BackendStatePathRoot(c.clusterID, systemID),
				"endpoints",
				endpointName,
			),
			Encrypt: true,
		},
		Modules: map[string]interface{}{
			"endpoint": endpointModule,
		},
	}
}

func workDirectory(endpoint *crv1.Endpoint) string {
	return "/tmp/lattice-controller-manager/controllers/cloud/aws/endpoint/terraform/" + endpoint.Name
}
