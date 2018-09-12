package systemlifecycle

import (
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	kubeutil "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/util/kubernetes"

	"github.com/deckarep/golang-set"
	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/labels"
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
	namespace, err := c.kubeNamespaceLister.Get(systemNamespace)
	if err != nil {
		return
	}

	deploys, teardown := c.lifecycleActions.InProgressActions(namespace.UID)
	for _, id := range deploys {
		deploy, err := c.deployLister.Deploys(systemNamespace).Get(string(id))
		if err != nil {
			continue
		}

		c.enqueueDeploy(deploy)
	}

	if teardown != nil {
		teardown, err := c.teardownLister.Teardowns(systemNamespace).Get(string(*teardown))
		if err != nil {
			return
		}

		c.enqueueTeardown(teardown)
	}
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

	deploys, err := c.owningDeploys(build)
	if err != nil {
		return
	}

	for _, deploy := range deploys {
		c.enqueueDeploy(&deploy)
	}
}

func (c *Controller) owningDeploys(build *latticev1.Build) ([]latticev1.Deploy, error) {
	owningDeploys := mapset.NewSet()
	for _, owner := range build.OwnerReferences {
		// not a lattice.mlab.com owner (probably shouldn't happen)
		if owner.APIVersion != latticev1.SchemeGroupVersion.String() {
			continue
		}

		// not a build owner (probably shouldn't happen)
		if owner.Kind != latticev1.DeployKind.Kind {
			continue
		}

		owningDeploys.Add(owner.UID)
	}

	deploys, err := c.deployLister.List(labels.Everything())
	if err != nil {
		return nil, err
	}

	var matchingDeploys []latticev1.Deploy
	for _, deploy := range deploys {
		if owningDeploys.Contains(deploy.UID) {
			matchingDeploys = append(matchingDeploys, *deploy)
		}
	}

	return matchingDeploys, nil
}
