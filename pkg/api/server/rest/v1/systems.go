package v1

import (
	"fmt"
	"net/http"

	v1server "github.com/mlab-lattice/lattice/pkg/api/server/v1"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	v1rest "github.com/mlab-lattice/lattice/pkg/api/v1/rest"
	"github.com/mlab-lattice/lattice/pkg/definition/resolver"
	definitionv1 "github.com/mlab-lattice/lattice/pkg/definition/v1"
	"github.com/mlab-lattice/lattice/pkg/util/git"

	"io"
	"strconv"

	"github.com/gin-gonic/gin"
)

const (
	systemIdentifier = "system_id"
	buildIdentifier  = "build_id"
)

var systemIdentifierPathComponent = fmt.Sprintf(":%v", systemIdentifier)
var systemPath = fmt.Sprintf(v1rest.SystemPathFormat, systemIdentifierPathComponent)

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

// CreateSystem godoc
// @ID create-system
// @Summary Create a new system
// @Description create system
// @Router /v1/systems [post]
// @Param account body rest.CreateSystemRequest true "Create system"
// @Accept  json
// @Produce  json
// @Success 200 {object} v1.System
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

// ListSystems godoc
// @ID list-systems
// @Summary Lists systems
// @Description list systems
// @Router /v1/systems [get]
// @Accept  json
// @Produce  json
// @Success 200 {array} v1.System
func (api *LatticeAPI) handleListSystems(c *gin.Context) {
	systems, err := api.backend.ListSystems()
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, systems)
}

// GetSystem godoc
// @ID get-system
// @Summary Get system
// @Description get system
// @Router /v1/systems/{id} [get]
// @Param id path string true "System ID"
// @Accept  json
// @Produce  json
// @Success 200 {object} v1.System
func (api *LatticeAPI) handleGetSystem(c *gin.Context) {
	systemID := v1.SystemID(c.Param(systemIdentifier))
	system, err := api.backend.GetSystem(systemID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, system)
}

// DeleteSystem godoc
// @ID delete-system
// @Summary Delete system
// @Description get system
// @Router /v1/systems/{id} [delete]
// @Accept  json
// @Produce  json
// @Param id path string true "System ID"
// @Success 200 {object} v1.Result
func (api *LatticeAPI) handleDeleteSystem(c *gin.Context) {
	systemID := v1.SystemID(c.Param(systemIdentifier))

	err := api.backend.DeleteSystem(systemID)
	if err != nil {
		handleError(c, err)
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

	logOptions := v1.NewContainerLogOptions()
	logOptions.Follow = follow
	logOptions.Timestamps = timestamps
	logOptions.Previous = previous
	logOptions.Tail = tail
	logOptions.Since = since
	logOptions.SinceTime = sinceTime

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

func getSystemDefinitionRoot(
	backend v1server.Interface,
	sysResolver resolver.SystemResolver,
	systemID v1.SystemID,
	version v1.SystemVersion,
) (*definitionv1.SystemNode, error) {
	system, err := backend.GetSystem(systemID)
	if err != nil {
		return nil, err
	}

	systemDefURI := fmt.Sprintf(
		"%v#%v/%s",
		system.DefinitionURL,
		version,
		definitionv1.SystemDefinitionRootPathDefault,
	)

	root, err := sysResolver.ResolveDefinition(systemDefURI, &git.Options{})
	if err != nil {
		return nil, err
	}

	if def, ok := root.(*definitionv1.SystemNode); ok {
		return def, nil
	}

	return nil, fmt.Errorf("definition is not a system")
}

func getSystemVersions(backend v1server.Interface, sysResolver resolver.SystemResolver, systemID v1.SystemID) ([]string, error) {
	system, err := backend.GetSystem(systemID)
	if err != nil {
		return nil, err
	}

	return sysResolver.ListDefinitionVersions(system.DefinitionURL, &git.Options{})
}

type Result struct {
}
