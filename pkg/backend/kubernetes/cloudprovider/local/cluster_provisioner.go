package local

import (
	"fmt"
	"os/user"
	"strings"
	"time"

	kubeconstants "github.com/mlab-lattice/system/pkg/backend/kubernetes/constants"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/util/minikube"
	"github.com/mlab-lattice/system/pkg/constants"
	"github.com/mlab-lattice/system/pkg/managerapi/client/rest"

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
	clusterNamePrefixMinikube = "lattice-local-"
)

var (
	localDNSControllerArgList = []string{
		"-v", "5",
		"--logtostderr",
		"--dnsmasq-config-path", DnsmasqConfigFile,
		"--hosts-file-path", DNSHostsFile,
	}

	dnsNannyArgList = []string{
		"-v=2",
		"-logtostderr",
		"-restartDnsmasq=true",
		"-configDir=" + DNSConfigDirectory,
	}

	dnsmasqArgList = []string{
		"-k", // Keep in foreground so as to not immediately exit.
		"--hostsdir=" + DNSConfigDirectory,             // Read all the hosts from this directory. File changes read automatically by dnsmasq.
		"--conf-dir=" + DNSConfigDirectory + ",*.conf", // Read all *.conf files in the directory as dns config files
	}
)

type ClusterProvisionerOptions struct {
}

type DefaultLocalClusterProvisioner struct {
	latticeContainerRegistry   string
	latticeContainerRepoPrefix string
	mec                        *minikube.ExecContext
}

func NewClusterProvisioner(latticeContainerRegistry, latticeContainerRepoPrefix, workingDir string, options *ClusterProvisionerOptions) (*DefaultLocalClusterProvisioner, error) {
	mec, err := minikube.NewMinikubeExecContext(fmt.Sprintf("%v/logs", workingDir))
	if err != nil {
		return nil, err
	}

	provisioner := &DefaultLocalClusterProvisioner{
		latticeContainerRegistry:   latticeContainerRegistry,
		latticeContainerRepoPrefix: latticeContainerRepoPrefix,
		mec: mec,
	}
	return provisioner, nil
}

func (p *DefaultLocalClusterProvisioner) Provision(clusterID string, initialSystemDefinitionURL *string) (string, error) {
	prefixedName := clusterNamePrefixMinikube + clusterID
	result, logFilename, err := p.mec.Start(prefixedName)
	if err != nil {
		return "", err
	}

	fmt.Printf("Running minikube start (pid: %v, log file: %v)\n", result.Pid, filepath.Join(*p.mec.LogPath, logFilename))

	err = result.Wait()
	if err != nil {
		return "", err
	}

	address, err := p.address(clusterID)
	if err != nil {
		return "", err
	}

	err = p.bootstrap(address, initialSystemDefinitionURL, clusterID)
	if err != nil {
		return "", err
	}

	fmt.Println("Waiting for Cluster Manager to be ready...")
	clusterClient := rest.NewClient(address)
	err = wait.Poll(1*time.Second, 300*time.Second, func() (bool, error) {
		ok, _ := clusterClient.Status()
		return ok, nil
	})

	if err != nil {
		return "", err
	}

	return address, nil
}

func (p *DefaultLocalClusterProvisioner) address(clusterID string) (string, error) {
	address, err := p.mec.IP(clusterNamePrefixMinikube + clusterID)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("http://%v", address), nil
}

func (p *DefaultLocalClusterProvisioner) bootstrap(address string, initialSystemDefinitionURL *string, name string) error {
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
			Namespace: kubeconstants.NamespaceDefault,
		},
	}

	fmt.Printf("%v\n", kubeClientset)
	fmt.Printf("%v\n", kubeconstants.NamespaceDefault)
	fmt.Printf("%v\n", bootstrapSA)

	_, err = kubeClientset.
		CoreV1().
		ServiceAccounts(kubeconstants.NamespaceDefault).
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

	dnsNannyArgList := append(append(dnsNannyArgList, "--"), dnsmasqArgList...)
	// Use ':' as the separator here, as ',' is included in the --conf-dir argument
	dnsNannyArgs := "local-dns-server-args=" + strings.Join(dnsNannyArgList, ":")
	dnsControllerArgs := "local-dns-controller-args=" + strings.Join(localDNSControllerArgList, ",")

	jobName := "bootstrap-lattice"
	var backoffLimit int32 = 2
	var bootstrapArgs []string

	if initialSystemDefinitionURL != nil {
		bootstrapArgs = append(
			bootstrapArgs,
			"--initial-system-definition-url", *initialSystemDefinitionURL,
		)
	}

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
							Name:  "bootstrap-lattice",
							Image: p.getLatticeContainerImage(constants.DockerImageLatticectl),
							Args: append(
								bootstrapArgs,
								[]string{
									"cluster", "kubernetes", "bootstrap",
									"--lattice-controller-manager-image", p.getLatticeContainerImage(kubeconstants.DockerImageLatticeControllerManager),
									"--manager-api-image", p.getLatticeContainerImage(kubeconstants.DockerImageManagerAPIRest),
									"--cloud-provider", "local",
									"--cloud-provider-var", "cluster-ip=" + address,
									"--cloud-provider-var", "dns-controller-image=" + p.getLatticeContainerImage(DockerImageDNSController),
									"--cloud-provider-var", "dns-controller-args=" + dnsControllerArgs,
									"--cloud-provider-var", "dnsmasq-nanny-image=" + DockerImageDnsmasqNanny,
									"--cloud-provider-var", "dnsmasq-nanny-args=" + dnsNannyArgs,
									"--component-builder-image", p.getLatticeContainerImage(kubeconstants.DockerImageComponentBuilder),
									"--component-build-docker-artifact-registry", "lattice-local",
									"--component-build-docker-artifact-repository-per-image=true",
									"--component-build-docker-artifact-push=false",
									"--service-mesh", "envoy",
									"--service-mesh-var", fmt.Sprintf("prepare-image=%v", p.getLatticeContainerImage(constants.DockerImageEnvoyPrepare)),
									"--service-mesh-var", fmt.Sprintf("xds-api-image=%v", p.getLatticeContainerImage(constants.DockerImageEnvoyXDSAPIRestPerNode)),
									"--service-mesh-var", "redirect-cidr-block=172.16.0.0/16",
								}...,
							),
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
		Jobs(kubeconstants.NamespaceDefault).
		Create(&job)
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return err
	}

	fmt.Println("Polling bootstrap job status")
	err = wait.Poll(1*time.Second, 300*time.Second, func() (bool, error) {
		j, err := kubeClientset.BatchV1().Jobs(kubeconstants.NamespaceDefault).Get(job.Name, metav1.GetOptions{})
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
	return kubeClientset.CoreV1().ServiceAccounts(kubeconstants.NamespaceDefault).Delete(bootstrapSA.Name, nil)
}

func (p *DefaultLocalClusterProvisioner) Deprovision(name string, force bool) error {
	result, logFilename, err := p.mec.Delete(clusterNamePrefixMinikube + name)
	if err != nil {
		return err
	}

	fmt.Printf("Running minikube delete (pid: %v, log file: %v)\n", result.Pid, filepath.Join(*p.mec.LogPath, logFilename))

	return result.Wait()
}

func (p *DefaultLocalClusterProvisioner) getLatticeContainerImage(image string) string {
	return p.latticeContainerRegistry + "/" + p.latticeContainerRepoPrefix + image
}
