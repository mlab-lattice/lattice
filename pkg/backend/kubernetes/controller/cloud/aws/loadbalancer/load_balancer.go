package loadbalancer

import (
	"fmt"
	"reflect"
	"strings"

	awscloudprovider "github.com/mlab-lattice/system/pkg/backend/kubernetes/cloudprovider/aws"
	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	kubetf "github.com/mlab-lattice/system/pkg/backend/kubernetes/terraform/aws"
	kubeutil "github.com/mlab-lattice/system/pkg/backend/kubernetes/util/kubernetes"
	tf "github.com/mlab-lattice/system/pkg/terraform"
	awstfprovider "github.com/mlab-lattice/system/pkg/terraform/provider/aws"
)

const (
	terraformOutputDNSName = "dns_name"
)

func (c *Controller) provisionLoadBalancer(loadBalancer *crv1.LoadBalancer) (*crv1.LoadBalancer, error) {
	config, err := c.loadBalancerConfig(loadBalancer)
	if err != nil {
		return nil, err
	}

	err = tf.Apply(workDirectory(loadBalancer), config)
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

func (c *Controller) deprovisionLoadBalancer(loadBalancer *crv1.LoadBalancer) error {
	config, err := c.loadBalancerConfig(loadBalancer)
	if err != nil {
		return err
	}

	return tf.Destroy(workDirectory(loadBalancer), config)
}

func (c *Controller) loadBalancerConfig(loadBalancer *crv1.LoadBalancer) (*tf.Config, error) {
	systemID, err := kubeutil.SystemID(loadBalancer.Namespace)
	if err != nil {
		return nil, err
	}

	service, err := c.serviceLister.Services(loadBalancer.Namespace).Get(loadBalancer.Name)
	if err != nil {
		return nil, err
	}

	kubeServiceName := kubeutil.GetKubeServiceNameForLoadBalancer(loadBalancer.Name)
	kubeService, err := c.kubeServiceLister.Services(loadBalancer.Namespace).Get(kubeServiceName)
	if err != nil {
		return nil, err
	}

	ports := map[int32]int32{}
	for _, port := range kubeService.Spec.Ports {
		ports[port.Port] = port.NodePort
	}

	loadBalancerModule := kubetf.NewApplicationLoadBalancerModule(
		c.terraformModuleRoot,
		c.awsCloudProvider.Region(),
		string(c.clusterID),
		string(systemID),
		c.awsCloudProvider.VPCID(),
		strings.Join(c.awsCloudProvider.SubnetIDs(), ","),
		loadBalancer.Name,
		service.Annotations[awscloudprovider.AnnotationKeyNodePoolAutoscalingGroupName],
		service.Annotations[awscloudprovider.AnnotationKeyNodePoolSecurityGroupID],
		ports,
	)

	config := &tf.Config{
		Provider: awstfprovider.Provider{
			Region: c.awsCloudProvider.Region(),
		},
		Backend: tf.S3BackendConfig{
			Region: c.awsCloudProvider.Region(),
			Bucket: c.terraformBackendOptions.S3.Bucket,
			Key: fmt.Sprintf(
				"%v/%v",
				kubetf.GetS3BackendSystemStatePathRoot(c.clusterID, systemID),
				loadBalancer.Name,
			),
			Encrypt: true,
		},
		Modules: map[string]interface{}{
			"load-balancer": loadBalancerModule,
		},
		Output: map[string]tf.ConfigOutput{
			terraformOutputDNSName: {
				Value: fmt.Sprintf("${module.load-balancer.%v}", terraformOutputDNSName),
			},
		},
	}
	return config, nil
}

type loadBalancerInfo struct {
	DNSName string
}

func (c *Controller) currentLoadBalancerInfo(loadBalancer *crv1.LoadBalancer) (loadBalancerInfo, error) {
	outputVars := []string{terraformOutputDNSName}

	config, err := c.loadBalancerConfig(loadBalancer)
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
	loadBalancer *crv1.LoadBalancer,
	state crv1.LoadBalancerState,
) (*crv1.LoadBalancer, error) {
	status := crv1.LoadBalancerStatus{
		State: state,
	}

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

func (c *Controller) loadBalancerAnnotations(loadBalancer *crv1.LoadBalancer) (map[string]string, error) {
	info, err := c.currentLoadBalancerInfo(loadBalancer)
	if err != nil {
		return nil, err
	}

	annotations := map[string]string{
		awscloudprovider.AnnotationKeyLoadBalancerDNSName: info.DNSName,
	}
	return annotations, nil
}

func workDirectory(loadBalancer *crv1.LoadBalancer) string {
	return "/tmp/lattice-controller-manager/controllers/cloud/aws/node-pool/terraform/" + loadBalancer.Namespace + "/" + loadBalancer.Name
}
