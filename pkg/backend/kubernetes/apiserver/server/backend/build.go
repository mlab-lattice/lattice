package backend

import (
	"fmt"
	"strings"

	"github.com/mlab-lattice/system/pkg/api/v1"
	kubeconstants "github.com/mlab-lattice/system/pkg/backend/kubernetes/constants"
	latticev1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	kubeutil "github.com/mlab-lattice/system/pkg/backend/kubernetes/util/kubernetes"
	"github.com/mlab-lattice/system/pkg/definition"
	"github.com/mlab-lattice/system/pkg/definition/tree"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/satori/go.uuid"
)

func (kb *KubernetesBackend) Build(
	systemID v1.SystemID,
	definitionRoot tree.Node,
	version v1.SystemVersion,
) (v1.BuildID, error) {
	systemBuild, err := systemBuild(systemID, definitionRoot, version)
	if err != nil {
		return "", err
	}

	namespace := kubeutil.SystemNamespace(kb.latticeID, systemID)
	result, err := kb.latticeClient.LatticeV1().Builds(namespace).Create(systemBuild)
	if err != nil {
		return "", err
	}

	return v1.BuildID(result.Name), err
}

func systemBuild(
	systemID v1.SystemID,
	definitionRoot tree.Node,
	version v1.SystemVersion,
) (*latticev1.Build, error) {
	labels := map[string]string{
		kubeconstants.LatticeNamespaceLabel: string(systemID),
		kubeconstants.LabelKeySystemVersion: string(version),
	}

	services := map[tree.NodePath]latticev1.BuildSpecServiceInfo{}
	for path, svcNode := range definitionRoot.Services() {
		services[path] = latticev1.BuildSpecServiceInfo{
			Definition: svcNode.Definition().(definition.Service),
		}
	}

	sysB := &latticev1.Build{
		ObjectMeta: metav1.ObjectMeta{
			Name:   uuid.NewV4().String(),
			Labels: labels,
		},
		Spec: latticev1.BuildSpec{
			DefinitionRoot: definitionRoot,
			Services:       services,
		},
		Status: latticev1.BuildStatus{
			State: latticev1.BuildStatePending,
		},
	}

	return sysB, nil
}

func (kb *KubernetesBackend) ListBuilds(systemID v1.SystemID) ([]v1.Build, error) {
	namespace := kubeutil.SystemNamespace(kb.latticeID, systemID)
	fmt.Println("listing system builds")
	buildList, err := kb.latticeClient.LatticeV1().Builds(namespace).List(metav1.ListOptions{})
	if err != nil {
		fmt.Printf("got error: %v", err)
		return nil, err
	}

	builds := make([]v1.Build, 0, len(buildList.Items))
	for _, build := range buildList.Items {
		externalBuild, err := transformSystemBuild(&build)
		if err != nil {
			return nil, err
		}

		builds = append(builds, externalBuild)
	}

	return builds, nil
}

func (kb *KubernetesBackend) GetBuild(
	systemID v1.SystemID,
	buildID v1.BuildID,
) (*v1.Build, bool, error) {
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
	systemID v1.SystemID,
	buildID v1.BuildID,
) (*latticev1.Build, bool, error) {
	namespace := kubeutil.SystemNamespace(kb.latticeID, systemID)
	result, err := kb.latticeClient.LatticeV1().Builds(namespace).Get(string(buildID), metav1.GetOptions{})
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

func transformSystemBuild(build *latticev1.Build) (v1.Build, error) {
	externalBuild := v1.Build{
		ID:       v1.BuildID(build.Name),
		State:    getSystemBuildState(build.Status.State),
		Version:  v1.SystemVersion(build.Labels[kubeconstants.LabelKeySystemVersion]),
		Services: map[tree.NodePath]v1.ServiceBuild{},
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
			return v1.Build{}, err
		}

		externalServiceBuild, err := transformServiceBuild(build.Namespace, serviceBuildName, &serviceBuildStatus)
		if err != nil {
			return v1.Build{}, err
		}

		externalBuild.Services[service] = externalServiceBuild
	}

	return externalBuild, nil
}

func getSystemBuildState(state latticev1.BuildState) v1.BuildState {
	switch state {
	case latticev1.BuildStatePending:
		return v1.BuildStatePending
	case latticev1.BuildStateRunning:
		return v1.BuildStateRunning
	case latticev1.BuildStateSucceeded:
		return v1.BuildStateSucceeded
	case latticev1.BuildStateFailed:
		return v1.BuildStateFailed
	default:
		panic("unreachable")
	}
}
