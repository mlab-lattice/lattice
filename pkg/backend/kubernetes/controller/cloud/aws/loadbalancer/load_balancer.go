package loadbalancer

import (
	"fmt"
	"reflect"
	"strings"

	awscloudprovider "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/cloudprovider/aws"
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	kubetf "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/terraform/aws"
	kubeutil "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/util/kubernetes"
	tf "github.com/mlab-lattice/lattice/pkg/util/terraform"
	awstfprovider "github.com/mlab-lattice/lattice/pkg/util/terraform/provider/aws"

	"k8s.io/apimachinery/pkg/api/errors"

	"github.com/golang/glog"
)

const (
	finalizerName = "load-balancer.aws.cloud-provider.lattice.mlab.com"

	terraformOutputDNSName = "dns_name"
)

func (c *Controller) syncDeletedLoadBalancer(loadBalancer *latticev1.LoadBalancer) error {
	if err := c.deleteKubeService(loadBalancer); err != nil && !errors.IsNotFound(err) {
		return err
	}

	if err := c.deprovisionLoadBalancer(loadBalancer); err != nil {
		return err
	}

	_, err := c.removeFinalizer(loadBalancer)
	return err
}

func (c *Controller) nodePoolProvisioned(loadBalancer *latticev1.LoadBalancer) (bool, error) {
	nodePool, err := c.nodePoolLister.NodePools(loadBalancer.Namespace).Get(loadBalancer.Name)
	if err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		}

		return false, err
	}

	return nodePool.Status.State == latticev1.NodePoolStateStable, nil
}

func (c *Controller) provisionLoadBalancer(loadBalancer *latticev1.LoadBalancer) (*latticev1.LoadBalancer, error) {
	loadBalancerModule, err := c.loadBalancerModule(loadBalancer)
	if err != nil {
		return nil, err
	}

	config, err := c.loadBalancerConfig(loadBalancer, loadBalancerModule)
	if err != nil {
		return nil, err
	}

	_, err = tf.Apply(workDirectory(loadBalancer), config)
	if err != nil {
		return nil, err
	}

	annotations, err := c.loadBalancerAnnotations(loadBalancer)
	if err != nil {
		return nil, err
	}

	if !reflect.DeepEqual(loadBalancer.Annotations, annotations) {
		// Copy so the shared cache isn't mutated
		loadBalancer = loadBalancer.DeepCopy()
		loadBalancer.Annotations = annotations

		loadBalancer, err = c.latticeClient.LatticeV1().LoadBalancers(loadBalancer.Namespace).Update(loadBalancer)
		if err != nil {
			return nil, err
		}
	}

	return loadBalancer, nil
}

func (c *Controller) deprovisionLoadBalancer(loadBalancer *latticev1.LoadBalancer) error {
	config, err := c.loadBalancerConfig(loadBalancer, nil)
	if err != nil {
		return err
	}

	_, err = tf.Destroy(workDirectory(loadBalancer), config)
	return err
}

func (c *Controller) loadBalancerConfig(
	loadBalancer *latticev1.LoadBalancer,
	loadBalancerModule *kubetf.ApplicationLoadBalancer,
) (*tf.Config, error) {
	systemID, err := kubeutil.SystemID(loadBalancer.Namespace)
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
				"%v/%v",
				kubetf.GetS3BackendSystemStatePathRoot(c.latticeID, systemID),
				loadBalancer.Name,
			),
			Encrypt: true,
		},
	}

	if loadBalancerModule != nil {
		config.Modules = map[string]interface{}{
			"load-balancer": loadBalancerModule,
		}

		config.Output = map[string]tf.ConfigOutput{
			terraformOutputDNSName: {
				Value: fmt.Sprintf("${module.load-balancer.%v}", terraformOutputDNSName),
			},
		}
	}

	return config, nil
}

func (c *Controller) loadBalancerModule(loadBalancer *latticev1.LoadBalancer) (*kubetf.ApplicationLoadBalancer, error) {
	service, err := c.serviceLister.Services(loadBalancer.Namespace).Get(loadBalancer.Name)
	if err != nil {
		return nil, err
	}

	kubeServiceName := kubeutil.GetKubeServiceNameForLoadBalancer(loadBalancer.Name)
	kubeService, err := c.kubeServiceLister.Services(loadBalancer.Namespace).Get(kubeServiceName)
	if err != nil {
		return nil, err
	}

	servicePorts, err := c.serviceMesh.ServicePorts(service)
	if err != nil {
		return nil, err
	}

	nodePorts := map[int32]int32{}
	for _, port := range kubeService.Spec.Ports {
		nodePorts[servicePorts[port.Port]] = port.NodePort
	}

	systemID, err := kubeutil.SystemID(loadBalancer.Namespace)
	if err != nil {
		return nil, err
	}

	nodePool, err := c.nodePoolLister.NodePools(loadBalancer.Namespace).Get(loadBalancer.Name)
	if err != nil {
		return nil, err
	}

	loadBalancerModule := kubetf.NewApplicationLoadBalancerModule(
		c.terraformModuleRoot,
		c.awsCloudProvider.Region(),
		string(c.latticeID),
		string(systemID),
		c.awsCloudProvider.VPCID(),
		strings.Join(c.awsCloudProvider.SubnetIDs(), ","),
		loadBalancer.Name,
		nodePool.Annotations[awscloudprovider.AnnotationKeyNodePoolAutoscalingGroupName],
		nodePool.Annotations[awscloudprovider.AnnotationKeyNodePoolSecurityGroupID],
		nodePorts,
	)
	return loadBalancerModule, nil
}

type loadBalancerInfo struct {
	DNSName string
}

func (c *Controller) currentLoadBalancerInfo(loadBalancer *latticev1.LoadBalancer) (loadBalancerInfo, error) {
	outputVars := []string{terraformOutputDNSName}

	loadBalancerModule, err := c.loadBalancerModule(loadBalancer)
	if err != nil {
		return loadBalancerInfo{}, err
	}

	config, err := c.loadBalancerConfig(loadBalancer, loadBalancerModule)
	if err != nil {
		return loadBalancerInfo{}, err
	}

	values, err := tf.Output(workDirectory(loadBalancer), config, outputVars)
	if err != nil {
		return loadBalancerInfo{}, err
	}

	info := loadBalancerInfo{
		DNSName: values[terraformOutputDNSName],
	}
	return info, nil
}

func (c *Controller) updateLoadBalancerStatus(
	loadBalancer *latticev1.LoadBalancer,
	status latticev1.LoadBalancerStatus,
) (*latticev1.LoadBalancer, error) {
	if reflect.DeepEqual(loadBalancer.Status, status) {
		return loadBalancer, nil
	}

	// Copy the service so the shared cache isn't mutated
	loadBalancer = loadBalancer.DeepCopy()
	loadBalancer.Status = status

	return c.latticeClient.LatticeV1().LoadBalancers(loadBalancer.Namespace).Update(loadBalancer)

	// TODO: switch to this when https://github.com/kubernetes/kubernetes/issues/38113 is merged
	// TODO: also watch https://github.com/kubernetes/kubernetes/pull/55168
	//return c.latticeClient.LatticeV1().LoadBalancers(loadBalancer.Namespace).UpdateStatus(loadBalancer)
}

func (c *Controller) loadBalancerAnnotations(loadBalancer *latticev1.LoadBalancer) (map[string]string, error) {
	info, err := c.currentLoadBalancerInfo(loadBalancer)
	if err != nil {
		return nil, err
	}

	annotations := map[string]string{
		awscloudprovider.AnnotationKeyLoadBalancerDNSName: info.DNSName,
	}
	return annotations, nil
}

func (c *Controller) addFinalizer(loadBalancer *latticev1.LoadBalancer) (*latticev1.LoadBalancer, error) {
	// Check to see if the finalizer already exists. If so nothing needs to be done.
	for _, finalizer := range loadBalancer.Finalizers {
		if finalizer == finalizerName {
			glog.V(5).Infof("LoadBalancer %v has %v finalizer", loadBalancer.Name, finalizerName)
			return loadBalancer, nil
		}
	}

	// Add the finalizer to the list and update.
	// If this fails due to a race the Endpoint should get requeued by the controller, so
	// not a big deal.
	loadBalancer.Finalizers = append(loadBalancer.Finalizers, finalizerName)
	glog.V(5).Infof("LoadBalancer %v missing %v finalizer, adding it", loadBalancer.Name, finalizerName)

	return c.latticeClient.LatticeV1().LoadBalancers(loadBalancer.Namespace).Update(loadBalancer)
}

func (c *Controller) removeFinalizer(loadBalancer *latticev1.LoadBalancer) (*latticev1.LoadBalancer, error) {
	// Build up a list of all the finalizers except the aws service controller finalizer.
	var finalizers []string
	found := false
	for _, finalizer := range loadBalancer.Finalizers {
		if finalizer == finalizerName {
			found = true
			continue
		}
		finalizers = append(finalizers, finalizer)
	}

	// If the finalizer wasn't part of the list, nothing to do.
	if !found {
		return loadBalancer, nil
	}

	// The finalizer was in the list, so we should remove it.
	loadBalancer.Finalizers = finalizers
	return c.latticeClient.LatticeV1().LoadBalancers(loadBalancer.Namespace).Update(loadBalancer)
}

func workDirectory(loadBalancer *latticev1.LoadBalancer) string {
	return fmt.Sprintf(
		"/tmp/lattice-controller-manager/controllers/cloud/aws/load-balancer/terraform/%v/%v",
		loadBalancer.Namespace,
		loadBalancer.Name,
	)
}
