package model

import (
	"encoding/json"
	"fmt"

	"github.com/gofiber/fiber/v2"
)

type APIError struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description,omitempty"`
}

var ResponseNYI = Response{fiber.StatusNotImplemented, APIErrorNYI}

// Predefined errors
var (
	APIErrorUnknownIssuer    = APIError{ErrorInvalidRequest, "The provided issuer is not supported"}
	APIErrorStateMismatch    = APIError{ErrorInvalidRequest, "State mismatched"}
	APIErrorUnknownOIDCFlow  = APIError{ErrorInvalidGrant, "Unknown oidc_flow"}
	APIErrorUnknownGrantType = APIError{ErrorInvalidGrant, "Unknown grant_type"}
	APIErrorNoRefreshToken   = APIError{ErrorOIDC, "Did not receive a refresh token"}
	APIErrorNYI              = APIError{ErrorNYI, ""}
)

// Predefined OAuth2/OIDC errors
const (
	ErrorInvalidRequest       = "invalid_request"
	ErrorInvalidClient        = "invalid_client"
	ErrorInvalidGrant         = "invalid_grant"
	ErrorUnauthorizedClient   = "unauthorized_client"
	ErrorUnsupportedGrantType = "unsupported_grant_type"
	ErrorInvalidScope         = "invalid_scope"
	ErrorInvalidToken         = "invalid_token"
	ErrorInsufficientScope    = "insufficient_scope"
)

// Additional Mytoken errors
const (
	ErrorInternal = "internal_server_error"
	ErrorOIDC     = "oidc_error"
	ErrorNYI      = "not_yet_implemented"
)

func InternalServerError(errorDescription string) APIError {
	return APIError{
		Error:            ErrorInternal,
		ErrorDescription: errorDescription,
	}
}

func OIDCError(oidcError, oidcErrorDescription string) APIError {
	error := oidcError
	if oidcErrorDescription != "" {
		error = fmt.Sprintf("%s: %s", oidcError, oidcErrorDescription)
	}
	return APIError{
		Error:            ErrorOIDC,
		ErrorDescription: error,
	}
}

func OIDCErrorFromBody(body []byte) (error APIError, ok bool) {
	bodyError := APIError{}
	if err := json.Unmarshal(body, &bodyError); err != nil {
		return
	}
	error = OIDCError(bodyError.Error, bodyError.ErrorDescription)
	ok = true
	return
}

func BadRequestError(errorDescription string) APIError {
	return APIError{
		Error:            ErrorInvalidRequest,
		ErrorDescription: errorDescription,
	}
}
