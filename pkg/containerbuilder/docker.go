package containerbuilder

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	definitionv1 "github.com/mlab-lattice/lattice/pkg/definition/v1"
	"github.com/mlab-lattice/lattice/pkg/util/tar"

	dockertypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/fatih/color"
)

func (b *Builder) buildDockerImageContainer(image *definitionv1.DockerImage) error {
	sourceDockerImageFQN, err := getDockerImageFQNFromDockerImageBlock(image)
	if err != nil {
		return err
	}

	err = b.pullDockerImage(sourceDockerImageFQN)
	if err != nil {
		return err
	}

	err = b.tagDockerImage(sourceDockerImageFQN)
	if err != nil {
		return err
	}

	// If the image is not to be pushed, there's no more to do
	if !b.DockerOptions.Push {
		return nil
	}

	return b.pushDockerImage()
}

func (b *Builder) buildDockerImage(sourceDirectory, baseImage, dockerfileCommand string) error {
	color.Blue("Building docker image...")

	if b.StatusUpdater != nil {
		// For now ignore status update errors, don't need to fail a build because the status could
		// not be updated.
		b.StatusUpdater.UpdateProgress(b.BuildID, b.SystemID, v1.ContainerBuildPhaseBuildingDockerImage)
	}

	// Get Dockerfile contents and write them to the directory
	dockerfileContents, err := b.getDockerfileContents(sourceDirectory, baseImage, dockerfileCommand)
	if err != nil {
		return newErrorInternal("could not get Dockerfile contents: " + err.Error())
	}

	err = ioutil.WriteFile(filepath.Join(b.WorkingDir, "Dockerfile"), []byte(dockerfileContents), 0444)
	if err != nil {
		return newErrorInternal("Dockerfile could not be written: " + err.Error())
	}

	// Tar up the directory to send it as the build context to the docker daemon
	buildContext, err := tar.ArchiveDirectory(b.WorkingDir)
	if err != nil {
		return newErrorInternal("docker build context could not be tar-ed: " + err.Error())
	}

	// Tag the image to be built with the desired FQN
	dockerImageFQN := getDockerImageFQN(b.DockerOptions.Registry, b.DockerOptions.Repository, b.DockerOptions.Tag)
	buildOptions := dockertypes.ImageBuildOptions{
		Tags: []string{dockerImageFQN},
	}

	response, err := b.DockerClient.ImageBuild(context.Background(), buildContext, buildOptions)
	if err != nil {
		// The build should at least be able to be sent to the daemon even if the user has an error, so
		// if this fails, label it as internal.
		return newErrorInternal("docker image build request failed: " + err.Error())
	}
	defer response.Body.Close()

	// A little help here from https://github.com/docker/cli/blob/1ff73f867df382cb5a19df4579da3570f4daaff5/cli/command/image/build.go#L393-L426
	err = jsonmessage.DisplayJSONMessagesStream(response.Body, os.Stdout, os.Stdout.Fd(), true, nil)
	if err != nil {
		if jerr, ok := err.(*jsonmessage.JSONError); ok {
			// Build failed with a message, report this message as a user error.
			return newErrorUser("docker image build failed: " + jerr.Message)
		}

		// If the displaying of the stream failed, it cannot be told whether the build succeeded or failed.
		// Report this as a user error rather than swallowing it to take no chances.
		return newErrorInternal("docker image build stream failed: " + err.Error())
	}

	color.Green("✓ Success!")
	fmt.Println()

	// If the image is not to be pushed, there's no more to do
	if !b.DockerOptions.Push {
		return nil
	}

	return b.pushDockerImage()
}

func (b *Builder) getDockerfileContents(sourceDirectory, baseImage, dockerfileCommand string) (string, error) {
	relativeSourceDirectory, err := filepath.Rel(b.WorkingDir, sourceDirectory)
	if err != nil {
		return "", newErrorInternal("could not get relative source directory: " + err.Error())
	}

	dockerfileContents := fmt.Sprintf(`FROM %v

RUN mkdir -p /usr/src/app
WORKDIR /usr/src/app

COPY %v /usr/src/app

%v`,
		baseImage,
		relativeSourceDirectory,
		dockerfileCommand,
	)

	return dockerfileContents, nil
}

func getDockerImageFQNFromDockerImageBlock(image *definitionv1.DockerImage) (string, error) {
	if image == nil {
		return "", newErrorInternal("cannot get docker image FQN from nil image")
	}

	return getDockerImageFQN(image.Registry, image.Repository, image.Tag), nil
}

func (b *Builder) pullDockerImage(dockerImageFQN string) error {
	color.Blue("Pulling docker image...")

	if b.StatusUpdater != nil {
		// For now ignore status update errors, don't need to fail a build because the status could
		// not be updated.
		b.StatusUpdater.UpdateProgress(b.BuildID, b.SystemID, v1.ContainerBuildPhasePullingDockerImage)
	}

	// TODO: add support for registry creds
	pullOptions := dockertypes.ImagePullOptions{}
	responseBody, err := b.DockerClient.ImagePull(context.Background(), dockerImageFQN, pullOptions)
	if err != nil {
		return newErrorUser("pulling docker image failed: " + err.Error())
	}
	defer responseBody.Close()

	// A little help here from https://github.com/docker/cli/blob/1ff73f867df382cb5a19df4579da3570f4daaff5/cli/command/image/build.go#L393-L426
	err = jsonmessage.DisplayJSONMessagesStream(responseBody, os.Stdout, os.Stdout.Fd(), true, nil)
	if err != nil {
		if jerr, ok := err.(*jsonmessage.JSONError); ok {
			// Build failed with a message, report this message as an internal error.
			return newErrorUser("docker image pull failed: " + jerr.Message)
		}

		// If the displaying of the stream failed, it cannot be told whether the push succeeded or failed.
		// Report this as a user error rather than swallowing it to take no chances.
		return newErrorInternal("docker image pull stream failed: " + err.Error())
	}

	color.Green("✓ Success!")
	fmt.Println()

	return nil
}

func (b *Builder) tagDockerImage(sourceDockerImageFQN string) error {
	color.Blue("Tagging docker image...")

	targetDockerImageFQN := getDockerImageFQN(b.DockerOptions.Registry, b.DockerOptions.Repository, b.DockerOptions.Tag)

	err := b.DockerClient.ImageTag(context.Background(), sourceDockerImageFQN, targetDockerImageFQN)
	if err != nil {
		return newErrorInternal("failed to tag docker image: " + err.Error())
	}

	color.Green("✓ Success!")
	fmt.Println()

	return nil
}

func (b *Builder) pushDockerImage() error {
	color.Blue("Pushing docker image...")

	if b.StatusUpdater != nil {
		// For now ignore status update errors, don't need to fail a build because the status could
		// not be updated.
		b.StatusUpdater.UpdateProgress(b.BuildID, b.SystemID, v1.ContainerBuildPhasePushingDockerImage)
	}

	// Assumes the image has already been built and tagged.
	dockerImageFQN := getDockerImageFQN(b.DockerOptions.Registry, b.DockerOptions.Repository, b.DockerOptions.Tag)

	// Include creds if they were passed in
	pushOptions := dockertypes.ImagePushOptions{}
	if b.DockerOptions.RegistryAuthProvider != nil {
		user, pass, err := b.DockerOptions.RegistryAuthProvider.GetLoginCredentials(b.DockerOptions.Registry)
		if err != nil {
			return newErrorInternal("failed to retrieve registry auth token: " + err.Error())
		}

		// A little help here from https://github.com/docker/cli/blob/042575aac918c90e9838c67c9ac9e2ff2810c326/cli/command/image/trust.go#L168-L180
		// and https://github.com/docker/cli/blob/74af31be7f2d956a021f097af894ed9adf89272f/cli/command/registry.go#L40-L47
		authConfig := dockertypes.AuthConfig{
			Username:      user,
			Password:      pass,
			ServerAddress: "https://" + b.DockerOptions.Registry,
		}

		buf, err := json.Marshal(authConfig)
		if err != nil {
			return newErrorInternal("marshalling auth config to JSON failed: " + err.Error())
		}

		pushOptions.RegistryAuth = base64.URLEncoding.EncodeToString(buf)
	}

	responseBody, err := b.DockerClient.ImagePush(context.Background(), dockerImageFQN, pushOptions)
	if err != nil {
		return newErrorInternal("pushing docker image failed: " + err.Error())
	}
	defer responseBody.Close()

	// A little help here from https://github.com/docker/cli/blob/1ff73f867df382cb5a19df4579da3570f4daaff5/cli/command/image/build.go#L393-L426
	err = jsonmessage.DisplayJSONMessagesStream(responseBody, os.Stdout, os.Stdout.Fd(), true, nil)
	if err != nil {
		if jerr, ok := err.(*jsonmessage.JSONError); ok {
			// Build failed with a message, report this message as an internal error.
			return newErrorInternal("docker image push failed: " + jerr.Message)
		}

		// If the displaying of the stream failed, it cannot be told whether the push succeeded or failed.
		// Report this as a user error rather than swallowing it to take no chances.
		return newErrorInternal("docker image push stream failed: " + err.Error())
	}

	color.Green("✓ Success!")
	fmt.Println()

	return nil
}

func getDockerImageFQN(registry, repository, tag string) string {
	if registry == "" {
		return fmt.Sprintf("%v:%v", repository, tag)
	}
	return fmt.Sprintf("%v/%v:%v", registry, repository, tag)
}
