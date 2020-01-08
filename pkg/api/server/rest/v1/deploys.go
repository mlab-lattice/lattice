package v1

import (
	"fmt"

	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	v1rest "github.com/mlab-lattice/lattice/pkg/api/v1/rest"
	reflectutil "github.com/mlab-lattice/lattice/pkg/util/reflect"
)

const deployIdentifier = "deploy_id"

var (
	deployIdentifierPathComponent = fmt.Sprintf(":%v", deployIdentifier)
	deployPath                    = fmt.Sprintf(v1rest.DeployPathFormat, systemIdentifierPathComponent, deployIdentifierPathComponent)
)

func (api *LatticeAPI) setupDeployEndpoints() {
	deploysPath := fmt.Sprintf(v1rest.DeploysPathFormat, systemIdentifierPathComponent)
	// deploy
	api.router.POST(deploysPath, api.handleDeploySystem)

	// list-deploys
	api.router.GET(deploysPath, api.handleListDeploys)

	// get-deploy
	api.router.GET(deployPath, api.handleGetDeploy)

}

// swagger:operation POST /systems/{system}/deploys deploys DeploySystem
//
// Deploy system
//
// Deploy system
// ---
//     consumes:
//     - application/json
//     produces:
//     - application/json
//
//     parameters:
//       - description: System ID
//         in: path
//         name: system
//         required: true
//         type: string
//       - in: body
//         schema:
//           "$ref": "#/definitions/DeployRequest"
//     responses:
//         '200':
//           description: Build object
//           schema:
//             "$ref": "#/definitions/Deploy"

// handleDeploySystem handler for DeploySystem
func (api *LatticeAPI) handleDeploySystem(c *gin.Context) {
	systemID := v1.SystemID(c.Param(systemIdentifier))

	var req v1rest.DeployRequest
	if err := c.BindJSON(&req); err != nil {
		handleBadRequestBody(c)
		return
	}

	err := reflectutil.ValidateUnion(&req)
	if err != nil {
		switch err.(type) {
		case *reflectutil.InvalidUnionNoFieldSetError, *reflectutil.InvalidUnionMultipleFieldSetError:
			c.Status(http.StatusBadRequest)

		default:
			handleInternalError(c, err)
		}
		return
	}

	var deploy *v1.Deploy
	switch {
	case req.BuildID != nil:
		deploy, err = api.backend.Systems().Deploys(systemID).CreateFromBuild(*req.BuildID)

	case req.Path != nil:
		deploy, err = api.backend.Systems().Deploys(systemID).CreateFromPath(*req.Path)

	case req.Version != nil:
		deploy, err = api.backend.Systems().Deploys(systemID).CreateFromVersion(*req.Version)
	}

	if err != nil {
		v1err, ok := err.(*v1.Error)
		if !ok {
			handleInternalError(c, err)
			return
		}

		switch v1err.Code {
		case v1.ErrorCodeInvalidSystemID, v1.ErrorCodeInvalidBuildID:
			c.JSON(http.StatusNotFound, v1err)

		case v1.ErrorCodeSystemDeleting, v1.ErrorCodeSystemPending:
			c.JSON(http.StatusConflict, v1err)

		default:
			handleInternalError(c, err)
		}
		return
	}

	c.JSON(http.StatusCreated, deploy)
}

// swagger:operation GET /systems/{system}/deploys deploys ListDeploys
//
// Lists deploys
//
// Lists deploys
// ---
//     consumes:
//     - application/json
//     produces:
//     - application/json
//
//     parameters:
//       - description: System ID
//         in: path
//         name: system
//         required: true
//         type: string
//
//     responses:
//         '200':
//           description: deploy list
//           schema:
//             type: array
//             items:
//               "$ref": "#/definitions/Deploy"
//

// handleListDeploys handler for ListDeploys
func (api *LatticeAPI) handleListDeploys(c *gin.Context) {
	systemID := v1.SystemID(c.Param(systemIdentifier))

	deploys, err := api.backend.Systems().Deploys(systemID).List()
	if err != nil {
		v1err, ok := err.(*v1.Error)
		if !ok {
			handleInternalError(c, err)
			return
		}

		switch v1err.Code {
		case v1.ErrorCodeInvalidSystemID:
			c.JSON(http.StatusNotFound, v1err)

		case v1.ErrorCodeSystemDeleting, v1.ErrorCodeSystemPending:
			c.JSON(http.StatusConflict, v1err)

		default:
			handleInternalError(c, err)
		}
		return
	}

	c.JSON(http.StatusOK, deploys)
}

// swagger:operation GET /systems/{system}/deploys/{deployId} deploys GetDeploy
//
// Get deploy
//
// Get deploy
// ---
//     consumes:
//     - application/json
//     produces:
//     - application/json
//
//     parameters:
//       - description: System ID
//         in: path
//         name: system
//         required: true
//         type: string
//       - description: Deploy ID
//         in: path
//         name: deployId
//         required: true
//         type: string
//
//     responses:
//         '200':
//           description: Deploy Object
//           schema:
//             "$ref": "#/definitions/Deploy"
//

// handleGetDeploy handler for GetDeploy
func (api *LatticeAPI) handleGetDeploy(c *gin.Context) {
	systemID := v1.SystemID(c.Param(systemIdentifier))
	deployID := v1.DeployID(c.Param(deployIdentifier))

	deploy, err := api.backend.Systems().Deploys(systemID).Get(deployID)
	if err != nil {
		v1err, ok := err.(*v1.Error)
		if !ok {
			handleInternalError(c, err)
			return
		}

		switch v1err.Code {
		case v1.ErrorCodeInvalidSystemID, v1.ErrorCodeInvalidDeployID:
			c.JSON(http.StatusNotFound, v1err)

		case v1.ErrorCodeSystemDeleting, v1.ErrorCodeSystemPending:
			c.JSON(http.StatusConflict, v1err)

		default:
			handleInternalError(c, err)
		}
		return
	}

	c.JSON(http.StatusOK, deploy)
}
