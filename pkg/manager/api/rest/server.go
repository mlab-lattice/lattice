package rest

import (
	"fmt"
	"net/http"

	systemdefinition "github.com/mlab-lattice/core/pkg/system/definition"
	systemdefinitionblock "github.com/mlab-lattice/core/pkg/system/definition/block"
	systemtree "github.com/mlab-lattice/core/pkg/system/tree"
	coretypes "github.com/mlab-lattice/core/pkg/types"

	"github.com/mlab-lattice/system/pkg/manager/backend"

	"github.com/gin-gonic/gin"
)

type restServer struct {
	router  *gin.Engine
	backend backend.Interface
}

func RunNewRestServer(b backend.Interface, port int32) {
	s := restServer{
		router:  gin.Default(),
		backend: b,
	}

	s.mountHandlers()
	s.router.Run(fmt.Sprintf(":%v", port))
}

type systemVersionResponse struct {
	Id         string                   `json:"id"`
	Definition *systemdefinition.System `json:"definition"`
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

	namespaces := r.router.Group("/namespaces")
	{
		namespace := namespaces.Group("/:namespace_id")
		{
			// Versions
			versions := namespace.Group("/versions")
			{
				versions.GET("", func(c *gin.Context) {
					//namespace := c.Param("namespace_id")
					versions := getSystemVersions()
					c.JSON(http.StatusOK, versions)
				})

				versions.GET("/:version_id", func(c *gin.Context) {
					//namespace := c.Param("namespace_id")
					version := c.Param("version_id")
					sysDef, err := getSystemRoot(version)
					if err != nil {
						c.String(http.StatusInternalServerError, err.Error())
						return
					}

					c.JSON(http.StatusOK, systemVersionResponse{
						Id:         version,
						Definition: sysDef,
					})
				})
			}

			// Builds
			builds := namespace.Group("/builds")
			{
				// build-system
				builds.POST("", func(c *gin.Context) {
					namespace := c.Param("namespace_id")
					var req buildSystemRequest
					if err := c.BindJSON(&req); err != nil {
						c.String(http.StatusInternalServerError, err.Error())
						return
					}

					root, err := getSystemRoot(req.Version)
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

			// Rollouts
			rollouts := namespace.Group("/rollouts")
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
						root, err := getSystemRoot(*req.Version)
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

			// Teardowns
			teardowns := namespace.Group("/teardowns")
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

			// Services
			services := namespace.Group("/services")
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

				// get-rollout
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
	}
}

func getSystemRoot(version string) (*systemdefinition.System, error) {
	var wwwCommit string
	if version == "v1.0.0" {

		wwwCommit = "3d2acf451eb031694915c8aa755555a09037db80"
	} else if version == "v2.0.0" {
		wwwCommit = "3d2acf451eb031694915c8aa755555a09037db80"
	} else {
		return nil, fmt.Errorf("invalid version %v", version)
	}
	serviceCommit := "7651060c28687531116837cdeb685d5e5be5a02b"

	language := "node:boron"

	clientInstallCommand := "npm install && npm run build"
	serviceInstallCommand := "npm install"
	var one int32 = 1
	t2small := "t2.small"

	sysDefinition := &systemdefinition.System{
		Meta: systemdefinitionblock.Metadata{
			Name: "petflix",
			Type: systemdefinition.SystemType,
		},
		Subsystems: []systemdefinition.Interface{
			systemdefinition.Interface(&systemdefinition.Service{
				Meta: systemdefinitionblock.Metadata{
					Name: "www",
					Type: systemdefinition.ServiceType,
				},
				Components: []*systemdefinitionblock.Component{
					{
						Name: "http",
						Ports: []*systemdefinitionblock.ComponentPort{
							{
								Name:     "http",
								Port:     8080,
								Protocol: systemdefinitionblock.HttpProtocol,
								ExternalAccess: &systemdefinitionblock.ExternalAccess{
									Public: true,
								},
							},
						},
						Build: systemdefinitionblock.ComponentBuild{
							GitRepository: &systemdefinitionblock.GitRepository{
								Url:    "https://github.com/mlab-lattice/example-petflix-www",
								Commit: &wwwCommit,
							},
							Language: &language,
							Command:  &clientInstallCommand,
						},
						Exec: systemdefinitionblock.ComponentExec{
							Command: []string{
								"node",
								"app.js",
							},
							Environment: map[string]string{
								"PETFLIX_API_URI": "http://api.petflix",
							},
						},
					},
				},
				Resources: systemdefinitionblock.Resources{
					NumInstances: &one,
					InstanceType: &t2small,
				},
			}),
			systemdefinition.Interface(&systemdefinition.Service{
				Meta: systemdefinitionblock.Metadata{
					Name: "api",
					Type: systemdefinition.ServiceType,
				},
				Components: []*systemdefinitionblock.Component{
					{
						Name: "http",
						Ports: []*systemdefinitionblock.ComponentPort{
							{
								Name:     "http",
								Port:     80,
								Protocol: systemdefinitionblock.HttpProtocol,
							},
						},
						Build: systemdefinitionblock.ComponentBuild{
							GitRepository: &systemdefinitionblock.GitRepository{
								Url:    "https://github.com/mlab-lattice/example-petflix-service",
								Commit: &serviceCommit,
							},
							Language: &language,
							Command:  &serviceInstallCommand,
						},
						Exec: systemdefinitionblock.ComponentExec{
							Command: []string{
								"node",
								"index.js",
							},
							Environment: map[string]string{
								"PORT":        "80",
								"MONGODB_URI": "connection_string_goes_here",
							},
						},
					},
				},
				Resources: systemdefinitionblock.Resources{
					NumInstances: &one,
					InstanceType: &t2small,
				},
			}),
		},
	}

	return sysDefinition, nil
	//return systemtree.NewNode(sysDefinition, nil)
}

func getSystemVersions() []string {
	return []string{"v1.0.0"}
}
