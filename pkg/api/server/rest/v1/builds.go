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

// swagger:operation POST /systems/{system}/builds builds BuildSystem
//
// Build system
//
// Builds the system
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
//           "$ref": "#/definitions/BuildRequest"
//     responses:
//         '200':
//           description: Build object
//           schema:
//             "$ref": "#/definitions/Build"

// handleBuildSystem handler for BuildSystem
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
			handleInternalError(c, err)
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

	c.JSON(http.StatusCreated, build)

}

// swagger:operation GET /systems/{system}/builds builds ListBuilds
//
// Lists builds
//
// Lists builds for a system
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
//           description: build list
//           schema:
//             type: array
//             items:
//               "$ref": "#/definitions/Build"
//

// handleListBuilds handler for ListBuilds
func (api *LatticeAPI) handleListBuilds(c *gin.Context) {
	systemID := v1.SystemID(c.Param(systemIdentifier))

	builds, err := api.backend.Systems().Builds(systemID).List()
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

	c.JSON(http.StatusOK, builds)
}

// swagger:operation GET /systems/{system}/build/{buildId} builds GetBuild
//
// Get build
//
// Get build
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
//       - description: Build ID
//         in: path
//         name: buildId
//         required: true
//         type: string
//
//     responses:
//         '200':
//           description: Build Object
//           schema:
//             "$ref": "#/definitions/Build"
//

// handleGetBuild handler for GetBuild
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

		case v1.ErrorCodeSystemDeleting, v1.ErrorCodeSystemPending:
			c.JSON(http.StatusConflict, v1err)

		default:
			handleInternalError(c, err)
		}
		return
	}

	c.JSON(http.StatusOK, build)
}

// swagger:operation GET /systems/{system}/jobs/{jobId}/logs jobs getJobLogs
//
// Get job logs
//
// Returns a build log stream
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
//       - description: Buikld ID
//         in: path
//         name: buildId
//         required: true
//         type: string
//       - description: Sidecar
//         in: query
//         name: sidecar
//         required: false
//         type: string
//       - description: Follow
//         in: query
//         name: follow
//         required: false
//         type: boolean
//       - description: Previous
//         in: query
//         name: previous
//         required: false
//         type: boolean
//       - description: Timestamps
//         in: query
//         name: timestamps
//         required: false
//         type: boolean
//       - description: Tail
//         in: query
//         name: tail
//         required: false
//         type: int
//       - description: Since
//         in: query
//         name: since
//         required: false
//         type: string
//       - description: Since Time
//         in: query
//         name: sinceTime
//         required: false
//         type: string
//     responses:
//         '200':
//           description: log stream
//           schema:
//             type: string

// handleGetJobLogs handler for GetJobLogs
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
// @Failure 404 ""
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
			handleInternalError(c, err)
			return
		}

		switch v1err.Code {
		case v1.ErrorCodeInvalidSystemID, v1.ErrorCodeInvalidBuildID,
			v1.ErrorCodeInvalidPath, v1.ErrorCodeInvalidSidecar:
			c.JSON(http.StatusNotFound, v1err)

		case v1.ErrorCodeSystemDeleting, v1.ErrorCodeSystemPending:
			c.JSON(http.StatusConflict, v1err)

		default:
			handleInternalError(c, err)
		}
		return
	}

	if log == nil {
		c.Status(http.StatusOK)
		return
	}

	serveLogFile(log, c)
}
