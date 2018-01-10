package bootstrap

import (
	"fmt"

	"github.com/mlab-lattice/system/pkg/backend/kubernetes/cloudprovider"
	kubeconstants "github.com/mlab-lattice/system/pkg/backend/kubernetes/constants"
	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	latticeclientset "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/generated/clientset/versioned"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/system/bootstrap/bootstrapper"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/servicemesh"
	kubeutil "github.com/mlab-lattice/system/pkg/backend/kubernetes/util/kubernetes"
	"github.com/mlab-lattice/system/pkg/definition/tree"
	"github.com/mlab-lattice/system/pkg/types"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	kubeclientset "k8s.io/client-go/kubernetes"
)

func Bootstrap(
	clusterID types.ClusterID,
	systemID types.SystemID,
	definitionURL string,
	serviceMesh servicemesh.Interface,
	cloudProvider cloudprovider.Interface,
	kubeClient kubeclientset.Interface,
	latticeClient latticeclientset.Interface,
) (*bootstrapper.SystemResources, error) {
	resources := GetBootstrapResources(clusterID, systemID, definitionURL, serviceMesh, cloudProvider)

	fmt.Println("seeding namespaces")
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

	fmt.Println("seeding service account")
	var serviceAccounts []*corev1.ServiceAccount
	for _, sa := range resources.ServiceAccounts {
		sa, err = kubeClient.CoreV1().ServiceAccounts(namespace.Name).Create(sa)
		if err != nil {
			if !errors.IsAlreadyExists(err) {
				return nil, err
			}

			sa, err = kubeClient.CoreV1().ServiceAccounts(namespace.Name).Get(sa.Name, metav1.GetOptions{})
			if err != nil {
				return nil, err
			}
		}

		serviceAccounts = append(serviceAccounts, sa)
	}

	fmt.Println("seeding role bindings")
	var roleBindings []*rbacv1.RoleBinding
	for _, roleBinding := range resources.RoleBindings {
		roleBinding, err = kubeClient.RbacV1().RoleBindings(namespace.Name).Create(roleBinding)
		if err != nil {
			if !errors.IsAlreadyExists(err) {
				return nil, err
			}

			roleBinding, err = kubeClient.RbacV1().RoleBindings(namespace.Name).Get(roleBinding.Name, metav1.GetOptions{})
			if err != nil {
				return nil, err
			}
		}

		roleBindings = append(roleBindings, roleBinding)
	}

	fmt.Println("seeding daemon sets")
	var daemonSets []*appsv1.DaemonSet
	for _, daemonSet := range resources.DaemonSets {
		daemonSet, err = kubeClient.AppsV1().DaemonSets(namespace.Name).Create(daemonSet)
		if err != nil {
			if !errors.IsAlreadyExists(err) {
				return nil, err
			}

			daemonSet, err = kubeClient.AppsV1().DaemonSets(namespace.Name).Get(daemonSet.Name, metav1.GetOptions{})
			if err != nil {
				return nil, err
			}
		}

		daemonSets = append(daemonSets, daemonSet)
	}

	fmt.Println("seeding system")
	system, err := latticeClient.LatticeV1().Systems(namespace.Name).Create(resources.System)
	if err != nil {
		return nil, err
	}

	resources = &bootstrapper.SystemResources{
		System:          system,
		Namespace:       namespace,
		ServiceAccounts: serviceAccounts,
		RoleBindings:    roleBindings,
		DaemonSets:      daemonSets,
	}
	return resources, err
}

func GetBootstrapResources(
	clusterID types.ClusterID,
	systemID types.SystemID,
	definitionURL string,
	serviceMesh servicemesh.Interface,
	cloudProvider cloudprovider.Interface,
) *bootstrapper.SystemResources {
	namespace := &corev1.Namespace{
		// Include TypeMeta so if this is a dry run it will be printed out
		TypeMeta: metav1.TypeMeta{
			Kind:       "Namespace",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: kubeutil.SystemNamespace(clusterID, systemID),
			Labels: map[string]string{
				kubeconstants.LabelKeyLatticeClusterID: string(clusterID),
			},
		},
	}

	componentBuilderSA := &corev1.ServiceAccount{
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

	componentBuilderRB := &rbacv1.RoleBinding{
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

	system := &crv1.System{
		// Include TypeMeta so if this is a dry run it will be printed out
		TypeMeta: metav1.TypeMeta{
			Kind:       "System",
			APIVersion: crv1.GroupName + "/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      string(systemID),
			Namespace: namespace.Name,
			Labels: map[string]string{
				kubeconstants.LabelKeyLatticeClusterID: string(clusterID),
			},
		},
		Spec: crv1.SystemSpec{
			DefinitionURL: definitionURL,
			Services:      map[tree.NodePath]crv1.SystemSpecServiceInfo{},
		},
		Status: crv1.SystemStatus{
			State: crv1.SystemStateStable,
		},
	}

	resources := &bootstrapper.SystemResources{
		System:          system,
		Namespace:       namespace,
		ServiceAccounts: []*corev1.ServiceAccount{componentBuilderSA},
		RoleBindings:    []*rbacv1.RoleBinding{componentBuilderRB},
	}

	serviceMesh.BootstrapSystemResources(resources)
	cloudProvider.BootstrapSystemResources(resources)

	return resources
}
