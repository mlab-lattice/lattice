package rest

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/mlab-lattice/system/pkg/api/v1"
	"github.com/mlab-lattice/system/pkg/util/rest"
)

func HandleErrorStatusCode(statusCode int, body io.Reader) error {
	switch statusCode {
	case http.StatusBadRequest, http.StatusConflict, http.StatusNotFound:
		errorDecoder := &v1ErrorDecoder{}
		if err := rest.UnmarshalBodyJSON(body, errorDecoder); err != nil {
			return err
		}
		return handleV1Error(errorDecoder)

	default:
		return handleUnexpectedErrorStatusCode(statusCode)
	}
}

func handleUnexpectedErrorStatusCode(statusCode int) error {
	return fmt.Errorf("unexpected status code %v", statusCode)
}

func handleV1Error(errorDecoder *v1ErrorDecoder) v1.Error {
	var v1Error v1.Error = v1.NewUnknownError()

	var err error
	switch errorDecoder.Code {
	case v1.ErrorCodeInvalidSystemOptions:
		target := &v1.InvalidSystemOptionsError{}
		err = json.Unmarshal(errorDecoder.Error, &target)
		v1Error = target

	case v1.ErrorCodeSystemAlreadyExists:
		target := &v1.SystemAlreadyExistsError{}
		err = json.Unmarshal(errorDecoder.Error, &target)
		v1Error = target

	case v1.ErrorCodeInvalidSystemID:
		target := &v1.InvalidSystemIDError{}
		err = json.Unmarshal(errorDecoder.Error, &target)
		v1Error = target

	case v1.ErrorCodeInvalidSystemVersion:
		target := &v1.InvalidSystemVersionError{}
		err = json.Unmarshal(errorDecoder.Error, &target)
		v1Error = target

	case v1.ErrorCodeInvalidBuildID:
		target := &v1.InvalidBuildIDError{}
		err = json.Unmarshal(errorDecoder.Error, &target)
		v1Error = target

	case v1.ErrorCodeInvalidDeployID:
		target := &v1.InvalidDeployIDError{}
		err = json.Unmarshal(errorDecoder.Error, &target)
		v1Error = target

	case v1.ErrorCodeInvalidTeardownID:
		target := &v1.InvalidTeardownIDError{}
		err = json.Unmarshal(errorDecoder.Error, &target)
		v1Error = target

	case v1.ErrorCodeInvalidServiceID:
		target := &v1.InvalidServiceIDError{}
		err = json.Unmarshal(errorDecoder.Error, &target)
		v1Error = target

	case v1.ErrorCodeInvalidSystemSecret:
		target := &v1.InvalidSystemSecretError{}
		err = json.Unmarshal(errorDecoder.Error, &target)
		v1Error = target

	case v1.ErrorCodeConflict:
		target := &v1.ConflictError{}
		err = json.Unmarshal(errorDecoder.Error, &target)
		v1Error = target
	}

	if err != nil {
		return v1.NewUnknownError()
	}
	return v1Error
}

type v1ErrorDecoder struct {
	Code  v1.ErrorCode    `json:"code"`
	Error json.RawMessage `json:"error"`
}
