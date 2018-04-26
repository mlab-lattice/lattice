package build

import (
	"fmt"

	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/runtime"

	"k8s.io/client-go/tools/cache"

	"github.com/deckarep/golang-set"
	"github.com/golang/glog"
)

func (c *Controller) handleBuildAdd(obj interface{}) {
	build := obj.(*latticev1.Build)

	if build.DeletionTimestamp != nil {
		c.handleBuildDelete(build)
		return
	}

	c.handleBuildEvent(build, "added")
}

func (c *Controller) handleBuildUpdate(old, cur interface{}) {
	build := cur.(*latticev1.Build)
	c.handleBuildEvent(build, "updated")
}

func (c *Controller) handleBuildDelete(obj interface{}) {
	build, ok := obj.(*latticev1.Build)

	// When a delete is dropped, the relist will notice a pod in the store not
	// in the list, leading to the insertion of a tombstone object which contains
	// the deleted key/value.
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			runtime.HandleError(fmt.Errorf("couldn't get object from tombstone %#v", obj))
			return
		}
		build, ok = tombstone.Obj.(*latticev1.Build)
		if !ok {
			runtime.HandleError(fmt.Errorf("tombstone contained object that is not a build %#v", obj))
			return
		}
	}

	c.handleBuildEvent(build, "deleted")
}

func (c *Controller) handleBuildEvent(build *latticev1.Build, verb string) {
	glog.V(4).Infof("%s %s", build.Description(c.namespacePrefix), verb)
	c.enqueue(build)
}

// handleServiceBuildAdd enqueues the System that manages a Service when the Service is created.
func (c *Controller) handleServiceBuildAdd(obj interface{}) {
	serviceBuild := obj.(*latticev1.ServiceBuild)

	if serviceBuild.DeletionTimestamp != nil {
		// only orphaned service builds should be deleted
		return
	}

	c.handleServiceBuildEvent(serviceBuild, "added")
}

// handleServiceBuildUpdate figures out what Build manages a Service when the
// Service is updated and enqueues them.
func (c *Controller) handleServiceBuildUpdate(old, cur interface{}) {
	serviceBuild := cur.(*latticev1.ServiceBuild)
	c.handleServiceBuildEvent(serviceBuild, "updated")
}

func (c *Controller) handleServiceBuildEvent(serviceBuild *latticev1.ServiceBuild, verb string) {
	glog.V(4).Infof("%s %s", serviceBuild.Description(c.namespacePrefix), verb)

	builds, err := c.owningBuilds(serviceBuild)
	if err != nil {
		// FIXME: send error event?
		return
	}

	for _, build := range builds {
		c.enqueue(&build)
	}
}

func (c *Controller) owningBuilds(serviceBuild *latticev1.ServiceBuild) ([]latticev1.Build, error) {
	owningBuilds := mapset.NewSet()
	for _, owner := range serviceBuild.OwnerReferences {
		// not a lattice.mlab.com owner (probably shouldn't happen)
		if owner.APIVersion != latticev1.SchemeGroupVersion.String() {
			continue
		}

		// not a build owner (probably shouldn't happen)
		if owner.Kind != latticev1.BuildKind.Kind {
			continue
		}

		owningBuilds.Add(owner.UID)
	}

	builds, err := c.buildLister.List(labels.Everything())
	if err != nil {
		return nil, err
	}

	var matchingBuilds []latticev1.Build
	for _, build := range builds {
		if owningBuilds.Contains(build.UID) {
			matchingBuilds = append(matchingBuilds, *build)
		}
	}

	return matchingBuilds, nil
}
