package build

import (
	"fmt"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	crv1 "github.com/mlab-lattice/kubernetes-integration/pkg/api/customresource/v1"
)

func getBuildJob(b *crv1.Build) *batchv1.Job {
	name := fmt.Sprintf("lattice-build-%s", b.Name)
	labels := map[string]string{
		"mlab.lattice.com/build": "true",
	}
	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Labels:          labels,
			Name:            name,
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(b, controllerKind)},
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
				},
				Spec: corev1.PodSpec{
					// FIXME: only needed for minikube
					ImagePullSecrets: []corev1.LocalObjectReference{
						{
							Name: "pull",
						},
					},
					Volumes: []corev1.Volume{
						// FIXME: this is minikube specific
						{
							Name: "workdir",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/data/builder",
								},
							},
						},
						// FIXME: only need this locally for minikube. ec2 instances have instance roles
						{
							Name: "awsconf",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/home/docker/.aws",
								},
							},
						},
						{
							Name: "dockersock",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/var/run/docker.sock",
								},
							},
						},
					},
					InitContainers: []corev1.Container{
						// FIXME: this is only supporting commit git repos
						{
							Name:    "pull-git-repo",
							Image:   "XXX_PULL_GIT_REPO_IMAGE_GOES_HERE",
							Command: []string{"./pull-git-repo.sh"},
							Env: []corev1.EnvVar{
								{
									Name:  "WORK_DIR",
									Value: "/var/run/builder",
								},
								{
									Name:  "GIT_URL",
									Value: b.Spec.GitRepository.Url,
								},
								{
									Name:  "GIT_CHECKOUT_TARGET",
									Value: *b.Spec.GitRepository.Commit,
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "workdir",
									MountPath: "/var/run/builder",
								},
							},
						},
						// FIXME: this is only needed when pushing to ecr
						{
							Name:    "get-ecr-creds",
							Image:   "XXX_GET_ECR_CREDS_IMAGE_GOES_HERE",
							Command: []string{"./get-ecr-creds.sh"},
							Env: []corev1.EnvVar{
								{
									Name:  "WORK_DIR",
									Value: "/var/run/builder",
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "workdir",
									MountPath: "/var/run/builder",
								},
								{
									Name:      "awsconf",
									MountPath: "/root/.aws",
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:    "build-docker-image",
							Image:   "XXX_BUILD_DOCKER_IMAGE_IMAGE_GOES_HERE",
							Command: []string{"./build-docker-image.sh"},
							Env: []corev1.EnvVar{
								{
									Name:  "WORK_DIR",
									Value: "/var/run/builder",
								},
								{
									Name:  "DOCKER_REGISTRY",
									Value: "XXX_DOCKER_REGISTRY_GOES_HERE",
								},
								{
									Name:  "DOCKER_REPOSITORY",
									Value: "XXX_DOCKER_REPO_GOES_HERE",
								},
								{
									Name:  "DOCKER_IMAGE_TAG",
									Value: "XXX_DOCKER_IMAGE_TAG_GOES_HERE",
								},
								{
									Name:  "DOCKER_BASE_IMAGE",
									Value: "XXX_DOCKER_BASE_IMAGE_GOES_HERE",
								},
								{
									Name:  "BUILD_CMD",
									Value: "XXX_BUILD_CMD_GOES_HERE",
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "workdir",
									MountPath: "/var/run/builder",
								},
								{
									Name:      "dockersock",
									MountPath: "/var/run/docker.sock",
								},
							},
						},
					},
					RestartPolicy: corev1.RestartPolicyNever,
				},
			},
		},
	}
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
