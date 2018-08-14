package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/mlab-lattice/lattice/bazel-lattice/external/com_github_swaggo_gin_swagger/swaggerFiles"
	_ "github.com/mlab-lattice/lattice/pkg/api/docs"
	v1server "github.com/mlab-lattice/lattice/pkg/api/server/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/resolver"
	"github.com/swaggo/gin-swagger"
)

type LatticeAPI struct {
	router      *gin.RouterGroup
	backend     v1server.Interface
	sysResolver resolver.SystemResolver
}

func newLatticeAPI(router *gin.RouterGroup, backend v1server.Interface, sysResolver resolver.SystemResolver) *LatticeAPI {
	return &LatticeAPI{
		router:      router,
		backend:     backend,
		sysResolver: sysResolver,
	}
}

// @title Lattice API
// @version 1.0
// @description This is a sample server celler server.
// @termsOfService http://swagger.io/terms/
// @license.name Apache 2.0
// @host localhost:8876
// @BasePath /v1
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name apiKey
func (api *LatticeAPI) setupAPI() {
	api.setupSystemEndpoints()
	api.setupBuildEndpoints()
	api.setupDeployEndpoints()
	api.setupNoodPoolEndpoints()
	api.setupServicesEndpoints()
	api.setupJobsEndpoints()
	api.setupTeardownEndpoints()
	api.setupSecretsEndpoints()
	api.setupVersionsEndpoints()
	api.router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}
