package v1

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	v1rest "github.com/mlab-lattice/lattice/pkg/api/v1/rest"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	reflectutil "github.com/mlab-lattice/lattice/pkg/util/reflect"
)

var (
	buildIdentifierPathComponent = fmt.Sprintf(":%v", buildIdentifier)
	buildsPath                   = fmt.Sprintf(v1rest.BuildsPathFormat, systemIdentifierPathComponent)
	buildPath                    = fmt.Sprintf(v1rest.BuildPathFormat, systemIdentifierPathComponent, buildIdentifierPathComponent)
	buildsLogPath                = fmt.Sprintf(v1rest.BuildLogsPathFormat, systemIdentifierPathComponent, buildIdentifierPathComponent)
)

func (api *LatticeAPI) setupBuildEndpoints() {
	// build-system
	api.router.POST(buildsPath, api.handleBuildSystem)

	// list-builds
	api.router.GET(buildsPath, api.handleListBuilds)

	// get-build
	api.router.GET(buildPath, api.handleGetBuild)

	// get-build-logs
	api.router.GET(buildsLogPath, api.handleGetBuildLogs)

}

// handleBuildSystem handler for build-system
// @ID build-system
// @Summary Build system
// @Description Builds the system
// @Router /systems/{system}/builds [post]
// @Security ApiKeyAuth
// @Tags builds
// @Security ApiKeyAuth
// @Param system path string true "System ID"
// @Param buildRequest body rest.BuildRequest true "Create build"
// @Accept  json
// @Produce  json
// @Success 200 {object} v1.Build
// @Failure 400 {object} v1.ErrorResponse
func (api *LatticeAPI) handleBuildSystem(c *gin.Context) {
	systemID := v1.SystemID(c.Param(systemIdentifier))

	var req v1rest.BuildRequest
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
			c.Status(http.StatusInternalServerError)
		}
		return
	}

	var build *v1.Build
	switch {
	case req.Path != nil:
		build, err = api.backend.Systems().Builds(systemID).CreateFromPath(*req.Path)

	case req.Version != nil:
		build, err = api.backend.Systems().Builds(systemID).CreateFromVersion(*req.Version)
	}

	if err != nil {
		v1err, ok := err.(*v1.Error)
		if !ok {
			c.Status(http.StatusInternalServerError)
			return
		}

		switch v1err.Code {
		case v1.ErrorCodeInvalidSystemID, v1.ErrorCodeInvalidSystemVersion:
			c.JSON(http.StatusNotFound, v1err)

		default:
			c.Status(http.StatusInternalServerError)
		}
		return
	}

	c.JSON(http.StatusCreated, build)

}

// handleListBuilds handler for list-builds
// @ID list-builds
// @Summary List builds
// @Description List all builds for the system
// @Router /systems/{system}/builds [get]
// @Security ApiKeyAuth
// @Tags builds
// @Param system path string true "System ID"
// @Accept  json
// @Produce  json
// @Success 200 {array} v1.Build
func (api *LatticeAPI) handleListBuilds(c *gin.Context) {
	systemID := v1.SystemID(c.Param(systemIdentifier))

	builds, err := api.backend.Systems().Builds(systemID).List()
	if err != nil {
		v1err, ok := err.(*v1.Error)
		if !ok {
			c.Status(http.StatusInternalServerError)
			return
		}

		switch v1err.Code {
		case v1.ErrorCodeInvalidSystemID:
			c.JSON(http.StatusNotFound, v1err)

		default:
			c.Status(http.StatusInternalServerError)
		}
		return
	}

	c.JSON(http.StatusOK, builds)
}

// handleGetBuild handler for get-build
// @ID get-build
// @Summary Get build
// @Description Gets the build object
// @Router /systems/{system}/builds/{id} [get]
// @Security ApiKeyAuth
// @Tags builds
// @Param system path string true "System ID"
// @Param id path string true "Build ID"
// @Accept  json
// @Produce  json
// @Success 200 {object} v1.Build
// @Failure 404 {object} v1.ErrorResponse
func (api *LatticeAPI) handleGetBuild(c *gin.Context) {
	systemID := v1.SystemID(c.Param(systemIdentifier))
	buildID := v1.BuildID(c.Param(buildIdentifier))

	build, err := api.backend.Systems().Builds(systemID).Get(buildID)
	if err != nil {
		v1err, ok := err.(*v1.Error)
		if !ok {
			c.Status(http.StatusInternalServerError)
			return
		}

		switch v1err.Code {
		case v1.ErrorCodeInvalidSystemID, v1.ErrorCodeInvalidBuildID:
			c.JSON(http.StatusNotFound, v1err)

		default:
			c.Status(http.StatusInternalServerError)
		}
		return
	}

	c.JSON(http.StatusOK, build)
}

// handleGetBuildLogs handler for get-build-logs
// @ID get-build-logs
// @Summary Get build logs
// @Description Retrieves/Streams logs for build
// @Router /systems/{system}/builds/{id}/logs  [get]
// @Security ApiKeyAuth
// @Tags builds
// @Param system path string true "System ID"
// @Param id path string true "Build ID"
// @Param path query string true "Node Path"
// @Param sidecar query string false "Sidecar"
// @Param follow query string bool "Follow"
// @Param previous query boolean false "Previous"
// @Param timestamps query boolean false "Timestamps"
// @Param tail query integer false "tail"
// @Param since query string false "Since"
// @Param sinceTime query string false "Since Time"
// @Accept  json
// @Produce  json
// @Success 200 {string} string "log stream"
// @Failure 404 {object} v1.ErrorResponse
func (api *LatticeAPI) handleGetBuildLogs(c *gin.Context) {
	systemID := v1.SystemID(c.Param(systemIdentifier))
	buildID := v1.BuildID(c.Param(buildIdentifier))
	pathStr := c.Query("path")

	sidecarQuery, sidecarSet := c.GetQuery("sidecar")
	var sidecar *string
	if sidecarSet {
		sidecar = &sidecarQuery
	}

	if pathStr == "" {
		c.Status(http.StatusBadRequest)
		return
	}

	path, err := tree.NewPath(pathStr)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	logOptions, err := requestedLogOptions(c)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	log, err := api.backend.Systems().Builds(systemID).Logs(buildID, path, sidecar, logOptions)
	if err != nil {
		v1err, ok := err.(*v1.Error)
		if !ok {
			c.Status(http.StatusInternalServerError)
			return
		}

		switch v1err.Code {
		case v1.ErrorCodeInvalidSystemID, v1.ErrorCodeInvalidBuildID,
			v1.ErrorCodeInvalidPath, v1.ErrorCodeInvalidSidecar:
			c.JSON(http.StatusNotFound, v1err)

		default:
			c.Status(http.StatusInternalServerError)
		}
		return
	}

	if log == nil {
		c.Status(http.StatusOK)
		return
	}

	serveLogFile(log, c)
}
