package v1

import (
	"github.com/gin-gonic/gin"

	v1server "github.com/mlab-lattice/lattice/pkg/api/server/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/resolver"
	"github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"
)

type LatticeAPI struct {
	router   *gin.RouterGroup
	backend  v1server.Interface
	resolver resolver.ComponentResolver
}

func newLatticeAPI(router *gin.RouterGroup, backend v1server.Interface, resolver resolver.ComponentResolver) *LatticeAPI {
	return &LatticeAPI{
		router:   router,
		backend:  backend,
		resolver: resolver,
	}
}

// @title Lattice API Docs
// @version 1.0
// @description This document describes the lattice API.
// @termsOfService TBD
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
