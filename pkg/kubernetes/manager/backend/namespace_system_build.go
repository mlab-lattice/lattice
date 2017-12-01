package backend

import (
	"strings"

	coreconstants "github.com/mlab-lattice/core/pkg/constants"
	systemdefinition "github.com/mlab-lattice/core/pkg/system/definition"
	systemtree "github.com/mlab-lattice/core/pkg/system/tree"
	coretypes "github.com/mlab-lattice/core/pkg/types"

	"github.com/mlab-lattice/system/pkg/kubernetes/constants"
	crv1 "github.com/mlab-lattice/system/pkg/kubernetes/customresource/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/satori/go.uuid"
)

func (kb *KubernetesBackend) BuildSystem(ln coretypes.LatticeNamespace, definitionRoot systemtree.Node, v coretypes.SystemVersion) (coretypes.SystemBuildID, error) {
	systemBuild, err := getNewSystemBuild(ln, definitionRoot, v)
	if err != nil {
		return "", err
	}

	result := &crv1.SystemBuild{}
	err = kb.LatticeResourceClient.Post().
		Namespace(constants.NamespaceLatticeInternal).
		Resource(crv1.SystemBuildResourcePlural).
		Body(systemBuild).
		Do().
		Into(result)

	return coretypes.SystemBuildID(result.Name), err
}

func getNewSystemBuild(ln coretypes.LatticeNamespace, definitionRoot systemtree.Node, v coretypes.SystemVersion) (*crv1.SystemBuild, error) {
	labels := map[string]string{
		constants.LatticeNamespaceLabel: string(ln),
		crv1.SystemVersionLabelKey:      string(v),
	}

	services := map[systemtree.NodePath]crv1.SystemBuildServicesInfo{}
	for path, svcNode := range definitionRoot.Services() {
		services[path] = crv1.SystemBuildServicesInfo{
			Definition: *(svcNode.Definition().(*systemdefinition.Service)),
		}
	}

	sysB := &crv1.SystemBuild{
		ObjectMeta: metav1.ObjectMeta{
			Name:   uuid.NewV4().String(),
			Labels: labels,
		},
		Spec: crv1.SystemBuildSpec{
			LatticeNamespace: ln,
			DefinitionRoot:   definitionRoot,
			Services:         services,
		},
		Status: crv1.SystemBuildStatus{
			State: crv1.SystemBuildStatePending,
		},
	}

	return sysB, nil
}

func (kb *KubernetesBackend) ListSystemBuilds(ln coretypes.LatticeNamespace) ([]coretypes.SystemBuild, error) {
	result := &crv1.SystemBuildList{}
	err := kb.LatticeResourceClient.Get().
		Namespace(constants.NamespaceLatticeInternal).
		Resource(crv1.SystemBuildResourcePlural).
		Do().
		Into(result)

	if err != nil {
		return nil, err
	}

	builds := []coretypes.SystemBuild{}
	for _, build := range result.Items {
		// TODO: add this to the query
		if strings.Compare(build.Labels[constants.LatticeNamespaceLabel], string(ln)) != 0 {
			continue
		}

		builds = append(builds, transformSystemBuild(&build))
	}

	return builds, nil
}

func (kb *KubernetesBackend) GetSystemBuild(ln coretypes.LatticeNamespace, bid coretypes.SystemBuildID) (*coretypes.SystemBuild, bool, error) {
	build, exists, err := kb.getInternalSystemBuild(ln, bid)
	if err != nil || !exists {
		return nil, exists, err
	}

	coreBuild := transformSystemBuild(build)
	return &coreBuild, true, nil
}

func (kb *KubernetesBackend) getInternalSystemBuild(ln coretypes.LatticeNamespace, bid coretypes.SystemBuildID) (*crv1.SystemBuild, bool, error) {
	result := &crv1.SystemBuild{}
	err := kb.LatticeResourceClient.Get().
		Namespace(constants.NamespaceLatticeInternal).
		Resource(crv1.SystemBuildResourcePlural).
		Name(string(bid)).
		Do().
		Into(result)

	if err != nil {
		if errors.IsNotFound(err) {
			return nil, false, nil
		}
		return nil, false, err
	}

	// TODO: add this to the query
	if strings.Compare(result.Labels[constants.LatticeNamespaceLabel], string(ln)) != 0 {
		return nil, false, nil
	}

	return result, true, nil
}

func transformSystemBuild(build *crv1.SystemBuild) coretypes.SystemBuild {
	sysb := coretypes.SystemBuild{
		ID:            coretypes.SystemBuildID(build.Name),
		State:         getSystemBuildState(build.Status.State),
		Version:       coretypes.SystemVersion(build.Labels[crv1.SystemVersionLabelKey]),
		ServiceBuilds: map[systemtree.NodePath]*coretypes.ServiceBuild{},
	}

	for service, svcbInfo := range build.Spec.Services {
		if svcbInfo.BuildName == nil {
			sysb.ServiceBuilds[service] = nil
			continue
		}
		id := coretypes.ServiceBuildID(*svcbInfo.BuildName)

		state := coreconstants.ServiceBuildStatePending
		if svcbInfo.BuildState != nil {
			state = getServiceBuildState(*svcbInfo.BuildState)
		}

		svcb := &coretypes.ServiceBuild{
			ID:              id,
			State:           state,
			ComponentBuilds: map[string]*coretypes.ComponentBuild{},
		}

		for component, cbInfo := range svcbInfo.Components {
			if cbInfo.BuildName == nil {
				svcb.ComponentBuilds[component] = nil
				continue
			}
			id := coretypes.ComponentBuildID(*cbInfo.BuildName)

			state := coreconstants.ComponentBuildStatePending
			if cbInfo.BuildState != nil {
				state = getComponentBuildState(*cbInfo.BuildState)
			}

			var failureMessage *string
			if cbInfo.FailureInfo != nil {
				failMessage := getComponentBuildFailureMessage(*cbInfo.FailureInfo)
				failureMessage = &failMessage
			}

			svcb.ComponentBuilds[component] = &coretypes.ComponentBuild{
				ID:                id,
				State:             state,
				LastObservedPhase: cbInfo.LastObservedPhase,
				FailureMessage:    failureMessage,
			}
		}

		sysb.ServiceBuilds[service] = svcb
	}

	return sysb
}

func getSystemBuildState(state crv1.SystemBuildState) coretypes.SystemBuildState {
	switch state {
	case crv1.SystemBuildStatePending:
		return coreconstants.SystemBuildStatePending
	case crv1.SystemBuildStateRunning:
		return coreconstants.SystemBuildStateRunning
	case crv1.SystemBuildStateSucceeded:
		return coreconstants.SystemBuildStateSucceeded
	case crv1.SystemBuildStateFailed:
		return coreconstants.SystemBuildStateFailed
	default:
		panic("unreachable")
	}
}
