package flannel

import (
	"fmt"

	kubeconstants "github.com/mlab-lattice/system/pkg/backend/kubernetes/constants"
	clusterbootstrapper "github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/cluster/bootstrap/bootstrapper"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/scale/scheme/appsv1beta2"
)

type Options struct {
	CIDRBlock string
}

func NewFlannelNetworkingProvider(options *Options) *DefaultFlannelNetworkingProvider {
	return &DefaultFlannelNetworkingProvider{
		CIDRBlock: options.CIDRBlock,
	}
}

type DefaultFlannelNetworkingProvider struct {
	CIDRBlock string
}

func (np *DefaultFlannelNetworkingProvider) BootstrapClusterResources(resources *clusterbootstrapper.ClusterResources) {
	// Translated from: https://github.com/coreos/flannel/blob/317b7d199e3fe937f04ecb39beed025e47316430/Documentation/kube-flannel.yml
	serviceAccount := &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceAccount",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "flannel",
			Namespace: kubeconstants.NamespaceKubeSystem,
		},
	}

	clusterRole := &rbacv1.ClusterRole{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRole",
			APIVersion: rbacv1.GroupName + "/metav1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "flannel",
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"pods"},
				Verbs:     []string{"get"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"nodes"},
				Verbs:     []string{"list", "watch"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"nodes/status"},
				Verbs:     []string{"patch"},
			},
		},
	}

	clusterRoleBinding := &rbacv1.ClusterRoleBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRoleBinding",
			APIVersion: rbacv1.GroupName + "/metav1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "flannel",
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      rbacv1.ServiceAccountKind,
				Name:      serviceAccount.Name,
				Namespace: serviceAccount.Namespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "ClusterRole",
			Name:     clusterRole.Name,
		},
	}

	netConf := fmt.Sprintf(`{
	"Network": "%v",
	"Backend": {"Type": "vxlan"}
}`, np.CIDRBlock)

	configMap := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kube-flannel-cfg",
			Namespace: kubeconstants.NamespaceKubeSystem,
		},
		Data: map[string]string{
			"cni-conf.json": `{
	"name": "cbr0",
	"type": "flannel",
	"delegate": {"isDefaultGateway": true}
}`,
			"net-conf.json": netConf,
		},
	}

	truth := true
	dsLabels := map[string]string{
		"system.kubernetes.io/flannel": "true",
	}
	daemonSet := &appsv1.DaemonSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DaemonSet",
			APIVersion: appsv1beta2.GroupName + "/v1beta2",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kube-flannel-daemonSet",
			Namespace: kubeconstants.NamespaceKubeSystem,
			Labels:    dsLabels,
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: dsLabels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "kube-flannel-daemonSet",
					Labels: dsLabels,
				},
				Spec: corev1.PodSpec{
					HostNetwork: true,
					Tolerations: []corev1.Toleration{
						kubeconstants.TolerateAllTaints,
					},
					ServiceAccountName: serviceAccount.Name,
					InitContainers: []corev1.Container{
						{
							Name:    "install-cni",
							Image:   "quay.io/coreos/flannel:v0.9.0-amd64",
							Command: []string{"np"},
							Args: []string{
								"-f",
								"/etc/kube-flannel/cni-conf.json",
								"/etc/cni/net.d/10-flannel.conf",
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "cni",
									MountPath: "/etc/cni/net.d",
								},
								{
									Name:      "flannel-cfg",
									MountPath: "/etc/kube-flannel/",
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:  "kube-flannel",
							Image: "quay.io/coreos/flannel:v0.9.0-amd64",
							Command: []string{
								"/opt/bin/flanneld",
								"--ip-masq",
								"--kube-subnet-mgr",
							},
							SecurityContext: &corev1.SecurityContext{
								Privileged: &truth,
							},
							Env: []corev1.EnvVar{
								{
									Name: "POD_NAME",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.name",
										},
									},
								},
								{
									Name: "POD_NAMESPACE",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.namespace",
										},
									},
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "run",
									MountPath: "/run",
								},
								{
									Name:      "flannel-cfg",
									MountPath: "/etc/kube-flannel/",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "run",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/run",
								},
							},
						},
						{
							Name: "cni",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/etc/cni/net.d",
								},
							},
						},
						{
							Name: "flannel-cfg",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: configMap.Name,
									},
								},
							},
						},
					},
				},
			},
		},
	}

	resources.ServiceAccounts = append(resources.ServiceAccounts, serviceAccount)
	resources.ClusterRoles = append(resources.ClusterRoles, clusterRole)
	resources.ClusterRoleBindings = append(resources.ClusterRoleBindings, clusterRoleBinding)
	resources.ConfigMaps = append(resources.ConfigMaps, configMap)
	resources.DaemonSets = append(resources.DaemonSets, daemonSet)
}
