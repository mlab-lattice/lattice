package v1

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	v1rest "github.com/mlab-lattice/lattice/pkg/api/v1/rest"
)

const nodePoolIdentifier = "node_pool_path"

var nodePoolIdentifierPathComponent = fmt.Sprintf(":%v", nodePoolIdentifier)
var nodePoolPath = fmt.Sprintf(v1rest.NodePoolPathFormat, systemIdentifierPathComponent, nodePoolIdentifierPathComponent)
var nodePoolsPath = fmt.Sprintf(v1rest.NodePoolsPathFormat, systemIdentifierPathComponent)

func (api *LatticeAPI) setupNoodPoolEndpoints() {
	// list-node-pools
	api.router.GET(nodePoolsPath, api.handleListNodePools)

	// get-node-pool
	api.router.GET(nodePoolPath, api.handleGetNodePool)

}

// ListNodePools godoc
// @ID list-node-pools
// @Summary Lists node pools
// @Description list node pools
// @Router /v1/systems/{systemId}/node-pools [get]
// @Param systemId path string true "System ID"
// @Accept  json
// @Produce  json
// @Success 200 {array} v1.NodePool
func (api *LatticeAPI) handleListNodePools(c *gin.Context) {
	systemID := v1.SystemID(c.Param(systemIdentifier))

	nodePools, err := api.backend.ListNodePools(systemID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, nodePools)
}

// GetNodePool godoc
// @ID get-node-pool
// @Summary Get node pool
// @Description get node pool
// @Router /v1/systems/{systemId}/node-pools/{id} [get]
// @Param systemId path string true "System ID"
// @Param id path string true "NodePool ID"
// @Accept  json
// @Produce  json
// @Success 200 {object} v1.NodePool
func (api *LatticeAPI) handleGetNodePool(c *gin.Context) {
	systemID := v1.SystemID(c.Param(systemIdentifier))
	escapedNodePoolPath := c.Param(nodePoolIdentifier)

	nodePoolPathString, err := url.PathUnescape(escapedNodePoolPath)
	if err != nil {
		// FIXME: send invalid nodePool error
		c.Status(http.StatusBadRequest)
		return
	}

	path, err := v1.ParseNodePoolPath(nodePoolPathString)
	if err != nil {
		// FIXME: send invalid nodePool error
		c.Status(http.StatusBadRequest)
		return
	}

	nodePool, err := api.backend.GetNodePool(systemID, path)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, nodePool)
}
