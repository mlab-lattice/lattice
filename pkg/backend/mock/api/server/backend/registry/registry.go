package registry

import (
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/component/resolver"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	definitionv1 "github.com/mlab-lattice/lattice/pkg/definition/v1"
	"github.com/satori/go.uuid"
	"sync"
)

func New() *Registry {
	return &Registry{Systems: make(map[v1.SystemID]*SystemRecord)}
}

type Registry struct {
	sync.Mutex

	Systems map[v1.SystemID]*SystemRecord
}

type SystemRecord struct {
	System     *v1.System
	Definition *resolver.ResolutionTree

	Builds map[v1.BuildID]*BuildInfo

	Deploys map[v1.DeployID]*v1.Deploy

	Jobs map[v1.JobID]*v1.Job

	NodePools map[tree.PathSubcomponent]*v1.NodePool

	Secrets map[tree.PathSubcomponent]*v1.Secret

	Services     map[v1.ServiceID]*ServiceInfo
	ServicePaths map[tree.Path]v1.ServiceID

	Teardowns map[v1.TeardownID]*v1.Teardown
}

type BuildInfo struct {
	Build      *v1.Build
	Definition *resolver.ResolutionTree
}

type ServiceInfo struct {
	Service    *v1.Service
	Definition *definitionv1.Service
}

// both the backend and the controller need the ability to create builds, so it is
// implemented here.
func (r *Registry) CreateBuild(p *tree.Path, v *v1.Version, record *SystemRecord) *v1.Build {
	build := &v1.Build{
		ID:      v1.BuildID(uuid.NewV4().String()),
		State:   v1.BuildStatePending,
		Path:    p,
		Version: v,
	}
	record.Builds[build.ID] = &BuildInfo{
		Build: build,
	}

	return build
}
