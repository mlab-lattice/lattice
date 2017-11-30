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
	components := r.router.Group("/admin/master/:master_node_id/components")
	{
		// get-master-components
		components.GET("", func(c *gin.Context) {
			nodeId := c.Param("master_node_id")

			components, err := r.backend.GetMasterNodeComponents(nodeId)
			if err != nil {
				c.String(http.StatusInternalServerError, err.Error())
				return
			}

			c.JSON(http.StatusOK, components)
		})

		component := components.Group("/:component_id")
		{
			// get-master-component-logs
			component.GET("/logs", func(c *gin.Context) {
				nodeId := c.Param("master_node_id")
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
				}

				log, err := r.backend.GetMasterNodeComponentLog(nodeId, component, follow)
				if err != nil {
					c.String(http.StatusInternalServerError, err.Error())
					return
				}

				c.JSON(http.StatusOK, log)
			})

			// restart-master-component
			component.POST("/restart", func(c *gin.Context) {
				nodeId := c.Param("master_node_id")
				component := c.Param("component_id")

				err := r.backend.RestartMasterNodeComponent(nodeId, component)

				if err != nil {
					c.String(http.StatusInternalServerError, err.Error())
					return
				}

				c.String(http.StatusOK, "")
			})
		}
	}
}
