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

const secretIdentifier = "secret_path"

var (
	secretIdentifierPathComponent = fmt.Sprintf(":%v", secretIdentifier)
	secretPath                    = fmt.Sprintf(v1rest.SystemSecretPathFormat, systemIdentifierPathComponent, secretIdentifierPathComponent)
	secretsPath                   = fmt.Sprintf(v1rest.SystemSecretsPathFormat, systemIdentifierPathComponent)
)

func (api *LatticeAPI) setupSecretsEndpoints() {

	// list-secrets
	api.router.GET(secretsPath, api.handleListSecrets)

	// get-secret
	api.router.GET(secretPath, api.handleGetSecret)

	// set-secret
	api.router.PATCH(secretPath, api.handleSetSecret)

	// unset-secret
	api.router.DELETE(secretPath, api.handleUnsetSecret)

}

// swagger:operation POST /systems/{system}/secrets secrets SetSecret
//
// Set secret
//
// Set secret
// ---
//     consumes:
//     - application/json
//     produces:
//     - application/json
//
//     parameters:
//       - description: System ID
//         in: path
//         name: system
//         required: true
//         type: string
//       - in: body
//         schema:
//           "$ref": "#/definitions/SetSecretRequest"
//     responses:
//         '200':
//           description: Secret object
//           schema:
//             "$ref": "#/definitions/Secret"

// handleSetSecret handler for SetSecret
func (api *LatticeAPI) handleSetSecret(c *gin.Context) {
	var req v1rest.SetSecretRequest
	if err := c.BindJSON(&req); err != nil {
		handleBadRequestBody(c)
		return
	}

	systemID := v1.SystemID(c.Param(systemIdentifier))
	escapedSecretPath := c.Param(secretIdentifier)

	secretPathString, err := url.PathUnescape(escapedSecretPath)
	if err != nil {
		c.JSON(http.StatusBadRequest, v1.NewInvalidPathError())
		return
	}

	path, err := tree.NewPathSubcomponent(secretPathString)
	if err != nil {
		c.JSON(http.StatusBadRequest, v1.NewInvalidPathError())
		return
	}

	err = api.backend.Systems().Secrets(systemID).Set(path, req.Value)
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

		case v1.ErrorCodeConflict:
			c.JSON(http.StatusConflict, v1err)

		default:
			handleInternalError(c, err)
		}
		return
	}

	c.Status(http.StatusOK)
}

// swagger:operation GET /systems/{system}/secrets secrets ListSecrets
//
// Lists secrets
//
// Lists secrets
// ---
//     consumes:
//     - application/json
//     produces:
//     - application/json
//
//     parameters:
//       - description: System ID
//         in: path
//         name: system
//         required: true
//         type: string
//
//     responses:
//         '200':
//           description: secret list
//           schema:
//             type: array
//             items:
//               "$ref": "#/definitions/Secret"
//

// handleListSecrets handler for ListSecrets
func (api *LatticeAPI) handleListSecrets(c *gin.Context) {
	systemID := v1.SystemID(c.Param(systemIdentifier))

	secrets, err := api.backend.Systems().Secrets(systemID).List()
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

	c.JSON(http.StatusOK, secrets)
}

// swagger:operation GET /systems/{system}/secrets/{secretPath} secrets GetSecret
//
// Get secret
//
// Get secret
// ---
//     consumes:
//     - application/json
//     produces:
//     - application/json
//
//     parameters:
//       - description: System ID
//         in: path
//         name: system
//         required: true
//         type: string
//       - description: Secret Path
//         in: path
//         name: secretPath
//         required: true
//         type: string
//
//     responses:
//         '200':
//           description: Secret Object
//           schema:
//             "$ref": "#/definitions/Secret"
//

// handleGetSecret handler for GetSecret
func (api *LatticeAPI) handleGetSecret(c *gin.Context) {
	systemID := v1.SystemID(c.Param(systemIdentifier))
	escapedSecretPath := c.Param(secretIdentifier)

	secretPathString, err := url.PathUnescape(escapedSecretPath)
	if err != nil {
		c.JSON(http.StatusBadRequest, v1.NewInvalidPathError())
		return
	}

	path, err := tree.NewPathSubcomponent(secretPathString)
	if err != nil {
		c.JSON(http.StatusBadRequest, v1.NewInvalidPathError())
		return
	}

	secret, err := api.backend.Systems().Secrets(systemID).Get(path)
	if err != nil {
		v1err, ok := err.(*v1.Error)
		if !ok {
			handleInternalError(c, err)
			return
		}

		switch v1err.Code {
		case v1.ErrorCodeInvalidSystemID, v1.ErrorCodeInvalidSecret:
			c.JSON(http.StatusNotFound, v1err)

		case v1.ErrorCodeSystemDeleting, v1.ErrorCodeSystemPending:
			c.JSON(http.StatusConflict, v1err)

		default:
			handleInternalError(c, err)
		}
		return
	}

	c.JSON(http.StatusOK, secret)
}

// swagger:operation DELETE /systems/{systemId}/secrets/{secretPath} secrets DeleteSecret
//
// Delete secret
//
// Delete secret
// ---
//     consumes:
//     - application/json
//     produces:
//     - application/json
//
//     parameters:
//       - description: System ID
//         in: path
//         name: systemId
//         required: true
//         type: string
//       - description: Secret Path
//         in: path
//         name: secretPath
//         required: true
//         type: string
//
//     responses:
//         '200':
//

// handleUnsetSecret handler for DeleteSecret
func (api *LatticeAPI) handleUnsetSecret(c *gin.Context) {
	systemID := v1.SystemID(c.Param(systemIdentifier))
	escapedSecretPath := c.Param(secretIdentifier)
	secretPathString, err := url.PathUnescape(escapedSecretPath)
	if err != nil {
		c.JSON(http.StatusBadRequest, v1.NewInvalidPathError())
		return
	}

	path, err := tree.NewPathSubcomponent(secretPathString)
	if err != nil {
		c.JSON(http.StatusBadRequest, v1.NewInvalidPathError())
		return
	}

	err = api.backend.Systems().Secrets(systemID).Unset(path)
	if err != nil {
		v1err, ok := err.(*v1.Error)
		if !ok {
			handleInternalError(c, err)
			return
		}

		switch v1err.Code {
		case v1.ErrorCodeInvalidSystemID, v1.ErrorCodeInvalidSecret:
			c.JSON(http.StatusNotFound, v1err)

		case v1.ErrorCodeSystemDeleting, v1.ErrorCodeSystemPending:
			c.JSON(http.StatusConflict, v1err)

		case v1.ErrorCodeConflict:
			c.JSON(http.StatusConflict, v1err)

		default:
			handleInternalError(c, err)
		}
		return
	}

	c.Status(http.StatusOK)

}
