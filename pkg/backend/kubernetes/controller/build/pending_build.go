package build

import (
	"fmt"
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	kubeutil "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/util/kubernetes"
	"github.com/mlab-lattice/lattice/pkg/definition/resolver"
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

	rr, err := c.componentResolver.ResolveReference(systemID, tree.RootPath(), nil, ref, resolver.DepthInfinite)
	if err != nil {
		return err
	}

	root, err := definitionv1.NewNode(rr.Component, "", nil)
	if err != nil {
		return err
	}

	systemNode, ok := root.(*definitionv1.SystemNode)
	if !ok {
		return fmt.Errorf("system resolved for %v is not a system", build.Description(c.namespacePrefix))
	}

	_, err = c.updateBuildStatus(
		build,
		latticev1.BuildStateAccepted,
		systemNode,
		rr.Info,
		"",
		nil,
		nil,
		nil,
		nil,
		nil,
	)
	return err
}
