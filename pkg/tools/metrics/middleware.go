package metrics

import (
	"github.com/gin-gonic/gin"
	"go.opencensus.io/plugin/ochttp"
)

type wrappedGinRouter struct {
	gin.IRoutes
}

func WrapGinRouter(r gin.IRoutes) gin.IRoutes {
	return &wrappedGinRouter{r}
}

var _ gin.IRoutes = (*wrappedGinRouter)(nil)

func withMonitoring(relativePath string, handlers ...gin.HandlerFunc) []gin.HandlerFunc {
	h := []gin.HandlerFunc{func(c *gin.Context) {
		ctx := c.Request.Context()
		ochttp.SetRoute(ctx, relativePath)
		c.Request = c.Request.WithContext(ctx)
	}}

	h = append(h, handlers...)
	return h
}

func (w *wrappedGinRouter) Use(handlers ...gin.HandlerFunc) gin.IRoutes {
	w.IRoutes.Use(handlers...)
	return w
}

func (w *wrappedGinRouter) Handle(httpMethod, relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes {
	w.IRoutes.Handle(httpMethod, relativePath, withMonitoring(relativePath, handlers...)...)
	return w
}

func (w *wrappedGinRouter) Any(relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes {
	w.IRoutes.Any(relativePath, withMonitoring(relativePath, handlers...)...)
	return w
}

func (w *wrappedGinRouter) GET(relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes {
	w.IRoutes.GET(relativePath, withMonitoring(relativePath, handlers...)...)
	return w
}

func (w *wrappedGinRouter) POST(relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes {
	w.IRoutes.POST(relativePath, withMonitoring(relativePath, handlers...)...)
	return w
}

func (w *wrappedGinRouter) DELETE(relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes {
	w.IRoutes.DELETE(relativePath, withMonitoring(relativePath, handlers...)...)
	return w
}

func (w *wrappedGinRouter) PUT(relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes {
	w.IRoutes.PUT(relativePath, withMonitoring(relativePath, handlers...)...)
	return w
}

func (w *wrappedGinRouter) OPTIONS(relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes {
	w.IRoutes.OPTIONS(relativePath, withMonitoring(relativePath, handlers...)...)
	return w
}
