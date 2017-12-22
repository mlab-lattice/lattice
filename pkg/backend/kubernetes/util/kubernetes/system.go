package kubernetes

import (
	"fmt"
	"strings"

	kubeconstants "github.com/mlab-lattice/system/pkg/backend/kubernetes/constants"
	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	latticeclientset "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/generated/clientset/versioned"
	"github.com/mlab-lattice/system/pkg/definition/tree"
	"github.com/mlab-lattice/system/pkg/types"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	kubeclientset "k8s.io/client-go/kubernetes"

	"github.com/ghodss/yaml"
)

type SystemResources struct {
	System          *crv1.System
	Namespace       *corev1.Namespace
	ServiceAccounts []corev1.ServiceAccount
	RoleBindings    []rbacv1.RoleBinding
}

func (r *SystemResources) String() (string, error) {
	header := "---\n"
	output := ""

	data, err := yaml.Marshal(r.Namespace)
	if err != nil {
		return "", err
	}

	output += fmt.Sprintf("%v%v", header, string(data))

	for _, serviceAccount := range r.ServiceAccounts {
		data, err := yaml.Marshal(serviceAccount)
		if err != nil {
			return "", err
		}

		output += fmt.Sprintf("%v%v", header, string(data))
	}

	for _, roleBinding := range r.RoleBindings {
		data, err := yaml.Marshal(roleBinding)
		if err != nil {
			return "", err
		}

		output += fmt.Sprintf("%v%v", header, string(data))
	}

	data, err = yaml.Marshal(r.Namespace)
	if err != nil {
		return "", err
	}

	output += fmt.Sprintf("%v%v", header, string(data))

	return output, nil
}

func CreateNewSystem(
	clusterID types.ClusterID,
	systemID types.SystemID,
	definitionURL string,
	kubeClient kubeclientset.Interface,
	latticeClient latticeclientset.Interface,
) (*SystemResources, error) {
	resources := NewSystem(clusterID, systemID, definitionURL)
	return CreateNewSystemResources(resources, kubeClient, latticeClient)
}

func CreateNewSystemResources(resources *SystemResources, kubeClient kubeclientset.Interface, latticeClient latticeclientset.Interface) (*SystemResources, error) {
	namespace, err := kubeClient.CoreV1().Namespaces().Create(resources.Namespace)
	if err != nil {
		if !errors.IsAlreadyExists(err) {
			return nil, err
		}

		namespace, err = kubeClient.CoreV1().Namespaces().Get(namespace.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
	}

	var serviceAccounts []corev1.ServiceAccount
	for _, sa := range resources.ServiceAccounts {
		sa, err := kubeClient.CoreV1().ServiceAccounts(namespace.Name).Create(&sa)
		if err != nil {
			if !errors.IsAlreadyExists(err) {
				return nil, err
			}

			sa, err = kubeClient.CoreV1().ServiceAccounts(namespace.Name).Get(sa.Name, metav1.GetOptions{})
			if err != nil {
				return nil, err
			}
		}

		serviceAccounts = append(serviceAccounts, *sa)
	}

	var roleBindings []rbacv1.RoleBinding
	for _, roleBinding := range resources.RoleBindings {
		roleBinding, err := kubeClient.RbacV1().RoleBindings(namespace.Name).Create(&roleBinding)
		if err != nil {
			if !errors.IsAlreadyExists(err) {
				return nil, err
			}

			roleBinding, err = kubeClient.RbacV1().RoleBindings(namespace.Name).Get(roleBinding.Name, metav1.GetOptions{})
			if err != nil {
				return nil, err
			}
		}

		roleBindings = append(roleBindings, *roleBinding)
	}

	system, err := latticeClient.LatticeV1().Systems(namespace.Name).Create(resources.System)
	if err != nil {
		return nil, err
	}

	resources = &SystemResources{
		System:          system,
		Namespace:       namespace,
		ServiceAccounts: serviceAccounts,
		RoleBindings:    roleBindings,
	}
	return resources, err
}

func NewSystem(
	clusterID types.ClusterID,
	systemID types.SystemID,
	definitionURL string,
) *SystemResources {
	system := &crv1.System{
		// Include TypeMeta so if this is a dry run it will be printed out
		TypeMeta: metav1.TypeMeta{
			Kind:       "System",
			APIVersion: crv1.GroupName + "/v1",
		},
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
		// Include TypeMeta so if this is a dry run it will be printed out
		TypeMeta: metav1.TypeMeta{
			Kind:       "Namespace",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: SystemNamespace(clusterID, systemID),
		},
	}

	componentBuilderSA := corev1.ServiceAccount{
		// Include TypeMeta so if this is a dry run it will be printed out
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceAccount",
			APIVersion: rbacv1.GroupName + "/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      kubeconstants.ServiceAccountComponentBuilder,
			Namespace: namespace.Name,
		},
	}

	componentBuilderRB := rbacv1.RoleBinding{
		// Include TypeMeta so if this is a dry run it will be printed out
		TypeMeta: metav1.TypeMeta{
			Kind:       "RoleBinding",
			APIVersion: rbacv1.GroupName + "/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      kubeconstants.InternalComponentComponentBuilder,
			Namespace: componentBuilderSA.Namespace,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      rbacv1.ServiceAccountKind,
				Name:      componentBuilderSA.Name,
				Namespace: componentBuilderSA.Namespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "ClusterRole",
			Name:     kubeconstants.InternalComponentComponentBuilder,
		},
	}

	return &SystemResources{
		System:          system,
		Namespace:       namespace,
		ServiceAccounts: []corev1.ServiceAccount{componentBuilderSA},
		RoleBindings:    []rbacv1.RoleBinding{componentBuilderRB},
	}
}

func SystemID(namespace string) (types.ClusterID, error) {
	parts := strings.Split(namespace, "-")
	if len(parts) < 3 {
		return "", fmt.Errorf("unexpected system namespace format: %v", namespace)
	}

	return types.ClusterID(strings.Join(parts[2:], "-")), nil
}
