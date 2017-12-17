package kubernetes

import (
	"fmt"
	"strings"

	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	latticeclientset "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/generated/clientset/versioned"
	"github.com/mlab-lattice/system/pkg/definition/tree"
	"github.com/mlab-lattice/system/pkg/types"

	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	kubeclientset "k8s.io/client-go/kubernetes"
)

func CreateNewSystem(
	clusterID types.ClusterID,
	systemID types.SystemID,
	definitionURL string,
	kubeClient kubeclientset.Interface,
	latticeClient latticeclientset.Interface,
) (*crv1.System, *corev1.Namespace, error) {
	system, namespace := NewSystem(clusterID, systemID, definitionURL)

	namespace, err := kubeClient.CoreV1().Namespaces().Create(namespace)
	if err != nil {
		if !errors.IsAlreadyExists(err) {
			return nil, nil, err
		}

		namespace, err = kubeClient.CoreV1().Namespaces().Get(namespace.Name, metav1.GetOptions{})
		if err != nil {
			return nil, nil, err
		}
	}

	system, err = latticeClient.LatticeV1().Systems(namespace.Name).Create(system)
	return system, namespace, err
}

func NewSystem(clusterID types.ClusterID, systemID types.SystemID, definitionURL string) (*crv1.System, *corev1.Namespace) {
	system := &crv1.System{
		ObjectMeta: metav1.ObjectMeta{
			Name: string(systemID),
		},
		Spec: crv1.SystemSpec{
			DefinitionURL: definitionURL,
			Services:      map[tree.NodePath]crv1.SystemSpecServiceInfo{},
		},
		Status: crv1.SystemStatus{
			State: crv1.SystemStateStable,
		},
	}

	namespace := &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Namespace",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: SystemNamespace(clusterID, systemID),
		},
	}

	return system, namespace
}

func SystemID(namespace string) (types.ClusterID, error) {
	parts := strings.Split(namespace, "-")
	if len(parts) < 3 {
		return "", fmt.Errorf("unexpected system namespace format: %v", namespace)
	}

	return types.ClusterID(strings.Join(parts[2:], "-")), nil
}
