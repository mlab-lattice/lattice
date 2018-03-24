package system

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/mlab-lattice/system/pkg/apiserver/server"
	"github.com/mlab-lattice/system/pkg/definition"
	"github.com/mlab-lattice/system/pkg/definition/resolver"
	"github.com/mlab-lattice/system/pkg/definition/tree"
	"github.com/mlab-lattice/system/pkg/types"
	"github.com/mlab-lattice/system/pkg/util/git"

	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
)

type CreateRequest struct {
	ID            types.SystemID `json:"id"`
	DefinitionURL string         `json:"definitionUrl"`
}

func MountHandlers(router *gin.Engine, backend server.Backend, sysResolver *resolver.SystemResolver) {
	systems := router.Group("/systems")
	{
		// create-system
		systems.POST("", func(c *gin.Context) {
			var req CreateRequest
			if err := c.BindJSON(&req); err != nil {
				handleInternalError(c, err)
				return
			}

			system, err := backend.CreateSystem(req.ID, req.DefinitionURL)
			if err != nil {
				handleInternalError(c, err)
				return
			}

			if err != nil {
				handleInternalError(c, err)
				return
			}

			c.JSON(http.StatusCreated, system)
		})

		// list-systems
		systems.GET("", func(c *gin.Context) {
			systems, err := backend.ListSystems()

			if err != nil {
				handleInternalError(c, err)
				return
			}

			c.JSON(http.StatusOK, systems)
		})

		// get-system
		systems.GET("/:system_id", func(c *gin.Context) {
			systemID := c.Param("system_id")

			system, exists, err := backend.GetSystem(types.SystemID(systemID))
			if err != nil {
				handleInternalError(c, err)
				return
			}

			if !exists {
				c.String(http.StatusNotFound, "")
				return
			}

			c.JSON(http.StatusOK, system)
		})

		// delete-system
		systems.DELETE("/:system_id", func(c *gin.Context) {
			systemID := c.Param("system_id")

			err := backend.DeleteSystem(types.SystemID(systemID))
			if err != nil {
				handleInternalError(c, err)
				return
			}

			c.Status(http.StatusOK)
		})
	}

	mountVersionHandlers(router, backend, sysResolver)
	mountBuildHandlers(router, backend, sysResolver)
	mountDeployHandlers(router, backend, sysResolver)
	mountTeardownHandlers(router, backend)
	mountServiceHandlers(router, backend)
	mountSecretHandlers(router, backend)
}

type VersionResponse struct {
	ID         string               `json:"id"`
	Definition definition.Interface `json:"definition"`
}

func mountVersionHandlers(router *gin.Engine, backend server.Backend, sysResolver *resolver.SystemResolver) {
	versions := router.Group("/systems/:system_id/versions")
	{
		// list-system-versions
		versions.GET("", func(c *gin.Context) {
			systemID := c.Param("system_id")

			versions, err := getSystemVersions(backend, sysResolver, systemID)
			if err != nil {
				handleInternalError(c, err)
				return
			}

			c.JSON(http.StatusOK, versions)
		})

		// get-system-version
		versions.GET("/:version_id", func(c *gin.Context) {
			systemID := c.Param("system_id")
			version := c.Param("version_id")

			definitionRoot, err := getSystemDefinitionRoot(backend, sysResolver, systemID, version)
			if err != nil {
				handleInternalError(c, err)
				return
			}

			c.JSON(http.StatusOK, VersionResponse{
				ID: version,
				// FIXME: this probalby won't work
				Definition: definitionRoot,
			})
		})
	}
}

type BuildRequest struct {
	Version string `json:"version"`
}

type BuildResponse struct {
	ID types.BuildID `json:"id"`
}

func mountBuildHandlers(router *gin.Engine, backend server.Backend, sysResolver *resolver.SystemResolver) {
	systemBuilds := router.Group("/systems/:system_id/builds")
	{
		// build-system
		systemBuilds.POST("", func(c *gin.Context) {
			systemID := c.Param("system_id")

			var req BuildRequest
			if err := c.BindJSON(&req); err != nil {
				handleInternalError(c, err)
				return
			}

			root, err := getSystemDefinitionRoot(backend, sysResolver, systemID, req.Version)
			if err != nil {
				handleInternalError(c, err)
				return
			}

			bid, err := backend.Build(
				types.SystemID(systemID),
				root,
				types.SystemVersion(req.Version),
			)

			if err != nil {
				handleInternalError(c, err)
				return
			}

			c.JSON(http.StatusCreated, BuildResponse{
				ID: bid,
			})
		})

		// list-system-builds
		systemBuilds.GET("", func(c *gin.Context) {
			systemID := c.Param("system_id")

			bs, err := backend.ListBuilds(types.SystemID(systemID))
			if err != nil {
				handleInternalError(c, err)
				return
			}

			c.JSON(http.StatusOK, bs)
		})

		// get-system-build
		systemBuilds.GET("/:build_id", func(c *gin.Context) {
			systemID := c.Param("system_id")
			buildID := c.Param("build_id")

			b, exists, err := backend.GetBuild(types.SystemID(systemID), types.BuildID(buildID))
			if err != nil {
				handleInternalError(c, err)
				return
			}

			if !exists {
				c.String(http.StatusNotFound, "")
				return
			}

			c.JSON(http.StatusOK, b)
		})
	}
}

type DeployRequest struct {
	Version *string        `json:"version,omitempty"`
	BuildID *types.BuildID `json:"buildId,omitempty"`
}

type DeployResponse struct {
	ID types.DeployID `json:"id"`
}

func mountDeployHandlers(router *gin.Engine, backend server.Backend, sysResolver *resolver.SystemResolver) {
	deploys := router.Group("/systems/:system_id/deploys")
	{
		// roll-out-system
		deploys.POST("", func(c *gin.Context) {
			systemID := c.Param("system_id")

			var req DeployRequest
			if err := c.BindJSON(&req); err != nil {
				handleInternalError(c, err)
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

			var deployID types.DeployID
			var err error
			if req.Version != nil {
				root, err := getSystemDefinitionRoot(backend, sysResolver, systemID, *req.Version)
				if err != nil {
					handleInternalError(c, err)
					return
				}

				deployID, err = backend.DeployVersion(
					types.SystemID(systemID),
					root,
					types.SystemVersion(*req.Version),
				)
			} else {
				deployID, err = backend.DeployBuild(
					types.SystemID(systemID),
					types.BuildID(*req.BuildID),
				)
			}

			if err != nil {
				handleInternalError(c, err)
				return
			}

			c.JSON(http.StatusCreated, DeployResponse{
				ID: deployID,
			})
		})

		// list-deploys
		deploys.GET("", func(c *gin.Context) {
			systemID := c.Param("system_id")

			rollouts, err := backend.ListDeploys(types.SystemID(systemID))
			if err != nil {
				handleInternalError(c, err)
				return
			}

			c.JSON(http.StatusOK, rollouts)
		})

		// get-rollout
		deploys.GET("/:deploy_id", func(c *gin.Context) {
			systemID := c.Param("system_id")
			deployID := c.Param("deploy_id")

			rollout, exists, err := backend.GetDeploy(types.SystemID(systemID), types.DeployID(deployID))
			if err != nil {
				handleInternalError(c, err)
				return
			}

			if !exists {
				c.String(http.StatusNotFound, "")
				return
			}

			c.JSON(http.StatusOK, rollout)
		})
	}
}

type TearDownResponse struct {
	ID types.TeardownID `json:"id"`
}

func mountTeardownHandlers(router *gin.Engine, backend server.Backend) {
	teardowns := router.Group("/systems/:system_id/teardowns")
	{
		// tear-down-system
		teardowns.POST("", func(c *gin.Context) {
			systemID := c.Param("system_id")

			teardownID, err := backend.TearDown(types.SystemID(systemID))

			if err != nil {
				handleInternalError(c, err)
				return
			}

			c.JSON(http.StatusCreated, TearDownResponse{
				ID: teardownID,
			})
		})

		// list-teardowns
		teardowns.GET("", func(c *gin.Context) {
			systemID := c.Param("system_id")

			teardowns, err := backend.ListTeardowns(types.SystemID(systemID))
			if err != nil {
				handleInternalError(c, err)
				return
			}

			c.JSON(http.StatusOK, teardowns)
		})

		// get-teardown
		teardowns.GET("/:teardown_id", func(c *gin.Context) {
			systemID := c.Param("system_id")
			teardownID := c.Param("teardown_id")

			teardown, exists, err := backend.GetTeardown(types.SystemID(systemID), types.TeardownID(teardownID))
			if err != nil {
				handleInternalError(c, err)
				return
			}

			if !exists {
				c.String(http.StatusNotFound, "")
				return
			}

			c.JSON(http.StatusOK, teardown)
		})
	}
}

func mountServiceHandlers(router *gin.Engine, backend server.Backend) {
	services := router.Group("/systems/:system_id/services")
	{
		// list-services
		services.GET("", func(c *gin.Context) {
			systemID := c.Param("system_id")

			services, err := backend.ListServices(types.SystemID(systemID))
			if err != nil {
				handleInternalError(c, err)
				return
			}

			c.JSON(http.StatusOK, services)
		})

		// get-service
		services.GET("/:svc_domain", func(c *gin.Context) {
			systemID := c.Param("system_id")
			serviceDomain := c.Param("svc_domain")

			servicePath, err := tree.NodePathFromDomain(serviceDomain)
			if err != nil {
				handleInternalError(c, err)
				return
			}

			service, err := backend.GetService(types.SystemID(systemID), servicePath)
			if err != nil {
				handleInternalError(c, err)
				return
			}

			if service == nil {
				c.String(http.StatusNotFound, "")
				return
			}

			c.JSON(http.StatusOK, service)
		})
	}
}

type SetSecretRequest struct {
	Value string `json:"value"`
}

func mountSecretHandlers(router *gin.Engine, backend server.Backend) {
	secrets := router.Group("/systems/:system_id/secrets")
	{
		// list-secrets
		secrets.GET("", func(c *gin.Context) {
			systemID := c.Param("system_id")

			services, err := backend.ListSecrets(types.SystemID(systemID))
			if err != nil {
				handleInternalError(c, err)
				return
			}

			c.JSON(http.StatusOK, services)
		})

		// get-secret
		secrets.GET("/:secret_path", func(c *gin.Context) {
			systemID := c.Param("system_id")
			secretPath := c.Param("secret_path")

			splitPath := strings.Split(secretPath, ":")
			if len(splitPath) != 2 {
				c.Status(http.StatusBadRequest)
				return
			}

			path, err := tree.NodePathFromDomain(splitPath[0])
			if err != nil {
				handleInternalError(c, err)
				return
			}

			name := splitPath[1]

			secret, exists, err := backend.GetSecret(types.SystemID(systemID), path, name)
			if err != nil {
				handleInternalError(c, err)
				return
			}

			if !exists {
				c.String(http.StatusNotFound, "")
				return
			}

			c.JSON(http.StatusOK, secret)
		})

		// set-secret
		secrets.PATCH("/:secret_path", func(c *gin.Context) {
			var req SetSecretRequest
			if err := c.BindJSON(&req); err != nil {
				handleInternalError(c, err)
				return
			}

			systemID := c.Param("system_id")
			secretPath := c.Param("secret_path")

			splitPath := strings.Split(secretPath, ":")
			if len(splitPath) != 2 {
				c.Status(http.StatusBadRequest)
				return
			}

			path, err := tree.NodePathFromDomain(splitPath[0])
			if err != nil {
				handleInternalError(c, err)
				return
			}

			name := splitPath[1]

			err = backend.SetSecret(types.SystemID(systemID), path, name, req.Value)
			if err != nil {
				handleInternalError(c, err)
				return
			}

			c.Status(http.StatusOK)
		})

		// unset-secret
		secrets.DELETE("/:secret_path", func(c *gin.Context) {
			systemID := c.Param("system_id")
			secretPath := c.Param("secret_path")

			splitPath := strings.Split(secretPath, ":")
			if len(splitPath) != 2 {
				c.Status(http.StatusBadRequest)
				return
			}

			path, err := tree.NodePathFromDomain(splitPath[0])
			if err != nil {
				handleInternalError(c, err)
				return
			}

			name := splitPath[1]

			err = backend.UnsetSecret(types.SystemID(systemID), path, name)
			if err != nil {
				handleInternalError(c, err)
				return
			}

			c.Status(http.StatusOK)
		})
	}
}

func getSystemDefinitionRoot(backend server.Backend, sysResolver *resolver.SystemResolver, systemID string, version string) (tree.Node, error) {
	system, exists, err := backend.GetSystem(types.SystemID(systemID))
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, fmt.Errorf("system %v does not exist", systemID)
	}

	systemDefURI := fmt.Sprintf(
		"%v#%v/%s",
		system.DefinitionURL,
		version,
		definition.SystemDefinitionRootPathDefault,
	)

	return sysResolver.ResolveDefinition(systemDefURI, &git.Options{})
}

func getSystemVersions(backend server.Backend, sysResolver *resolver.SystemResolver, systemID string) ([]string, error) {
	system, exists, err := backend.GetSystem(types.SystemID(systemID))
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, fmt.Errorf("system %v does not exist", systemID)
	}

	return sysResolver.ListDefinitionVersions(system.DefinitionURL, &git.Options{})
}

func handleInternalError(c *gin.Context, err error) {
	glog.Errorf("encountered error: %v", err.Error())
	c.String(http.StatusInternalServerError, "")
}
