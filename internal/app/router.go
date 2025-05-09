package app

import (
	"metafarm/internal/storage"

	"github.com/gin-gonic/gin"
)

type serverRouter struct {
	*gin.Engine
}

func newServerRouter(storage storage.Storage, openaiKey string) *serverRouter {
	sr := serverRouter{
		Engine: gin.Default(),
	}
	sr.setupRoutes(storage, openaiKey)
	return &sr
}

func (s *serverRouter) setupRoutes(storage storage.Storage, openaiKey string) {
	api := s.Group("/api")
	{
		api.GET("/ping", pingHandler())

		api.POST("/analysis", analysisHandler(storage, openaiKey))
		api.GET("/analysis/:id", analysisResultHandler(storage))
	}
}
