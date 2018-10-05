package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
)

func handleBadRequestBody(c *gin.Context) {
	c.String(http.StatusBadRequest, "invalid request body")
}

func handleError(c *gin.Context, err error) {
	v1err, ok := err.(*v1.Error)
	if !ok {
		handleInternalError(c, err)
		return
	}

	switch v1err.Code {
	case v1.ErrorCodeInvalidBuildID,
		v1.ErrorCodeInvalidDeployID,
		v1.ErrorCodeInvalidJobID,
		v1.ErrorCodeInvalidJobRunID,
		v1.ErrorCodeInvalidSecret,
		v1.ErrorCodeInvalidServiceID,
		v1.ErrorCodeInvalidServiceInstanceID,
		v1.ErrorCodeInvalidTeardownID,
		v1.ErrorCodeInvalidSystemID,
		v1.ErrorCodeInvalidPath,
		v1.ErrorCodeInvalidSidecar,
		v1.ErrorCodeInvalidVersion:
		c.JSON(http.StatusNotFound, v1err)

	case v1.ErrorCodeSystemAlreadyExists,
		v1.ErrorCodeSystemDeleting,
		v1.ErrorCodeSystemFailed,
		v1.ErrorCodeSystemPending,
		v1.ErrorCodeConflict:
		c.JSON(http.StatusConflict, v1err)

	case v1.ErrorCodeInvalidSystemOptions,
		v1.ErrorCodeInvalidComponentType:
		c.JSON(http.StatusBadRequest, v1err)

	case v1.ErrorCodeUnknown:
		handleInternalError(c, err)

	default:
		handleInternalError(c, err)
	}
	return
}

func handleInternalError(c *gin.Context, err error) {
	c.Status(http.StatusInternalServerError)
	panic(err)
}
