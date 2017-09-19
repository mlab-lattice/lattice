package service

import (
	"fmt"
	crv1 "github.com/mlab-lattice/kubernetes-integration/pkg/api/customresource/v1"
	"reflect"
	"time"

	corev1 "k8s.io/api/core/v1"
	extensions "k8s.io/api/extensions/v1beta1"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"

	extensioninformers "k8s.io/client-go/informers/extensions/v1beta1"
	clientset "k8s.io/client-go/kubernetes"
	extensionlisters "k8s.io/client-go/listers/extensions/v1beta1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"github.com/golang/glog"
)

var controllerKind = crv1.SchemeGroupVersion.WithKind("Service")

// We'll use LatticeService to differentiate between kubernetes' Service
type ServiceController struct {
	syncHandler    func(bKey string) error
	enqueueService func(cb *crv1.Service)

	latticeResourceRestClient rest.Interface
	kubeClient                clientset.Interface

	serviceStore       cache.Store
	serviceStoreSynced cache.InformerSynced

	serviceBuildStore       cache.Store
	serviceBuildStoreSynced cache.InformerSynced

	componentBuildStore       cache.Store
	componentBuildStoreSynced cache.InformerSynced

	// TODO: switch to apps when stabilized https://github.com/kubernetes/features/issues/353
	deploymentLister       extensionlisters.DeploymentLister
	deploymentListerSynced cache.InformerSynced

	queue workqueue.RateLimitingInterface
}

func NewServiceController(
	kubeClient clientset.Interface,
	latticeResourceRestClient rest.Interface,
	serviceInformer cache.SharedInformer,
	serviceBuildInformer cache.SharedInformer,
	componentBuildInformer cache.SharedInformer,
	deploymentInformer extensioninformers.DeploymentInformer,
) *ServiceController {
	sc := &ServiceController{
		latticeResourceRestClient: latticeResourceRestClient,
		kubeClient:                kubeClient,
		queue:                     workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "service"),
	}

	sc.syncHandler = sc.syncService
	sc.enqueueService = sc.enqueue

	serviceInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    sc.addService,
		UpdateFunc: sc.updateService,
		DeleteFunc: sc.deleteService,
	})
	sc.serviceStore = serviceInformer.GetStore()
	sc.serviceStoreSynced = serviceInformer.HasSynced

	sc.serviceBuildStore = serviceBuildInformer.GetStore()
	sc.serviceBuildStoreSynced = serviceBuildInformer.HasSynced

	sc.componentBuildStore = componentBuildInformer.GetStore()
	sc.componentBuildStoreSynced = componentBuildInformer.HasSynced

	deploymentInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    sc.addDeployment,
		UpdateFunc: sc.updateDeployment,
		DeleteFunc: sc.deleteDeployment,
	})
	sc.deploymentLister = deploymentInformer.Lister()
	sc.deploymentListerSynced = deploymentInformer.Informer().HasSynced

	return sc
}

func (sc *ServiceController) addService(obj interface{}) {
	svc := obj.(*crv1.Service)
	glog.V(4).Infof("Adding Service %s", svc.Name)
	sc.enqueueService(svc)
}

func (sc *ServiceController) updateService(old, cur interface{}) {
	oldSvc := old.(*crv1.Service)
	curSvc := cur.(*crv1.Service)
	glog.V(4).Infof("Updating Service %s", oldSvc.Name)
	sc.enqueueService(curSvc)
}

func (sc *ServiceController) deleteService(obj interface{}) {
	svc, ok := obj.(*crv1.Service)
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			runtime.HandleError(fmt.Errorf("couldn't get object from tombstone %#v", obj))
			return
		}
		svc, ok = tombstone.Obj.(*crv1.Service)
		if !ok {
			runtime.HandleError(fmt.Errorf("tombstone contained object that is not a Service %#v", obj))
			return
		}
	}
	glog.V(4).Infof("Deleting Service %s", svc.Name)
	sc.enqueueService(svc)
}

// addDeployment enqueues the Service that manages a Deployment when the Deployment is created.
func (sc *ServiceController) addDeployment(obj interface{}) {
	d := obj.(*extensions.Deployment)

	if d.DeletionTimestamp != nil {
		// On a restart of the controller manager, it's possible for an object to
		// show up in a state that is already pending deletion.
		sc.deleteDeployment(d)
		return
	}

	// If it has a ControllerRef, that's all that matters.
	if controllerRef := metav1.GetControllerOf(d); controllerRef != nil {
		svc := sc.resolveControllerRef(d.Namespace, controllerRef)

		// Not a Service Deployment.
		if svc == nil {
			return
		}

		glog.V(4).Infof("Deployment %s added.", d.Name)
		sc.enqueueService(svc)
		return
	}

	// Otherwise, it's an orphan. These should never exist. All deployments should be run by some
	// controller.
	// FIXME: send warn event
}

// updateDeployment figures out what Service manages a Deployment when the Deployment
// is updated and enqueues it.
func (sc *ServiceController) updateDeployment(old, cur interface{}) {
	glog.V(5).Info("Got Deployment update")
	oldD := old.(*extensions.Deployment)
	curD := cur.(*extensions.Deployment)
	if curD.ResourceVersion == oldD.ResourceVersion {
		// Periodic resync will send update events for all known Deployments.
		// Two different versions of the same Deployment will always have different RVs.
		glog.V(5).Info("Deployment ResourceVersions are the same")
		return
	}

	curControllerRef := metav1.GetControllerOf(curD)
	oldControllerRef := metav1.GetControllerOf(oldD)
	controllerRefChanged := !reflect.DeepEqual(curControllerRef, oldControllerRef)
	if controllerRefChanged {
		// The ControllerRef was changed. If this is a Service Deployment, this shouldn't happen.
		if b := sc.resolveControllerRef(oldD.Namespace, oldControllerRef); b != nil {
			// FIXME: send error event here, this should not happen
		}
	}

	// If it has a ControllerRef, that's all that matters.
	if curControllerRef != nil {
		svc := sc.resolveControllerRef(curD.Namespace, curControllerRef)

		// Not a Service Deployment
		if svc == nil {
			return
		}

		sc.enqueueService(svc)
		return
	}

	// Otherwise, it's an orphan. These should never exist. All deployments should be run by some
	// controller.
	// FIXME: send warn event
}

// deleteDeployment enqueues the Service that manages a Deployment when
// the Deployment is deleted.
func (sc *ServiceController) deleteDeployment(obj interface{}) {
	d, ok := obj.(*extensions.Deployment)

	// When a delete is dropped, the relist will notice a pod in the store not
	// in the list, leading to the insertion of a tombstone object which contains
	// the deleted key/value.
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			runtime.HandleError(fmt.Errorf("couldn't get object from tombstone %#v", obj))
			return
		}
		d, ok = tombstone.Obj.(*extensions.Deployment)
		if !ok {
			runtime.HandleError(fmt.Errorf("tombstone contained object that is not a Deployment %#v", obj))
			return
		}
	}

	controllerRef := metav1.GetControllerOf(d)
	if controllerRef == nil {
		// No controller should care about orphans being deleted.
		return
	}

	svc := sc.resolveControllerRef(d.Namespace, controllerRef)

	// Not a Service Deployment
	if svc == nil {
		return
	}

	glog.V(4).Infof("Deployment %s deleted.", d.Name)
	sc.enqueueService(svc)
}

func (sc *ServiceController) enqueue(svc *crv1.Service) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(svc)
	if err != nil {
		runtime.HandleError(fmt.Errorf("couldn't get key for object %#v: %v", svc, err))
		return
	}

	sc.queue.Add(key)
}

// resolveControllerRef returns the controller referenced by a ControllerRef,
// or nil if the ControllerRef could not be resolved to a matching controller
// of the correct Kind.
func (sc *ServiceController) resolveControllerRef(namespace string, controllerRef *metav1.OwnerReference) *crv1.Service {
	// We can't look up by Name, so look up by Name and then verify Name.
	// Don't even try to look up by Name if it's the wrong Kind.
	if controllerRef.Kind != controllerKind.Kind {
		return nil
	}

	svcKey := fmt.Sprintf("%v/%v", namespace, controllerRef.Name)
	svcObj, exists, err := sc.serviceStore.GetByKey(svcKey)
	if err != nil || !exists {
		return nil
	}

	svc := svcObj.(*crv1.Service)

	if svc.UID != controllerRef.UID {
		// The controller we found with this Name is not the same one that the
		// ControllerRef points to.
		return nil
	}
	return svc
}

func (sc *ServiceController) Run(workers int, stopCh <-chan struct{}) {
	// don't let panics crash the process
	defer runtime.HandleCrash()
	// make sure the work queue is shutdown which will trigger workers to end
	defer sc.queue.ShutDown()

	glog.Infof("Starting service controller")
	defer glog.Infof("Shutting down service controller")

	// wait for your secondary caches to fill before starting your work
	if !cache.WaitForCacheSync(stopCh, sc.serviceStoreSynced, sc.serviceBuildStoreSynced, sc.componentBuildStoreSynced, sc.deploymentListerSynced) {
		return
	}

	glog.V(4).Info("Caches synced.")

	// start up your worker threads based on threadiness.  Some controllers
	// have multiple kinds of workers
	for i := 0; i < workers; i++ {
		// runWorker will loop until "something bad" happens.  The .Until will
		// then rekick the worker after one second
		go wait.Until(sc.runWorker, time.Second, stopCh)
	}

	// wait until we're told to stop
	<-stopCh
}

func (sc *ServiceController) runWorker() {
	// hot loop until we're told to stop.  processNextWorkItem will
	// automatically wait until there's work available, so we don't worry
	// about secondary waits
	for sc.processNextWorkItem() {
	}
}

// processNextWorkItem deals with one key off the queue.  It returns false
// when it's time to quit.
func (sc *ServiceController) processNextWorkItem() bool {
	// pull the next work item from queue.  It should be a key we use to lookup
	// something in a cache
	key, quit := sc.queue.Get()
	if quit {
		return false
	}
	// you always have to indicate to the queue that you've completed a piece of
	// work
	defer sc.queue.Done(key)

	// do your work on the key.  This method will contains your "do stuff" logic
	err := sc.syncHandler(key.(string))
	if err == nil {
		// if you had no error, tell the queue to stop tracking history for your
		// key. This will reset things like failure counts for per-item rate
		// limiting
		sc.queue.Forget(key)
		return true
	}

	// there was a failure so be sure to report it.  This method allows for
	// pluggable error handling which can be used for things like
	// cluster-monitoring
	runtime.HandleError(fmt.Errorf("%v failed with : %v", key, err))

	// since we failed, we should requeue the item to work on later.  This
	// method will add a backoff to avoid hotlooping on particular items
	// (they're probably still not going to work right away) and overall
	// controller protection (everything I've done is broken, this controller
	// needs to calm down or it can starve other useful work) cases.
	sc.queue.AddRateLimited(key)

	return true
}

// syncService will sync the Service with the given key.
// This function is not meant to be invoked concurrently with the same key.
func (sc *ServiceController) syncService(key string) error {
	glog.Flush()
	startTime := time.Now()
	glog.V(4).Infof("Started syncing Service %q (%v)", key, startTime)
	defer func() {
		glog.V(4).Infof("Finished syncing Service %q (%v)", key, time.Now().Sub(startTime))
	}()

	svcObj, exists, err := sc.serviceStore.GetByKey(key)
	if errors.IsNotFound(err) || !exists {
		glog.V(2).Infof("Service %v has been deleted", key)
		return nil
	}
	if err != nil {
		return err
	}

	svc := svcObj.(*crv1.Service)

	// TODO: probably want to add something like this once finalizers are supported for CRD
	// https://github.com/kubernetes/kubernetes/pull/51469
	//if svc.DeletionTimestamp != nil {
	//	return svc.syncStatusOnly()
	//}

	// TODO: probably need to change this when adding Blue/Green rollouts or canaries, there will probably be
	// 		 multiple deployments per Service.
	d, err := sc.getDeploymentForService(svc)
	if err != nil {
		return err
	}

	svcCopy := svc.DeepCopy()

	if d == nil {
		return sc.createServiceDeployment(svcCopy)
	}

	return sc.syncServiceWithDeployment(svcCopy, d)
}

func (sc *ServiceController) getDeploymentForService(svc *crv1.Service) (*extensions.Deployment, error) {
	// List all Deployments to find in the Service's namespace to find the Deployment the Service manages.
	deployments, err := sc.deploymentLister.Deployments(svc.Namespace).List(labels.Everything())
	if err != nil {
		return nil, err
	}

	matchingDeployments := []*extensions.Deployment{}
	svcControllerRef := metav1.NewControllerRef(svc, controllerKind)

	for _, deployment := range deployments {
		dControllerRef := metav1.GetControllerOf(deployment)

		if reflect.DeepEqual(svcControllerRef, dControllerRef) {
			matchingDeployments = append(matchingDeployments, deployment)
		}
	}

	if len(matchingDeployments) == 0 {
		return nil, nil
	}

	if len(matchingDeployments) > 1 {
		// TODO: maybe handle this better. Could choose one to make the source of truth.
		return nil, fmt.Errorf("Service %v has multiple Deployments", svc.Name)
	}

	return matchingDeployments[0], nil
}

func (sc *ServiceController) createServiceDeployment(svc *crv1.Service) error {
	svcBuild, err := sc.getSvcBuildForSvc(svc)
	if err != nil {
		return err
	}

	d, err := sc.getDeployment(svc, svcBuild)
	if err != nil {
		return err
	}

	dResp, err := sc.kubeClient.ExtensionsV1beta1().Deployments(svc.Namespace).Create(d)
	if err != nil {
		// FIXME: send warn event
		return err
	}

	glog.V(4).Infof("Created Deployment %s", dResp.Name)
	// FIXME: send normal event
	return sc.syncServiceWithDeployment(svc, dResp)
}

func (sc *ServiceController) getSvcBuildForSvc(svc *crv1.Service) (*crv1.ServiceBuild, error) {
	svcBuildKey := svc.Namespace + "/" + svc.Spec.BuildName
	svcBuildObj, exists, err := sc.serviceBuildStore.GetByKey(svcBuildKey)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, fmt.Errorf("Service %v Service %v is not in Service Store", svc.Name, svcBuildKey)
	}

	svcBuild := svcBuildObj.(*crv1.ServiceBuild)
	return svcBuild, nil
}

func (sc *ServiceController) syncServiceWithDeployment(svc *crv1.Service, d *extensions.Deployment) error {
	newStatus := calculateServiceStatus(d)

	if reflect.DeepEqual(svc.Status, newStatus) {
		return nil
	}

	svc.Status = newStatus

	err := sc.latticeResourceRestClient.Put().
		Namespace(svc.Namespace).
		Resource(crv1.ServiceResourcePlural).
		Name(svc.Name).
		Body(svc).
		Do().
		Into(nil)

	return err
}

// TODO: this is overly simplistic
func calculateServiceStatus(d *extensions.Deployment) crv1.ServiceStatus {
	progressing := false
	failure := false

	for _, condition := range d.Status.Conditions {
		switch condition.Type {
		case extensions.DeploymentProgressing:
			if condition.Status == corev1.ConditionTrue {
				progressing = true
			}
		case extensions.DeploymentReplicaFailure:
			if condition.Status == corev1.ConditionTrue {
				failure = true
			}
		}
	}

	if failure {
		return crv1.ServiceStatus{
			State: crv1.ServiceStateRolloutFailed,
		}
	}

	if progressing {
		return crv1.ServiceStatus{
			State: crv1.ServiceStateRollingOut,
		}
	}

	return crv1.ServiceStatus{
		State: crv1.ServiceStateRolloutSucceeded,
	}
}
