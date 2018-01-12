package rest

import (
	"fmt"
	"net/http"

	"github.com/mlab-lattice/system/pkg/constants"
	"github.com/mlab-lattice/system/pkg/definition"
	"github.com/mlab-lattice/system/pkg/definition/resolver"
	"github.com/mlab-lattice/system/pkg/definition/tree"
	"github.com/mlab-lattice/system/pkg/managerapi/server"
	"github.com/mlab-lattice/system/pkg/types"

	"github.com/gin-gonic/gin"
	"github.com/mlab-lattice/system/pkg/util/git"
)

func (r *restServer) mountSystemHandlers() {
	systems := r.router.Group("/systems")
	{
		// list-system-versions
		systems.GET("", func(c *gin.Context) {
			systems, err := r.backend.ListSystems()

			if err != nil {
				handleInternalError(c, err)
				return
			}

			c.JSON(http.StatusOK, systems)
		})

		// get-system-version
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
	}

	r.mountSystemVersionHandlers()
	r.mountSystemSystemBuildHandlers()
	r.mountSystemServiceBuildHandlers()
	r.mountSystemComponentBuildHandlers()
	r.mountSystemRolloutHandlers()
	r.mountSystemTeardownHandlers()
	r.mountSystemServiceHandlers()
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
				ID:         version,
				Definition: definitionRoot.Definition(),
			})
		})
	}
}

type buildSystemRequest struct {
	Version string `json:"version"`
}

type buildSystemResponse struct {
	BuildID types.SystemBuildID `json:"buildId"`
}

func (r *restServer) mountSystemSystemBuildHandlers() {
	systemBuilds := r.router.Group("/systems/:system_id/system-builds")
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

			bid, err := r.backend.BuildSystem(
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

			bs, err := r.backend.ListSystemBuilds(types.SystemID(systemID))
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

			b, exists, err := r.backend.GetSystemBuild(types.SystemID(systemID), types.SystemBuildID(buildID))
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

func (r *restServer) mountSystemServiceBuildHandlers() {
	serviceBuilds := r.router.Group("/systems/:system_id/service-builds")
	{
		// list-service-builds
		serviceBuilds.GET("", func(c *gin.Context) {
			systemID := c.Param("system_id")

			builds, err := r.backend.ListServiceBuilds(types.SystemID(systemID))
			if err != nil {
				handleInternalError(c, err)
				return
			}

			c.JSON(http.StatusOK, builds)
		})

		// get-service-build
		serviceBuilds.GET("/:build_id", func(c *gin.Context) {
			systemID := c.Param("system_id")
			buildID := c.Param("build_id")

			build, exists, err := r.backend.GetServiceBuild(types.SystemID(systemID), types.ServiceBuildID(buildID))
			if err != nil {
				handleInternalError(c, err)
				return
			}

			if !exists {
				c.String(http.StatusNotFound, "")
				return
			}

			c.JSON(http.StatusOK, build)
		})
	}
}

func (r *restServer) mountSystemComponentBuildHandlers() {
	componentBuilds := r.router.Group("/systems/:system_id/component-builds")
	{
		// list-component-builds
		componentBuilds.GET("", func(c *gin.Context) {
			systemID := c.Param("system_id")

			builds, err := r.backend.ListComponentBuilds(types.SystemID(systemID))
			if err != nil {
				handleInternalError(c, err)
				return
			}

			c.JSON(http.StatusOK, builds)
		})

		// get-system-build
		componentBuilds.GET("/:build_id", func(c *gin.Context) {
			systemID := c.Param("system_id")
			buildID := c.Param("build_id")

			build, exists, err := r.backend.GetComponentBuild(types.SystemID(systemID), types.ComponentBuildID(buildID))
			if err != nil {
				handleInternalError(c, err)
				return
			}

			if !exists {
				c.String(http.StatusNotFound, "")
				return
			}

			c.JSON(http.StatusOK, build)
		})

		// get-system-build-logs
		componentBuilds.GET("/:build_id/logs", func(c *gin.Context) {
			systemID := c.Param("system_id")
			buildID := c.Param("build_id")
			followQuery := c.DefaultQuery("follow", "false")

			var follow bool
			switch followQuery {
			case "false":
				follow = false
			case "true":
				follow = true
			default:
				c.String(http.StatusBadRequest, "invalid value for follow query")
			}

			log, exists, err := r.backend.GetComponentBuildLogs(types.SystemID(systemID), types.ComponentBuildID(buildID), follow)
			if exists == false {
				switch err.(type) {
				case *server.UserError:
					c.String(http.StatusNotFound, "")
				default:
					handleInternalError(c, err)
				}
				return
			}

			logEndpoint(c, log, follow)
		})
	}
}

type rollOutSystemRequest struct {
	Version *string              `json:"version,omitempty"`
	BuildID *types.SystemBuildID `json:"buildId,omitempty"`
}

type rollOutSystemResponse struct {
	RolloutID types.SystemRolloutID `json:"rolloutId"`
}

func (r *restServer) mountSystemRolloutHandlers() {
	rollouts := r.router.Group("/systems/:system_id/rollouts")
	{
		// roll-out-system
		rollouts.POST("", func(c *gin.Context) {
			systemID := c.Param("system_id")

			var req rollOutSystemRequest
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

			var rolloutID types.SystemRolloutID
			var err error
			if req.Version != nil {
				root, err := r.getSystemDefinitionRoot(systemID, *req.Version)
				if err != nil {
					handleInternalError(c, err)
					return
				}

				rolloutID, err = r.backend.RollOutSystem(
					types.SystemID(systemID),
					root,
					types.SystemVersion(*req.Version),
				)
			} else {
				rolloutID, err = r.backend.RollOutSystemBuild(
					types.SystemID(systemID),
					types.SystemBuildID(*req.BuildID),
				)
			}

			if err != nil {
				handleInternalError(c, err)
				return
			}

			c.JSON(http.StatusCreated, rollOutSystemResponse{
				RolloutID: rolloutID,
			})
		})

		// list-rollouts
		rollouts.GET("", func(c *gin.Context) {
			systemID := c.Param("system_id")

			rollouts, err := r.backend.ListSystemRollouts(types.SystemID(systemID))
			if err != nil {
				handleInternalError(c, err)
				return
			}

			c.JSON(http.StatusOK, rollouts)
		})

		// get-rollout
		rollouts.GET("/:rollout_id", func(c *gin.Context) {
			systemID := c.Param("system_id")
			rolloutID := c.Param("rollout_id")

			rollout, exists, err := r.backend.GetSystemRollout(types.SystemID(systemID), types.SystemRolloutID(rolloutID))
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
	TeardownID types.SystemTeardownID `json:"teardownId"`
}

func (r *restServer) mountSystemTeardownHandlers() {
	teardowns := r.router.Group("/systems/:system_id/teardowns")
	{
		// tear-down-system
		teardowns.POST("", func(c *gin.Context) {
			systemID := c.Param("system_id")

			teardownID, err := r.backend.TearDownSystem(types.SystemID(systemID))

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

			teardowns, err := r.backend.ListSystemTeardowns(types.SystemID(systemID))
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

			teardown, exists, err := r.backend.GetSystemTeardown(types.SystemID(systemID), types.SystemTeardownID(teardownID))
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

func (r *restServer) getSystemDefinitionRoot(systemID string, version string) (tree.Node, error) {
	system, exists, err := r.backend.GetSystem(types.SystemID(systemID))
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, fmt.Errorf("System %v does not exist", systemID)
	}

	systemDefUri := fmt.Sprintf("%v#%v/%s", system.DefinitionURL, version,
		constants.SystemDefinitionRootPathDefault)

	return r.resolver.ResolveDefinition(systemDefUri, &git.Options{})
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
