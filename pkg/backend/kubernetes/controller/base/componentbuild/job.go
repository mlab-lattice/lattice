package componentbuild

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	kubeconstants "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/constants"
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	kubeutil "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/util/kubernetes"
	"github.com/mlab-lattice/lattice/pkg/util/docker"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
)

const (
	jobWorkingDirectory           = "/var/run/builder"
	jobWorkingDirectoryVolumeName = "workdir"

	jobDockerSocketVolumePath = "/var/run/docker.sock"
	jobDockerSocketPath       = "/var/run/docker.sock"
	jobDockerSocketVolumeName = "dockersock"

	jobDockerFqnAnnotationKey = "docker-image-fqn"
)

// getJobForBuild uses ControllerRefManager to retrieve the Job for a ComponentBuild
func (c *Controller) getJobForBuild(build *latticev1.ComponentBuild) (*batchv1.Job, error) {
	selector := labels.NewSelector()
	requirement, err := labels.NewRequirement(kubeconstants.LabelKeyComponentBuildID, selection.Equals, []string{build.Name})
	if err != nil {
		return nil, err
	}

	selector = selector.Add(*requirement)
	jobs, err := c.jobLister.Jobs(build.Namespace).List(selector)
	if err != nil {
		return nil, err
	}

	if len(jobs) == 1 {
		return jobs[0], nil
	}

	if len(jobs) > 1 {
		return nil, fmt.Errorf("ComponentBuild %v has multiple Jobs", build.Name)
	}

	// Didn't find anything in the cache. Will do a full API query to see if one exists.
	listOptions := metav1.ListOptions{
		LabelSelector: selector.String(),
	}
	jobItems, err := c.kubeClient.BatchV1().Jobs(build.Namespace).List(listOptions)
	if err != nil {
		return nil, err
	}

	if len(jobItems.Items) == 0 {
		return nil, nil
	}

	if len(jobItems.Items) > 1 {
		return nil, fmt.Errorf("ComponentBuild %v has multiple Jobs", build.Name)
	}

	return &jobItems.Items[0], nil
}

func (c *Controller) createNewJob(build *latticev1.ComponentBuild) (*batchv1.Job, error) {
	job, err := c.newJob(build)
	if err != nil {
		return nil, err
	}

	return c.kubeClient.BatchV1().Jobs(build.Namespace).Create(job)
}

func (c *Controller) newJob(build *latticev1.ComponentBuild) (*batchv1.Job, error) {
	// Need a consistent view of our config while generating the Job
	c.configLock.RLock()
	defer c.configLock.RUnlock()

	name := jobName(build)

	spec, dockerImageFQN, err := c.jobSpec(build)
	if err != nil {
		return nil, err
	}

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				jobDockerFqnAnnotationKey: dockerImageFQN,
			},
			Labels: map[string]string{
				kubeconstants.LabelKeyComponentBuildID: build.Name,
			},
			Name:            name,
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(build, controllerKind)},
		},
		Spec: spec,
	}
	return job, nil
}

func jobName(build *latticev1.ComponentBuild) string {
	return fmt.Sprintf("lattice-build-%s", build.Name)
}

func (c *Controller) jobSpec(build *latticev1.ComponentBuild) (batchv1.JobSpec, string, error) {
	buildContainer, dockerImageFQN, err := c.getBuildContainer(build)
	if err != nil {
		return batchv1.JobSpec{}, "", err
	}

	name := jobName(build)
	workingDirectoryVolumeSource := c.cloudProvider.ComponentBuildWorkDirectoryVolumeSource(name)

	var zero int32
	spec := &batchv1.JobSpec{
		// Only †ry to run the build once
		BackoffLimit: &zero,
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
				Labels: map[string]string{
					kubeconstants.LabelKeyComponentBuildID: build.Name,
				},
			},
			Spec: corev1.PodSpec{
				Tolerations: []corev1.Toleration{
					// Can tolerate build node taint even in local case
					kubeconstants.TolerationBuildNode,
				},
				Volumes: []corev1.Volume{
					{
						Name:         jobWorkingDirectoryVolumeName,
						VolumeSource: workingDirectoryVolumeSource,
					},
					{
						Name: jobDockerSocketVolumeName,
						VolumeSource: corev1.VolumeSource{
							HostPath: &corev1.HostPathVolumeSource{
								Path: jobDockerSocketVolumePath,
							},
						},
					},
				},
				Containers:         []corev1.Container{*buildContainer},
				ServiceAccountName: kubeconstants.ServiceAccountComponentBuilder,
				RestartPolicy:      corev1.RestartPolicyNever,
				DNSPolicy:          corev1.DNSDefault,
				Affinity: &corev1.Affinity{
					NodeAffinity: &corev1.NodeAffinity{
						RequiredDuringSchedulingIgnoredDuringExecution: &kubeconstants.NodeSelectorBuildNode,
					},
				},
			},
		},
	}

	spec = c.cloudProvider.TransformComponentBuildJobSpec(spec)

	return *spec, dockerImageFQN, nil
}

func (c *Controller) getBuildContainer(build *latticev1.ComponentBuild) (*corev1.Container, string, error) {
	buildJSON, err := json.Marshal(&build.Spec.BuildDefinitionBlock)
	if err != nil {
		return nil, "", err
	}

	repo := c.config.ComponentBuild.DockerArtifact.Repository
	tag := build.Annotations[kubeconstants.AnnotationKeyComponentBuildDefinitionHash]
	if c.config.ComponentBuild.DockerArtifact.RepositoryPerImage {
		repo = build.Annotations[kubeconstants.AnnotationKeyComponentBuildDefinitionHash]
		tag = fmt.Sprint(time.Now().Unix())
	}

	systemID, err := kubeutil.SystemID(build.Namespace)
	if err != nil {
		return nil, "", err
	}

	args := []string{
		"--component-build-id", build.Name,
		"--lattice-id", string(c.latticeID),
		"--system-id", string(systemID),
		"--component-build-definition", string(buildJSON),
		"--docker-registry", c.config.ComponentBuild.DockerArtifact.Registry,
		"--docker-repository", repo,
		"--docker-tag", tag,
		"--work-directory", jobWorkingDirectory,
	}

	if c.config.ComponentBuild.DockerArtifact.Push {
		args = append(args, "--docker-push")
	}

	if c.config.ComponentBuild.DockerArtifact.RegistryAuthType != nil {
		args = append(args, "--docker-registry-auth-type", *c.config.ComponentBuild.DockerArtifact.RegistryAuthType)
	}

	buildContainer := &corev1.Container{
		Name:  "build",
		Image: c.config.ComponentBuild.Builder.Image,
		Args:  args,
		Env: []corev1.EnvVar{
			{
				Name:  docker.APIVersionEnvironmentVariable,
				Value: c.config.ComponentBuild.Builder.DockerAPIVersion,
			},
		},
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      jobWorkingDirectoryVolumeName,
				MountPath: jobWorkingDirectory,
			},
			{
				Name:      jobDockerSocketVolumeName,
				MountPath: jobDockerSocketPath,
			},
		},
	}

	if build.Spec.BuildDefinitionBlock.GitRepository != nil && build.Spec.BuildDefinitionBlock.GitRepository.SSHKey != nil {
		// FIXME: add support for references
		secretParts := strings.Split(*build.Spec.BuildDefinitionBlock.GitRepository.SSHKey.Name, ":")
		if len(secretParts) != 2 {
			return nil, "", fmt.Errorf("invalid secret format for ssh_key")
		}

		secretPath := secretParts[0]
		secretName := secretParts[1]

		buildContainer.Env = append(buildContainer.Env, corev1.EnvVar{
			Name: "GIT_REPO_SSH_KEY",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: secretPath,
					},
					Key: secretName,
				},
			},
		})
	}

	dockerImageFQN := fmt.Sprintf(
		"%v/%v:%v",
		c.config.ComponentBuild.DockerArtifact.Registry,
		repo,
		tag,
	)

	return buildContainer, dockerImageFQN, nil
}

func jobStatus(j *batchv1.Job) (finished bool, succeeded bool) {
	for _, c := range j.Status.Conditions {
		if c.Type == batchv1.JobComplete && c.Status == corev1.ConditionTrue {
			finished = true
			succeeded = true
			return
		}
		if (c.Type == batchv1.JobComplete || c.Type == batchv1.JobFailed) && c.Status == corev1.ConditionTrue {
			finished = true
			return
		}
	}
	return
}
