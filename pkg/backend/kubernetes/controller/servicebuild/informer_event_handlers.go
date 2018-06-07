package servicebuild

import (
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"

	"github.com/deckarep/golang-set"
	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/labels"
)

func (c *Controller) handleServiceBuildAdd(obj interface{}) {
	build := obj.(*latticev1.ServiceBuild)

	if build.DeletionTimestamp != nil {
		c.handleServiceBuildDelete(build)
		return
	}

	c.handleServiceBuildEvent(build, "added")
}

func (c *Controller) handleServiceBuildUpdate(old, cur interface{}) {
	build := cur.(*latticev1.ServiceBuild)
	c.handleServiceBuildEvent(build, "updated")
}

func (c *Controller) handleServiceBuildDelete(obj interface{}) {
	build := obj.(*latticev1.ServiceBuild)
	c.handleServiceBuildEvent(build, "deleted")
}

func (c *Controller) handleServiceBuildEvent(build *latticev1.ServiceBuild, verb string) {
	glog.V(4).Infof("%s %s", build.Description(c.namespacePrefix), verb)
	c.enqueue(build)
}

func (c *Controller) handleComponentBuildAdd(obj interface{}) {
	componentBuild := obj.(*latticev1.ContainerBuild)

	if componentBuild.DeletionTimestamp != nil {
		// only orphaned component builds should be deleted
		return
	}

	c.handleComponentBuildEvent(componentBuild, "added")
}

// handleComponentBuildUpdate enqueues any ContainerBuilds which may be interested in it when
// a Definition is updated.
func (c *Controller) handleComponentBuildUpdate(old, cur interface{}) {
	componentBuild := cur.(*latticev1.ContainerBuild)
	c.handleComponentBuildEvent(componentBuild, "updated")
}

func (c *Controller) handleComponentBuildEvent(componentBuild *latticev1.ContainerBuild, verb string) {
	glog.V(4).Infof("%s %s", componentBuild.Description(c.namespacePrefix), verb)

	serviceBuilds, err := c.owningServiceBuilds(componentBuild)
	if err != nil {
		// FIXME: send error event?
		return
	}

	for _, serviceBuild := range serviceBuilds {
		c.enqueue(&serviceBuild)
	}
}

func (c *Controller) owningServiceBuilds(componentBuild *latticev1.ContainerBuild) ([]latticev1.ServiceBuild, error) {
	owningBuilds := mapset.NewSet()
	for _, owner := range componentBuild.OwnerReferences {
		// not a lattice.mlab.com owner (probably shouldn't happen)
		if owner.APIVersion != latticev1.SchemeGroupVersion.String() {
			continue
		}

		// not a service build owner (probably shouldn't happen)
		if owner.Kind != latticev1.ServiceBuildKind.Kind {
			continue
		}

		owningBuilds.Add(owner.UID)
	}

	builds, err := c.serviceBuildLister.List(labels.Everything())
	if err != nil {
		return nil, err
	}

	var matchingBuilds []latticev1.ServiceBuild
	for _, build := range builds {
		if owningBuilds.Contains(build.UID) {
			matchingBuilds = append(matchingBuilds, *build)
		}
	}

	return matchingBuilds, nil
}
