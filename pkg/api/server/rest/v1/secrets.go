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

// handleSetSecret handler for set-secret
// @ID set-secret
// @Summary set secret
// @Description Sets a new secret
// @Router /systems/{system}/secrets [post]
// @Security ApiKeyAuth
// @Tags secrets
// @Param system path string true "System ID"
// @Param secretRequest body rest.SetSecretRequest true "Create secret"
// @Accept  json
// @Produce  json
// @Success 200 {object} v1.Secret
// @Failure 400 ""
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

// handleListSecrets handler for list-secrets
// @ID list-secrets
// @Summary Lists secrets
// @Description Lists all secrets
// @Router /systems/{system}/secrets [get]
// @Security ApiKeyAuth
// @Tags secrets
// @Param system path string true "System ID"
// @Accept  json
// @Produce  json
// @Success 200 {array} v1.Secret
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

// handleGetSecret handler for get-secret
// @ID get-secret
// @Summary Get secret
// @Description Gets the secret object
// @Router /systems/{system}/secrets/{secret} [get]
// @Security ApiKeyAuth
// @Tags secrets
// @Param system path string true "System ID"
// @Param secret path string true "Secret Path"
// @Accept  json
// @Produce  json
// @Success 200 {object} v1.Secret
// @Failure 404 ""
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

// handleUnsetSecret handler for delete-secret
// @ID delete-secret
// @Summary Delete secret
// @Description Unsets the specified secret
// @Router /systems/{system}/secrets/{secret} [delete]
// @Security ApiKeyAuth
// @Tags secrets
// @Accept  json
// @Produce  json
// @Param system path string true "System ID"
// @Param secret path string true "Secret Path"
// @Success 200 ""
// @Failure 404 ""
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
