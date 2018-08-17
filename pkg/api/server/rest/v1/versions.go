package v1

import (
	"fmt"

	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	v1rest "github.com/mlab-lattice/lattice/pkg/api/v1/rest"
)

func (api *LatticeAPI) setupVersionsEndpoints() {
	systemIDPathComponent := fmt.Sprintf(":%v", systemIdentifier)
	versionsPath := fmt.Sprintf(v1rest.VersionsPathFormat, systemIDPathComponent)

	// list-system-versions
	api.router.GET(versionsPath, api.handleListSystemVersions)

}

// ListSystemVersions godoc
// @ID list-system-versions
// @Summary Lists system versions
// @Description list teardowns
// @Router /systems/{system}/versions [get]
// @Tags versions
// @Param system path string true "System ID"
// @Accept  json
// @Produce  json
// @Success 200 {array} v1.SystemVersion
func (api *LatticeAPI) handleListSystemVersions(c *gin.Context) {
	systemID := v1.SystemID(c.Param(systemIdentifier))

	versionStrings, err := getSystemVersions(api.backend, api.sysResolver, systemID)
	if err != nil {
		handleError(c, err)
		return
	}

	versions := make([]v1.SystemVersion, 0)
	for _, version := range versionStrings {
		versions = append(versions, v1.SystemVersion(version))
	}

	c.JSON(http.StatusOK, versions)
}
