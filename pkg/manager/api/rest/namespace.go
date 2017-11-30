package rest

import (
	"fmt"
	"net/http"

	coreconstants "github.com/mlab-lattice/core/pkg/constants"
	systemdefinition "github.com/mlab-lattice/core/pkg/system/definition"
	systemresolver "github.com/mlab-lattice/core/pkg/system/resolver"
	systemtree "github.com/mlab-lattice/core/pkg/system/tree"
	coretypes "github.com/mlab-lattice/core/pkg/types"

	"github.com/mlab-lattice/system/pkg/manager/backend"

	"github.com/gin-gonic/gin"
)

func (r *restServer) mountNamespaceHandlers() {
	r.mountNamespaceVersionHandlers()
	r.mountNamespaceSystemBuildHandlers()
	r.mountNamespaceComponentBuildHandlers()
	r.mountNamespaceRolloutHandlers()
	r.mountNamespaceTeardownHandlers()
	r.mountNamespaceServiceHandlers()
}

type systemVersionResponse struct {
	Id         string                     `json:"id"`
	Definition systemdefinition.Interface `json:"definition"`
}

func (r *restServer) mountNamespaceVersionHandlers() {
	versions := r.router.Group("/namespaces/:namespace_id/versions")
	{
		// list-system-versions
		versions.GET("", func(c *gin.Context) {
			namespace := c.Param("namespace_id")
			versions, err := r.getSystemVersions(namespace)
			if err != nil {
				c.String(http.StatusInternalServerError, err.Error())
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
				c.String(http.StatusInternalServerError, err.Error())
				return
			}

			c.JSON(http.StatusOK, systemVersionResponse{
				Id:         version,
				Definition: sysDef.Definition(),
			})
		})
	}
}

type buildSystemRequest struct {
	Version string `json:"version"`
}

type buildSystemResponse struct {
	BuildId coretypes.SystemBuildID `json:"buildId"`
}

func (r *restServer) mountNamespaceSystemBuildHandlers() {
	sysbs := r.router.Group("/namespaces/:namespace_id/system-builds")
	{
		// build-system
		sysbs.POST("", func(c *gin.Context) {
			namespace := c.Param("namespace_id")
			var req buildSystemRequest
			if err := c.BindJSON(&req); err != nil {
				c.String(http.StatusInternalServerError, err.Error())
				return
			}

			root, err := r.getSystemRoot(namespace, req.Version)
			if err != nil {
				c.String(http.StatusInternalServerError, err.Error())
				return
			}

			bid, err := r.backend.BuildSystem(
				coretypes.LatticeNamespace(namespace),
				root,
				coretypes.SystemVersion(req.Version),
			)

			if err != nil {
				c.String(http.StatusInternalServerError, err.Error())
				return
			}

			c.JSON(http.StatusOK, buildSystemResponse{
				BuildId: bid,
			})
		})

		// list-system-builds
		sysbs.GET("", func(c *gin.Context) {
			namespace := c.Param("namespace_id")

			bs, err := r.backend.ListSystemBuilds(coretypes.LatticeNamespace(namespace))
			if err != nil {
				c.String(http.StatusInternalServerError, err.Error())
				return
			}

			c.JSON(http.StatusOK, bs)
		})

		// get-system-build
		sysbs.GET("/:build_id", func(c *gin.Context) {
			namespace := c.Param("namespace_id")
			bid := c.Param("build_id")

			b, exists, err := r.backend.GetSystemBuild(coretypes.LatticeNamespace(namespace), coretypes.SystemBuildID(bid))
			if err != nil {
				c.String(http.StatusInternalServerError, err.Error())
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

			bs, err := r.backend.ListComponentBuilds(coretypes.LatticeNamespace(namespace))
			if err != nil {
				c.String(http.StatusInternalServerError, err.Error())
				return
			}

			c.JSON(http.StatusOK, bs)
		})

		// get-system-build
		sysbs.GET("/:build_id", func(c *gin.Context) {
			namespace := c.Param("namespace_id")
			bid := c.Param("build_id")

			b, exists, err := r.backend.GetComponentBuild(coretypes.LatticeNamespace(namespace), coretypes.ComponentBuildID(bid))
			if err != nil {
				c.String(http.StatusInternalServerError, err.Error())
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

			log, exists, err := r.backend.GetComponentBuildLogs(coretypes.LatticeNamespace(namespace), coretypes.ComponentBuildID(bid), follow)
			if exists == false {
				switch err.(type) {
				case *backend.UserError:
					c.String(http.StatusNotFound, err.Error())
				default:
					c.String(http.StatusInternalServerError, "")
				}
				return
			}

			logEndpoint(c, log, follow)
		})
	}
}

type rollOutSystemRequest struct {
	Version *string                  `json:"version,omitempty"`
	BuildId *coretypes.SystemBuildID `json:"buildId,omitempty"`
}

type rollOutSystemResponse struct {
	RolloutId coretypes.SystemRolloutID `json:"rolloutId"`
}

func (r *restServer) mountNamespaceRolloutHandlers() {
	rollouts := r.router.Group("/namespaces/:namespace_id/rollouts")
	{
		// roll-out-system
		rollouts.POST("", func(c *gin.Context) {
			namespace := c.Param("namespace_id")
			var req rollOutSystemRequest
			if err := c.BindJSON(&req); err != nil {
				c.String(http.StatusInternalServerError, err.Error())
				return
			}

			if req.Version != nil && req.BuildId != nil {
				c.String(http.StatusBadRequest, "can only specify version or buildId")
				return
			}

			if req.Version == nil && req.BuildId == nil {
				c.String(http.StatusBadRequest, "must specify version or buildId")
				return
			}

			var rid coretypes.SystemRolloutID
			var err error
			if req.Version != nil {
				root, err := r.getSystemRoot(namespace, *req.Version)
				if err != nil {
					c.String(http.StatusInternalServerError, err.Error())
					return
				}

				rid, err = r.backend.RollOutSystem(
					coretypes.LatticeNamespace(namespace),
					root,
					coretypes.SystemVersion(*req.Version),
				)
			} else {
				rid, err = r.backend.RollOutSystemBuild(
					coretypes.LatticeNamespace(namespace),
					coretypes.SystemBuildID(*req.BuildId),
				)
			}

			if err != nil {
				c.String(http.StatusInternalServerError, err.Error())
				return
			}

			c.JSON(http.StatusOK, rollOutSystemResponse{
				RolloutId: rid,
			})
		})

		// list-rollouts
		rollouts.GET("", func(c *gin.Context) {
			namespace := c.Param("namespace_id")

			rs, err := r.backend.ListSystemRollouts(coretypes.LatticeNamespace(namespace))
			if err != nil {
				c.String(http.StatusInternalServerError, err.Error())
				return
			}

			c.JSON(http.StatusOK, rs)
		})

		// get-rollout
		rollouts.GET("/:rollout_id", func(c *gin.Context) {
			namespace := c.Param("namespace_id")
			rid := c.Param("rollout_id")

			r, exists, err := r.backend.GetSystemRollout(coretypes.LatticeNamespace(namespace), coretypes.SystemRolloutID(rid))
			if err != nil {
				c.String(http.StatusInternalServerError, err.Error())
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
	TeardownId coretypes.SystemTeardownID `json:"teardownId"`
}

func (r *restServer) mountNamespaceTeardownHandlers() {
	teardowns := r.router.Group("/namespaces/:namespace_id/teardowns")
	{
		// tear-down-system
		teardowns.POST("", func(c *gin.Context) {
			namespace := c.Param("namespace_id")

			tid, err := r.backend.TearDownSystem(coretypes.LatticeNamespace(namespace))

			if err != nil {
				c.String(http.StatusInternalServerError, err.Error())
				return
			}

			c.JSON(http.StatusOK, tearDownSystemResponse{
				TeardownId: tid,
			})
		})

		// list-teardowns
		teardowns.GET("", func(c *gin.Context) {
			namespace := c.Param("namespace_id")

			ts, err := r.backend.ListSystemTeardowns(coretypes.LatticeNamespace(namespace))
			if err != nil {
				c.String(http.StatusInternalServerError, err.Error())
				return
			}

			c.JSON(http.StatusOK, ts)
		})

		// get-teardown
		teardowns.GET("/:teardown_id", func(c *gin.Context) {
			namespace := c.Param("namespace_id")
			tid := c.Param("teardown_id")

			t, exists, err := r.backend.GetSystemTeardown(coretypes.LatticeNamespace(namespace), coretypes.SystemTeardownID(tid))
			if err != nil {
				c.String(http.StatusInternalServerError, err.Error())
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

			svcs, err := r.backend.ListSystemServices(coretypes.LatticeNamespace(namespace))
			if err != nil {
				c.String(http.StatusInternalServerError, err.Error())
				return
			}

			c.JSON(http.StatusOK, svcs)
		})

		// get-service
		services.GET("/:svc_domain", func(c *gin.Context) {
			namespace := c.Param("namespace_id")
			svcDomain := c.Param("svc_domain")

			svcPath, err := systemtree.NodePathFromDomain(svcDomain)
			if err != nil {
				c.String(http.StatusInternalServerError, err.Error())
				return
			}

			svc, err := r.backend.GetSystemService(coretypes.LatticeNamespace(namespace), svcPath)
			if err != nil {
				c.String(http.StatusInternalServerError, err.Error())
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

func (r *restServer) getSystemRoot(ln string, version string) (systemtree.Node, error) {
	url, err := r.backend.GetSystemUrl(coretypes.LatticeNamespace(ln))
	if err != nil {
		return nil, err
	}

	return r.resolver.ResolveDefinition(
		fmt.Sprintf("%v#%v", url, version),
		coreconstants.SystemDefinitionRootPathDefault,
		&systemresolver.GitResolveOptions{},
	)
}

func (r *restServer) getSystemVersions(ln string) ([]string, error) {
	url, err := r.backend.GetSystemUrl(coretypes.LatticeNamespace(ln))
	if err != nil {
		return nil, err
	}

	return r.resolver.ListDefinitionVersions(url, &systemresolver.GitResolveOptions{})
}
