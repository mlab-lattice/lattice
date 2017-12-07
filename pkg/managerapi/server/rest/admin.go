package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (r *restServer) mountAdminHandlers() {
	r.mountAdminMasterHandlers()
}

func (r *restServer) mountAdminMasterHandlers() {
	r.mountAdminMasterNodeHandlers()
}

func (r *restServer) mountAdminMasterNodeHandlers() {
	// FIXME: move nodeID into query
	components := r.router.Group("/admin/master/components")
	{
		// get-master-components
		components.GET("", func(c *gin.Context) {
			components, err := r.backend.GetMasterComponents()
			if err != nil {
				handleInternalError(c, err)
				return
			}

			c.JSON(http.StatusOK, components)
		})

		component := components.Group("/:component_id")
		{
			// get-master-component-logs
			component.GET("/logs", func(c *gin.Context) {
				nodeID := c.Query("nodeId")
				component := c.Param("component_id")
				followQuery := c.DefaultQuery("follow", "false")

				var follow bool
				switch followQuery {
				case "false":
					follow = false
				case "true":
					follow = true
				default:
					c.String(http.StatusBadRequest, "invalid value for follow query")
					return
				}

				log, exists, err := r.backend.GetMasterComponentLog(nodeID, component, follow)
				if err != nil {
					handleInternalError(c, err)
					return
				}

				if exists == false {
					c.String(http.StatusNotFound, "")
					return
				}

				logEndpoint(c, log, follow)
			})

			// restart-master-component
			component.POST("/restart", func(c *gin.Context) {
				nodeID := c.Query("nodeId")
				component := c.Param("component_id")

				exists, err := r.backend.RestartMasterComponent(nodeID, component)

				if err != nil {
					handleInternalError(c, err)
					return
				}

				if !exists {
					c.String(http.StatusNotFound, "")
					return
				}

				c.String(http.StatusOK, "")
			})
		}
	}
}
