// Lattice API Documentation
//
// Welcome to lattice API.
//
// Terms Of Service:
//
// there are no TOS at this moment, use at your own risk we take no responsibility
//
//     Schemes: http, https
//     Host: <your lattice host>
//     BasePath: /v1
//     Version: 0.0.1
//     License: MIT http://opensource.org/licenses/MIT
//     Contact: mLab Lattice Team<team@mlab-lattice.org> http://mlab-lattice.org
//
//     Consumes:
//     - application/json
//
//     Produces:
//     - application/json
//
//     Security:
//     - api_key:
//
//     SecurityDefinitions:
//     api_key:
//          type: apiKey
//          name: apiKey
//          in: header
//
// swagger:meta
package v1

import (
	"github.com/gin-gonic/gin"

	backendv1 "github.com/mlab-lattice/lattice/pkg/api/server/backend/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/component/resolver"

	"github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"
)

type LatticeAPI struct {
	router   *gin.RouterGroup
	backend  backendv1.Interface
	resolver resolver.Interface
}

func newLatticeAPI(router *gin.RouterGroup, backend backendv1.Interface, resolver resolver.Interface) *LatticeAPI {
	return &LatticeAPI{
		router:   router,
		backend:  backend,
		resolver: resolver,
	}
}

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
