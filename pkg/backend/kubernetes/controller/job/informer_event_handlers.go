package job

import (
	"fmt"

	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/cloudprovider"
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/servicemesh"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"

	"k8s.io/client-go/tools/cache"

	"github.com/golang/glog"
)

func (c *Controller) handleConfigAdd(obj interface{}) {
	config := obj.(*latticev1.Config)
	err := c.handleConfigEvent(config, "added")
	if err != nil {
		return
	}

	c.configLock.Lock()
	defer c.configLock.Unlock()
	if !c.configSet {
		c.configSet = true
		close(c.configSetChan)
	}
}

func (c *Controller) handleConfigUpdate(old, cur interface{}) {
	config := cur.(*latticev1.Config)
	c.handleConfigEvent(config, "updated")
}

func (c *Controller) handleConfigEvent(config *latticev1.Config, verb string) error {
	glog.V(4).Infof("config %v/%v %v", config.Namespace, config.Name, verb)

	c.configLock.Lock()
	defer c.configLock.Unlock()
	c.config = config.DeepCopy().Spec

	err := c.newCloudProvider()
	if err != nil {
		glog.Errorf("error creating cloud provider: %v", err)
		// FIXME: what to do here?
		return err
	}

	err = c.newServiceMesh()
	if err != nil {
		glog.Errorf("error creating service mesh: %v", err)
		// FIXME: what to do here?
		return err
	}

	return nil
}

func (c *Controller) newCloudProvider() error {
	options, err := cloudprovider.OverlayConfigOptions(c.staticCloudProviderOptions, &c.config.CloudProvider)
	if err != nil {
		return err
	}

	cloudProvider, err := cloudprovider.NewCloudProvider(
		c.namespacePrefix,
		c.kubeClient,
		c.kubeInformerFactory,
		c.latticeInformerFactory,
		options,
	)
	if err != nil {
		return err
	}

	c.cloudProvider = cloudProvider
	return nil
}

func (c *Controller) newServiceMesh() error {
	options, err := servicemesh.OverlayConfigOptions(c.staticServiceMeshOptions, &c.config.ServiceMesh)
	if err != nil {
		return err
	}

	serviceMesh, err := servicemesh.NewServiceMesh(options)
	if err != nil {
		return err
	}

	c.serviceMesh = serviceMesh
	return nil
}

func (c *Controller) handleJobRunAdd(obj interface{}) {
	jobRun := obj.(*latticev1.JobRun)

	if jobRun.DeletionTimestamp != nil {
		// On a restart of the controller manager, it's possible for an object to
		// show up in a state that is already pending deletion.
		c.handleJobRunDelete(jobRun)
		return
	}

	c.handleJobRunEvent(jobRun, "added")
}

func (c *Controller) handleJobRunUpdate(old, cur interface{}) {
	jobRun := cur.(*latticev1.JobRun)
	c.handleJobRunEvent(jobRun, "updated")
}

func (c *Controller) handleJobRunDelete(obj interface{}) {
	jobRun, ok := obj.(*latticev1.JobRun)

	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			runtime.HandleError(fmt.Errorf("couldn't get object from tombstone %#v", obj))
			return
		}
		jobRun, ok = tombstone.Obj.(*latticev1.JobRun)
		if !ok {
			runtime.HandleError(fmt.Errorf("tombstone contained object that is not a job run %#v", obj))
			return
		}
	}

	c.handleJobRunEvent(jobRun, "deleted")
}

func (c *Controller) handleJobRunEvent(jobRun *latticev1.JobRun, verb string) {
	glog.V(4).Infof("%v %v", jobRun.Description(c.namespacePrefix), verb)
	c.enqueue(jobRun)
}

func (c *Controller) handleNodePoolAdd(obj interface{}) {
	nodePool := obj.(*latticev1.NodePool)

	if nodePool.DeletionTimestamp != nil {
		// On a restart of the controller manager, it's possible for an object to
		// show up in a state that is already pending deletion.
		c.handleNodePoolDelete(nodePool)
		return
	}

	c.handleNodePoolEvent(nodePool, "added")
}

func (c *Controller) handleNodePoolUpdate(old, cur interface{}) {
	nodePool := cur.(*latticev1.NodePool)
	c.handleNodePoolEvent(nodePool, "added")
}

func (c *Controller) handleNodePoolDelete(obj interface{}) {
	nodePool, ok := obj.(*latticev1.NodePool)

	// When a delete is dropped, the relist will notice a pod in the store not
	// in the list, leading to the insertion of a tombstone object which contains
	// the deleted key/value.
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			runtime.HandleError(fmt.Errorf("couldn't get object from tombstone %#v", obj))
			return
		}
		nodePool, ok = tombstone.Obj.(*latticev1.NodePool)
		if !ok {
			runtime.HandleError(fmt.Errorf("tombstone contained object that is not a node pool %#v", obj))
			return
		}
	}

	c.handleNodePoolEvent(nodePool, "deleted")
}

func (c *Controller) handleNodePoolEvent(nodePool *latticev1.NodePool, verb string) {
	glog.V(4).Infof("%v %v", nodePool.Description(c.namespacePrefix), verb)

	jobs, err := c.nodePoolJobRuns(nodePool)
	if err != nil {
		// FIXME: log/send warn event
		return
	}

	for _, service := range jobs {
		c.enqueue(&service)
	}
}

// handleKubeJobAdd enqueues the Service that manages a Deployment when the Deployment is created.
func (c *Controller) handleKubeJobAdd(obj interface{}) {
	job := obj.(*batchv1.Job)

	if job.DeletionTimestamp != nil {
		// On a restart of the controller manager, it's possible for an object to
		// show up in a state that is already pending deletion.
		c.handleDeploymentDelete(job)
		return
	}

	c.handleJobEvent(job, "added")
}

// handleDeploymentUpdate figures out what Service manages a Deployment when the Deployment
// is updated and enqueues it.
func (c *Controller) handleDeploymentUpdate(old, cur interface{}) {
	job := cur.(*batchv1.Job)
	c.handleJobEvent(job, "updated")
}

// handleDeploymentDelete enqueues the Service that manages a Deployment when
// the Deployment is deleted.
func (c *Controller) handleDeploymentDelete(obj interface{}) {
	job, ok := obj.(*batchv1.Job)

	// When a delete is dropped, the relist will notice a pod in the store not
	// in the list, leading to the insertion of a tombstone object which contains
	// the deleted key/value.
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			runtime.HandleError(fmt.Errorf("couldn't get object from tombstone %#v", obj))
			return
		}
		job, ok = tombstone.Obj.(*batchv1.Job)
		if !ok {
			runtime.HandleError(fmt.Errorf("tombstone contained object that is not a job %#v", obj))
			return
		}
	}

	c.handleJobEvent(job, "deleted")
}

func (c *Controller) handleJobEvent(job *batchv1.Job, verb string) {
	glog.V(4).Infof("job %v/%v %v", job.Namespace, job.Name, verb)

	// see if the job has a jobRun as a controller owning reference
	if controllerRef := metav1.GetControllerOf(job); controllerRef != nil {
		jobRun := c.resolveControllerRef(job.Namespace, controllerRef)

		// Not a Service Deployment.
		if jobRun == nil {
			return
		}

		c.enqueue(jobRun)
		return
	}

	// Otherwise, it's an orphan. These shouldn't exist within a lattice controlled namespace.
	// TODO: maybe log/send warn event if there's an orphan job in a lattice controlled namespace
}

func (c *Controller) handlePodDelete(obj interface{}) {
	pod, ok := obj.(*corev1.Pod)

	// When a delete is dropped, the relist will notice a pod in the store not
	// in the list, leading to the insertion of a tombstone object which contains
	// the deleted key/value.
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			runtime.HandleError(fmt.Errorf("couldn't get object from tombstone %#v", obj))
			return
		}
		pod, ok = tombstone.Obj.(*corev1.Pod)
		if !ok {
			runtime.HandleError(fmt.Errorf("tombstone contained object that is not a pod %#v", obj))
			return
		}
	}

	jobRunID, ok := pod.Labels[latticev1.JobRunIDLabelKey]
	if !ok {
		// not a jobRun pod
		return
	}

	jobRun, err := c.jobRunLister.JobRuns(pod.Namespace).Get(jobRunID)
	if err != nil {
		if errors.IsNotFound(err) {
			// jobRun doesn't exist anymore, so it doesn't care about this
			// pod's deletion
			return
		}

		runtime.HandleError(fmt.Errorf("error retrieving job run %v/%v", pod.Namespace, jobRunID))
		return
	}

	glog.V(4).Infof("pod %s/%s deleted", pod.Namespace, pod.Name)
	c.enqueue(jobRun)
}

// resolveControllerRef returns the controller referenced by a ControllerRef,
// or nil if the ControllerRef could not be resolved to a matching controller
// of the correct Kind.
func (c *Controller) resolveControllerRef(namespace string, controllerRef *metav1.OwnerReference) *latticev1.JobRun {
	// We can't look up by Name, so look up by Name and then verify Name.
	// Don't even try to look up by Name if it's the wrong Kind.
	if controllerRef.Kind != latticev1.JobRunKind.Kind {
		return nil
	}

	jobRun, err := c.jobRunLister.JobRuns(namespace).Get(controllerRef.Name)
	if err != nil {
		// FIXME(kevinrosendahl): send error?
		return nil
	}

	if jobRun.UID != controllerRef.UID {
		// The controller we found with this Name is not the same one that the
		// ControllerRef points to.
		return nil
	}
	return jobRun
}
