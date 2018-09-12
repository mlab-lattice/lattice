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

	var path tree.Path
	var cmpnt component.Interface
	var ctx *git.CommitReference
	if build.Spec.Path == nil {
		path = tree.RootPath()
		tag := string(*build.Spec.Version)
		cmpnt = &definitionv1.Reference{
			GitRepository: &definitionv1.GitRepositoryReference{
				GitRepository: &definitionv1.GitRepository{
					URL: system.Spec.DefinitionURL,
					Tag: &tag,
				},
			},
		}
	} else {
		path = *build.Spec.Path
		if system.Spec.Definition == nil {
			_, err = c.updateBuildStatus(
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
			return err
		}

		info, ok := system.Spec.Definition.Get(path)
		if !ok {
			_, err = c.updateBuildStatus(
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
			return err
		}

		cmpnt = info.Component
		ctx = info.Commit
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
