package errors

import (
	"fmt"
	"io"
	"net/http"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/util/rest"
)

func HandleErrorStatusCode(statusCode int, body io.Reader) error {
	switch statusCode {
	case http.StatusBadRequest, http.StatusConflict, http.StatusNotFound:
		v1Err := &v1.Error{}
		if err := rest.UnmarshalBodyJSON(body, v1Err); err != nil {
			return err
		}

		return v1Err

	default:
		return handleUnexpectedErrorStatusCode(statusCode)
	}
}

func handleUnexpectedErrorStatusCode(statusCode int) error {
	return fmt.Errorf("unexpected status code %v", statusCode)
}
