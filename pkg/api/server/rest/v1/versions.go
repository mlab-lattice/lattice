package v1

import (
	"fmt"
	"net/http"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	v1rest "github.com/mlab-lattice/lattice/pkg/api/v1/rest"
	definitionv1 "github.com/mlab-lattice/lattice/pkg/definition/v1"

	"github.com/gin-gonic/gin"
)

func (api *LatticeAPI) setupVersionsEndpoints() {
	systemIDPathComponent := fmt.Sprintf(":%v", systemIdentifier)
	versionsPath := fmt.Sprintf(v1rest.VersionsPathFormat, systemIDPathComponent)

	// list-system-versions
	api.router.GET(versionsPath, api.handleListSystemVersions)

}

// handleListSystemVersions handler for list-system-versions
// @ID list-system-versions
// @Summary Lists system versions
// @Description List all versions of the specified system
// @Router /systems/{system}/versions [get]
// @Security ApiKeyAuth
// @Tags versions
// @Param system path string true "System ID"
// @Accept  json
// @Produce  json
// @Success 200 {array} v1.SystemVersion
func (api *LatticeAPI) handleListSystemVersions(c *gin.Context) {
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
			c.JSON(http.StatusNotFound, v1err)

		default:
			c.Status(http.StatusInternalServerError)
		}
		return
	}

	ref := &definitionv1.Reference{
		GitRepository: &definitionv1.GitRepositoryReference{
			GitRepository: &definitionv1.GitRepository{
				URL: system.DefinitionURL,
			},
		},
	}

	v, err := api.resolver.Versions(ref, nil)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	versions := make([]v1.SystemVersion, 0)
	for _, version := range v {
		versions = append(versions, v1.SystemVersion(version))
	}

	c.JSON(http.StatusOK, versions)
}
