package app

import (
	"fmt"

	"github.com/mlab-lattice/system/pkg/backend/kubernetes/constants"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	appsv1beta2 "k8s.io/api/apps/v1beta2"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
)

func seedCloudSpecific() {
	fmt.Println("Seeding cloud specific resources...")
	seedFlannel()
}

func seedFlannel() {
	// Translated from: https://github.com/coreos/flannel/blob/317b7d199e3fe937f04ecb39beed025e47316430/Documentation/kube-flannel.yml
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "flannel",
			Namespace: constants.NamespaceKubeSystem,
		},
	}

	pollKubeResourceCreation(func() (interface{}, error) {
		return kubeClient.CoreV1().ServiceAccounts(sa.Namespace).Create(sa)
	})

	cr := &rbacv1.ClusterRole{
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

	pollKubeResourceCreation(func() (interface{}, error) {
		return kubeClient.RbacV1().ClusterRoles().Create(cr)
	})

	crb := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: "flannel",
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      rbacv1.ServiceAccountKind,
				Name:      sa.Name,
				Namespace: sa.Namespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "ClusterRole",
			Name:     cr.Name,
		},
	}

	pollKubeResourceCreation(func() (interface{}, error) {
		return kubeClient.RbacV1().ClusterRoleBindings().Create(crb)
	})

	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kube-flannel-cfg",
			Namespace: constants.NamespaceKubeSystem,
		},
		Data: map[string]string{
			"cni-conf.json": `{
	"name": "cbr0",
	"type": "flannel",
	"delegate": {"isDefaultGateway": true}
}`,
			"net-conf.json": `{
	"Network": "10.200.0.0/16",
	"Backend": {"Type": "vxlan"}
}`,
		},
	}

	pollKubeResourceCreation(func() (interface{}, error) {
		return kubeClient.CoreV1().ConfigMaps(cm.Namespace).Create(cm)
	})

	truth := true
	dsLabels := map[string]string{
		"system.kubernetes.io/flannel": "true",
	}
	ds := &appsv1beta2.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kube-flannel-ds",
			Namespace: constants.NamespaceKubeSystem,
			Labels:    dsLabels,
		},
		Spec: appsv1beta2.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: dsLabels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "kube-flannel-ds",
					Labels: dsLabels,
				},
				Spec: corev1.PodSpec{
					HostNetwork: true,
					// Do not forget to add new tolerations here
					Tolerations: []corev1.Toleration{
						constants.TolerateAllTaints,
					},
					ServiceAccountName: sa.Name,
					InitContainers: []corev1.Container{
						{
							Name:    "install-cni",
							Image:   "quay.io/coreos/flannel:v0.9.0-amd64",
							Command: []string{"cp"},
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
										Name: cm.Name,
									},
								},
							},
						},
					},
				},
			},
		},
	}

	pollKubeResourceCreation(func() (interface{}, error) {
		return kubeClient.AppsV1beta2().DaemonSets(ds.Namespace).Create(ds)
	})
}
