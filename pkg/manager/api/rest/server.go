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

type restServer struct {
	router   *gin.Engine
	backend  backend.Interface
	resolver systemresolver.SystemResolver
}

func RunNewRestServer(b backend.Interface, port int32, workingDirectory string) {
	s := restServer{
		router:  gin.Default(),
		backend: b,
		resolver: systemresolver.SystemResolver{
			WorkDirectory: workingDirectory + "/resolver",
		},
	}

	s.mountHandlers()
	s.router.Run(fmt.Sprintf(":%v", port))
}

type systemVersionResponse struct {
	Id         string                     `json:"id"`
	Definition systemdefinition.Interface `json:"definition"`
}

type buildSystemRequest struct {
	Version string `json:"version"`
}

type buildSystemResponse struct {
	BuildId coretypes.SystemBuildId `json:"buildId"`
}

type rollOutSystemRequest struct {
	Version *string                  `json:"version,omitempty"`
	BuildId *coretypes.SystemBuildId `json:"buildId,omitempty"`
}

type rollOutSystemResponse struct {
	RolloutId coretypes.SystemRolloutId `json:"rolloutId"`
}

type tearDownSystemResponse struct {
	TeardownId coretypes.SystemTeardownId `json:"teardownId"`
}

func (r *restServer) mountHandlers() {
	// Status
	r.router.GET("/status", func(c *gin.Context) {
		c.String(http.StatusOK, "")
	})

	r.mountVersionHandlers()
	r.mountBuildHandlers()
	r.mountRolloutHandlers()
	r.mountTeardownHandlers()
	r.mountServiceHandlers()
}

func (r *restServer) mountVersionHandlers() {
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

func (r *restServer) mountBuildHandlers() {
	builds := r.router.Group("/namespaces/:namespace_id/builds")
	{
		// build-system
		builds.POST("", func(c *gin.Context) {
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

		// list-builds
		builds.GET("", func(c *gin.Context) {
			namespace := c.Param("namespace_id")

			bs, err := r.backend.ListSystemBuilds(coretypes.LatticeNamespace(namespace))
			fmt.Printf("%#v\n", bs)
			if err != nil {
				c.String(http.StatusInternalServerError, err.Error())
				return
			}

			c.JSON(http.StatusOK, bs)
		})

		// get-build
		builds.GET("/:build_id", func(c *gin.Context) {
			namespace := c.Param("namespace_id")
			bid := c.Param("build_id")

			b, exists, err := r.backend.GetSystemBuild(coretypes.LatticeNamespace(namespace), coretypes.SystemBuildId(bid))
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

func (r *restServer) mountRolloutHandlers() {
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

			var rid coretypes.SystemRolloutId
			var err error
			if req.Version != nil {
				// FIXME: get system definition
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
					coretypes.SystemBuildId(*req.BuildId),
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

			r, exists, err := r.backend.GetSystemRollout(coretypes.LatticeNamespace(namespace), coretypes.SystemRolloutId(rid))
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

func (r *restServer) mountTeardownHandlers() {
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

			t, exists, err := r.backend.GetSystemTeardown(coretypes.LatticeNamespace(namespace), coretypes.SystemTeardownId(tid))
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

func (r *restServer) mountServiceHandlers() {
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
		&systemresolver.ResolveOptions{},
	)
}

func (r *restServer) getSystemVersions(ln string) ([]string, error) {
	url, err := r.backend.GetSystemUrl(coretypes.LatticeNamespace(ln))
	if err != nil {
		return nil, err
	}

	return r.resolver.ListDefinitionVersions(url, &systemresolver.ResolveOptions{})
}
