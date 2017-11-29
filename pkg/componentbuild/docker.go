package componentbuild

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"golang.org/x/net/context"

	systemdefinitionblock "github.com/mlab-lattice/core/pkg/system/definition/block"

	tarutil "github.com/mlab-lattice/system/pkg/util/tar"

	dockertypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/jsonmessage"
)

func (b *Builder) buildDockerImage(sourceDirectory string) error {
	fmt.Println("Building docker image...")

	// Get Dockerfile contents and write them to the directory
	dockerfileContents, err := b.getDockerfileContents(sourceDirectory)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filepath.Join(b.WorkingDir, "Dockerfile"), []byte(dockerfileContents), 0444)
	if err != nil {
		return newErrorInternal("Dockerfile could not be written: " + err.Error())
	}

	// Tar up the directory to send it as the build context to the docker daemon
	buildContext, err := tarutil.ArchiveDirectory(b.WorkingDir)
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

	// If the image is not to be pushed, there's no more to do
	if !b.DockerOptions.Push {
		return nil
	}

	return b.pushDockerImage()
}

func (b *Builder) pushDockerImage() error {
	// Assumes the image has already been built and tagged.
	dockerImageFQN := getDockerImageFQN(b.DockerOptions.Registry, b.DockerOptions.Repository, b.DockerOptions.Tag)

	// Include creds if they were passed in
	pushOptions := dockertypes.ImagePushOptions{}
	if b.DockerOptions.RegistryAuth != nil {
		pushOptions.RegistryAuth = *b.DockerOptions.RegistryAuth
	}

	out, err := b.DockerClient.ImagePush(context.Background(), dockerImageFQN, pushOptions)
	if err != nil {
		return newErrorInternal("pushing docker image failed: " + err.Error())
	}
	defer out.Close()

	// Ignoring the potential error here, shouldn't fail the build because we couldn't log to stdout
	io.Copy(os.Stdout, out)

	return nil
}

func (b *Builder) getDockerfileContents(sourceDirectory string) (string, error) {
	if b.ComponentBuildBlock.Command == nil {
		return "", newErrorUser("component build command cannot be nil")
	}

	baseDockerImage, err := b.getBaseDockerImage()
	if err != nil {
		return "", err
	}

	relativeSourceDirectory, err := filepath.Rel(b.WorkingDir, sourceDirectory)
	if err != nil {
		return "", newErrorInternal("could not get relative source directory: " + err.Error())
	}

	dockerfileContents := fmt.Sprintf(`FROM %v

RUN mkdir -p /usr/src/app
WORKDIR /usr/src/app

COPY %v /usr/src/app

RUN %v`,
		baseDockerImage,
		relativeSourceDirectory,
		*b.ComponentBuildBlock.Command,
	)

	return dockerfileContents, nil
}

func (b *Builder) getBaseDockerImage() (string, error) {
	if b.ComponentBuildBlock.BaseDockerImage != nil {
		return getDockerImageFQNFromDockerImageBlock(b.ComponentBuildBlock.BaseDockerImage)
	}

	if b.ComponentBuildBlock.Language != nil {
		return *b.ComponentBuildBlock.Language, nil
	}

	return "", newErrorUser("component build must have base_docker_image or language")
}

func getDockerImageFQNFromDockerImageBlock(image *systemdefinitionblock.DockerImage) (string, error) {
	if image == nil {
		return "", newErrorInternal("cannot get docker image FQN from nil image")
	}

	return getDockerImageFQN(image.Registry, image.Repository, image.Tag), nil
}

func getDockerImageFQN(registry, repository, tag string) string {
	return fmt.Sprintf("%v/%v:%v", registry, repository, tag)
}
