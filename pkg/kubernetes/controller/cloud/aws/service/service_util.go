package service

import (
	crv1 "github.com/mlab-lattice/system/pkg/kubernetes/customresource/v1"
)

const (
	kubeFinalizerAWSServiceController = "aws.cloud.controllers.lattice.mlab.com/service"
)

func (sc *ServiceController) addFinalizer(svc *crv1.Service) error {
	// Check to see if the finalizer already exists. If so nothing needs to be done.
	for _, finalizer := range svc.Finalizers {
		if finalizer == kubeFinalizerAWSServiceController {
			return nil
		}
	}

	// Add the finalizer to the list and update.
	// If this fails due to a race the Service should get requeued by the controller, so
	// not a big deal.
	// TODO: investigate how long the requeue backoff could potentially delay this.
	// I *think* in theory it should only be able to lose this race once, since the
	// kubernetes-service-controller will have to wait for this controller to provision
	// services before making any more progress.
	// If it is a big deal, we could try to handle this error here, re-retrieve the Service
	// and try again. Not sure if there's gotchas around that (e.g. that Service being queued
	// and handled concurrently, which breaks some invariants. I don't think this should be
	// the case)
	svc.Finalizers = append(svc.Finalizers, kubeFinalizerAWSServiceController)
	return sc.latticeResourceRestClient.Put().
		Namespace(svc.Namespace).
		Resource(crv1.ServiceResourcePlural).
		Name(svc.Name).
		Body(svc).
		Do().
		Into(svc)
}
