package local

import (
	"fmt"
	"os/user"
	"time"

	"github.com/mlab-lattice/lattice/pkg/api/client/rest"
	kubeconstants "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/constants"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/servicemesh"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/util/minikube"

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
	dnsControllerArgList = []string{
		"-v", "5",
		"--logtostderr",
		"--dnsmasq-config-path", DnsmasqConfigFile,
		"--hosts-file-path", DNSHostsFile,
	}

	// arguments for the dnsmasq nanny itself
	dnsmasqNannyArgList = []string{
		"-v=2",
		"-logtostderr",
		"-restartDnsmasq=true",
		"-configDir=" + DNSConfigDirectory,
	}

	// arguments passed through dnsmasq-nanny to dnsmasq
	dnsmasqArgList = []string{
		"-k", // Keep in foreground so as to not immediately exit.
		"--hostsdir=" + DNSConfigDirectory,             // Read all the hosts from this directory. File changes read automatically by dnsmasq.
		"--conf-dir=" + DNSConfigDirectory + ",*.conf", // Read all *.conf files in the directory as dns config files
	}
)

type LatticeProvisionerOptions struct {
}

type DefaultLocalLatticeProvisioner struct {
	latticeContainerRegistry   string
	latticeContainerRepoPrefix string
	mec                        *minikube.ExecContext
}

func NewLatticeProvisioner(latticeContainerRegistry, latticeContainerRepoPrefix, workingDir string, options *LatticeProvisionerOptions) (*DefaultLocalLatticeProvisioner, error) {
	mec, err := minikube.NewMinikubeExecContext(fmt.Sprintf("%v/logs", workingDir))
	if err != nil {
		return nil, err
	}

	provisioner := &DefaultLocalLatticeProvisioner{
		latticeContainerRegistry:   latticeContainerRegistry,
		latticeContainerRepoPrefix: latticeContainerRepoPrefix,
		mec: mec,
	}
	return provisioner, nil
}

func (p *DefaultLocalLatticeProvisioner) Provision(latticeID string) (string, error) {
	prefixedName := clusterNamePrefixMinikube + latticeID
	result, logFilename, err := p.mec.Start(prefixedName)
	if err != nil {
		return "", err
	}

	fmt.Printf("Running minikube start (pid: %v, log file: %v)\n", result.Pid, filepath.Join(*p.mec.LogPath, logFilename))

	err = result.Wait()
	if err != nil {
		return "", err
	}

	address, err := p.address(latticeID)
	if err != nil {
		return "", err
	}

	err = p.bootstrap(address, latticeID)
	if err != nil {
		return "", err
	}

	fmt.Println("Waiting for API server to be ready...")
	clusterClient := rest.NewClient(address)
	err = wait.Poll(1*time.Second, 300*time.Second, func() (bool, error) {
		ok, _ := clusterClient.Health()
		return ok, nil
	})

	if err != nil {
		return "", err
	}

	return address, nil
}

func (p *DefaultLocalLatticeProvisioner) address(latticeID string) (string, error) {
	address, err := p.mec.IP(clusterNamePrefixMinikube + latticeID)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("http://%v", address), nil
}

func (p *DefaultLocalLatticeProvisioner) bootstrap(address string, name string) error {
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

	jobName := "bootstrap-lattice"
	var backoffLimit int32 = 2
	var bootstrapArgs []string

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
							Image: p.getLatticeContainerImage("latticectl"),
							Args: append(
								bootstrapArgs,
								[]string{
									"kubernetes:bootstrap",
									"--lattice-id", "local",
									"--controller-manager-var", fmt.Sprintf("image=%v", p.getLatticeContainerImage("kubernetes-lattice-controller-manager")),
									"--controller-manager-var", "args=-v=5",
									"--api-var", fmt.Sprintf("image=%v", p.getLatticeContainerImage("kubernetes-api-server-rest")),
									"--component-builder-var", fmt.Sprintf("image=%v", p.getLatticeContainerImage("kubernetes-component-builder")),
									"--component-builder-var", "docker-api-version=1.35",
									"--component-build-docker-artifact-var", "registry=lattice-local",
									"--component-build-docker-artifact-var", "repository-per-image=true",
									"--component-build-docker-artifact-var", "push=false",
									"--service-mesh", servicemesh.Envoy,
									"--service-mesh-var", fmt.Sprintf("prepare-image=%v", p.getLatticeContainerImage("kubernetes-envoy-prepare")),
									"--service-mesh-var", fmt.Sprintf("xds-api-image=%v", p.getLatticeContainerImage("kubernetes-envoy-xds-api-rest-per-node")),
									"--service-mesh-var", "redirect-cidr-block=172.16.0.0/16",
									"--cloud-provider", "local",
									"--cloud-provider-var", "IP=" + address,
									"--cloud-provider-var", fmt.Sprintf("dns-var=controller-image=%v", p.getLatticeContainerImage(DockerImageDNSController)),
									"--cloud-provider-var", fmt.Sprintf("dns-var=dnsmasq-nanny-image=%v", DockerImageDnsmasqNanny),
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

	// add dns controller args
	for _, arg := range dnsControllerArgList {
		job.Spec.Template.Spec.Containers[0].Args = append(
			job.Spec.Template.Spec.Containers[0].Args,
			"--cloud-provider-var",
			fmt.Sprintf("dns-var=controller-args=%v", arg),
		)
	}

	// add dnsmasq nanny args
	for _, arg := range dnsmasqNannyArgList {
		job.Spec.Template.Spec.Containers[0].Args = append(
			job.Spec.Template.Spec.Containers[0].Args,
			"--cloud-provider-var",
			fmt.Sprintf("dns-var=dnsmasq-nanny-args=%v", arg),
		)
	}

	// dnsmasq nanny expects its args, then a --, then the args to pass through to dnsmasq
	job.Spec.Template.Spec.Containers[0].Args = append(
		job.Spec.Template.Spec.Containers[0].Args,
		"--cloud-provider-var",
		"dns-var=dnsmasq-nanny-args=--",
	)
	for _, arg := range dnsmasqArgList {
		job.Spec.Template.Spec.Containers[0].Args = append(
			job.Spec.Template.Spec.Containers[0].Args,
			"--cloud-provider-var",
			fmt.Sprintf("dns-var=dnsmasq-nanny-args=%v", arg),
		)
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

func (p *DefaultLocalLatticeProvisioner) Deprovision(name string, force bool) error {
	result, logFilename, err := p.mec.Delete(clusterNamePrefixMinikube + name)
	if err != nil {
		return err
	}

	fmt.Printf("Running minikube delete (pid: %v, log file: %v)\n", result.Pid, filepath.Join(*p.mec.LogPath, logFilename))

	return result.Wait()
}

func (p *DefaultLocalLatticeProvisioner) getLatticeContainerImage(image string) string {
	return fmt.Sprintf("%v/%v-%v", p.latticeContainerRegistry, p.latticeContainerRepoPrefix, image)
}
