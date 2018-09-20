package build

import (
	"fmt"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	kubeutil "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/util/kubernetes"
	"github.com/mlab-lattice/lattice/pkg/definition/component"
	"github.com/mlab-lattice/lattice/pkg/definition/component/resolver"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	definitionv1 "github.com/mlab-lattice/lattice/pkg/definition/v1"
	"github.com/mlab-lattice/lattice/pkg/util/git"
	reflectutil "github.com/mlab-lattice/lattice/pkg/util/reflect"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	// find and resolve the build's component
	path, cmpnt, ctx, version, err := c.getBuildComponent(system, build)
	if err != nil {
		return err
	}

	t, err := c.componentResolver.Resolve(cmpnt, systemID, path, ctx, resolver.DepthInfinite)
	if err != nil {
		_, err := c.updateBuildStatus(
			build,
			latticev1.BuildStateFailed,
			fmt.Sprintf("error resolving system: %v", err),
			nil,
			nil,
			&path,
			&version,
			nil,
			nil,
			nil,
			nil,
		)
		return err
	}

	// ensure that the component is a system if it's at the root
	if path.IsRoot() {
		root, ok := t.Get(tree.RootPath())
		if !ok {
			_, err := c.updateBuildStatus(
				build,
				latticev1.BuildStateFailed,
				"system does not have root",
				nil,
				nil,
				&path,
				&version,
				nil,
				nil,
				nil,
				nil,
			)
			return err
		}

		_, ok = root.Component.(*definitionv1.System)
		if !ok {
			_, err := c.updateBuildStatus(
				build,
				latticev1.BuildStateFailed,
				"root component must be a system",
				nil,
				t,
				&path,
				&version,
				nil,
				nil,
				nil,
				nil,
			)
			return err
		}
	}

	now := metav1.Now()
	_, err = c.updateBuildStatus(
		build,
		latticev1.BuildStateAccepted,
		"",
		nil,
		t,
		&path,
		&version,
		&now,
		nil,
		nil,
		nil,
	)
	return err
}

func (c *Controller) getBuildComponent(
	system *latticev1.System,
	build *latticev1.Build,
) (
	tree.Path, component.Interface,
	*git.CommitReference,
	v1.Version,
	error,
) {
	// if the build is a version build, return a reference pointing
	// at the version's tag on the system's definition repo
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

		return tree.RootPath(), ref, nil, *build.Spec.Version, nil
	}

	// if it's a path build, first check to make sure that the system
	// currently has a deployed definition
	path := *build.Spec.Path
	if system.Spec.Definition == nil {
		_, err := c.updateBuildStatus(
			build,
			latticev1.BuildStateFailed,
			fmt.Sprintf("system %v does not have any components, cannot build the system based off a path", system.Name),
			nil,
			nil,
			&path,
			nil,
			nil,
			nil,
			nil,
			nil,
		)
		return "", nil, nil, "", err
	}

	version, ok := system.DefinitionVersionLabel()
	if !ok {
		version = v1.Version("unknown")
	}

	// if we're just rebuilding the whole system, we can exit early for this simple case
	if path == tree.RootPath() {
		info, ok := system.Spec.Definition.Get(path)
		if !ok {
			_, err := c.updateBuildStatus(
				build,
				latticev1.BuildStateFailed,
				fmt.Sprintf("system %v does not contain %v", system.Name, path.String()),
				nil,
				nil,
				&path,
				&version,
				nil,
				nil,
				nil,
				nil,
			)
			return "", nil, nil, "", err
		}

		return path, info.Component, info.Commit, version, nil
	}

	// otherwise, if the path is not the root, get the path's parent
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
			&path,
			&version,
			nil,
			nil,
			nil,
			nil,
		)
		return "", nil, nil, version, err
	}

	// since the parent is necessarily an internal node, ensure that it is a system
	s, ok := parentInfo.Component.(*definitionv1.System)
	if !ok {
		_, err := c.updateBuildStatus(
			build,
			latticev1.BuildStateFailed,
			fmt.Sprintf("system %v internal node %v is not a system", system.Name, parent.String()),
			nil,
			nil,
			&path,
			&version,
			nil,
			nil,
			nil,
			nil,
		)
		return "", nil, nil, "", err
	}

	// get the path's component from its parent system
	// this may end up being a reference or an inlined component
	// if it is a reference, it and all its descendant references
	// will end up being fully re-resolved, allowing for new version tags
	// or branch commits, etc if it is inlined, all its descendant will end up being re-resolved
	cmpnt, ok := s.Components[name]
	if !ok {
		_, err := c.updateBuildStatus(
			build,
			latticev1.BuildStateFailed,
			fmt.Sprintf("system %v does not contain %v", system.Name, path.String()),
			nil,
			nil,
			&path,
			&version,
			nil,
			nil,
			nil,
			nil,
		)
		return "", nil, nil, "", err
	}

	return path, cmpnt, parentInfo.Commit, version, nil
}
