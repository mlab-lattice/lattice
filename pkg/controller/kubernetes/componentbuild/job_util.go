package componentbuild

import (
	"fmt"
	"reflect"
	"time"

	coreconstants "github.com/mlab-lattice/core/pkg/constants"
	systemdefinitionblock "github.com/mlab-lattice/core/pkg/system/definition/block"

	crv1 "github.com/mlab-lattice/kubernetes-integration/pkg/api/customresource/v1"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
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
func (cbc *ComponentBuildController) getJobForBuild(cb *crv1.ComponentBuild) (*batchv1.Job, error) {
	// List all Jobs to find in the ComponentBuild's namespace to find the Job the ComponentBuild manages.
	jList, err := cbc.jobLister.Jobs(cb.Namespace).List(labels.Everything())
	if err != nil {
		return nil, err
	}

	matchingJobs := []*batchv1.Job{}
	cbControllerRef := metav1.NewControllerRef(cb, controllerKind)

	for _, j := range jList {
		jControllerRef := metav1.GetControllerOf(j)

		if reflect.DeepEqual(cbControllerRef, jControllerRef) {
			matchingJobs = append(matchingJobs, j)
		}
	}

	if len(matchingJobs) == 0 {
		return nil, nil
	}

	if len(matchingJobs) > 1 {
		return nil, fmt.Errorf("ComponentBuild %v has multiple Jobs", cb.Name)
	}

	return matchingJobs[0], nil
}

func (cbc *ComponentBuildController) getBuildJob(cb *crv1.ComponentBuild) (*batchv1.Job, error) {
	// Need a consistent view of our config while generating the Job
	cbc.configLock.RLock()
	defer cbc.configLock.RUnlock()

	name := getBuildJobName(cb)

	// FIXME: get job spec for build.DockerImage as well
	jSpec, dockerImageFqn, err := cbc.getGitRepositoryBuildJobSpec(cb)
	if err != nil {
		return nil, err
	}

	jLabels := map[string]string{
		crv1.ComponentBuildJobLabelKey: "true",
	}
	jAnnotations := map[string]string{
		jobDockerFqnAnnotationKey: dockerImageFqn,
	}

	j := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Annotations:     jAnnotations,
			Labels:          jLabels,
			Name:            name,
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(cb, controllerKind)},
		},
		Spec: jSpec,
	}
	return j, nil
}

func getBuildJobName(cb *crv1.ComponentBuild) string {
	return fmt.Sprintf("lattice-build-%s", cb.Name)
}

func (cbc *ComponentBuildController) getGitRepositoryBuildJobSpec(cb *crv1.ComponentBuild) (batchv1.JobSpec, string, error) {
	pullGitRepoContainer := cbc.getPullGitRepoContainer(cb)
	buildDockerImageContainer, dockerImageFqn := cbc.getBuildDockerImageContainer(cb)
	name := getBuildJobName(cb)

	provider, err := crv1.GetProviderFromConfigSpec(cbc.config)
	if err != nil {
		return batchv1.JobSpec{}, "", err
	}

	var volumeHostPathPrefix string
	switch provider {
	case coreconstants.ProviderLocal:
		volumeHostPathPrefix = workingDirectoryVolumeHostPathPrefixLocal
	case coreconstants.ProviderAWS:
		volumeHostPathPrefix = workingDirectoryVolumeHostPathPrefixCloud
	default:
		panic(fmt.Sprintf("unsupported provider: %s", provider))
	}

	workingDirectoryVolumeSource := corev1.VolumeSource{
		HostPath: &corev1.HostPathVolumeSource{
			Path: volumeHostPathPrefix + "/" + name,
		},
	}

	jobSpec := batchv1.JobSpec{
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
			Spec: corev1.PodSpec{
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
				InitContainers: []corev1.Container{
					pullGitRepoContainer,
				},
				Containers: []corev1.Container{
					buildDockerImageContainer,
				},
				// TODO: add failure policy once it is supported: https://github.com/kubernetes/kubernetes/issues/30243
				RestartPolicy: corev1.RestartPolicyNever,
				DNSPolicy:     corev1.DNSDefault,
			},
		},
	}

	return jobSpec, dockerImageFqn, nil
}

func (cbc *ComponentBuildController) getPullGitRepoContainer(cb *crv1.ComponentBuild) corev1.Container {
	pullGitRepoContainer := corev1.Container{
		Name:            "pull-git-repo",
		Image:           cbc.config.ComponentBuild.PullGitRepoImage,
		ImagePullPolicy: corev1.PullIfNotPresent,
		Command:         []string{"./pull-git-repo.sh"},
		Env: []corev1.EnvVar{
			{
				Name:  "WORK_DIR",
				Value: jobWorkingDirectory,
			},
			{
				Name:  "GIT_URL",
				Value: cb.Spec.BuildDefinitionBlock.GitRepository.Url,
			},
		},
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      jobWorkingDirectoryVolumeName,
				MountPath: jobWorkingDirectory,
			},
		},
	}

	if cb.Spec.BuildDefinitionBlock.GitRepository.Commit != nil {
		pullGitRepoContainer.Env = append(
			pullGitRepoContainer.Env,
			corev1.EnvVar{
				Name:  "GIT_CHECKOUT_TARGET",
				Value: *cb.Spec.BuildDefinitionBlock.GitRepository.Commit,
			},
		)
	} else {
		pullGitRepoContainer.Env = append(
			pullGitRepoContainer.Env,
			corev1.EnvVar{
				Name:  "GIT_CHECKOUT_TARGET",
				Value: *cb.Spec.BuildDefinitionBlock.GitRepository.Tag,
			},
		)
	}

	return pullGitRepoContainer
}

func (cbc *ComponentBuildController) getBuildDockerImageContainer(cb *crv1.ComponentBuild) (corev1.Container, string) {
	buildDockerImageContainer := corev1.Container{
		Name:            "build-docker-image",
		Image:           cbc.config.ComponentBuild.BuildDockerImage,
		ImagePullPolicy: corev1.PullIfNotPresent,
		Command:         []string{"./build-docker-image.sh"},
		Env: []corev1.EnvVar{
			{
				Name:  "WORK_DIR",
				Value: jobWorkingDirectory,
			},
			{
				Name:  "DOCKER_REGISTRY",
				Value: cbc.config.ComponentBuild.DockerConfig.Registry,
			},
			{
				Name:  "BUILD_CMD",
				Value: *cb.Spec.BuildDefinitionBlock.Command,
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

	repo := cbc.config.ComponentBuild.DockerConfig.Repository
	tag := cb.Name
	if cbc.config.ComponentBuild.DockerConfig.RepositoryPerImage {
		repo = cb.Name
		tag = fmt.Sprint(time.Now().Unix())
	}

	dockerImageFqn := fmt.Sprintf(
		"%v/%v:%v",
		cbc.config.ComponentBuild.DockerConfig.Registry,
		repo,
		tag,
	)

	buildDockerImageContainer.Env = append(
		buildDockerImageContainer.Env,
		// TODO: should this be Namespace/Name? should builds be namespaced?
		corev1.EnvVar{
			Name:  "DOCKER_REPOSITORY",
			Value: repo,
		},
		corev1.EnvVar{
			Name:  "DOCKER_IMAGE_TAG",
			Value: tag,
		},
	)

	push := "0"
	if cbc.config.ComponentBuild.DockerConfig.Push {
		push = "1"
	}
	buildDockerImageContainer.Env = append(
		buildDockerImageContainer.Env,
		corev1.EnvVar{
			Name:  "DOCKER_PUSH",
			Value: push,
		},
	)

	var baseImage string
	if cb.Spec.BuildDefinitionBlock.Language != nil {
		// TODO: insert custom language images when we have them
		baseImage = *cb.Spec.BuildDefinitionBlock.Language
	} else {
		baseImage = getDockerImageFqn(cb.Spec.BuildDefinitionBlock.DockerImage)
	}
	buildDockerImageContainer.Env = append(
		buildDockerImageContainer.Env,
		corev1.EnvVar{
			Name:  "DOCKER_BASE_IMAGE",
			Value: baseImage,
		},
	)

	return buildDockerImageContainer, dockerImageFqn
}

func getDockerImageFqn(di *systemdefinitionblock.DockerImage) string {
	return fmt.Sprintf("%v/%v:%v", di.Registry, di.Repository, di.Tag)
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
