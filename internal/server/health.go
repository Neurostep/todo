package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (r *api) healthz(c *gin.Context) {
	c.String(http.StatusOK, "OK")
}

func (r *api) readyz(c *gin.Context) {
	c.String(http.StatusOK, "OK")
}
