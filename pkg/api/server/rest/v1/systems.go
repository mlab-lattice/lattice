package v1

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	v1server "github.com/mlab-lattice/lattice/pkg/api/server/v1"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	v1rest "github.com/mlab-lattice/lattice/pkg/api/v1/rest"
	"github.com/mlab-lattice/lattice/pkg/definition/resolver"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	definitionv1 "github.com/mlab-lattice/lattice/pkg/definition/v1"
	"github.com/mlab-lattice/lattice/pkg/util/git"

	"io"
	"strconv"

	"github.com/gin-gonic/gin"
)

func mountSystemHandlers(router *gin.RouterGroup, backend v1server.Interface, sysResolver *resolver.SystemResolver) {

	// create-system
	router.POST(v1rest.SystemsPath, func(c *gin.Context) {
		var req v1rest.CreateSystemRequest
		if err := c.BindJSON(&req); err != nil {
			handleBadRequestBody(c)
			return
		}

		system, err := backend.CreateSystem(req.ID, req.DefinitionURL)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusCreated, system)
	})

	// list-systems
	router.GET(v1rest.SystemsPath, func(c *gin.Context) {
		systems, err := backend.ListSystems()
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, systems)
	})

	systemIdentifier := "system_id"
	systemIdentifierPathComponent := fmt.Sprintf(":%v", systemIdentifier)
	systemPath := fmt.Sprintf(v1rest.SystemPathFormat, systemIdentifierPathComponent)

	// get-system
	router.GET(systemPath, func(c *gin.Context) {
		systemID := v1.SystemID(c.Param(systemIdentifier))

		system, err := backend.GetSystem(systemID)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, system)
	})

	// delete-system
	router.DELETE(systemPath, func(c *gin.Context) {
		systemID := v1.SystemID(c.Param(systemIdentifier))

		err := backend.DeleteSystem(systemID)
		if err != nil {
			handleError(c, err)
			return
		}

		c.Status(http.StatusOK)
	})

	mountBuildHandlers(router, backend, sysResolver)
	mountDeployHandlers(router, backend, sysResolver)
	mountNodePoolHandlers(router, backend)
	mountServiceHandlers(router, backend)
	mountJobHandlers(router, backend)
	mountSecretHandlers(router, backend)
	mountTeardownHandlers(router, backend)
	mountVersionHandlers(router, backend, sysResolver)
}

func mountBuildHandlers(router *gin.RouterGroup, backend v1server.Interface, sysResolver *resolver.SystemResolver) {
	systemIdentifier := "system_id"
	systemIdentifierPathComponent := fmt.Sprintf(":%v", systemIdentifier)
	buildsPath := fmt.Sprintf(v1rest.BuildsPathFormat, systemIdentifierPathComponent)

	// build-system
	router.POST(buildsPath, func(c *gin.Context) {
		systemID := v1.SystemID(c.Param(systemIdentifier))

		var req v1rest.BuildRequest
		if err := c.BindJSON(&req); err != nil {
			handleBadRequestBody(c)
			return
		}

		root, err := getSystemDefinitionRoot(backend, sysResolver, systemID, req.Version)
		if err != nil {
			handleError(c, err)
			return
		}

		build, err := backend.Build(
			systemID,
			root,
			req.Version,
		)

		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusCreated, build)
	})

	// list-builds
	router.GET(buildsPath, func(c *gin.Context) {
		systemID := v1.SystemID(c.Param(systemIdentifier))

		builds, err := backend.ListBuilds(systemID)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, builds)
	})

	buildIdentifier := "build_id"
	buildIdentifierPathComponent := fmt.Sprintf(":%v", buildIdentifier)
	buildPath := fmt.Sprintf(v1rest.BuildPathFormat, systemIdentifierPathComponent, buildIdentifierPathComponent)

	// get-build
	router.GET(buildPath, func(c *gin.Context) {
		systemID := v1.SystemID(c.Param(systemIdentifier))
		buildID := v1.BuildID(c.Param(buildIdentifier))

		build, err := backend.GetBuild(systemID, buildID)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, build)
	})

	buildsLogPath := fmt.Sprintf(
		v1rest.BuildLogsPathFormat,
		systemIdentifierPathComponent,
		buildIdentifierPathComponent,
	)

	// get-build-logs
	router.GET(buildsLogPath, func(c *gin.Context) {
		systemID := v1.SystemID(c.Param(systemIdentifier))
		buildID := v1.BuildID(c.Param(buildIdentifier))
		path := c.Query("path")

		sidecarQuery, sidecarSet := c.GetQuery("sidecar")
		var sidecar *string
		if sidecarSet {
			sidecar = &sidecarQuery
		}

		if path == "" {
			c.Status(http.StatusBadRequest)
			return
		}

		nodePath, err := tree.NewNodePath(path)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		logOptions, err := requestedLogOptions(c)

		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		log, err := backend.BuildLogs(systemID, buildID, nodePath, sidecar, logOptions)
		if err != nil {
			handleError(c, err)
			return
		}

		if log == nil {
			c.Status(http.StatusOK)
			return
		}

		serveLogFile(log, c)
	})
}

func mountDeployHandlers(router *gin.RouterGroup, backend v1server.Interface, sysResolver *resolver.SystemResolver) {
	systemIdentifier := "system_id"
	systemIdentifierPathComponent := fmt.Sprintf(":%v", systemIdentifier)
	deploysPath := fmt.Sprintf(v1rest.DeploysPathFormat, systemIdentifierPathComponent)

	// deploy
	router.POST(deploysPath, func(c *gin.Context) {
		systemID := v1.SystemID(c.Param(systemIdentifier))

		var req v1rest.DeployRequest
		if err := c.BindJSON(&req); err != nil {
			handleBadRequestBody(c)
			return
		}

		if req.Version != nil && req.BuildID != nil {
			c.String(http.StatusBadRequest, "can only specify version or buildId")
			return
		}

		if req.Version == nil && req.BuildID == nil {
			c.String(http.StatusBadRequest, "must specify version or buildId")
			return
		}

		var deploy *v1.Deploy
		var err error
		if req.Version != nil {
			root, err := getSystemDefinitionRoot(backend, sysResolver, systemID, *req.Version)
			if err != nil {
				handleError(c, err)
				return
			}

			deploy, err = backend.DeployVersion(
				systemID,
				root,
				*req.Version,
			)
		} else {
			deploy, err = backend.DeployBuild(
				systemID,
				*req.BuildID,
			)
		}

		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusCreated, deploy)
	})

	// list-deploys
	router.GET(deploysPath, func(c *gin.Context) {
		systemID := v1.SystemID(c.Param(systemIdentifier))

		deploys, err := backend.ListDeploys(systemID)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, deploys)
	})

	deployIdentifier := "deploy_id"
	deployIdentifierPathComponent := fmt.Sprintf(":%v", deployIdentifier)
	deployPath := fmt.Sprintf(v1rest.DeployPathFormat, systemIdentifierPathComponent, deployIdentifierPathComponent)

	// get-deploy
	router.GET(deployPath, func(c *gin.Context) {
		systemID := v1.SystemID(c.Param(systemIdentifier))
		deployID := v1.DeployID(c.Param(deployIdentifier))

		deploy, err := backend.GetDeploy(v1.SystemID(systemID), v1.DeployID(deployID))
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, deploy)
	})
}

func mountNodePoolHandlers(router *gin.RouterGroup, backend v1server.Interface) {
	systemIdentifier := "system_id"
	systemIdentifierPathComponent := fmt.Sprintf(":%v", systemIdentifier)
	nodePoolsPath := fmt.Sprintf(v1rest.NodePoolsPathFormat, systemIdentifierPathComponent)

	// list-node-pools
	router.GET(nodePoolsPath, func(c *gin.Context) {
		systemID := v1.SystemID(c.Param(systemIdentifier))

		nodePools, err := backend.ListNodePools(systemID)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, nodePools)
	})

	nodePoolIdentifier := "node_pool_path"
	nodePoolIdentifierPathComponent := fmt.Sprintf(":%v", nodePoolIdentifier)
	nodePoolPath := fmt.Sprintf(v1rest.NodePoolPathFormat, systemIdentifierPathComponent, nodePoolIdentifierPathComponent)

	// get-node-pool
	router.GET(nodePoolPath, func(c *gin.Context) {
		systemID := v1.SystemID(c.Param(systemIdentifier))
		escapedNodePoolPath := c.Param(nodePoolIdentifier)

		nodePoolPathString, err := url.PathUnescape(escapedNodePoolPath)
		if err != nil {
			// FIXME: send invalid nodePool error
			c.Status(http.StatusBadRequest)
			return
		}

		path, err := v1.ParseNodePoolPath(nodePoolPathString)
		if err != nil {
			// FIXME: send invalid nodePool error
			c.Status(http.StatusBadRequest)
			return
		}

		nodePool, err := backend.GetNodePool(systemID, path)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, nodePool)
	})
}

func mountServiceHandlers(router *gin.RouterGroup, backend v1server.Interface) {
	systemIdentifier := "system_id"
	systemIdentifierPathComponent := fmt.Sprintf(":%v", systemIdentifier)
	servicesPath := fmt.Sprintf(v1rest.ServicesPathFormat, systemIdentifierPathComponent)

	// list-services
	router.GET(servicesPath, func(c *gin.Context) {
		systemID := v1.SystemID(c.Param(systemIdentifier))
		servicePathParam := c.Query("path")

		// check if its a query by service path

		if servicePathParam != "" {
			servicePath, err := tree.NewNodePath(servicePathParam)
			if err != nil {
				handleError(c, err)
				return
			}

			service, err := backend.GetServiceByPath(systemID, servicePath)

			if err != nil {
				handleError(c, err)
				return
			}

			if service == nil {
				c.Status(http.StatusBadRequest)
				return
			}

			c.JSON(http.StatusOK, []*v1.Service{service})
			return
		}

		// otherwise its just a normal list services request
		services, err := backend.ListServices(systemID)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, services)
	})

	serviceIdentifier := "service_id"
	serviceIdentifierPathComponent := fmt.Sprintf(":%v", serviceIdentifier)
	servicePath := fmt.Sprintf(v1rest.ServicePathFormat, systemIdentifierPathComponent, serviceIdentifierPathComponent)

	// get-service
	router.GET(servicePath, func(c *gin.Context) {
		systemID := v1.SystemID(c.Param(systemIdentifier))
		serviceID := v1.ServiceID(c.Param(serviceIdentifier))

		service, err := backend.GetService(systemID, serviceID)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, service)
	})

	// service component log path

	serviceLogPath := fmt.Sprintf(
		v1rest.ServiceLogsPathFormat,
		systemIdentifierPathComponent,
		serviceIdentifierPathComponent,
	)

	router.GET(serviceLogPath, func(c *gin.Context) {
		systemID := v1.SystemID(c.Param(systemIdentifier))
		serviceId := v1.ServiceID(c.Param(serviceIdentifier))
		instance := c.Query("instance")

		sidecarQuery, sidecarSet := c.GetQuery("sidecar")
		var sidecar *string
		if sidecarSet {
			sidecar = &sidecarQuery
		}

		logOptions, err := requestedLogOptions(c)

		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		log, err := backend.ServiceLogs(systemID, serviceId, sidecar, instance, logOptions)

		if err != nil {
			handleError(c, err)
			return
		}

		if log == nil {
			c.Status(http.StatusOK)
			return
		}

		serveLogFile(log, c)

	})
}

func mountJobHandlers(router *gin.RouterGroup, backend v1server.Interface) {
	systemIdentifier := "system_id"
	systemIdentifierPathComponent := fmt.Sprintf(":%v", systemIdentifier)
	jobsPath := fmt.Sprintf(v1rest.JobsPathFormat, systemIdentifierPathComponent)

	// run-job
	router.POST(jobsPath, func(c *gin.Context) {
		systemID := v1.SystemID(c.Param(systemIdentifier))

		var req v1rest.RunJobRequest
		if err := c.BindJSON(&req); err != nil {
			handleBadRequestBody(c)
			return
		}

		job, err := backend.RunJob(systemID, req.Path)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusCreated, job)
	})

	// list-jobs
	router.GET(jobsPath, func(c *gin.Context) {
		systemID := v1.SystemID(c.Param(systemIdentifier))

		jobs, err := backend.ListJobs(systemID)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, jobs)
	})

	jobIdentifier := "job_id"
	jobIdentifierPathComponent := fmt.Sprintf(":%v", jobIdentifier)
	jobPath := fmt.Sprintf(v1rest.JobPathFormat, systemIdentifierPathComponent, jobIdentifierPathComponent)

	// get-job
	router.GET(jobPath, func(c *gin.Context) {
		systemID := v1.SystemID(c.Param(systemIdentifier))
		jobID := v1.JobID(c.Param(jobIdentifier))

		job, err := backend.GetJob(systemID, jobID)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, job)
	})

	jobLogPath := fmt.Sprintf(
		v1rest.JobLogsPathFormat,
		systemIdentifierPathComponent,
		jobIdentifierPathComponent,
	)

	// get-job-logs
	router.GET(jobLogPath, func(c *gin.Context) {
		systemID := v1.SystemID(c.Param(systemIdentifier))
		jobID := v1.JobID(c.Param(jobIdentifier))

		sidecarQuery, sidecarSet := c.GetQuery("sidecar")
		var sidecar *string
		if sidecarSet {
			sidecar = &sidecarQuery
		}

		logOptions, err := requestedLogOptions(c)

		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		log, err := backend.JobLogs(systemID, jobID, sidecar, logOptions)
		if err != nil {
			handleError(c, err)
			return
		}

		if log == nil {
			c.Status(http.StatusOK)
			return
		}

		serveLogFile(log, c)
	})
}

// requestedLogOptions
func requestedLogOptions(c *gin.Context) (*v1.ContainerLogOptions, error) {
	// follow
	follow, err := strconv.ParseBool(c.DefaultQuery("follow", "false"))
	if err != nil {
		return nil, err
	}
	// previous
	previous, err := strconv.ParseBool(c.DefaultQuery("previous", "false"))
	if err != nil {
		return nil, err
	}
	//timestamps
	timestamps, err := strconv.ParseBool(c.DefaultQuery("timestamps", "false"))
	if err != nil {
		return nil, err
	}
	// tail
	var tail *int64
	tailStr := c.Query("tail")
	if tailStr != "" {
		lines, err := strconv.ParseInt(tailStr, 10, 64)
		if err != nil {
			return nil, err
		}
		tail = &lines
	}

	// since
	since := c.Query("since")

	// sinceTime
	sinceTime := c.Query("sinceTime")

	logOptions := v1.NewContainerLogOptions()
	logOptions.Follow = follow
	logOptions.Timestamps = timestamps
	logOptions.Previous = previous
	logOptions.Tail = tail
	logOptions.Since = since
	logOptions.SinceTime = sinceTime

	return logOptions, nil
}

// serveLogFile
func serveLogFile(log io.ReadCloser, c *gin.Context) {
	defer log.Close()

	buff := make([]byte, 1024)

	c.Stream(func(w io.Writer) bool {
		n, err := log.Read(buff)
		if err != nil {
			return false
		}

		w.Write(buff[:n])
		return true
	})
}

func mountSecretHandlers(router *gin.RouterGroup, backend v1server.Interface) {
	systemIdentifier := "system_id"
	systemIdentifierPathComponent := fmt.Sprintf(":%v", systemIdentifier)
	secretsPath := fmt.Sprintf(v1rest.SystemSecretsPathFormat, systemIdentifierPathComponent)

	// list-secrets
	router.GET(secretsPath, func(c *gin.Context) {
		systemID := v1.SystemID(c.Param(systemIdentifier))

		secrets, err := backend.ListSystemSecrets(systemID)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, secrets)
	})

	secretIdentifier := "secret_path"
	secretIdentifierPathComponent := fmt.Sprintf(":%v", secretIdentifier)
	secretPath := fmt.Sprintf(v1rest.SystemSecretPathFormat, systemIdentifierPathComponent, secretIdentifierPathComponent)

	// get-secret
	router.GET(secretPath, func(c *gin.Context) {
		systemID := v1.SystemID(c.Param(systemIdentifier))
		escapedSecretPath := c.Param(secretIdentifier)

		secretPathString, err := url.PathUnescape(escapedSecretPath)
		if err != nil {
			// FIXME: send invalid secret error
			c.Status(http.StatusBadRequest)
			return
		}

		splitPath := strings.Split(secretPathString, ":")
		if len(splitPath) != 2 {
			// FIXME: send invalid secret error
			c.Status(http.StatusBadRequest)
			return
		}

		path, err := tree.NewNodePath(splitPath[0])
		if err != nil {
			// FIXME: send invalid secret error
			c.Status(http.StatusBadRequest)
			return
		}

		name := splitPath[1]

		secret, err := backend.GetSystemSecret(systemID, path, name)
		if err != nil {
			// FIXME: send invalid secret error
			c.Status(http.StatusBadRequest)
			return
		}

		c.JSON(http.StatusOK, secret)
	})

	// set-secret
	router.PATCH(secretPath, func(c *gin.Context) {
		var req v1rest.SetSecretRequest
		if err := c.BindJSON(&req); err != nil {
			handleBadRequestBody(c)
			return
		}

		systemID := v1.SystemID(c.Param(systemIdentifier))
		escapedSecretPath := c.Param(secretIdentifier)

		secretPathString, err := url.PathUnescape(escapedSecretPath)
		if err != nil {
			// FIXME: send invalid secret error
			c.Status(http.StatusBadRequest)
			return
		}

		splitPath := strings.Split(secretPathString, ":")
		if len(splitPath) != 2 {
			// FIXME: send invalid secret error
			c.Status(http.StatusBadRequest)
			return
		}

		path, err := tree.NewNodePath(splitPath[0])
		if err != nil {
			// FIXME: send invalid secret error
			c.Status(http.StatusBadRequest)
			return
		}

		name := splitPath[1]

		err = backend.SetSystemSecret(systemID, path, name, req.Value)
		if err != nil {
			handleError(c, err)
			return
		}

		c.Status(http.StatusOK)
	})

	// unset-secret
	router.DELETE(secretPath, func(c *gin.Context) {
		systemID := v1.SystemID(c.Param(systemIdentifier))
		escapedSecretPath := c.Param(secretIdentifier)

		secretPathString, err := url.PathUnescape(escapedSecretPath)
		if err != nil {
			// FIXME: send invalid secret error
			c.Status(http.StatusBadRequest)
			return
		}

		splitPath := strings.Split(secretPathString, ":")
		if len(splitPath) != 2 {
			// FIXME: send invalid secret error
			c.Status(http.StatusBadRequest)
			return
		}

		path, err := tree.NewNodePath(splitPath[0])
		if err != nil {
			// FIXME: send invalid secret error
			c.Status(http.StatusBadRequest)
			return
		}

		name := splitPath[1]

		err = backend.UnsetSystemSecret(systemID, path, name)
		if err != nil {
			handleError(c, err)
			return
		}

		c.Status(http.StatusOK)
	})
}

func mountTeardownHandlers(router *gin.RouterGroup, backend v1server.Interface) {
	systemIdentifier := "system_id"
	systemIdentifierPathComponent := fmt.Sprintf(":%v", systemIdentifier)
	teardownsPath := fmt.Sprintf(v1rest.TeardownsPathFormat, systemIdentifierPathComponent)

	// tear-down-system
	router.POST(teardownsPath, func(c *gin.Context) {
		systemID := v1.SystemID(c.Param(systemIdentifier))

		teardown, err := backend.TearDown(systemID)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusCreated, teardown)
	})

	// list-teardowns
	router.GET(teardownsPath, func(c *gin.Context) {
		systemID := v1.SystemID(c.Param(systemIdentifier))

		teardowns, err := backend.ListTeardowns(systemID)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, teardowns)
	})

	teardownIdentifier := "teardown_id"
	teardownIdentifierPathComponent := fmt.Sprintf(":%v", teardownIdentifier)
	teardownPath := fmt.Sprintf(v1rest.TeardownPathFormat, systemIdentifierPathComponent, teardownIdentifierPathComponent)

	// get-teardown
	router.GET(teardownPath, func(c *gin.Context) {
		systemID := v1.SystemID(c.Param(systemIdentifier))
		teardownID := v1.TeardownID(c.Param(teardownIdentifier))

		teardown, err := backend.GetTeardown(systemID, teardownID)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, teardown)
	})
}

func mountVersionHandlers(router *gin.RouterGroup, backend v1server.Interface, sysResolver *resolver.SystemResolver) {
	systemIDIdentifier := "system_id"
	systemIDPathComponent := fmt.Sprintf(":%v", systemIDIdentifier)
	versionsPath := fmt.Sprintf(v1rest.VersionsPathFormat, systemIDPathComponent)

	// list-system-versions
	router.GET(versionsPath, func(c *gin.Context) {
		systemID := v1.SystemID(c.Param(systemIDIdentifier))

		versionStrings, err := getSystemVersions(backend, sysResolver, systemID)
		if err != nil {
			handleError(c, err)
			return
		}

		versions := make([]v1.SystemVersion, 0)
		for _, version := range versionStrings {
			versions = append(versions, v1.SystemVersion(version))
		}

		c.JSON(http.StatusOK, versions)
	})
}

func getSystemDefinitionRoot(
	backend v1server.Interface,
	sysResolver *resolver.SystemResolver,
	systemID v1.SystemID,
	version v1.SystemVersion,
) (*definitionv1.SystemNode, error) {
	system, err := backend.GetSystem(systemID)
	if err != nil {
		return nil, err
	}

	systemDefURI := fmt.Sprintf(
		"%v#%v/%s",
		system.DefinitionURL,
		version,
		definitionv1.SystemDefinitionRootPathDefault,
	)

	root, err := sysResolver.ResolveDefinition(systemDefURI, &git.Options{})
	if err != nil {
		return nil, err
	}

	if def, ok := root.(*definitionv1.SystemNode); ok {
		return def, nil
	}

	return nil, fmt.Errorf("definition is not a system")
}

func getSystemVersions(backend v1server.Interface, sysResolver *resolver.SystemResolver, systemID v1.SystemID) ([]string, error) {
	system, err := backend.GetSystem(systemID)
	if err != nil {
		return nil, err
	}

	return sysResolver.ListDefinitionVersions(system.DefinitionURL, &git.Options{})
}
