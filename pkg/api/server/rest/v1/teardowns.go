package v1

import (
	"fmt"

	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	v1rest "github.com/mlab-lattice/lattice/pkg/api/v1/rest"
)

const teardownIdentifier = "teardown_id"

var teardownsPath = fmt.Sprintf(v1rest.TeardownsPathFormat, systemIdentifierPathComponent)
var teardownIdentifierPathComponent = fmt.Sprintf(":%v", teardownIdentifier)
var teardownPath = fmt.Sprintf(v1rest.TeardownPathFormat, systemIdentifierPathComponent,
	teardownIdentifierPathComponent)

func (api *LatticeAPI) setupTeardownEndpoints() {

	// tear-down-system
	api.router.POST(teardownsPath, api.handleTeardownSystem)

	// list-teardowns
	api.router.GET(teardownsPath, api.handleListTeardowns)

	// get-teardown
	api.router.GET(teardownPath, api.handleGetTeardown)

}

// handleTeardownSystem handler for teardown-system
// @ID teardown-system
// @Summary Teardown system
// @Description Tears the system down
// @Router /systems/{system}/teardowns [post]
// @Security ApiKeyAuth
// @Tags teardowns
// @Param system path string true "System ID"
// @Accept  json
// @Produce  json
// @Success 200 {object} v1.Teardown
// @Failure 400 {object} v1.ErrorResponse
// @Failure 404 {object} v1.ErrorResponse
func (api *LatticeAPI) handleTeardownSystem(c *gin.Context) {
	systemID := v1.SystemID(c.Param(systemIdentifier))

	teardown, err := api.backend.TearDown(systemID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, teardown)
}

// handleListTeardowns handler for list-teardowns
// @ID list-teardowns
// @Summary Lists teardowns
// @Description Lists all teardowns made to the system
// @Router /systems/{system}/teardowns [get]
// @Security ApiKeyAuth
// @Tags teardowns
// @Param system path string true "System ID"
// @Accept  json
// @Produce  json
// @Success 200 {array} v1.Teardown
func (api *LatticeAPI) handleListTeardowns(c *gin.Context) {
	systemID := v1.SystemID(c.Param(systemIdentifier))

	teardowns, err := api.backend.ListTeardowns(systemID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, teardowns)
}

// handleGetTeardown handler for get-teardown
// @ID get-teardown
// @Summary Get teardown
// @Description Gets the teardown object
// @Router /systems/{system}/teardowns/{id} [get]
// @Security ApiKeyAuth
// @Tags teardowns
// @Param system path string true "System ID"
// @Param id path string true "Teardown ID"
// @Accept  json
// @Produce  json
// @Success 200 {object} v1.Teardown
// @Failure 404 {object} v1.ErrorResponse
func (api *LatticeAPI) handleGetTeardown(c *gin.Context) {
	systemID := v1.SystemID(c.Param(systemIdentifier))
	teardownID := v1.TeardownID(c.Param(teardownIdentifier))

	teardown, err := api.backend.GetTeardown(systemID, teardownID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, teardown)
}
