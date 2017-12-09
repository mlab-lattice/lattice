package provisioner

import (
	"fmt"
	"os/user"
	"time"

	"github.com/mlab-lattice/system/pkg/backend/kubernetes/constants"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/util/minikube"

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

type LocalProvisioner struct {
	dockerAPIVersion           string
	latticeContainerRegistry   string
	latticeContainerRepoPrefix string
	mec                        *minikube.ExecContext
}

const (
	defaultDockerAPIVersion  = "1.24"
	systemNamePrefixMinikube = "lattice-local-"
)

func NewLocalProvisioner(dockerAPIVersion, latticeContainerRegistry, latticeContainerRepoPrefix, logPath string) (*LocalProvisioner, error) {
	mec, err := minikube.NewMinikubeExecContext(logPath)
	if err != nil {
		return nil, err
	}

	if dockerAPIVersion == "" {
		dockerAPIVersion = defaultDockerAPIVersion
	}

	lp := &LocalProvisioner{
		dockerAPIVersion:           dockerAPIVersion,
		latticeContainerRegistry:   latticeContainerRegistry,
		latticeContainerRepoPrefix: latticeContainerRepoPrefix,
		mec: mec,
	}
	return lp, nil
}

func (lp *LocalProvisioner) Provision(name, url string) error {
	prefixedName := systemNamePrefixMinikube + name
	result, logFilename, err := lp.mec.Start(prefixedName)
	if err != nil {
		return err
	}

	fmt.Printf("Running minikube start (pid: %v, log file: %v)\n", result.Pid, filepath.Join(*lp.mec.LogPath, logFilename))

	err = result.Wait()
	if err != nil {
		return err
	}

	address, err := lp.Address(name)
	if err != nil {
		return err
	}

	err = lp.bootstrap(address, url, name)
	if err != nil {
		return err
	}

	fmt.Println("Waiting for System Environment Manager to be ready...")
	return pollForSystemEnvironmentReadiness(address)
}

func (lp *LocalProvisioner) Address(name string) (string, error) {
	return lp.mec.IP(systemNamePrefixMinikube + name)
}

func (lp *LocalProvisioner) bootstrap(address, url, name string) error {
	fmt.Println("Bootstrapping")
	usr, err := user.Current()
	if err != nil {
		return err
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
		return err
	}

	kubeClientset := clientset.NewForConfigOrDie(config)

	fmt.Println("Creating bootstrap SA")
	bootstrapSA := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "bootstrap-lattice",
			Namespace: constants.NamespaceDefault,
		},
	}

	_, err = kubeClientset.
		CoreV1().
		ServiceAccounts(constants.NamespaceDefault).
		Create(bootstrapSA)
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return err
	}

	bootstrapClusterAdminRoleBind := rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: "bootstrap-lattice-cluster-admin",
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
		return err
	}

	jobName := "bootstrap-lattice"
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
							Name:  "kubernetes-bootstrap-lattice",
							Image: lp.latticeContainerRegistry + "/" + lp.latticeContainerRepoPrefix + constants.DockerImageBootstrapKubernetes,
							Args: []string{
								"--docker-api-version", lp.dockerAPIVersion,
								"--provider", "local",
								"--provider-var", "system-ip=" + address,
								"--system-definition-url", url,
								"--system-id", systemNamePrefixMinikube + name,
								"--lattice-container-registry", lp.latticeContainerRegistry,
								"--lattice-container-repo-prefix", lp.latticeContainerRepoPrefix,
								"--component-build-registry", "lattice-local",
							},
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
		Jobs(constants.NamespaceDefault).
		Create(&job)
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return err
	}

	fmt.Println("Polling bootstrap job status")
	err = wait.Poll(1*time.Second, 300*time.Second, func() (bool, error) {
		j, err := kubeClientset.BatchV1().Jobs(constants.NamespaceDefault).Get(job.Name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		if j.Status.Succeeded == 1 {
			return true, nil
		}

		if j.Status.Failed >= backoffLimit {
			return false, fmt.Errorf("surpassed backoffLimit")
		}

		return false, nil
	})
	if err != nil {
		return err
	}

	fmt.Println("Deleting bootstrap SA")
	return kubeClientset.CoreV1().ServiceAccounts(constants.NamespaceDefault).Delete(bootstrapSA.Name, nil)
}

func (lp *LocalProvisioner) Deprovision(name string) error {
	result, logFilename, err := lp.mec.Delete(systemNamePrefixMinikube + name)
	if err != nil {
		return err
	}

	fmt.Printf("Running minikube delete (pid: %v, log file: %v)\n", result.Pid, filepath.Join(*lp.mec.LogPath, logFilename))

	return result.Wait()
}
