package v1

import (
	"fmt"
	"net/http"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	v1rest "github.com/mlab-lattice/lattice/pkg/api/v1/rest"

	"github.com/gin-gonic/gin"
)

const teardownIdentifier = "teardown_id"

var (
	teardownsPath                   = fmt.Sprintf(v1rest.TeardownsPathFormat, systemIdentifierPathComponent)
	teardownIdentifierPathComponent = fmt.Sprintf(":%v", teardownIdentifier)
	teardownPath                    = fmt.Sprintf(v1rest.TeardownPathFormat, systemIdentifierPathComponent, teardownIdentifierPathComponent)
)

func (api *LatticeAPI) setupTeardownEndpoints() {

	// tear-down-system
	api.router.POST(teardownsPath, api.handleTeardownSystem)

	// list-teardowns
	api.router.GET(teardownsPath, api.handleListTeardowns)

	// get-teardown
	api.router.GET(teardownPath, api.handleGetTeardown)

}

// swagger:operation POST /systems/{system}/teardown teardowns TeardownSystem
//
// Teardown system
//
// Teardown system
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
//     responses:
//         '200':
//           description: Teardown object
//           schema:
//             "$ref": "#/definitions/Teardown"

// handleTeardownSystem handler for TeardownSystem
func (api *LatticeAPI) handleTeardownSystem(c *gin.Context) {
	systemID := v1.SystemID(c.Param(systemIdentifier))

	teardown, err := api.backend.Systems().Teardowns(systemID).Create()
	if err != nil {
		v1err, ok := err.(*v1.Error)
		if !ok {
			handleInternalError(c, err)
			return
		}

		switch v1err.Code {
		case v1.ErrorCodeSystemDeleting, v1.ErrorCodeSystemPending:
			c.JSON(http.StatusConflict, v1err)

		default:
			handleInternalError(c, err)
		}
		return
	}

	c.JSON(http.StatusCreated, teardown)
}

// swagger:operation GET /systems/{system}/teardowns teardowns ListTeardowns
//
// Lists teardowns
//
// Lists teardowns
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
//           description: teardown list
//           schema:
//             type: array
//             items:
//               "$ref": "#/definitions/Teardown"
//

// handleListTeardowns handler for ListTeardowns
func (api *LatticeAPI) handleListTeardowns(c *gin.Context) {
	systemID := v1.SystemID(c.Param(systemIdentifier))

	teardowns, err := api.backend.Systems().Teardowns(systemID).List()
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

	c.JSON(http.StatusOK, teardowns)
}

// swagger:operation GET /systems/{system}/teardowns/{teardownId} teardowns GetTeardown
//
// Get teardown
//
// Get teardown
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
//       - description: Teardown ID
//         in: path
//         name: teardownId
//         required: true
//         type: string
//
//     responses:
//         '200':
//           description: Teardown Object
//           schema:
//             "$ref": "#/definitions/Teardown"
//

// handleGetTeardown handler for GetTeardown
func (api *LatticeAPI) handleGetTeardown(c *gin.Context) {
	systemID := v1.SystemID(c.Param(systemIdentifier))
	teardownID := v1.TeardownID(c.Param(teardownIdentifier))

	teardown, err := api.backend.Systems().Teardowns(systemID).Get(teardownID)
	if err != nil {
		v1err, ok := err.(*v1.Error)
		if !ok {
			handleInternalError(c, err)
			return
		}

		switch v1err.Code {
		case v1.ErrorCodeInvalidSystemID, v1.ErrorCodeInvalidTeardownID:
			c.JSON(http.StatusNotFound, v1err)

		case v1.ErrorCodeSystemDeleting, v1.ErrorCodeSystemPending:
			c.JSON(http.StatusConflict, v1err)

		default:
			handleInternalError(c, err)
		}
		return
	}

	c.JSON(http.StatusOK, teardown)
}
