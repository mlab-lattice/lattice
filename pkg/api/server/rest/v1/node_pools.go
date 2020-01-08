package v1

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	v1rest "github.com/mlab-lattice/lattice/pkg/api/v1/rest"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"

	"github.com/gin-gonic/gin"
)

const nodePoolIdentifier = "node_pool_path"

var (
	nodePoolIdentifierPathComponent = fmt.Sprintf(":%v", nodePoolIdentifier)
	nodePoolPath                    = fmt.Sprintf(v1rest.NodePoolPathFormat, systemIdentifierPathComponent, nodePoolIdentifierPathComponent)
	nodePoolsPath                   = fmt.Sprintf(v1rest.NodePoolsPathFormat, systemIdentifierPathComponent)
)

func (api *LatticeAPI) setupNoodPoolEndpoints() {
	// list-node-pools
	api.router.GET(nodePoolsPath, api.handleListNodePools)

	// get-node-pool
	api.router.GET(nodePoolPath, api.handleGetNodePool)

}

// swagger:operation GET /systems/{system}/node-pools node-pools ListNodePools
//
// Lists node-pools
//
// Lists node-pools
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
//           description: node-pool list
//           schema:
//             type: array
//             items:
//               "$ref": "#/definitions/NodePool"
//
// handleListNodePools handler for ListNodePools
func (api *LatticeAPI) handleListNodePools(c *gin.Context) {
	systemID := v1.SystemID(c.Param(systemIdentifier))

	nodePools, err := api.backend.Systems().NodePools(systemID).List()
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

	c.JSON(http.StatusOK, nodePools)
}

// swagger:operation GET /systems/{system}/node-pools/{nodePoolId} node-pools GetNodePool
//
// Get node-pool
//
// Get node-pool
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
//         name: nodePoolId
//         required: true
//         type: string
//
//     responses:
//         '200':
//           description: Build Object
//           schema:
//             "$ref": "#/definitions/NodePool"
//
// handleGetNodePool handler for GetNodePool
func (api *LatticeAPI) handleGetNodePool(c *gin.Context) {
	systemID := v1.SystemID(c.Param(systemIdentifier))
	escapedNodePoolPath := c.Param(nodePoolIdentifier)

	nodePoolPathString, err := url.PathUnescape(escapedNodePoolPath)
	if err != nil {
		c.JSON(http.StatusBadRequest, v1.NewInvalidPathError())
		return
	}

	path, err := tree.NewPathSubcomponent(nodePoolPathString)
	if err != nil {
		c.JSON(http.StatusBadRequest, v1.NewInvalidPathError())
		return
	}

	nodePool, err := api.backend.Systems().NodePools(systemID).Get(path)
	if err != nil {
		v1err, ok := err.(*v1.Error)
		if !ok {
			handleInternalError(c, err)
			return
		}

		switch v1err.Code {
		case v1.ErrorCodeInvalidSystemID, v1.ErrorCodeInvalidPath:
			c.JSON(http.StatusNotFound, v1err)

		case v1.ErrorCodeSystemDeleting, v1.ErrorCodeSystemPending:
			c.JSON(http.StatusConflict, v1err)

		default:
			handleInternalError(c, err)
		}
		return
	}

	c.JSON(http.StatusOK, nodePool)
}
