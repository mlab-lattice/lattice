package service

import (
	"fmt"

	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	kubeutil "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/util/kubernetes"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Controller) syncDeletedService(service *latticev1.Service) error {
	address, err := c.address(service)
	if err != nil {
		return err
	}

	deployment, err := c.deployment(service)
	if err != nil {
		return err
	}

	deploymentStatus := &pendingDeploymentStatus
	if deployment != nil {
		deploymentStatus, err = c.getDeploymentStatus(service, deployment)
		if err != nil {
			return err
		}
	}

	// if the address still exists, delete it first so traffic stops being sent to the service
	if address != nil {
		message := "waiting for address to be deleted"

		// if the address is still deleting, nothing to do for now
		if address.DeletionTimestamp != nil {
			_, err = c.updateDeletedServiceStatus(service, &message, deploymentStatus, service.Status.Ports)
			return err
		}

		foregroundDelete := metav1.DeletePropagationForeground
		deleteOptions := &metav1.DeleteOptions{
			PropagationPolicy: &foregroundDelete,
		}

		err := c.latticeClient.LatticeV1().Addresses(address.Namespace).Delete(address.Name, deleteOptions)
		if err != nil {
			return fmt.Errorf(
				"error deleting %v for %v: %v",
				address.Description(c.namespacePrefix),
				service.Description(c.namespacePrefix),
				err,
			)
		}

		_, err = c.updateDeletedServiceStatus(service, &message, deploymentStatus, service.Status.Ports)
		return err
	}

	// if the deployment still exists, delete it once the address is deleted
	// FIXME: check to see if the deployment is deleted while pods are still terminating
	if deployment != nil {
		message := "waiting for instances to be deleted"

		// if the deployment is still deleting, nothing to do for now
		if deployment.DeletionTimestamp != nil {
			_, err = c.updateDeletedServiceStatus(service, &message, deploymentStatus, nil)
			return err
		}

		foregroundDelete := metav1.DeletePropagationForeground
		deleteOptions := &metav1.DeleteOptions{
			PropagationPolicy: &foregroundDelete,
		}

		err := c.kubeClient.AppsV1().Deployments(deployment.Namespace).Delete(deployment.Name, deleteOptions)
		if err != nil {
			return fmt.Errorf(
				"error deleting deployment %v for %v: %v",
				deployment.Name,
				service.Description(c.namespacePrefix),
				err,
			)
		}

		_, err = c.updateDeletedServiceStatus(service, &message, deploymentStatus, nil)
		return err
	}

	kubeServiceName := kubeutil.GetKubeServiceNameForService(service.Name)
	kubeService, err := c.kubeServiceLister.Services(service.Namespace).Get(kubeServiceName)
	if err != nil {
		if !errors.IsNotFound(err) {
			return fmt.Errorf("error getting kube service for %v: %v", service.Description(c.namespacePrefix), err)
		}

		kubeService = nil
	}

	if kubeService != nil {
		message := "waiting for internal resources to be deleted"

		// if the kube service is still deleting, nothing to do for now
		if kubeService.DeletionTimestamp != nil {
			_, err = c.updateDeletedServiceStatus(service, &message, deploymentStatus, nil)
			return err
		}

		foregroundDelete := metav1.DeletePropagationForeground
		deleteOptions := &metav1.DeleteOptions{
			PropagationPolicy: &foregroundDelete,
		}

		err := c.kubeClient.CoreV1().Services(kubeService.Namespace).Delete(kubeService.Name, deleteOptions)
		if err != nil {
			return fmt.Errorf(
				"error deleting kube service %v for %v: %v",
				kubeService.Name,
				service.Description(c.namespacePrefix),
				err,
			)
		}

		_, err = c.updateDeletedServiceStatus(service, &message, deploymentStatus, nil)
		return err
	}

	// if the deployment is deleted, update the service's node pool workload annotation
	// to indicate it is no longer running on any node pool epochs
	emptyAnnotation := make(latticev1.NodePoolAnnotationValue)
	service, err = c.updateServiceNodePoolAnnotation(service, emptyAnnotation)
	if err != nil {
		return err
	}

	selector, err := serviceNodePoolSelector(service)
	if err != nil {
		return err
	}

	nodePools, err := c.nodePoolLister.NodePools(service.Namespace).List(selector)
	if err != nil {
		err := fmt.Errorf(
			"error trying to get cached dedicated node pool for %v: %v",
			service.Description(c.namespacePrefix),
			err,
		)
		return err
	}

	// clean up the service's dedicated node pools if they exists
	for _, nodePool := range nodePools {
		// if the address is still deleting, nothing to do for now
		if nodePool.DeletionTimestamp != nil {
			continue
		}

		foregroundDelete := metav1.DeletePropagationForeground
		deleteOptions := &metav1.DeleteOptions{
			PropagationPolicy: &foregroundDelete,
		}

		err := c.latticeClient.LatticeV1().NodePools(nodePool.Namespace).Delete(nodePool.Name, deleteOptions)
		if err != nil {
			return fmt.Errorf(
				"error deleting %v for %v: %v",
				nodePool.Description(c.namespacePrefix),
				service.Description(c.namespacePrefix),
				err,
			)
		}
	}

	if len(nodePools) > 0 {
		message := "waiting for node pool to be deleted"
		_, err = c.updateDeletedServiceStatus(service, &message, deploymentStatus, nil)
		return err
	}

	_, err = c.removeFinalizer(service)
	return err
}

func (c *Controller) updateDeletedServiceStatus(
	service *latticev1.Service,
	message *string,
	deploymentStatus *deploymentStatus,
	ports map[int32]string,
) (*latticev1.Service, error) {
	return c.updateServiceStatus(
		service,
		latticev1.ServiceStateDeleting,
		message,
		nil,
		deploymentStatus.AvailableInstances,
		deploymentStatus.UpdatedInstances,
		deploymentStatus.StaleInstances,
		deploymentStatus.TerminatingInstances,
		ports,
	)
}
