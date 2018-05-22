package v1

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	v1server "github.com/mlab-lattice/lattice/pkg/api/server/v1"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	v1rest "github.com/mlab-lattice/lattice/pkg/api/v1/rest"
	"github.com/mlab-lattice/lattice/pkg/definition"
	"github.com/mlab-lattice/lattice/pkg/definition/resolver"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	"github.com/mlab-lattice/lattice/pkg/util/git"

	"github.com/gin-gonic/gin"
	"io"
	"io/ioutil"
	"strconv"
)

func mountSystemHandlers(router *gin.Engine, backend v1server.Interface, sysResolver *resolver.SystemResolver) {
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

	mountVersionHandlers(router, backend, sysResolver)
	mountBuildHandlers(router, backend, sysResolver)
	mountDeployHandlers(router, backend, sysResolver)
	mountNodePoolHandlers(router, backend)
	mountServiceHandlers(router, backend)
	mountSecretHandlers(router, backend)
	mountTeardownHandlers(router, backend)
}

func mountBuildHandlers(router *gin.Engine, backend v1server.Interface, sysResolver *resolver.SystemResolver) {
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

	componentIdentifier := "component"
	componentIdentifierPathComponent := fmt.Sprintf(":%v", componentIdentifier)
	componentLogPath := fmt.Sprintf(
		v1rest.BuildLogPathFormat,
		systemIdentifierPathComponent,
		buildIdentifierPathComponent,
		componentIdentifierPathComponent,
	)

	// get-build-logs
	router.GET(componentLogPath, func(c *gin.Context) {
		systemID := v1.SystemID(c.Param(systemIdentifier))
		buildID := v1.BuildID(c.Param(buildIdentifier))
		component := c.Param(componentIdentifier)
		followQuery := c.DefaultQuery("follow", "false")

		follow, err := strconv.ParseBool(followQuery)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		parts := strings.Split(component, ":")
		if len(parts) != 2 {
			c.Status(http.StatusBadRequest)
			return
		}

		path, err := tree.NewNodePath(parts[0])
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		component = parts[1]

		log, err := backend.BuildLogs(systemID, buildID, path, component, follow)
		if err != nil {
			handleError(c, err)
			return
		}

		if log == nil {
			c.Status(http.StatusOK)
			return
		}

		defer log.Close()
		if !follow {
			logContents, err := ioutil.ReadAll(log)
			if err != nil {
				c.String(http.StatusInternalServerError, "")
				return
			}
			c.String(http.StatusOK, string(logContents))
			return
		}

		// FIXME: totally arbitrary buffer size choice
		buf := make([]byte, 1024*4)
		c.Stream(func(w io.Writer) bool {
			n, err := log.Read(buf)
			if err != nil {
				return false
			}

			w.Write(buf[:n])
			return true
		})
	})
}

func mountDeployHandlers(router *gin.Engine, backend v1server.Interface, sysResolver *resolver.SystemResolver) {
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

func mountNodePoolHandlers(router *gin.Engine, backend v1server.Interface) {
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

func mountServiceHandlers(router *gin.Engine, backend v1server.Interface) {
	systemIdentifier := "system_id"
	systemIdentifierPathComponent := fmt.Sprintf(":%v", systemIdentifier)
	servicesPath := fmt.Sprintf(v1rest.ServicesPathFormat, systemIdentifierPathComponent)

	// list-services
	router.GET(servicesPath, func(c *gin.Context) {
		systemID := v1.SystemID(c.Param(systemIdentifier))

		services, err := backend.ListServices(systemID)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, services)
	})

	serviceIdentifier := "service_path"
	serviceIdentifierPathComponent := fmt.Sprintf(":%v", serviceIdentifier)
	servicePath := fmt.Sprintf(v1rest.ServicePathFormat, systemIdentifierPathComponent, serviceIdentifierPathComponent)

	// get-service
	router.GET(servicePath, func(c *gin.Context) {
		systemID := v1.SystemID(c.Param(systemIdentifier))
		escapedServicePath := c.Param(serviceIdentifier)

		servicePathString, err := url.PathUnescape(escapedServicePath)
		if err != nil {
			// FIXME: send invalid service error
			c.Status(http.StatusBadRequest)
			return
		}

		servicePath, err := tree.NewNodePath(servicePathString)
		if err != nil {
			// FIXME: send invalid service error
			c.Status(http.StatusBadRequest)
			return
		}

		service, err := backend.GetService(systemID, servicePath)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, service)
	})
}

func mountSecretHandlers(router *gin.Engine, backend v1server.Interface) {
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

func mountTeardownHandlers(router *gin.Engine, backend v1server.Interface) {
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

func mountVersionHandlers(router *gin.Engine, backend v1server.Interface, sysResolver *resolver.SystemResolver) {
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

func getSystemDefinitionRoot(backend v1server.Interface, sysResolver *resolver.SystemResolver, systemID v1.SystemID, version v1.SystemVersion) (tree.Node, error) {
	system, err := backend.GetSystem(systemID)
	if err != nil {
		return nil, err
	}

	systemDefURI := fmt.Sprintf(
		"%v#%v/%s",
		system.DefinitionURL,
		version,
		definition.SystemDefinitionRootPathDefault,
	)

	return sysResolver.ResolveDefinition(systemDefURI, &git.Options{})
}

func getSystemVersions(backend v1server.Interface, sysResolver *resolver.SystemResolver, systemID v1.SystemID) ([]string, error) {
	system, err := backend.GetSystem(systemID)
	if err != nil {
		return nil, err
	}

	return sysResolver.ListDefinitionVersions(system.DefinitionURL, &git.Options{})
}
