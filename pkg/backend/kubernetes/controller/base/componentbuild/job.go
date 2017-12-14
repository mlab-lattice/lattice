package componentbuild

import (
	"encoding/json"
	"fmt"
	"time"

	kubeconstants "github.com/mlab-lattice/system/pkg/backend/kubernetes/constants"
	crv1 "github.com/mlab-lattice/system/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	"github.com/mlab-lattice/system/pkg/constants"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
)

const (
	workingDirectoryVolumeHostPathPrefixLocal = "/data/component-builder"
	workingDirectoryVolumeHostPathPrefixCloud = "/var/lib/component-builder"

	jobWorkingDirectory           = "/var/run/builder"
	jobWorkingDirectoryVolumeName = "workdir"

	jobDockerSocketVolumePath = "/var/run/docker.sock"
	jobDockerSocketPath       = "/var/run/docker.sock"
	jobDockerSocketVolumeName = "dockersock"

	jobDockerFqnAnnotationKey = "docker-image-fqn"
)

// getJobForBuild uses ControllerRefManager to retrieve the Job for a ComponentBuild
func (c *Controller) getJobForBuild(cb *crv1.ComponentBuild) (*batchv1.Job, error) {
	selector := labels.NewSelector()
	requirement, err := labels.NewRequirement(kubeconstants.LabelKeyComponentBuildID, selection.Equals, []string{cb.Name})
	if err != nil {
		return nil, err
	}

	selector = selector.Add(*requirement)
	jobs, err := c.jobLister.Jobs(cb.Namespace).List(selector)
	if err != nil {
		return nil, err
	}

	if len(jobs) == 0 {
		return nil, nil
	}

	if len(jobs) > 1 {
		return nil, fmt.Errorf("ComponentBuild %v has multiple Jobs", cb.Name)
	}

	return jobs[0], nil
}

func (c *Controller) createNewJob(build *crv1.ComponentBuild) (*batchv1.Job, error) {
	job, err := c.newJob(build)
	if err != nil {
		return nil, err
	}

	return c.kubeClient.BatchV1().Jobs(build.Namespace).Create(job)
}

func (c *Controller) newJob(build *crv1.ComponentBuild) (*batchv1.Job, error) {
	// Need a consistent view of our config while generating the Job
	c.configLock.RLock()
	defer c.configLock.RUnlock()

	name := jobName(build)

	// FIXME: get job spec for build.DockerImage as well
	spec, dockerImageFQN, err := c.gitRepositoryBuildJobSpec(build)
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

func jobName(cb *crv1.ComponentBuild) string {
	return fmt.Sprintf("lattice-build-%s", cb.Name)
}

func (c *Controller) gitRepositoryBuildJobSpec(build *crv1.ComponentBuild) (batchv1.JobSpec, string, error) {
	buildContainer, dockerImageFQN, err := c.getBuildContainer(build)
	if err != nil {
		return batchv1.JobSpec{}, "", err
	}

	name := jobName(build)

	provider, err := crv1.GetProviderFromConfigSpec(c.config)
	if err != nil {
		return batchv1.JobSpec{}, "", err
	}

	var volumeHostPathPrefix string
	switch provider {
	case constants.ProviderLocal:
		volumeHostPathPrefix = workingDirectoryVolumeHostPathPrefixLocal
	case constants.ProviderAWS:
		volumeHostPathPrefix = workingDirectoryVolumeHostPathPrefixCloud
	default:
		return batchv1.JobSpec{}, "", fmt.Errorf("unsupported provider: %s", provider)
	}

	workingDirectoryVolumeSource := corev1.VolumeSource{
		HostPath: &corev1.HostPathVolumeSource{
			Path: volumeHostPathPrefix + "/" + name,
		},
	}

	var zero int32
	spec := batchv1.JobSpec{
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

	return spec, dockerImageFQN, nil
}

func (c *Controller) getBuildContainer(cb *crv1.ComponentBuild) (*corev1.Container, string, error) {
	buildJSON, err := json.Marshal(&cb.Spec.BuildDefinitionBlock)
	if err != nil {
		return nil, "", err
	}

	repo := c.config.ComponentBuild.DockerArtifact.Repository
	tag := cb.Annotations[kubeconstants.AnnotationKeyComponentBuildDefinitionHash]
	if c.config.ComponentBuild.DockerArtifact.RepositoryPerImage {
		repo = cb.Annotations[kubeconstants.AnnotationKeyComponentBuildDefinitionHash]
		tag = fmt.Sprint(time.Now().Unix())
	}

	args := []string{
		"--component-build-id", cb.Name,
		"--component-build-definition", string(buildJSON),
		"--docker-registry", c.config.ComponentBuild.DockerArtifact.Registry,
		"--docker-repository", repo,
		"--docker-tag", tag,
		"--work-directory", jobWorkingDirectory,
	}

	if c.config.ComponentBuild.DockerArtifact.Push {
		args = append(args, "--docker-push")
	}

	provider, err := crv1.GetProviderFromConfigSpec(c.config)
	if err != nil {
		return nil, "", err
	}

	switch provider {
	case constants.ProviderLocal:
		// nothing to do here
	case constants.ProviderAWS:
		args = append(args, "--docker-registry-auth-type", constants.DockerRegistryAuthAWSEC2Role)
	default:
		return nil, "", fmt.Errorf("unsupported provider: %s", provider)
	}

	buildContainer := &corev1.Container{
		Name:  "build",
		Image: c.config.ComponentBuild.Builder.Image,
		Args:  args,
		Env: []corev1.EnvVar{
			{
				Name:  kubeconstants.EnvVarNameDockerAPIVersion,
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
