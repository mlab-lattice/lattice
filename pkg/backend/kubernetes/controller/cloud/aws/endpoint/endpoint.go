package endpoint

import (
	"fmt"
	"reflect"

	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	kubetf "github.com/mlab-lattice/system/pkg/backend/kubernetes/terraform/aws"
	kubeutil "github.com/mlab-lattice/system/pkg/backend/kubernetes/util/kubernetes"
	tf "github.com/mlab-lattice/system/pkg/terraform"
	awstfprovider "github.com/mlab-lattice/system/pkg/terraform/provider/aws"

	"github.com/golang/glog"
)

const (
	finalizerName = "endpoint.aws.cloud-provider.lattice.mlab.com"
)

func (c *Controller) syncDeletedEndpoint(endpoint *crv1.Endpoint) error {
	err := c.deprovisionEndpoint(endpoint)
	if err != nil {
		return err
	}

	_, err = c.removeFinalizer(endpoint)
	return err
}

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

func (c *Controller) updateEndpointStatus(
	endpoint *crv1.Endpoint,
	status crv1.EndpointStatus,
) (*crv1.Endpoint, error) {
	if reflect.DeepEqual(endpoint.Status, status) {
		return endpoint, nil
	}

	// Copy the service so the shared cache isn't mutated
	endpoint = endpoint.DeepCopy()
	endpoint.Status = status

	return c.latticeClient.LatticeV1().Endpoints(endpoint.Namespace).Update(endpoint)

	// TODO: switch to this when https://github.com/kubernetes/kubernetes/issues/38113 is merged
	// TODO: also watch https://github.com/kubernetes/kubernetes/pull/55168
	//return c.latticeClient.LatticeV1().LoadBalancers(loadBalancer.Namespace).UpdateStatus(loadBalancer)
}

func (c *Controller) addFinalizer(endpoint *crv1.Endpoint) (*crv1.Endpoint, error) {
	// Check to see if the finalizer already exists. If so nothing needs to be done.
	for _, finalizer := range endpoint.Finalizers {
		if finalizer == finalizerName {
			glog.V(5).Infof("Endpoint %v has %v finalizer", endpoint.Name, finalizerName)
			return endpoint, nil
		}
	}

	// Add the finalizer to the list and update.
	// If this fails due to a race the Endpoint should get requeued by the controller, so
	// not a big deal.
	endpoint.Finalizers = append(endpoint.Finalizers, finalizerName)
	glog.V(5).Infof("Endpoint %v missing %v finalizer, adding it", endpoint.Name, finalizerName)

	return c.latticeClient.LatticeV1().Endpoints(endpoint.Namespace).Update(endpoint)
}

func (c *Controller) removeFinalizer(endpoint *crv1.Endpoint) (*crv1.Endpoint, error) {
	// Build up a list of all the finalizers except the aws service controller finalizer.
	var finalizers []string
	found := false
	for _, finalizer := range endpoint.Finalizers {
		if finalizer == finalizerName {
			found = true
			continue
		}
		finalizers = append(finalizers, finalizer)
	}

	// If the finalizer wasn't part of the list, nothing to do.
	if !found {
		return endpoint, nil
	}

	// The finalizer was in the list, so we should remove it.
	endpoint.Finalizers = finalizers
	return c.latticeClient.LatticeV1().Endpoints(endpoint.Namespace).Update(endpoint)
}

func workDirectory(endpoint *crv1.Endpoint) string {
	return "/tmp/lattice-controller-manager/controllers/cloud/aws/endpoint/terraform/" + endpoint.Namespace + "/" + endpoint.Name
}
