package rest

import (
	"fmt"
	"net/http"

	"github.com/mlab-lattice/system/pkg/constants"
	"github.com/mlab-lattice/system/pkg/definition"
	"github.com/mlab-lattice/system/pkg/definition/resolver"
	"github.com/mlab-lattice/system/pkg/definition/tree"
	"github.com/mlab-lattice/system/pkg/manager/backend"
	"github.com/mlab-lattice/system/pkg/types"

	"github.com/gin-gonic/gin"
)

func (r *restServer) mountNamespaceHandlers() {
	r.mountNamespaceVersionHandlers()
	r.mountNamespaceSystemBuildHandlers()
	r.mountNamespaceServiceBuildHandlers()
	r.mountNamespaceComponentBuildHandlers()
	r.mountNamespaceRolloutHandlers()
	r.mountNamespaceTeardownHandlers()
	r.mountNamespaceServiceHandlers()
}

type systemVersionResponse struct {
	ID         string               `json:"id"`
	Definition definition.Interface `json:"definition"`
}

func (r *restServer) mountNamespaceVersionHandlers() {
	versions := r.router.Group("/namespaces/:namespace_id/versions")
	{
		// list-system-versions
		versions.GET("", func(c *gin.Context) {
			namespace := c.Param("namespace_id")
			versions, err := r.getSystemVersions(namespace)
			if err != nil {
				handleInternalError(c, err)
				return
			}

			c.JSON(http.StatusOK, versions)
		})

		// get-system-version
		versions.GET("/:version_id", func(c *gin.Context) {
			namespace := c.Param("namespace_id")
			version := c.Param("version_id")
			sysDef, err := r.getSystemRoot(namespace, version)
			if err != nil {
				handleInternalError(c, err)
				return
			}

			c.JSON(http.StatusOK, systemVersionResponse{
				ID:         version,
				Definition: sysDef.Definition(),
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

func (r *restServer) mountNamespaceSystemBuildHandlers() {
	sysbs := r.router.Group("/namespaces/:namespace_id/system-builds")
	{
		// build-system
		sysbs.POST("", func(c *gin.Context) {
			namespace := c.Param("namespace_id")
			var req buildSystemRequest
			if err := c.BindJSON(&req); err != nil {
				handleInternalError(c, err)
				return
			}

			root, err := r.getSystemRoot(namespace, req.Version)
			if err != nil {
				handleInternalError(c, err)
				return
			}

			bid, err := r.backend.BuildSystem(
				types.LatticeNamespace(namespace),
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
		sysbs.GET("", func(c *gin.Context) {
			namespace := c.Param("namespace_id")

			bs, err := r.backend.ListSystemBuilds(types.LatticeNamespace(namespace))
			if err != nil {
				handleInternalError(c, err)
				return
			}

			c.JSON(http.StatusOK, bs)
		})

		// get-system-build
		sysbs.GET("/:build_id", func(c *gin.Context) {
			namespace := c.Param("namespace_id")
			bid := c.Param("build_id")

			b, exists, err := r.backend.GetSystemBuild(types.LatticeNamespace(namespace), types.SystemBuildID(bid))
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

func (r *restServer) mountNamespaceServiceBuildHandlers() {
	sysbs := r.router.Group("/namespaces/:namespace_id/service-builds")
	{
		// list-service-builds
		sysbs.GET("", func(c *gin.Context) {
			namespace := c.Param("namespace_id")

			bs, err := r.backend.ListServiceBuilds(types.LatticeNamespace(namespace))
			if err != nil {
				handleInternalError(c, err)
				return
			}

			c.JSON(http.StatusOK, bs)
		})

		// get-service-build
		sysbs.GET("/:build_id", func(c *gin.Context) {
			namespace := c.Param("namespace_id")
			bid := c.Param("build_id")

			b, exists, err := r.backend.GetServiceBuild(types.LatticeNamespace(namespace), types.ServiceBuildID(bid))
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

func (r *restServer) mountNamespaceComponentBuildHandlers() {
	sysbs := r.router.Group("/namespaces/:namespace_id/component-builds")
	{
		// list-component-builds
		sysbs.GET("", func(c *gin.Context) {
			namespace := c.Param("namespace_id")

			bs, err := r.backend.ListComponentBuilds(types.LatticeNamespace(namespace))
			if err != nil {
				handleInternalError(c, err)
				return
			}

			c.JSON(http.StatusOK, bs)
		})

		// get-system-build
		sysbs.GET("/:build_id", func(c *gin.Context) {
			namespace := c.Param("namespace_id")
			bid := c.Param("build_id")

			b, exists, err := r.backend.GetComponentBuild(types.LatticeNamespace(namespace), types.ComponentBuildID(bid))
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

		// get-system-build-logs
		sysbs.GET("/:build_id/logs", func(c *gin.Context) {
			namespace := c.Param("namespace_id")
			bid := c.Param("build_id")
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

			log, exists, err := r.backend.GetComponentBuildLogs(types.LatticeNamespace(namespace), types.ComponentBuildID(bid), follow)
			if exists == false {
				switch err.(type) {
				case *backend.UserError:
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

func (r *restServer) mountNamespaceRolloutHandlers() {
	rollouts := r.router.Group("/namespaces/:namespace_id/rollouts")
	{
		// roll-out-system
		rollouts.POST("", func(c *gin.Context) {
			namespace := c.Param("namespace_id")
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

			var rid types.SystemRolloutID
			var err error
			if req.Version != nil {
				root, err := r.getSystemRoot(namespace, *req.Version)
				if err != nil {
					handleInternalError(c, err)
					return
				}

				rid, err = r.backend.RollOutSystem(
					types.LatticeNamespace(namespace),
					root,
					types.SystemVersion(*req.Version),
				)
			} else {
				rid, err = r.backend.RollOutSystemBuild(
					types.LatticeNamespace(namespace),
					types.SystemBuildID(*req.BuildID),
				)
			}

			if err != nil {
				handleInternalError(c, err)
				return
			}

			c.JSON(http.StatusCreated, rollOutSystemResponse{
				RolloutID: rid,
			})
		})

		// list-rollouts
		rollouts.GET("", func(c *gin.Context) {
			namespace := c.Param("namespace_id")

			rs, err := r.backend.ListSystemRollouts(types.LatticeNamespace(namespace))
			if err != nil {
				handleInternalError(c, err)
				return
			}

			c.JSON(http.StatusOK, rs)
		})

		// get-rollout
		rollouts.GET("/:rollout_id", func(c *gin.Context) {
			namespace := c.Param("namespace_id")
			rid := c.Param("rollout_id")

			r, exists, err := r.backend.GetSystemRollout(types.LatticeNamespace(namespace), types.SystemRolloutID(rid))
			if err != nil {
				handleInternalError(c, err)
				return
			}

			if !exists {
				c.String(http.StatusNotFound, "")
				return
			}

			c.JSON(http.StatusOK, r)
		})
	}
}

type tearDownSystemResponse struct {
	TeardownID types.SystemTeardownID `json:"teardownId"`
}

func (r *restServer) mountNamespaceTeardownHandlers() {
	teardowns := r.router.Group("/namespaces/:namespace_id/teardowns")
	{
		// tear-down-system
		teardowns.POST("", func(c *gin.Context) {
			namespace := c.Param("namespace_id")

			tid, err := r.backend.TearDownSystem(types.LatticeNamespace(namespace))

			if err != nil {
				handleInternalError(c, err)
				return
			}

			c.JSON(http.StatusCreated, tearDownSystemResponse{
				TeardownID: tid,
			})
		})

		// list-teardowns
		teardowns.GET("", func(c *gin.Context) {
			namespace := c.Param("namespace_id")

			ts, err := r.backend.ListSystemTeardowns(types.LatticeNamespace(namespace))
			if err != nil {
				handleInternalError(c, err)
				return
			}

			c.JSON(http.StatusOK, ts)
		})

		// get-teardown
		teardowns.GET("/:teardown_id", func(c *gin.Context) {
			namespace := c.Param("namespace_id")
			tid := c.Param("teardown_id")

			t, exists, err := r.backend.GetSystemTeardown(types.LatticeNamespace(namespace), types.SystemTeardownID(tid))
			if err != nil {
				handleInternalError(c, err)
				return
			}

			if !exists {
				c.String(http.StatusNotFound, "")
				return
			}

			c.JSON(http.StatusOK, t)
		})
	}
}

func (r *restServer) mountNamespaceServiceHandlers() {
	services := r.router.Group("/namespaces/:namespace_id/services")
	{
		// list-services
		services.GET("", func(c *gin.Context) {
			namespace := c.Param("namespace_id")

			svcs, err := r.backend.ListSystemServices(types.LatticeNamespace(namespace))
			if err != nil {
				handleInternalError(c, err)
				return
			}

			c.JSON(http.StatusOK, svcs)
		})

		// get-service
		services.GET("/:svc_domain", func(c *gin.Context) {
			namespace := c.Param("namespace_id")
			svcDomain := c.Param("svc_domain")

			svcPath, err := tree.NodePathFromDomain(svcDomain)
			if err != nil {
				handleInternalError(c, err)
				return
			}

			svc, err := r.backend.GetSystemService(types.LatticeNamespace(namespace), svcPath)
			if err != nil {
				handleInternalError(c, err)
				return
			}

			if svc == nil {
				c.String(http.StatusNotFound, "")
				return
			}

			c.JSON(http.StatusOK, svc)
		})
	}
}

func (r *restServer) getSystemRoot(ln string, version string) (tree.Node, error) {
	url, err := r.backend.GetSystemURL(types.LatticeNamespace(ln))
	if err != nil {
		return nil, err
	}

	return r.resolver.ResolveDefinition(
		fmt.Sprintf("%v#%v", url, version),
		constants.SystemDefinitionRootPathDefault,
		&resolver.GitResolveOptions{},
	)
}

func (r *restServer) getSystemVersions(ln string) ([]string, error) {
	url, err := r.backend.GetSystemURL(types.LatticeNamespace(ln))
	if err != nil {
		return nil, err
	}

	return r.resolver.ListDefinitionVersions(url, &resolver.GitResolveOptions{})
}
