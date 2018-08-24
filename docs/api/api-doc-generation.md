# Lattice API doc generation

Lattice API docs are generated from annotations using `swag` go library. The main API description live in the following file:

https://github.com/mlab-lattice/lattice/blob/44617829f866ccf01b228464044eb7057e5649fb/pkg/api/server/rest/v1/api.go#L33

```go
 @title Lattice API Docs
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

```

Endpoint docs are annotated on each endpoint handler. For instance, systems endpoints are in
https://github.com/mlab-lattice/lattice/blob/44617829f866ccf01b228464044eb7057e5649fb/pkg/api/server/rest/v1/systems.go#L42

```go
// CreateSystem godoc
// @ID create-system
// @Summary Create system
// @Description Create a new system
// @Router /systems [post]
// @Tags systems
// @Param account body rest.CreateSystemRequest true "Create system"
// @Accept  json
// @Produce  json
// @Success 200 {object} v1.System
// @Failure 400 {object} v1.ErrorResponse
func (api *LatticeAPI) handleCreateSystem(c *gin.Context) {

	var req v1rest.CreateSystemRequest
	if err := c.BindJSON(&req); err != nil {
		handleBadRequestBody(c)
		return
	}

	system, err := api.backend.CreateSystem(req.ID, req.DefinitionURL)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, system)

}
```


To generate lattice api docs, you need to install the following tools first:

1- https://github.com/swaggo/swag: Converts Go annotations to swagger docs.
 
``$ go get -u github.com/swaggo/swag/cmd/swag``

2- https://github.com/apiaryio/swagger2blueprint: Converts swagger docs to API blueprint docs.

``$ npm install -g swagger2blueprint``

3- https://github.com/danielgtaylor/aglio: Renders API Blueprint

``$ npm install -g aglio``  

After you have all these installed, run:


``$ make api-docs``

This will generate `lattice-api.html` under lattice/pkg/api/docs