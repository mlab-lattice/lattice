package v1

import (
	"net/http"

	"github.com/mlab-lattice/lattice/pkg/api/v1"

	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
)

type ErrorResponse struct {
	Code  v1.ErrorCode `json:"code"`
	Error error        `json:"error"`
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
		Code:  code,
		Error: err,
	}

	c.JSON(statusCode, errResponse)
}

func handleBadRequestBody(c *gin.Context) {
	c.String(http.StatusBadRequest, "invalid request body")
}
