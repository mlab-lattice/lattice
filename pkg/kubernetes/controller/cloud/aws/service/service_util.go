package service

import (
	crv1 "github.com/mlab-lattice/system/pkg/kubernetes/customresource/v1"

	"github.com/golang/glog"
)

const (
	kubeFinalizerAWSServiceController = "aws.cloud.controllers.lattice.mlab.com/service"
)

func (sc *ServiceController) addFinalizer(svc *crv1.Service) error {
	// Check to see if the finalizer already exists. If so nothing needs to be done.
	for _, finalizer := range svc.Finalizers {
		if finalizer == kubeFinalizerAWSServiceController {
			glog.V(5).Infof("Service %v has %v finalizer", svc.Name, kubeFinalizerAWSServiceController)
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
	glog.V(5).Infof("Service %v missing %v finalizer, adding it", svc.Name, kubeFinalizerAWSServiceController)
	result, err := sc.latticeClient.V1().Services(svc.Namespace).Update(svc)
	if err != nil {
		return err
	}
	*svc = *result
	return nil
}

func (sc *ServiceController) removeFinalizer(svc *crv1.Service) error {
	// Build up a list of all the finalizers except the aws service controller finalizer.
	finalizers := []string{}
	found := false
	for _, finalizer := range svc.Finalizers {
		if finalizer == kubeFinalizerAWSServiceController {
			found = true
			continue
		}
		finalizers = append(finalizers, finalizer)
	}

	// If the finalizer wasn't part of the list, nothing to do.
	if !found {
		return nil
	}

	// The finalizer was in the list, so we should remove it.
	svc.Finalizers = finalizers
	result, err := sc.latticeClient.V1().Services(svc.Namespace).Update(svc)
	if err != nil {
		return err
	}
	*svc = *result
	return nil
}
