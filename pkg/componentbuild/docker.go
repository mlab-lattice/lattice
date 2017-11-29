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

	buildResponse, err := b.DockerClient.ImageBuild(context.Background(), buildContext, buildOptions)
	if err != nil {
		return newErrorUser("docker image build failed: " + err.Error())
	}
	defer buildResponse.Body.Close()

	// Ignoring the potential error here, shouldn't fail the build because we couldn't log to stdout
	io.Copy(os.Stdout, buildResponse.Body)

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

RUN ${BUILD_CMD}`,
		baseDockerImage,
		relativeSourceDirectory,
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
