package main

import (
	"flag"
	"fmt"
	"os/user"
	"time"

	"github.com/mlab-lattice/kubernetes-integration/pkg/util/minikube"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"path/filepath"
)

const (
	workingDir         = "/tmp/lattice/provision"
	devDockerRegistry  = "gcr.io/lattice-dev"
	bootstrapImageName = "bootstrap-kubernetes"
	defaultNamespace   = "default"
)

var (
	systemName string
)

func init() {
	flag.StringVar(&systemName, "system-name", "", "name of the system to provision")
	flag.Parse()
}

func main() {
	mec, err := minikube.NewMinikubeExecContext(workingDir, systemName)
	if err != nil {
		panic(err)
	}

	pid, logFilename, waitFunc, err := mec.Start()
	if err != nil {
		panic(err)
	}

	fmt.Printf("minikube start\npid: %v\nlogFilename: %v\n\n", pid, logFilename)

	err = waitFunc()
	if err != nil {
		panic(err)
	}

	bootstrap()
}

func bootstrap() {
	fmt.Println("Bootstrapping")
	usr, err := user.Current()
	if err != nil {
		panic(err)
	}
	// TODO: support passing in the context when supported
	// https://github.com/kubernetes/minikube/issues/2100
	//configOverrides := &clientcmd.ConfigOverrides{CurrentContext: kubeContext}
	configOverrides := &clientcmd.ConfigOverrides{}
	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: filepath.Join(usr.HomeDir, ".kube/config")},
		configOverrides,
	).ClientConfig()

	if err != nil {
		panic(err)
	}

	kubeClientset := clientset.NewForConfigOrDie(config)

	fmt.Println("Creating bootstrap SA")
	bootstrapSA := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kubernetes-bootstrapper",
			Namespace: defaultNamespace,
		},
	}

	_, err = kubeClientset.
		CoreV1().
		ServiceAccounts(defaultNamespace).
		Create(bootstrapSA)
	if err != nil && !apierrors.IsAlreadyExists(err) {
		panic(err)
	}

	bootstrapClusterAdminRoleBind := rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: "kubernetes-bootstrapper-cluster-admin",
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      rbacv1.ServiceAccountKind,
				Name:      bootstrapSA.Name,
				Namespace: bootstrapSA.Namespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "ClusterRole",
			Name:     "cluster-admin",
		},
	}
	_, err = kubeClientset.
		RbacV1().
		ClusterRoleBindings().
		Create(&bootstrapClusterAdminRoleBind)
	if err != nil && !apierrors.IsAlreadyExists(err) {
		panic(err)
	}

	jobName := "lattice-bootstrap-kubernetes"
	var backoffLimit int32 = 2
	job := batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: jobName,
		},
		Spec: batchv1.JobSpec{
			BackoffLimit: &backoffLimit,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name: jobName,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:    "bootstrap-kubernetes",
							Image:   devDockerRegistry + "/" + bootstrapImageName,
							Command: []string{"/app/cmd/bootstrap-kubernetes/go_image.binary"},
							Args:    []string{"-provider", "local", "-user-system-url", "github.com/foo/bar"},
						},
					},
					RestartPolicy:      corev1.RestartPolicyNever,
					DNSPolicy:          corev1.DNSDefault,
					ServiceAccountName: bootstrapSA.Name,
				},
			},
		},
	}

	fmt.Println("Creating bootstrap job")
	_, err = kubeClientset.
		BatchV1().
		Jobs(defaultNamespace).
		Create(&job)
	if err != nil && !apierrors.IsAlreadyExists(err) {
		panic(err)
	}

	fmt.Println("Polling bootstrap job status")
	err = wait.Poll(1*time.Second, 300*time.Second, func() (bool, error) {
		j, err := kubeClientset.BatchV1().Jobs(defaultNamespace).Get(job.Name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		if j.Status.Succeeded == 1 {
			return true, nil
		}

		if j.Status.Failed >= backoffLimit {
			return false, fmt.Errorf("suprassed backoffLimit")
		}

		return false, nil
	})
	if err != nil {
		panic(err)
	}

	fmt.Println("Deleting bootstrap SA")
	err = kubeClientset.CoreV1().ServiceAccounts(defaultNamespace).Delete(bootstrapSA.Name, nil)
	if err != nil {
		panic(err)
	}
}
