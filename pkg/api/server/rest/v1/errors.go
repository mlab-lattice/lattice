package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func handleBadRequestBody(c *gin.Context) {
	c.String(http.StatusBadRequest, "invalid request body")
}
