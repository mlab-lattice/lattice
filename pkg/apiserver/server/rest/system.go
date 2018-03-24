package rest

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/mlab-lattice/system/pkg/definition"
	"github.com/mlab-lattice/system/pkg/definition/tree"
	"github.com/mlab-lattice/system/pkg/types"
	"github.com/mlab-lattice/system/pkg/util/git"

	"github.com/gin-gonic/gin"
)

type createSystemRequest struct {
	ID            types.SystemID `json:"id"`
	DefinitionURL string         `json:"definitionUrl"`
}

func (r *restServer) mountSystemHandlers() {
	systems := r.router.Group("/systems")
	{
		// create-system
		systems.POST("", func(c *gin.Context) {
			var req createSystemRequest
			if err := c.BindJSON(&req); err != nil {
				handleInternalError(c, err)
				return
			}

			system, err := r.backend.CreateSystem(req.ID, req.DefinitionURL)
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
			systems, err := r.backend.ListSystems()

			if err != nil {
				handleInternalError(c, err)
				return
			}

			c.JSON(http.StatusOK, systems)
		})

		// get-system
		systems.GET("/:system_id", func(c *gin.Context) {
			systemID := c.Param("system_id")

			system, exists, err := r.backend.GetSystem(types.SystemID(systemID))
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

			err := r.backend.DeleteSystem(types.SystemID(systemID))
			if err != nil {
				handleInternalError(c, err)
				return
			}

			c.Status(http.StatusOK)
		})
	}

	r.mountSystemVersionHandlers()
	r.mountSystemSystemBuildHandlers()
	r.mountSystemDeployHandlers()
	r.mountSystemTeardownHandlers()
	r.mountSystemServiceHandlers()
	r.mountSystemSecretHandlers()
}

type systemVersionResponse struct {
	ID         string               `json:"id"`
	Definition definition.Interface `json:"definition"`
}

func (r *restServer) mountSystemVersionHandlers() {
	versions := r.router.Group("/systems/:system_id/versions")
	{
		// list-system-versions
		versions.GET("", func(c *gin.Context) {
			systemID := c.Param("system_id")

			versions, err := r.getSystemVersions(systemID)
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

			definitionRoot, err := r.getSystemDefinitionRoot(systemID, version)
			if err != nil {
				handleInternalError(c, err)
				return
			}

			c.JSON(http.StatusOK, systemVersionResponse{
				ID: version,
				// FIXME: this probalby won't work
				Definition: definitionRoot,
			})
		})
	}
}

type buildSystemRequest struct {
	Version string `json:"version"`
}

type buildSystemResponse struct {
	BuildID types.BuildID `json:"buildId"`
}

func (r *restServer) mountSystemSystemBuildHandlers() {
	systemBuilds := r.router.Group("/systems/:system_id/builds")
	{
		// build-system
		systemBuilds.POST("", func(c *gin.Context) {
			systemID := c.Param("system_id")

			var req buildSystemRequest
			if err := c.BindJSON(&req); err != nil {
				handleInternalError(c, err)
				return
			}

			root, err := r.getSystemDefinitionRoot(systemID, req.Version)
			if err != nil {
				handleInternalError(c, err)
				return
			}

			bid, err := r.backend.Build(
				types.SystemID(systemID),
				root,
				types.SystemVersion(req.Version),
			)

			if err != nil {
				handleInternalError(c, err)
				return
			}

			c.JSON(http.StatusCreated, buildSystemResponse{
				BuildID: bid,
			})
		})

		// list-system-builds
		systemBuilds.GET("", func(c *gin.Context) {
			systemID := c.Param("system_id")

			bs, err := r.backend.ListBuilds(types.SystemID(systemID))
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

			b, exists, err := r.backend.GetBuild(types.SystemID(systemID), types.BuildID(buildID))
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

type deployRequest struct {
	Version *string        `json:"version,omitempty"`
	BuildID *types.BuildID `json:"buildId,omitempty"`
}

type deployResponse struct {
	DeployID types.DeployID `json:"deployId"`
}

func (r *restServer) mountSystemDeployHandlers() {
	deploys := r.router.Group("/systems/:system_id/deploys")
	{
		// roll-out-system
		deploys.POST("", func(c *gin.Context) {
			systemID := c.Param("system_id")

			var req deployRequest
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
				root, err := r.getSystemDefinitionRoot(systemID, *req.Version)
				if err != nil {
					handleInternalError(c, err)
					return
				}

				deployID, err = r.backend.DeployVersion(
					types.SystemID(systemID),
					root,
					types.SystemVersion(*req.Version),
				)
			} else {
				deployID, err = r.backend.DeployBuild(
					types.SystemID(systemID),
					types.BuildID(*req.BuildID),
				)
			}

			if err != nil {
				handleInternalError(c, err)
				return
			}

			c.JSON(http.StatusCreated, deployResponse{
				DeployID: deployID,
			})
		})

		// list-deploys
		deploys.GET("", func(c *gin.Context) {
			systemID := c.Param("system_id")

			rollouts, err := r.backend.ListDeploys(types.SystemID(systemID))
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

			rollout, exists, err := r.backend.GetDeploy(types.SystemID(systemID), types.DeployID(deployID))
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

type tearDownSystemResponse struct {
	TeardownID types.TeardownID `json:"teardownId"`
}

func (r *restServer) mountSystemTeardownHandlers() {
	teardowns := r.router.Group("/systems/:system_id/teardowns")
	{
		// tear-down-system
		teardowns.POST("", func(c *gin.Context) {
			systemID := c.Param("system_id")

			teardownID, err := r.backend.TearDown(types.SystemID(systemID))

			if err != nil {
				handleInternalError(c, err)
				return
			}

			c.JSON(http.StatusCreated, tearDownSystemResponse{
				TeardownID: teardownID,
			})
		})

		// list-teardowns
		teardowns.GET("", func(c *gin.Context) {
			systemID := c.Param("system_id")

			teardowns, err := r.backend.ListTeardowns(types.SystemID(systemID))
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

			teardown, exists, err := r.backend.GetTeardown(types.SystemID(systemID), types.TeardownID(teardownID))
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

func (r *restServer) mountSystemServiceHandlers() {
	services := r.router.Group("/systems/:system_id/services")
	{
		// list-services
		services.GET("", func(c *gin.Context) {
			systemID := c.Param("system_id")

			services, err := r.backend.ListServices(types.SystemID(systemID))
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

			service, err := r.backend.GetService(types.SystemID(systemID), servicePath)
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

type setSecretRequest struct {
	Value string `json:"value"`
}

func (r *restServer) mountSystemSecretHandlers() {
	secrets := r.router.Group("/systems/:system_id/secrets")
	{
		// list-secrets
		secrets.GET("", func(c *gin.Context) {
			systemID := c.Param("system_id")

			services, err := r.backend.ListSecrets(types.SystemID(systemID))
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

			secret, exists, err := r.backend.GetSecret(types.SystemID(systemID), path, name)
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
			var req setSecretRequest
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

			err = r.backend.SetSecret(types.SystemID(systemID), path, name, req.Value)
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

			err = r.backend.UnsetSecret(types.SystemID(systemID), path, name)
			if err != nil {
				handleInternalError(c, err)
				return
			}

			c.Status(http.StatusOK)
		})
	}
}

func (r *restServer) getSystemDefinitionRoot(systemID string, version string) (tree.Node, error) {
	system, exists, err := r.backend.GetSystem(types.SystemID(systemID))
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, fmt.Errorf("System %v does not exist", systemID)
	}

	systemDefURI := fmt.Sprintf(
		"%v#%v/%s",
		system.DefinitionURL,
		version,
		definition.SystemDefinitionRootPathDefault,
	)

	return r.resolver.ResolveDefinition(systemDefURI, &git.Options{})
}

func (r *restServer) getSystemVersions(systemID string) ([]string, error) {
	system, exists, err := r.backend.GetSystem(types.SystemID(systemID))
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, fmt.Errorf("System %v does not exist", systemID)
	}

	return r.resolver.ListDefinitionVersions(system.DefinitionURL, &git.Options{})
}
