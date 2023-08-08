package http

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// Response Ответы сервиса
type Response struct {
}

// OK 200
func (r Response) OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, data)
}

// NoContent 204
func (r Response) NoContent(c *gin.Context) {
	c.JSON(http.StatusNoContent, nil)
}

func (r Response) Error(c *gin.Context, err error) {
	if err, ok := err.(*Error); ok {
		r.err(c, err.HttpCode(), err.AppCode(), err.Message(), err.Error())
		return
	}
	r.InternalServerError(c, "APP_ERROR", err.Error(), err.Error())
}

// InternalServerError 500
func (r Response) InternalServerError(c *gin.Context, applicationErrorCode, message, debug string) {
	r.err(c, http.StatusInternalServerError, applicationErrorCode, message, debug)
}

// BadRequestError 400
func (r Response) BadRequestError(c *gin.Context, debug string) {
	r.err(c, http.StatusBadRequest, "BAD_REQUEST", "bad request", debug)
}

// NotFoundError 404
func (r Response) NotFoundError(c *gin.Context, debug string) {
	r.err(c, http.StatusNotFound, "NOT_FOUND", "not found", debug)
}

// ForbiddenError 403
func (r Response) ForbiddenError(c *gin.Context, debug string) {
	r.err(c, http.StatusForbidden, "FORBIDDEN", "bad request", debug)
}

func (r Response) err(c *gin.Context, httpErrorCode int, applicationErrorCode, message, debug string) {
	c.JSON(httpErrorCode, gin.H{
		"applicationErrorCode": applicationErrorCode,
		"message":              message,
		"debug":                debug,
	})
}
