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

	system, err := api.backend.Systems().Create(req.ID, req.DefinitionURL)
	if err != nil {
		v1err, ok := err.(*v1.Error)
		if !ok {
			c.Status(http.StatusInternalServerError)
			return
		}

		switch v1err.Code {
		case v1.ErrorCodeSystemAlreadyExists:
			c.JSON(http.StatusConflict, v1err)

		case v1.ErrorCodeInvalidSystemOptions:
			c.JSON(http.StatusBadRequest, v1err)

		default:
			c.Status(http.StatusInternalServerError)
		}
		return
	}

	c.JSON(http.StatusCreated, system)

}

// handleListSystems handler for list-systems
// @ID list-systems
// @Summary List systems
// @Description List systems
// @Router /systems [get]
// @Security ApiKeyAuth
// @Tags systems
// @Accept  json
// @Produce  json
// @Success 200 {array} v1.System
func (api *LatticeAPI) handleListSystems(c *gin.Context) {
	systems, err := api.backend.Systems().List()
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, systems)
}

// handleGetSystem handler for get-system
// @ID get-system
// @Summary Get system
// @Description get system
// @Router /systems/{system} [get]
// @Security ApiKeyAuth
// @Tags systems
// @Param system path string true "System ID"
// @Accept  json
// @Produce  json
// @Success 200 {object} v1.System
// @Failure 404 {object} v1.ErrorResponse
func (api *LatticeAPI) handleGetSystem(c *gin.Context) {
	systemID := v1.SystemID(c.Param(systemIdentifier))

	system, err := api.backend.Systems().Get(systemID)
	if err != nil {
		v1err, ok := err.(*v1.Error)
		if !ok {
			c.Status(http.StatusInternalServerError)
			return
		}

		switch v1err.Code {
		case v1.ErrorCodeInvalidSystemID:
			c.JSON(http.StatusBadRequest, v1err)

		case v1.ErrorCodeSystemDeleting, v1.ErrorCodeSystemPending:
			c.JSON(http.StatusConflict, v1err)

		default:
			c.Status(http.StatusInternalServerError)
		}
		return
	}

	c.JSON(http.StatusOK, system)
}

// handleDeleteSystem handler for delete-system
// @ID delete-system
// @Summary Delete system
// @Description Delete system
// @Router /systems/{system} [delete]
// @Security ApiKeyAuth
// @Tags systems
// @Accept  json
// @Produce  json
// @Param system path string true "System ID"
// @Success 200 {object} v1.Result
// @Failure 404 {object} v1.ErrorResponse
func (api *LatticeAPI) handleDeleteSystem(c *gin.Context) {
	systemID := v1.SystemID(c.Param(systemIdentifier))

	err := api.backend.Systems().Delete(systemID)
	if err != nil {
		v1err, ok := err.(*v1.Error)
		if !ok {
			c.Status(http.StatusInternalServerError)
			return
		}

		switch v1err.Code {
		case v1.ErrorCodeConflict:
			c.JSON(http.StatusConflict, v1err)

		case v1.ErrorCodeSystemDeleting:
			c.JSON(http.StatusConflict, v1err)

		default:
			c.Status(http.StatusInternalServerError)
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
