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

// DeepCopyInto is a deepcopy function, copying the receiver, writing into out. in must be non-nil.
// This is implemented here as the deepcopy generator does not properly handle custom types that are
// maps.
// Originally it was calling out to val.DeepCopyInto(), where val was a *string.
func (in DockerBuildArgs) DeepCopyInto(out *DockerBuildArgs) {
	{
		in := &in
		*out = make(DockerBuildArgs, len(*in))
		for key, val := range *in {
			var outVal *string
			if val != nil {
				outVal = new(string)
				*outVal = *val
			}
			(*out)[key] = outVal
		}
		return
	}
}

type DockerBuild struct {
	BuildContext *DockerBuildContext `json:"build_context,omitempty"`
	DockerFile   *DockerFile         `json:"docker_file,omitempty"`
	BuildArgs    DockerBuildArgs     `json:"build_args,omitempty"`
	Options      *DockerBuildOptions `json:"options,omitempty"`
}

// DeepCopyInto is a deepcopy function, copying the receiver, writing into out. in must be non-nil.
// This is implemented here as the deepcopy generator does not properly handle custom types that are
// maps. See the comment in the method implementation.
// IMPORTANT: if you add any fields to DockerBuild you _must_ _must_ _must_ update this function.
// a good way to do that is to delete this function, run `make codegen.deepcopy`, copy the generated
// method to here, and add back in the manually adjusted portion outlined in the comment below.
func (in *DockerBuild) DeepCopyInto(out *DockerBuild) {
	*out = *in
	if in.BuildContext != nil {
		in, out := &in.BuildContext, &out.BuildContext
		if *in == nil {
			*out = nil
		} else {
			*out = new(DockerBuildContext)
			(*in).DeepCopyInto(*out)
		}
	}
	if in.DockerFile != nil {
		in, out := &in.DockerFile, &out.DockerFile
		if *in == nil {
			*out = nil
		} else {
			*out = new(DockerFile)
			(*in).DeepCopyInto(*out)
		}
	}
	if in.BuildArgs != nil {
		in, out := &in.BuildArgs, &out.BuildArgs
		// This was originally the inlined invalid version of DockerBuildArgs.DeepCopyInto.
		// Instead we've replaced it with just a call to DeepCopyInto.
		// Add this back if you need to regenerate this method.
		(*in).DeepCopyInto(out)
	}
	if in.Options != nil {
		in, out := &in.Options, &out.Options
		if *in == nil {
			*out = nil
		} else {
			*out = new(DockerBuildOptions)
			(*in).DeepCopyInto(*out)
		}
	}
	return
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
