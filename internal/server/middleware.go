package server

import (
	"mime"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-kit/kit/log"
)

func requireContentType(logger log.Logger, contentType string) gin.HandlerFunc {
	return func(c *gin.Context) {
		contentHeader := c.GetHeader("Content-Type")
		method := c.Request.Method
		if method == "POST" || method == "PUT" || method == "PATCH" {
			ct, _, _ := mime.ParseMediaType(contentHeader)
			if ct != contentType {
				respondErrors(c, logger, http.StatusBadRequest, newError("validation", "application/json is required"))
				return
			}
		}
		c.Next()
	}
}

func CORS(c *gin.Context) {
	origin := c.Request.Header.Get("Origin")
	if len(origin) == 0 {
		c.Header("Access-Control-Allow-Origin", "*")
	} else {
		c.Header("Access-Control-Allow-Origin", origin)
	}
	c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
	c.Header("Access-Control-Allow-Headers", "origin, content-type, accept")
	c.Header("Access-Control-Allow-Credentials", "true")
	c.Header("Allow", "HEAD,GET,POST,PUT,PATCH,DELETE,OPTIONS")
	c.Header("Access-Control-Max-Age", "86400") // 24 hours

	if c.Request.Method != "OPTIONS" {
		c.Next()
	} else {
		c.AbortWithStatus(http.StatusNoContent)
	}
}
