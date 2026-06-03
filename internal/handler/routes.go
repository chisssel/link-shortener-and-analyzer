package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/artem/url-shortener/internal/config"
	"github.com/artem/url-shortener/internal/model"
	"github.com/artem/url-shortener/internal/service"
)

func RegisterRoutes(
	r *gin.Engine,
	svc *service.ShortenerService,
	linkRepo model.LinkRepository,
) {
	cfg := config.Load()

	shorten := NewShortenHandler(svc, cfg, linkRepo)
	redirect := NewRedirectHandler(svc, linkRepo)
	analytics := NewAnalyticsHandler(svc)

	r.POST("/api/shorten", shorten.Create)
	r.GET("/api/links", shorten.List)
	r.DELETE("/api/link", shorten.Delete)
	r.GET("/api/link/:id/stats", analytics.GetStats)
	r.GET("/:code", redirect.Redirect)
}
