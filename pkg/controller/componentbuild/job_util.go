package componentbuild

import (
	"fmt"
	"time"

	providerutils "github.com/mlab-lattice/core/pkg/provider"
	"github.com/mlab-lattice/core/pkg/system/definition/block"

	crv1 "github.com/mlab-lattice/kubernetes-integration/pkg/api/customresource/v1"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	jobLocalWorkingDirectoryVolumePathPrefix = "/data/builder"
	jobWorkingDirectory                      = "/var/run/builder"
	jobWorkingDirectoryVolumeName            = "workdir"

	jobDockerSocketVolumePath = "/var/run/docker.sock"
	jobDockerSocketPath       = "/var/run/docker.sock"
	jobDockerSocketVolumeName = "dockersock"

	jobDockerFqnAnnotationKey = "docker-image-fqn"
)

func getBuildJobName(b *crv1.ComponentBuild) string {
	return fmt.Sprintf("lattice-build-%s", b.Name)
}

func (cbc *ComponentBuildController) getBuildJob(cBuild *crv1.ComponentBuild) *batchv1.Job {
	// Need a consistent view of our config while generating the Job
	cbc.configLock.RLock()
	defer cbc.configLock.RUnlock()

	name := getBuildJobName(cBuild)

	// FIXME: get job spec for build.DockerImage as well
	jobSpec, dockerImageFqn := cbc.getGitRepositoryBuildJobSpec(cBuild)

	labels := map[string]string{
		crv1.ComponentBuildJobLabelKey: "true",
	}
	annotations := map[string]string{
		jobDockerFqnAnnotationKey: dockerImageFqn,
	}

	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Annotations:     annotations,
			Labels:          labels,
			Name:            name,
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(cBuild, controllerKind)},
		},
		Spec: jobSpec,
	}
}

func (cbc *ComponentBuildController) getGitRepositoryBuildJobSpec(cBuild *crv1.ComponentBuild) (batchv1.JobSpec, string) {
	pullGitRepoContainer := cbc.getPullGitRepoContainer(cBuild)
	authorizeDockerContainer := cbc.getAuthorizeDockerContainer()
	buildDockerImageContainer, dockerImageFqn := cbc.getBuildDockerImageContainer(cBuild)
	name := getBuildJobName(cBuild)

	var workingDirectoryVolumeSource corev1.VolumeSource
	switch cbc.provider {
	case providerutils.Local:
		workingDirectoryVolumeSource = corev1.VolumeSource{
			HostPath: &corev1.HostPathVolumeSource{
				Path: fmt.Sprintf("%v/%v", jobLocalWorkingDirectoryVolumePathPrefix, name),
			},
		}
	default:
		panic("unreachable")
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

	if authorizeDockerContainer != nil {
		jobSpec.Template.Spec.InitContainers = append(
			jobSpec.Template.Spec.InitContainers,
			*authorizeDockerContainer,
		)
	}

	return jobSpec, dockerImageFqn
}

func (cbc *ComponentBuildController) getPullGitRepoContainer(cBuild *crv1.ComponentBuild) corev1.Container {
	pullGitRepoContainer := corev1.Container{
		Name:    "pull-git-repo",
		Image:   cbc.config.PullGitRepoImage,
		Command: []string{"./pull-git-repo.sh"},
		Env: []corev1.EnvVar{
			{
				Name:  "WORK_DIR",
				Value: jobWorkingDirectory,
			},
			{
				Name:  "GIT_URL",
				Value: cBuild.Spec.BuildDefinitionBlock.GitRepository.Url,
			},
		},
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      jobWorkingDirectoryVolumeName,
				MountPath: jobWorkingDirectory,
			},
		},
	}

	if cBuild.Spec.BuildDefinitionBlock.GitRepository.Commit != nil {
		pullGitRepoContainer.Env = append(
			pullGitRepoContainer.Env,
			corev1.EnvVar{
				Name:  "GIT_CHECKOUT_TARGET",
				Value: *cBuild.Spec.BuildDefinitionBlock.GitRepository.Commit,
			},
		)
	} else {
		pullGitRepoContainer.Env = append(
			pullGitRepoContainer.Env,
			corev1.EnvVar{
				Name:  "GIT_CHECKOUT_TARGET",
				Value: *cBuild.Spec.BuildDefinitionBlock.GitRepository.Tag,
			},
		)
	}

	return pullGitRepoContainer
}

func (cbc *ComponentBuildController) getAuthorizeDockerContainer() *corev1.Container {
	switch cbc.provider {
	case providerutils.AWS:
		authorizeEcrContainer := cbc.getAuthorizeEcrContainer()
		return &authorizeEcrContainer
	case providerutils.Local:
		return nil
	default:
		panic("unreachable")
	}
}

func (cbc *ComponentBuildController) getAuthorizeEcrContainer() corev1.Container {
	return corev1.Container{
		Name:    "get-ecr-creds",
		Image:   cbc.config.AuthorizeDockerImage,
		Command: []string{"./get-ecr-creds.sh"},
		Env: []corev1.EnvVar{
			{
				Name:  "WORK_DIR",
				Value: jobWorkingDirectory,
			},
		},
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      jobWorkingDirectoryVolumeName,
				MountPath: jobWorkingDirectory,
			},
		},
	}
}

func (cbc *ComponentBuildController) getBuildDockerImageContainer(cBuild *crv1.ComponentBuild) (corev1.Container, string) {
	buildDockerImageContainer := corev1.Container{
		Name:    "build-docker-image",
		Image:   cbc.config.BuildDockerImage,
		Command: []string{"./build-docker-image.sh"},
		Env: []corev1.EnvVar{
			{
				Name:  "WORK_DIR",
				Value: jobWorkingDirectory,
			},
			{
				Name:  "DOCKER_REGISTRY",
				Value: cbc.config.DockerConfig.Registry,
			},
			{
				Name:  "BUILD_CMD",
				Value: *cBuild.Spec.BuildDefinitionBlock.Command,
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

	repo := cbc.config.DockerConfig.Repository
	tag := cBuild.Name
	if cbc.config.DockerConfig.RepositoryPerImage {
		repo = cBuild.Name
		tag = fmt.Sprint(time.Now().Unix())
	}

	dockerImageFqn := fmt.Sprintf(
		"%v/%v:%v",
		cbc.config.DockerConfig.Registry,
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
	if cbc.config.DockerConfig.Push {
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
	if cBuild.Spec.BuildDefinitionBlock.Language != nil {
		// TODO: insert custom language images when we have them
		baseImage = *cBuild.Spec.BuildDefinitionBlock.Language
	} else {
		baseImage = getDockerImageFqn(cBuild.Spec.BuildDefinitionBlock.DockerImage)
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

func getDockerImageFqn(dockerImage *block.DockerImage) string {
	return fmt.Sprintf("%v/%v:%v", dockerImage.Registry, dockerImage.Repository, dockerImage.Tag)
}

func jobStatus(job *batchv1.Job) (finished bool, succeeded bool) {
	for _, c := range job.Status.Conditions {
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
