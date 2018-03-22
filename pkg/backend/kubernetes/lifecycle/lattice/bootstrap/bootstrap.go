package bootstrap

import (
	"fmt"

	"github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource"
	latticev1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	latticeclientset "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/generated/clientset/versioned"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/lattice/bootstrap/bootstrapper"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/lattice/bootstrap/bootstrapper/base"
	"github.com/mlab-lattice/system/pkg/types"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"

	kubeclientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"k8s.io/apimachinery/pkg/api/errors"
)

type Options struct {
	Config           latticev1.ConfigSpec
	MasterComponents base.MasterComponentOptions
	Terraform        base.TerraformOptions
}

func Bootstrap(
	latticeID types.LatticeID,
	cloudProviderName string,
	options *Options,
	bootstrappers []bootstrapper.Interface,
	kubeConfig *rest.Config,
	kubeClient kubeclientset.Interface,
	latticeClient latticeclientset.Interface,
) (*bootstrapper.Resources, error) {
	resources, err := GetBootstrapResources(latticeID, cloudProviderName, options, bootstrappers)
	if err != nil {
		return nil, err
	}

	// First seed Namespaces so any future resources that get seeded in a serviceAccount can succeed
	fmt.Println("seeding namespaces")
	var namespaces []*corev1.Namespace
	for _, namespace := range resources.Namespaces {
		var result *corev1.Namespace
		err = idempotentSeed("namespace "+namespace.Name, func() error {
			result, err = kubeClient.CoreV1().Namespaces().Create(namespace)
			return err
		})

		if err != nil {
			return nil, err
		}

		namespaces = append(namespaces, namespace)
	}
	resources.Namespaces = namespaces

	// Next, seed ServiceAccounts so any RBAC resources or DaemonSets that use them can succeed
	fmt.Println("seeding service accounts")
	var serviceAccounts []*corev1.ServiceAccount
	for _, serviceAccount := range resources.ServiceAccounts {
		var result *corev1.ServiceAccount
		err := idempotentSeed(fmt.Sprintf("service account %v/%v", serviceAccount.Namespace, serviceAccount.Name), func() error {
			result, err = kubeClient.CoreV1().ServiceAccounts(serviceAccount.Namespace).Create(serviceAccount)
			return err
		})

		if err != nil {
			return nil, err
		}
		serviceAccounts = append(serviceAccounts, serviceAccount)
	}
	resources.ServiceAccounts = serviceAccounts

	// Next, seed RBAC resources so there's no race between any DaemonSets that require access to resources
	// and their Roles being seeded
	fmt.Println("seeding roles")
	var roles []*rbacv1.Role
	for _, role := range resources.Roles {
		var result *rbacv1.Role
		err := idempotentSeed(fmt.Sprintf("role %v/%v", role.Namespace, role.Name), func() error {
			result, err = kubeClient.RbacV1().Roles(role.Namespace).Create(role)
			return err
		})

		if err != nil {
			return nil, err
		}
		roles = append(roles, result)
	}
	resources.Roles = roles

	fmt.Println("seeding role bindings")
	var roleBindings []*rbacv1.RoleBinding
	for _, roleBinding := range resources.RoleBindings {
		var result *rbacv1.RoleBinding
		err = idempotentSeed(fmt.Sprintf("role binding %v/%v", roleBinding.Namespace, roleBinding.Name), func() error {
			result, err = kubeClient.RbacV1().RoleBindings(roleBinding.Namespace).Create(roleBinding)
			return err
		})

		if err != nil {
			return nil, err
		}
		roleBindings = append(roleBindings, result)
	}
	resources.RoleBindings = roleBindings

	fmt.Println("seeding cluster roles")
	var clusterRoles []*rbacv1.ClusterRole
	for _, clusterRole := range resources.ClusterRoles {
		var result *rbacv1.ClusterRole
		err = idempotentSeed(fmt.Sprintf("cluster role %v", clusterRole.Name), func() error {
			result, err = kubeClient.RbacV1().ClusterRoles().Create(clusterRole)
			return err
		})

		if err != nil {
			return nil, err
		}
		clusterRoles = append(clusterRoles, result)
	}
	resources.ClusterRoles = clusterRoles

	fmt.Println("seeding cluster role bindings")
	var clusterRoleBindings []*rbacv1.ClusterRoleBinding
	for _, clusterRoleBinding := range resources.ClusterRoleBindings {
		var result *rbacv1.ClusterRoleBinding
		err = idempotentSeed(fmt.Sprintf("cluster role binding %v", clusterRoleBinding.Name), func() error {
			result, err = kubeClient.RbacV1().ClusterRoleBindings().Create(clusterRoleBinding)
			return err
		})

		if err != nil {
			return nil, err
		}
		clusterRoleBindings = append(clusterRoleBindings, result)
	}
	resources.ClusterRoleBindings = clusterRoleBindings

	// Next, seed custom resource definitions.
	fmt.Println("seeding custom resource definitions")
	definitions, err := customresource.CreateCustomResourceDefinitions(resources.CustomResourceDefinitions, kubeConfig)
	if err != nil {
		return nil, err
	}
	resources.CustomResourceDefinitions = definitions

	// Next, seed the Config
	fmt.Println("seeding lattice config")
	config := resources.Config
	err = idempotentSeed(fmt.Sprintf("config %v/%v", config.Namespace, config.Name), func() error {
		config, err = latticeClient.LatticeV1().Configs(config.Namespace).Create(config)
		return err
	})
	if err != nil {
		return nil, err
	}
	resources.Config = config

	// Next, seed config maps
	fmt.Println("seeding config maps")
	var configMaps []*corev1.ConfigMap
	for _, configMap := range resources.ConfigMaps {
		var result *corev1.ConfigMap
		err = idempotentSeed(fmt.Sprintf("config map %v/%v", configMap.Namespace, configMap.Name), func() error {
			result, err = kubeClient.CoreV1().ConfigMaps(configMap.Namespace).Create(configMap)
			return err
		})

		if err != nil {
			return nil, err
		}
		configMaps = append(configMaps, result)
	}
	resources.ConfigMaps = configMaps

	// Finally, seed any DaemonSets
	fmt.Println("seeding daemon sets")
	var daemonSets []*appsv1.DaemonSet
	for _, daemonSet := range resources.DaemonSets {
		var result *appsv1.DaemonSet
		err = idempotentSeed(fmt.Sprintf("daemon set %v/%v", daemonSet.Namespace, daemonSet.Name), func() error {
			result, err = kubeClient.AppsV1().DaemonSets(daemonSet.Namespace).Create(daemonSet)
			return err
		})

		if err != nil {
			return nil, err
		}
		daemonSets = append(daemonSets, result)
	}
	resources.DaemonSets = daemonSets

	//Seed any services.
	fmt.Println("seeding services")
	var services []*corev1.Service
	for _, service := range resources.Services {
		var result *corev1.Service
		err = idempotentSeed(fmt.Sprintf("service %v/%v", service.Namespace, service.Name), func() error {
			result, err = kubeClient.CoreV1().Services(service.Namespace).Create(service)
			return err
		})

		if err != nil {
			return nil, err
		}
		services = append(services, result)
	}
	resources.Services = services

	return resources, nil
}

func idempotentSeed(resourceDescription string, seedFunc func() error) error {
	err := seedFunc()
	if err == nil {
		return nil
	}

	if errors.IsAlreadyExists(err) {
		fmt.Printf("%v already existed, continuing\n", resourceDescription)
		return nil
	}

	return err
}

func GetBootstrapResources(
	latticeID types.LatticeID,
	cloudProviderName string,
	options *Options,
	bootstrappers []bootstrapper.Interface,
) (*bootstrapper.Resources, error) {
	baseOptions := &base.Options{
		Config:           options.Config,
		MasterComponents: options.MasterComponents,
		TerraformOptions: options.Terraform,
	}

	baseBootstrapper, err := base.NewBootstrapper(latticeID, cloudProviderName, baseOptions)
	if err != nil {
		return nil, err
	}

	resources := &bootstrapper.Resources{}
	baseBootstrapper.BootstrapClusterResources(resources)

	for _, b := range bootstrappers {
		b.BootstrapLatticeResources(resources)
	}

	return resources, nil
}
