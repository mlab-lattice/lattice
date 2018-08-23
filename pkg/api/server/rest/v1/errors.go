package v1

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
)

type ErrorResponse struct {
	Code    v1.ErrorCode `json:"code"`
	Message string       `json:"message" example:"status bad request"`
}

func handleError(c *gin.Context, err error) {
	v1Error, ok := err.(v1.Error)
	if !ok {
		glog.Errorf("encountered error: %v", err.Error())
		c.String(http.StatusInternalServerError, "")
		return
	}

	statusCode := http.StatusInternalServerError
	code := v1Error.Code()
	switch code {
	case v1.ErrorCodeInvalidSystemOptions, v1.ErrorCodeInvalidSystemVersion:
		statusCode = http.StatusBadRequest
	case v1.ErrorCodeSystemAlreadyExists, v1.ErrorCodeConflict:
		statusCode = http.StatusConflict
	case v1.ErrorCodeInvalidSystemID, v1.ErrorCodeInvalidBuildID,
		v1.ErrorCodeInvalidDeployID, v1.ErrorCodeInvalidTeardownID,
		v1.ErrorCodeInvalidServicePath, v1.ErrorCodeInvalidSystemSecret:
		statusCode = http.StatusNotFound
	}

	errResponse := &ErrorResponse{
		Code:    code,
		Message: fmt.Sprintf("%v", err),
	}

	c.JSON(statusCode, errResponse)
}

func handleBadRequestBody(c *gin.Context) {
	c.String(http.StatusBadRequest, "invalid request body")
}
