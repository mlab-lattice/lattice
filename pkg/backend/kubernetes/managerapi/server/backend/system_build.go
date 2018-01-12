package backend

import (
	"fmt"
	"strings"

	kubeconstants "github.com/mlab-lattice/system/pkg/backend/kubernetes/constants"
	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	kubeutil "github.com/mlab-lattice/system/pkg/backend/kubernetes/util/kubernetes"
	"github.com/mlab-lattice/system/pkg/definition"
	"github.com/mlab-lattice/system/pkg/definition/tree"
	"github.com/mlab-lattice/system/pkg/types"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/satori/go.uuid"
)

func (kb *KubernetesBackend) BuildSystem(id types.SystemID, definitionRoot tree.Node, v types.SystemVersion) (types.SystemBuildID, error) {
	systemBuild, err := systemBuild(id, definitionRoot, v)
	if err != nil {
		return "", err
	}

	namespace := kubeutil.SystemNamespace(kb.ClusterID, id)
	result, err := kb.LatticeClient.LatticeV1().SystemBuilds(namespace).Create(systemBuild)
	if err != nil {
		return "", err
	}

	return types.SystemBuildID(result.Name), err
}

func systemBuild(id types.SystemID, definitionRoot tree.Node, v types.SystemVersion) (*crv1.SystemBuild, error) {
	labels := map[string]string{
		kubeconstants.LatticeNamespaceLabel: string(id),
		kubeconstants.LabelKeySystemVersion: string(v),
	}

	services := map[tree.NodePath]crv1.SystemBuildSpecServiceInfo{}
	for path, svcNode := range definitionRoot.Services() {
		services[path] = crv1.SystemBuildSpecServiceInfo{
			Definition: svcNode.Definition().(definition.Service),
		}
	}

	sysB := &crv1.SystemBuild{
		ObjectMeta: metav1.ObjectMeta{
			Name:   uuid.NewV4().String(),
			Labels: labels,
		},
		Spec: crv1.SystemBuildSpec{
			DefinitionRoot: definitionRoot,
			Services:       services,
		},
		Status: crv1.SystemBuildStatus{
			State: crv1.SystemBuildStatePending,
		},
	}

	return sysB, nil
}

func (kb *KubernetesBackend) ListSystemBuilds(id types.SystemID) ([]types.SystemBuild, error) {
	namespace := kubeutil.SystemNamespace(kb.ClusterID, id)
	fmt.Println("listing system builds")
	buildList, err := kb.LatticeClient.LatticeV1().SystemBuilds(namespace).List(metav1.ListOptions{})
	if err != nil {
		fmt.Printf("got error: %v", err)
		return nil, err
	}

	var builds []types.SystemBuild
	for _, build := range buildList.Items {
		externalBuild, err := transformSystemBuild(&build)
		if err != nil {
			return nil, err
		}

		builds = append(builds, externalBuild)
	}

	return builds, nil
}

func (kb *KubernetesBackend) GetSystemBuild(id types.SystemID, bid types.SystemBuildID) (*types.SystemBuild, bool, error) {
	build, exists, err := kb.getInternalSystemBuild(id, bid)
	if err != nil || !exists {
		return nil, exists, err
	}

	externalBuild, err := transformSystemBuild(build)
	if err != nil {
		return nil, true, err
	}

	return &externalBuild, true, nil
}

func (kb *KubernetesBackend) getInternalSystemBuild(id types.SystemID, bid types.SystemBuildID) (*crv1.SystemBuild, bool, error) {
	namespace := kubeutil.SystemNamespace(kb.ClusterID, id)
	result, err := kb.LatticeClient.LatticeV1().SystemBuilds(namespace).Get(string(bid), metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, false, nil
		}
		return nil, false, err
	}

	// TODO: add this to the query
	if strings.Compare(result.Labels[kubeconstants.LatticeNamespaceLabel], string(id)) != 0 {
		return nil, false, nil
	}

	return result, true, nil
}

func transformSystemBuild(build *crv1.SystemBuild) (types.SystemBuild, error) {
	externalBuild := types.SystemBuild{
		ID:       types.SystemBuildID(build.Name),
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

func getSystemBuildState(state crv1.SystemBuildState) types.SystemBuildState {
	switch state {
	case crv1.SystemBuildStatePending:
		return types.SystemBuildStatePending
	case crv1.SystemBuildStateRunning:
		return types.SystemBuildStateRunning
	case crv1.SystemBuildStateSucceeded:
		return types.SystemBuildStateSucceeded
	case crv1.SystemBuildStateFailed:
		return types.SystemBuildStateFailed
	default:
		panic("unreachable")
	}
}
