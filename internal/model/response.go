package model

import (
	"github.com/gofiber/fiber/v2"
	"github.com/oidc-mytoken/api/v0"
	"github.com/oidc-mytoken/server/internal/utils/errorfmt"
	"github.com/oidc-mytoken/server/shared/model"
	"github.com/valyala/fasthttp"
)

// Response models a http server response
type Response struct {
	// The Http Status code of the response
	Status int
	// The response body, will be marshalled as json
	Response interface{}
	// Cookies that should be set
	Cookies []*fiber.Cookie
}

// Send sends this response using the passed fiber.Ctx
func (r Response) Send(ctx *fiber.Ctx) error {
	for _, c := range r.Cookies {
		ctx.Cookie(c)
	}
	if fasthttp.StatusCodeIsRedirect(r.Status) {
		return ctx.Redirect(r.Response.(string), r.Status)
	}
	return ctx.Status(r.Status).JSON(r.Response)
}

// ErrorToInternalServerErrorResponse creates an internal server error response from a golang error
func ErrorToInternalServerErrorResponse(err error) *Response {
	return &Response{
		Status:   fiber.StatusInternalServerError,
		Response: model.InternalServerError(errorfmt.Error(err)),
	}
}

// ErrorToBadRequestErrorResponse creates a bad request error response from a golang error
func ErrorToBadRequestErrorResponse(err error) *Response {
	return &Response{
		Status:   fiber.StatusBadRequest,
		Response: model.BadRequestError(errorfmt.Error(err)),
	}
}

// ResponseNYI is the server response when something is not yet implemented
var ResponseNYI = Response{Status: fiber.StatusNotImplemented, Response: api.ErrorNYI}
