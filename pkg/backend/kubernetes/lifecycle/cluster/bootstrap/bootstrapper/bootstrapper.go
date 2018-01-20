package bootstrapper

import (
	"fmt"

	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"

	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"

	"github.com/ghodss/yaml"
)

type Interface interface {
	BootstrapClusterResources(*ClusterResources)
}

type ClusterResources struct {
	Namespaces      []*corev1.Namespace
	ServiceAccounts []*corev1.ServiceAccount

	Roles               []*rbacv1.Role
	RoleBindings        []*rbacv1.RoleBinding
	ClusterRoles        []*rbacv1.ClusterRole
	ClusterRoleBindings []*rbacv1.ClusterRoleBinding

	CustomResourceDefinitions []*apiextensionsv1beta1.CustomResourceDefinition
	Config                    *crv1.Config
	ConfigMaps                []*corev1.ConfigMap

	DaemonSets []*appsv1.DaemonSet
	Services   []*corev1.Service
}

func (r *ClusterResources) String() (string, error) {
	header := "---\n"
	output := ""

	for _, namespace := range r.Namespaces {
		data, err := yaml.Marshal(namespace)
		if err != nil {
			return "", err
		}

		output += fmt.Sprintf("%v%v", header, string(data))
	}

	for _, serviceAccount := range r.ServiceAccounts {
		data, err := yaml.Marshal(serviceAccount)
		if err != nil {
			return "", err
		}

		output += fmt.Sprintf("%v%v", header, string(data))
	}

	for _, role := range r.Roles {
		data, err := yaml.Marshal(role)
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

	for _, clusterRole := range r.ClusterRoles {
		data, err := yaml.Marshal(clusterRole)
		if err != nil {
			return "", err
		}

		output += fmt.Sprintf("%v%v", header, string(data))
	}

	for _, clusterRoleBinding := range r.ClusterRoleBindings {
		data, err := yaml.Marshal(clusterRoleBinding)
		if err != nil {
			return "", err
		}

		output += fmt.Sprintf("%v%v", header, string(data))
	}

	for _, definition := range r.CustomResourceDefinitions {
		data, err := yaml.Marshal(definition)
		if err != nil {
			return "", err
		}

		output += fmt.Sprintf("%v%v", header, string(data))
	}

	data, err := yaml.Marshal(r.Config)
	if err != nil {
		return "", err
	}

	output += fmt.Sprintf("%v%v", header, string(data))

	for _, configMap := range r.ConfigMaps {
		data, err := yaml.Marshal(configMap)
		if err != nil {
			return "", err
		}

		output += fmt.Sprintf("%v%v", header, string(data))
	}

	for _, daemonSet := range r.DaemonSets {
		data, err := yaml.Marshal(daemonSet)
		if err != nil {
			return "", err
		}

		output += fmt.Sprintf("%v%v", header, string(data))
	}

	return output, nil
}
