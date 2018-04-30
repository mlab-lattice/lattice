package service

import (
	"fmt"

	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"

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

	// if the address still exists, delete it first so traffic stops being sent to the service
	if address != nil {
		// if the address is still deleting, nothing to do for now
		if address.DeletionTimestamp == nil {
			// FIXME: update status
			return nil
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

		// FIXME: update status
		return nil
	}

	// if the deployment still exists, delete it once the address is deleted
	// FIXME: check to see if the deployment is deleted while pods are still terminating
	if deployment != nil {
		// if the deployment is still deleting, nothing to do for now
		if address.DeletionTimestamp == nil {
			// FIXME: update status
			return nil
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

		// FIXME: update status
		return nil
	}

	// if the deployment is deleted, update the service's node pool workload annotation
	// to indicate it is no longer running on any node pool epochs
	emptyAnnotation := make(latticev1.NodePoolAnnotationValue)
	service, err = c.updateServiceNodePoolAnnotation(service, emptyAnnotation)
	if err != nil {
		return err
	}

	// clean up the service's dedicated node pool if one exists
	dedicatedNodePool, err := c.dedicatedNodePool(service)
	if err != nil {
		return err
	}

	if dedicatedNodePool != nil {
		// if the address is still deleting, nothing to do for now
		if address.DeletionTimestamp == nil {
			// FIXME: update status
			return nil
		}

		foregroundDelete := metav1.DeletePropagationForeground
		deleteOptions := &metav1.DeleteOptions{
			PropagationPolicy: &foregroundDelete,
		}

		err := c.latticeClient.LatticeV1().NodePools(dedicatedNodePool.Namespace).Delete(dedicatedNodePool.Name, deleteOptions)
		if err != nil {
			return fmt.Errorf(
				"error deleting %v for %v: %v",
				dedicatedNodePool.Description(c.namespacePrefix),
				service.Description(c.namespacePrefix),
				err,
			)
		}

		// FIXME: update status
		return nil
	}

	_, err = c.removeFinalizer(service)
	return err
}
