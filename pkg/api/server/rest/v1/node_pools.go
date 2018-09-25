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

// handleListNodePools handler for list-node-pools
// @ID list-node-pools
// @Summary Lists node pools
// @Description list node pools
// @Router /systems/{system}/node-pools [get]
// @Security ApiKeyAuth
// @Tags node-pools
// @Param system path string true "System ID"
// @Accept  json
// @Produce  json
// @Success 200 {array} v1.NodePool
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

// handleGetNodePool handler for get-node-pool
// @ID get-node-pool
// @Summary Get node pool
// @Description Gets the node pool object
// @Router /systems/{system}/node-pools/{id} [get]
// @Security ApiKeyAuth
// @Tags node-pools
// @Param system path string true "System ID"
// @Param id path string true "NodePool ID"
// @Accept  json
// @Produce  json
// @Success 200 {object} v1.NodePool
// @Failure 404 ""
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
