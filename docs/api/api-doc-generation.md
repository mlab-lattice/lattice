# Lattice API doc generation

Lattice API docs are generated with slate + swagger. The slate fork for lattice is located in https://github.com/mlab-lattice/lattice-api-docs


Swagger files are generated using `goswagger` framework from annotations. The main API description live in the following file:

https://github.com/mlab-lattice/lattice/blob/19b06b31b003aef1431c841bb05724af70a75495/pkg/api/server/rest/v1/api.go#L1-L31

```go
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
```

Endpoint docs are annotated on each endpoint handler. For instance, systems endpoints are in
https://github.com/mlab-lattice/lattice/blob/718f8c1c075998517d66bb6c1a24ba6af1a8ccfa/pkg/api/server/rest/v1/systems.go#L45

```go
// swagger:operation POST /systems systems CreateSystem
//
// Creates systems
//
// Creates new systems
// ---
//     consumes:
//     - application/json
//     produces:
//     - application/json
//
//     parameters:
//       - in: body
//         schema:
//           "$ref": "#/definitions/CreateSystemRequest"
//     responses:
//         default:
//           description: System object
//           schema:
//             "$ref": "#/definitions/System"
//

// handleCreateSystem handler for CreateSystem
func (api *LatticeAPI) handleCreateSystem(c *gin.Context) {

	var req v1rest.CreateSystemRequest
	if err := c.BindJSON(&req); err != nil {
		handleBadRequestBody(c)
		return
	}

	system, err := api.backend.Systems().Create(req.ID, req.DefinitionURL)
	if err != nil {
		v1err, ok := err.(*v1.Error)
		if !ok {
			handleInternalError(c, err)
			return
		}

		switch v1err.Code {
		case v1.ErrorCodeSystemAlreadyExists:
			c.JSON(http.StatusConflict, v1err)

		case v1.ErrorCodeInvalidSystemOptions:
			c.JSON(http.StatusBadRequest, v1err)

		default:
			handleInternalError(c, err)
		}
		return
	}

	c.JSON(http.StatusCreated, system)

}
```

Schemas docs are generated from go struct by annotating the go struct with `swagger:model [name]`. Here is the schema for `System` for instance:

```go
// swagger:model System
type System struct {
	ID SystemID `json:"id"`

	DefinitionURL string `json:"definitionUrl"`

	Status SystemStatus `json:"status"`
}
```

To generate lattice api docs, you need to install the following tools first:


1- Install golang (if you don't have it already) and add $GOPATH/bin to $PATH

2- Set `$GOPATH` if you don't have it. You can set it to `$HOME/go`.

3- Clone this repository under `$GOPATH/src/github.com/mlab-lattice`. So your clone should be in `$GOPATH/src/github.com/mlab-lattice/lattice`
` 
4- Install https://goswagger.io: Converts Go annotations to swagger.
 
``$ brew install go-swagger``


5- Install https://github.com/Mermade/widdershins: Generates slate docs from swagger.

``$ npm install -g widdershins``

6- Install ruby (if you don't have it already)

``
$ brew update
$ brew install ruby
``

7- https://bundler.io/ to be used to run middleman which will create static pages

``$ gem install bundler``  


After you have all these installed, run:


``$ make docs.api``

This will generate `$GOPATH/src/github.com/mlab-lattice/lattice/api-docs/build` which will contain the static pages for documentation.