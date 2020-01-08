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

// swagger:operation GET /systems/{system}/versions versions ListVersions
//
// Lists versions
//
// Lists versions
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
//           description: version list
//           schema:
//             type: array
//             items:
//               type: string
//

// handleListSystemVersions handler for ListVersions
func (api *LatticeAPI) handleListSystemVersions(c *gin.Context) {
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
			c.JSON(http.StatusNotFound, v1err)

		default:
			handleInternalError(c, err)
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
		handleInternalError(c, err)
		return
	}

	versions := make([]v1.Version, 0)
	for _, version := range v {
		versions = append(versions, v1.Version(version))
	}

	c.JSON(http.StatusOK, versions)
}
