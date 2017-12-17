package base

import (
	"fmt"
	"strconv"

	"github.com/mlab-lattice/system/pkg/backend/kubernetes/constants"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/lifecycle/bootstrapper/util"
	kubeutil "github.com/mlab-lattice/system/pkg/backend/kubernetes/util/kubernetes"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

func (b *DefaultBootstrapper) seedMasterComponents() ([]interface{}, error) {
	if !b.Options.DryRun {
		fmt.Println("Seeding master components")
	}

	seedMasterComponentFuncs := []func() ([]interface{}, error){
		b.seedLatticeControllerManager,
		b.seedManagerAPI,
	}

	var objects []interface{}
	for _, seedMasterComponentFunc := range seedMasterComponentFuncs {
		additionalObjects, err := seedMasterComponentFunc()
		if err != nil {
			return nil, err
		}
		objects = append(objects, additionalObjects...)
	}
	return objects, nil
}

func (b *DefaultBootstrapper) seedLatticeControllerManager() ([]interface{}, error) {
	// TODO: for now we'll make a DaemonSet that runs on all the master nodes (aka all nodes in local)
	//		 and rely on the fact that we don't support multiple master nodes on AWS yet. Eventually we'll
	//		 either have to figure out how to have multiple lattice-controller-managers running (e.g. use leaderelect
	//		 in client-go) or find the best way to ensure there's at most one version of something running (maybe
	//		 StatefulSets?).
	namespace := kubeutil.InternalNamespace(b.ClusterID)
	args := []string{"--provider", b.Provider, "--cluster-id", string(b.ClusterID)}
	args = append(args, b.Options.MasterComponents.LatticeControllerManager.Args...)
	labels := map[string]string{
		constants.MasterNodeLabelComponent: constants.MasterNodeComponentLatticeControllerManager,
	}

	latticeControllerManagerDaemonSet := &appsv1.DaemonSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DaemonSet",
			APIVersion: appsv1.GroupName + "/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.MasterNodeComponentLatticeControllerManager,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:   constants.MasterNodeComponentLatticeControllerManager,
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  constants.MasterNodeComponentLatticeControllerManager,
							Image: b.Options.MasterComponents.LatticeControllerManager.Image,
							Args:  args,
						},
					},
					DNSPolicy:          corev1.DNSDefault,
					ServiceAccountName: constants.ServiceAccountLatticeControllerManager,
					Tolerations: []corev1.Toleration{
						constants.TolerationMasterNode,
					},
					Affinity: &corev1.Affinity{
						NodeAffinity: &constants.NodeAffinityMasterNode,
					},
				},
			},
		},
	}

	if b.Options.DryRun {
		return []interface{}{latticeControllerManagerDaemonSet}, nil
	}

	result, err := util.IdempotentSeed(func() (interface{}, error) {
		return b.KubeClient.AppsV1().DaemonSets(namespace).Create(latticeControllerManagerDaemonSet)
	})
	if err != nil {
		return nil, err
	}

	return []interface{}{result}, nil
}

func (b *DefaultBootstrapper) seedManagerAPI() ([]interface{}, error) {
	// TODO: for now we'll make a DaemonSet that runs on all the master nodes (aka all nodes in local)
	//		 and rely on the fact that we don't support multiple master nodes on AWS yet. Eventually we'll
	//		 either have to figure out how to have multiple lattice-controller-managers running (e.g. use leaderelect
	//		 in client-go) or find the best way to ensure there's at most one version of something running (maybe
	//		 StatefulSets?).
	namespace := kubeutil.InternalNamespace(b.ClusterID)
	args := []string{"--port", strconv.Itoa(int(b.Options.MasterComponents.ManagerAPI.Port)), "--cluster-id", string(b.ClusterID)}
	args = append(args, b.Options.MasterComponents.ManagerAPI.Args...)
	labels := map[string]string{
		constants.MasterNodeLabelComponent: constants.MasterNodeComponentManagerAPI,
	}

	managerAPIDaemonSet := &appsv1.DaemonSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DaemonSet",
			APIVersion: appsv1.GroupName + "/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.MasterNodeComponentManagerAPI,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:   constants.MasterNodeComponentManagerAPI,
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  constants.MasterNodeComponentManagerAPI,
							Image: b.Options.MasterComponents.ManagerAPI.Image,
							Args:  args,
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									HostPort:      b.Options.MasterComponents.ManagerAPI.Port,
									ContainerPort: b.Options.MasterComponents.ManagerAPI.Port,
								},
							},
						},
					},
					HostNetwork:        b.Options.MasterComponents.ManagerAPI.HostNetwork,
					DNSPolicy:          corev1.DNSDefault,
					ServiceAccountName: constants.ServiceAccountLatticeControllerManager,
					Tolerations: []corev1.Toleration{
						constants.TolerationMasterNode,
					},
					Affinity: &corev1.Affinity{
						NodeAffinity: &constants.NodeAffinityMasterNode,
					},
				},
			},
		},
	}

	if b.Options.DryRun {
		return []interface{}{managerAPIDaemonSet}, nil
	}

	result, err := util.IdempotentSeed(func() (interface{}, error) {
		return b.KubeClient.AppsV1().DaemonSets(namespace).Create(managerAPIDaemonSet)
	})
	if err != nil {
		return nil, err
	}

	return []interface{}{result}, nil
}
