package systemlifecycle

import (
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	kubeutil "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/util/kubernetes"

	"github.com/golang/glog"
)

func (c *Controller) handleDeployAdd(obj interface{}) {
	deploy := obj.(*latticev1.Deploy)
	c.handleDeployEvent(deploy, "added")
}

func (c *Controller) handleDeployUpdate(old, cur interface{}) {
	deploy := cur.(*latticev1.Deploy)
	c.handleDeployEvent(deploy, "updated")
}

func (c *Controller) handleDeployEvent(deploy *latticev1.Deploy, verb string) {
	glog.V(4).Infof("%v %v", deploy.Description(c.namespacePrefix), verb)
	c.enqueueDeploy(deploy)
}

func (c *Controller) handleTeardownAdd(obj interface{}) {
	teardown := obj.(*latticev1.Teardown)
	c.handleTeardownEvent(teardown, "added")
}

func (c *Controller) handleTeardownUpdate(old, cur interface{}) {
	teardown := cur.(*latticev1.Teardown)
	c.handleTeardownEvent(teardown, "updated")
}

func (c *Controller) handleTeardownEvent(teardown *latticev1.Teardown, verb string) {
	glog.V(4).Infof("%v %v", teardown.Description(c.namespacePrefix), verb)
	c.enqueueTeardown(teardown)
}

func (c *Controller) handleSystemAdd(obj interface{}) {
	system := obj.(*latticev1.System)
	c.handleSystemEvent(system, "added")
}

func (c *Controller) handleSystemUpdate(old, cur interface{}) {
	system := cur.(*latticev1.System)
	c.handleSystemEvent(system, "updated")
}

func (c *Controller) handleSystemEvent(system *latticev1.System, verb string) {
	<-c.lifecycleActionsSynced

	glog.V(4).Infof("%v %v", system.Description(), verb)

	systemNamespace := kubeutil.SystemNamespace(c.namespacePrefix, v1.SystemID(system.Name))
	action, exists := c.getOwningAction(systemNamespace)
	if !exists {
		glog.V(4).Infof("%v has no owning actions, skipping", system.Description())
		return
	}

	if action == nil {
		// FIXME: send warn event
		return
	}

	if action.deploy != nil {
		c.enqueueDeploy(action.deploy)
		return
	}

	if action.teardown != nil {
		c.enqueueTeardown(action.teardown)
		return
	}

	// FIXME: Send warn event
}

func (c *Controller) handleBuildAdd(obj interface{}) {
	build := obj.(*latticev1.Build)
	c.handleBuildEvent(build, "updated")
}

func (c *Controller) handleBuildUpdate(old, cur interface{}) {
	build := cur.(*latticev1.Build)
	c.handleBuildEvent(build, "updated")
}

func (c *Controller) handleBuildEvent(build *latticev1.Build, verb string) {
	<-c.lifecycleActionsSynced

	glog.V(4).Infof("%v %v", build.Description(c.namespacePrefix), verb)

	action, exists := c.getOwningAction(build.Namespace)
	if !exists {
		// No ongoing action
		return
	}

	if action == nil {
		// FIXME: send warn event
		return
	}

	if action.deploy != nil {
		c.enqueueDeploy(action.deploy)
		return
	}

	// only need to update deploys on builds finishing
}
