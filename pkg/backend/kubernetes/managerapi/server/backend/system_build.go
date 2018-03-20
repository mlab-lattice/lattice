package backend

import (
	"fmt"
	"strings"

	kubeconstants "github.com/mlab-lattice/system/pkg/backend/kubernetes/constants"
	latticev1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	kubeutil "github.com/mlab-lattice/system/pkg/backend/kubernetes/util/kubernetes"
	"github.com/mlab-lattice/system/pkg/definition"
	"github.com/mlab-lattice/system/pkg/definition/tree"
	"github.com/mlab-lattice/system/pkg/types"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/satori/go.uuid"
)

func (kb *KubernetesBackend) BuildSystem(
	systemID types.SystemID,
	definitionRoot tree.Node,
	version types.SystemVersion,
) (types.BuildID, error) {
	systemBuild, err := systemBuild(systemID, definitionRoot, version)
	if err != nil {
		return "", err
	}

	namespace := kubeutil.SystemNamespace(kb.clusterID, systemID)
	result, err := kb.latticeClient.LatticeV1().SystemBuilds(namespace).Create(systemBuild)
	if err != nil {
		return "", err
	}

	return types.BuildID(result.Name), err
}

func systemBuild(
	systemID types.SystemID,
	definitionRoot tree.Node,
	version types.SystemVersion,
) (*latticev1.SystemBuild, error) {
	labels := map[string]string{
		kubeconstants.LatticeNamespaceLabel: string(systemID),
		kubeconstants.LabelKeySystemVersion: string(version),
	}

	services := map[tree.NodePath]latticev1.SystemBuildSpecServiceInfo{}
	for path, svcNode := range definitionRoot.Services() {
		services[path] = latticev1.SystemBuildSpecServiceInfo{
			Definition: svcNode.Definition().(definition.Service),
		}
	}

	sysB := &latticev1.SystemBuild{
		ObjectMeta: metav1.ObjectMeta{
			Name:   uuid.NewV4().String(),
			Labels: labels,
		},
		Spec: latticev1.SystemBuildSpec{
			DefinitionRoot: definitionRoot,
			Services:       services,
		},
		Status: latticev1.SystemBuildStatus{
			State: latticev1.SystemBuildStatePending,
		},
	}

	return sysB, nil
}

func (kb *KubernetesBackend) ListSystemBuilds(systemID types.SystemID) ([]types.SystemBuild, error) {
	namespace := kubeutil.SystemNamespace(kb.clusterID, systemID)
	fmt.Println("listing system builds")
	buildList, err := kb.latticeClient.LatticeV1().SystemBuilds(namespace).List(metav1.ListOptions{})
	if err != nil {
		fmt.Printf("got error: %v", err)
		return nil, err
	}

	builds := make([]types.SystemBuild, 0, len(buildList.Items))
	for _, build := range buildList.Items {
		externalBuild, err := transformSystemBuild(&build)
		if err != nil {
			return nil, err
		}

		builds = append(builds, externalBuild)
	}

	return builds, nil
}

func (kb *KubernetesBackend) GetSystemBuild(
	systemID types.SystemID,
	buildID types.BuildID,
) (*types.SystemBuild, bool, error) {
	build, exists, err := kb.getInternalSystemBuild(systemID, buildID)
	if err != nil || !exists {
		return nil, exists, err
	}

	externalBuild, err := transformSystemBuild(build)
	if err != nil {
		return nil, true, err
	}

	return &externalBuild, true, nil
}

func (kb *KubernetesBackend) getInternalSystemBuild(
	systemID types.SystemID,
	buildID types.BuildID,
) (*latticev1.SystemBuild, bool, error) {
	namespace := kubeutil.SystemNamespace(kb.clusterID, systemID)
	result, err := kb.latticeClient.LatticeV1().SystemBuilds(namespace).Get(string(buildID), metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, false, nil
		}
		return nil, false, err
	}

	// TODO: add this to the query
	if strings.Compare(result.Labels[kubeconstants.LatticeNamespaceLabel], string(systemID)) != 0 {
		return nil, false, nil
	}

	return result, true, nil
}

func transformSystemBuild(build *latticev1.SystemBuild) (types.SystemBuild, error) {
	externalBuild := types.SystemBuild{
		ID:       types.BuildID(build.Name),
		State:    getSystemBuildState(build.Status.State),
		Version:  types.SystemVersion(build.Labels[kubeconstants.LabelKeySystemVersion]),
		Services: map[tree.NodePath]types.ServiceBuild{},
	}

	for service, serviceBuildName := range build.Status.ServiceBuilds {
		serviceBuildStatus, ok := build.Status.ServiceBuildStatuses[serviceBuildName]
		if !ok {
			err := fmt.Errorf(
				"System build %v/%v has ServiceBuild %v but no Status for it",
				build.Namespace,
				build.Name,
				serviceBuildName,
			)
			return types.SystemBuild{}, err
		}

		externalServiceBuild, err := transformServiceBuild(build.Namespace, serviceBuildName, &serviceBuildStatus)
		if err != nil {
			return types.SystemBuild{}, err
		}

		externalBuild.Services[service] = externalServiceBuild
	}

	return externalBuild, nil
}

func getSystemBuildState(state latticev1.SystemBuildState) types.BuildState {
	switch state {
	case latticev1.SystemBuildStatePending:
		return types.BuildStatePending
	case latticev1.SystemBuildStateRunning:
		return types.BuildStateRunning
	case latticev1.SystemBuildStateSucceeded:
		return types.BuildStateSucceeded
	case latticev1.SystemBuildStateFailed:
		return types.BuildStateFailed
	default:
		panic("unreachable")
	}
}
