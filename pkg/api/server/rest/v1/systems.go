package v1

import (
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	v1rest "github.com/mlab-lattice/lattice/pkg/api/v1/rest"

	"github.com/gin-gonic/gin"
)

const (
	systemIdentifier = "system_id"
	buildIdentifier  = "build_id"
)

var (
	systemIdentifierPathComponent = fmt.Sprintf(":%v", systemIdentifier)
	systemPath                    = fmt.Sprintf(v1rest.SystemPathFormat, systemIdentifierPathComponent)
)

func (api *LatticeAPI) setupSystemEndpoints() {
	// create-system
	api.router.POST(v1rest.SystemsPath, api.handleCreateSystem)

	// list-systems
	api.router.GET(v1rest.SystemsPath, api.handleListSystems)

	// get-system
	api.router.GET(systemPath, api.handleGetSystem)

	// delete-system
	api.router.DELETE(systemPath, api.handleDeleteSystem)
}

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

// swagger:operation GET /systems systems ListSystems
//
// Lists systems
//
// List all systems
// ---
//     consumes:
//     - application/json
//     produces:
//     - application/json
//
//
//     responses:
//         '200':
//           description: system list
//           schema:
//             type: array
//             items:
//               "$ref": "#/definitions/System"
//
// handleListSystems handler for ListSystems
func (api *LatticeAPI) handleListSystems(c *gin.Context) {
	systems, err := api.backend.Systems().List()
	if err != nil {
		handleInternalError(c, err)
		return
	}

	c.JSON(http.StatusOK, systems)
}

// swagger:operation GET /systems/{systemId} systems GetSystem
//
// Get system
//
// Get system
// ---
//     consumes:
//     - application/json
//     produces:
//     - application/json
//
//     parameters:
//       - description: System ID
//         in: path
//         name: systemId
//         required: true
//         type: string
//
//     responses:
//         '200':
//           description: System Object
//           schema:
//             "$ref": "#/definitions/System"
//
// handleGetSystem handler for GetSystem
func (api *LatticeAPI) handleGetSystem(c *gin.Context) {
	systemID := v1.SystemID(c.Param(systemIdentifier))

	system, err := api.backend.Systems().Get(systemID)
	if err != nil {
		v1err, ok := err.(*v1.Error)
		if !ok {
			handleInternalError(c, err)
			return
		}

		switch v1err.Code {
		case v1.ErrorCodeInvalidSystemID:
			c.JSON(http.StatusBadRequest, v1err)

		case v1.ErrorCodeSystemDeleting, v1.ErrorCodeSystemPending:
			c.JSON(http.StatusConflict, v1err)

		default:
			handleInternalError(c, err)
		}
		return
	}

	c.JSON(http.StatusOK, system)
}

// swagger:operation DELETE /systems/{systemId} systems DeleteSystem
//
// Delete system
//
// Delete system
// ---
//     consumes:
//     - application/json
//     produces:
//     - application/json
//
//     parameters:
//       - description: System ID
//         in: path
//         name: systemId
//         required: true
//         type: string
//
//     responses:
//         '200':
//

// handleDeleteSystem handler for DeleteSystem
func (api *LatticeAPI) handleDeleteSystem(c *gin.Context) {
	systemID := v1.SystemID(c.Param(systemIdentifier))

	err := api.backend.Systems().Delete(systemID)
	if err != nil {
		v1err, ok := err.(*v1.Error)
		if !ok {
			handleInternalError(c, err)
			return
		}

		switch v1err.Code {
		case v1.ErrorCodeInvalidSystemID:
			c.JSON(http.StatusBadRequest, v1err)

		case v1.ErrorCodeConflict:
			c.JSON(http.StatusConflict, v1err)

		case v1.ErrorCodeSystemDeleting:
			c.JSON(http.StatusConflict, v1err)

		default:
			handleInternalError(c, err)
		}
		return
	}

	c.Status(http.StatusOK)

}

// requestedLogOptions
func requestedLogOptions(c *gin.Context) (*v1.ContainerLogOptions, error) {
	// follow
	follow, err := strconv.ParseBool(c.DefaultQuery("follow", "false"))
	if err != nil {
		return nil, err
	}
	// previous
	previous, err := strconv.ParseBool(c.DefaultQuery("previous", "false"))
	if err != nil {
		return nil, err
	}
	//timestamps
	timestamps, err := strconv.ParseBool(c.DefaultQuery("timestamps", "false"))
	if err != nil {
		return nil, err
	}
	// tail
	var tail *int64
	tailStr := c.Query("tail")
	if tailStr != "" {
		lines, err := strconv.ParseInt(tailStr, 10, 64)
		if err != nil {
			return nil, err
		}
		tail = &lines
	}

	// since
	since := c.Query("since")

	// sinceTime
	sinceTime := c.Query("sinceTime")

	logOptions := &v1.ContainerLogOptions{
		Follow:     follow,
		Timestamps: timestamps,
		Previous:   previous,
		Tail:       tail,
		Since:      since,
		SinceTime:  sinceTime,
	}

	return logOptions, nil
}

// serveLogFile
func serveLogFile(log io.ReadCloser, c *gin.Context) {
	defer log.Close()

	buff := make([]byte, 1024)

	c.Stream(func(w io.Writer) bool {
		n, err := log.Read(buff)
		if err != nil {
			return false
		}

		w.Write(buff[:n])
		return true
	})
}
