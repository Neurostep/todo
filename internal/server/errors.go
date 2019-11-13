package server

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/go-kit/kit/log"
	"gopkg.in/go-playground/validator.v8"
)

const errorMsg = "validation for '%s' failed on the '%s' tag"

type (
	Error struct {
		Label   string `json:"label"`
		Message string `json:"message"`
	}

	Errors struct {
		Errors []*Error `json:"errors"`
	}
)

func (e *Error) Error() string {
	return fmt.Sprintf("%s:%s", e.Label, e.Message)
}

func respondErrors(c *gin.Context, logger log.Logger, code int, errors ...*Error) {
	errs := &Errors{
		Errors: errors,
	}
	logger.Log("event", "respond error", "route", c.Request.URL,
		"method", c.Request.Method,
		"code", code, "errors", errs)
	c.AbortWithStatusJSON(code, errs)
}

func newError(label, message string) *Error {
	return &Error{
		Label:   label,
		Message: message,
	}
}

func extractBindErrors(err error) []*Error {
	switch e := err.(type) {
	case validator.ValidationErrors:
		res := []*Error{}
		for _, v := range e {
			res = append(res, &Error{
				Label:   v.Field,
				Message: fmt.Sprintf(errorMsg, v.Field, v.Tag),
			})
		}
		if len(res) == 0 {
			return []*Error{newError("validation", err.Error())}
		}
		return res
	}
	return []*Error{newError("bind_error", err.Error())}
}
