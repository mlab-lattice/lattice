package v1

const (
	DockerBuildDefaultPath = "."
)

type DockerImage struct {
	Registry   string `json:"registry,omitempty"`
	Repository string `json:"repository"`
	Tag        string `json:"tag"`
}

type DockerBuildContext struct {
	Location *Location `json:"location,omitempty"`
	Path     string    `json:"path,omitempty"`
}

type DockerFile struct {
	Location *Location `json:"location,omitempty"`
	Path     string    `json:"path,omitempty"`
}

type DockerBuildArgs map[string]*string

type DockerBuild struct {
	BuildContext *DockerBuildContext `json:"build_context,omitempty"`
	DockerFile   *DockerFile         `json:"docker_file,omitempty"`
	BuildArgs    DockerBuildArgs     `json:"build_args,omitempty"`
	Options      *DockerBuildOptions `json:"options,omitempty"`
}

type DockerBuildOptions struct {
	// do not use local cache for intermediate layers
	NoCache bool `json:"no_cache,omitempty"`
	// always attempt to pull a newer version of the image
	PullParent bool `json:"pull_parent,omitempty"`
	// add these to /etc/hosts of intermediate build images (overwritten on deploy)
	ExtraHosts []string `json:"extra_hosts,omitempty"`
	// when multiple build stages are present, this can be used to treat the stage it names as the
	// final stage for the build
	Target string `json:"target,omitempty"`

	// --------------------
	// experimental options
	// --------------------
	//
	// Platform: API 1.32+
	// Squash: API 1.25+
	//
	// -------------------
	// unsupported options
	// -------------------
	//
	// Dockerfile string: location should be controlled by us
	// SuppressOutput bool: output should be captured
	// RemoteContext string: context is managed by us, so not needed
	// Isolation container.Isolation: specify alternant isolation, only relevant on windows
	// BuildArgs map[string]*string: managed by us outside of build config
	// AuthConfigs map[string]AuthConfig: punting on this for now (see
	//             https://godoc.org/github.com/docker/docker/api/types#AuthConfig)
	// Context io.Reader: context is localized by us (could use this for tarball builds)
	// SessionID  string: not sure what this does
	// Platform string: i don't think we're supporting this? (see
	//                  https://github.com/moby/moby/issues/33854)
	// BuildID string: this seems like something we would want to control ourselves. can be used
	//                 to gracefully cancel a build
	// CacheFrom []string: CacheFrom specifies images that are used for matching cache. Images
	//                     specified here do not need to have a valid parent chain to match cache.
	// Remove bool: remove intermediate containers after successful build
	// ForceRemove bool: remove intermediate containers after build (successful or not)
	// CPUSetCPUs string: limit the build to a set of CPUs
	// CPUSetMems string: memory nodes to use on a NUMA system
	// CPUShares int64: limit the number of CPU cycles the build container gets when cycles are
	//                  constrained (default 1024)
	// CPUQuota int64: limit the CPU CFS (completely fair scheduler) quota
	// CPUPeriod int64: limit the CPU CFS (completely fair scheduler) period
	// Memory int64: limit memory available to build
	// MemorySwap int64: limit swap available to build (-1 for unlimited)
	// CgroupParent string: open issue for documentation here:
	//                      https://github.com/moby/moby/issues/12849
	//                      run containers used during the build process in this cgroup
	// NetworkMode string: set networking mode for RUN instructions during build (default is
	//                     "bridge", other options include: "none", "container:<name|id>", "host",
	//                     "<network-name|network-id>")
	// ShmSize int64: set the size of /dev/shm (i.e., tmpfs), format is "<number><unit>" where
	//                "<unit>" is "b", "k", "m", or "g" (default is 64m). can be used to speed
	//                builds up, supposedly.
	// Ulimits []*units.Ulimit: run intermediate containers using these ulimits
	// Squash bool: squash the resulting image's layers to the parent preserves the original
	//              image and creates a new one from the parent with all the changes applied
	//              to a single layer
	// SecurityOpt []string: set various security parameters for intermediate images (see
	//                       https://docs.docker.com/engine/reference/run/#security-configuration)
	// Version BuilderVersion: specifies the version of the unerlying builder to use. only two
	//                         options at present, "1" for v1 and "2" for experimental
	//                         BuilderBuildKit
	// Labels map[string]string: image metadata
	// Tags []string: additional tags
}
