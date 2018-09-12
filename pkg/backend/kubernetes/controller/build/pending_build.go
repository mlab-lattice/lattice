package build

import (
	"fmt"
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	kubeutil "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/util/kubernetes"
	"github.com/mlab-lattice/lattice/pkg/definition/component"
	"github.com/mlab-lattice/lattice/pkg/definition/component/resolver"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	definitionv1 "github.com/mlab-lattice/lattice/pkg/definition/v1"
	"github.com/mlab-lattice/lattice/pkg/util/git"
	reflectutil "github.com/mlab-lattice/lattice/pkg/util/reflect"
)

func (c *Controller) syncPendingBuild(build *latticev1.Build) error {
	err := reflectutil.ValidateUnion(&build.Spec)
	if err != nil {
		switch err.(type) {
		case *reflectutil.InvalidUnionNoFieldSetError:
			_, err := c.updateBuildStatus(
				build,
				latticev1.BuildStateFailed,
				"either path or version must be set",
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
			)
			return err

		case *reflectutil.InvalidUnionMultipleFieldSetError:
			_, err := c.updateBuildStatus(
				build,
				latticev1.BuildStateFailed,
				"only path or version can be set",
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
			)
			return err

		default:
			msg := err.Error()
			_, err := c.updateBuildStatus(
				build,
				latticev1.BuildStateFailed,
				"internal error",
				&msg,
				nil,
				nil,
				nil,
				nil,
				nil,
			)
			return err
		}
	}

	systemID, err := kubeutil.SystemID(c.namespacePrefix, build.Namespace)
	if err != nil {
		return err
	}

	system, err := c.systemLister.Systems(kubeutil.InternalNamespace(c.namespacePrefix)).Get(string(systemID))
	if err != nil {
		return err
	}

	path, cmpnt, ctx, err := c.getBuildComponent(system, build)
	if err != nil {
		return err
	}

	t, err := c.componentResolver.Resolve(cmpnt, systemID, path, ctx, resolver.DepthInfinite)
	if err != nil {
		return err
	}

	_, err = c.updateBuildStatus(
		build,
		latticev1.BuildStateAccepted,
		"",
		nil,
		t,
		nil,
		nil,
		nil,
		nil,
	)
	return err
}

func (c *Controller) getBuildComponent(
	system *latticev1.System,
	build *latticev1.Build,
) (tree.Path, component.Interface, *git.CommitReference, error) {
	if build.Spec.Path == nil {
		tag := string(*build.Spec.Version)
		ref := &definitionv1.Reference{
			GitRepository: &definitionv1.GitRepositoryReference{
				GitRepository: &definitionv1.GitRepository{
					URL: system.Spec.DefinitionURL,
					Tag: &tag,
				},
			},
		}

		return tree.RootPath(), ref, nil, nil
	}

	path := *build.Spec.Path
	if system.Spec.Definition == nil {
		_, err := c.updateBuildStatus(
			build,
			latticev1.BuildStateFailed,
			fmt.Sprintf("system %v does not have any components, cannot build a path", system.Name),
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
		)
		return "", nil, nil, err
	}

	if path == tree.RootPath() {
		info, ok := system.Spec.Definition.Get(path)
		if !ok {
			_, err := c.updateBuildStatus(
				build,
				latticev1.BuildStateFailed,
				fmt.Sprintf("system %v does not contain %v", system.Name, path.String()),
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
			)
			return "", nil, nil, err
		}

		return path, info.Component, info.Commit, nil
	}

	name, _ := path.Leaf()
	parent, _ := path.Parent()
	parentInfo, ok := system.Spec.Definition.Get(parent)
	if !ok {
		_, err := c.updateBuildStatus(
			build,
			latticev1.BuildStateFailed,
			fmt.Sprintf("system %v does not contain %v", system.Name, path.String()),
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
		)
		return "", nil, nil, err
	}

	s, ok := parentInfo.Component.(*definitionv1.System)
	if !ok {
		_, err := c.updateBuildStatus(
			build,
			latticev1.BuildStateFailed,
			fmt.Sprintf("system %v internal node %v is not a system", system.Name, parent.String()),
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
		)
		return "", nil, nil, err
	}

	cmpnt, ok := s.Components[name]
	if !ok {
		_, err := c.updateBuildStatus(
			build,
			latticev1.BuildStateFailed,
			fmt.Sprintf("system %v does not contain %v", system.Name, path.String()),
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
		)
		return "", nil, nil, err
	}

	return path, cmpnt, parentInfo.Commit, nil
}
