package build

import (
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	kubeutil "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/util/kubernetes"
	"github.com/mlab-lattice/lattice/pkg/definition/component/resolver"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	definitionv1 "github.com/mlab-lattice/lattice/pkg/definition/v1"
)

func (c *Controller) syncPendingBuild(build *latticev1.Build) error {
	systemID, err := kubeutil.SystemID(c.namespacePrefix, build.Namespace)
	if err != nil {
		return err
	}

	system, err := c.systemLister.Systems(kubeutil.InternalNamespace(c.namespacePrefix)).Get(string(systemID))
	if err != nil {
		return err
	}

	tag := string(build.Spec.Version)
	ref := &definitionv1.Reference{
		GitRepository: &definitionv1.GitRepositoryReference{
			GitRepository: &definitionv1.GitRepository{
				URL: system.Spec.DefinitionURL,
				Tag: &tag,
			},
		},
	}

	t, err := c.componentResolver.Resolve(ref, systemID, tree.RootPath(), nil, resolver.DepthInfinite)
	if err != nil {
		return err
	}

	_, err = c.updateBuildStatus(
		build,
		latticev1.BuildStateAccepted,
		t,
		"",
		nil,
		nil,
		nil,
		nil,
	)
	return err
}
