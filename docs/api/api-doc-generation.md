# Lattice API doc generation

Lattice API docs are generated with slate + swagger. The slate fork for lattice is located in https://github.com/mlab-lattice/lattice-api-docs


Swagger files are generated using `swag` tool from annotations. The main API description live in the following file:

https://github.com/mlab-lattice/lattice/blob/9e2b722210d01d338ad00d498c36fdc7c07b8b40/pkg/api/server/rest/v1/api.go#L33

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
https://github.com/mlab-lattice/lattice/blob/9e2b722210d01d338ad00d498c36fdc7c07b8b40/pkg/api/server/rest/v1/systems.go#L42

```go
// handleCreateSystem handler for create-system
// @ID create-system
// @Summary Create system
// @Description Create a new system
// @Router /systems [post]
// @Security ApiKeyAuth
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

1- Install golang (if you don't have it already) and add $GOPATH/bin to $PATH

2- Install https://github.com/swaggo/swag: Converts Go annotations to swagger.
 
``$ go get -u github.com/swaggo/swag/cmd/swag``

3- Make sure that `swag` is available in your $PATH. Run `swag --help`.

4- Install https://github.com/Mermade/widdershins: Generates slate docs from swagger.

``$ npm install -g widdershins``

5- Install ruby (if you don't have it already)

``
$ brew update
$ brew install ruby
``

6- https://bundler.io/ to be used to run middleman which will create static pages

``$ gem install bundler``  


After you have all these installed, run:


``$ make docs.api``

This will generate `lattice/api-docs/build` which will contain the static pages for documentation.