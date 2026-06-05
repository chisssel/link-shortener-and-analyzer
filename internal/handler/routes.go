package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/artem/url-shortener/internal/config"
	"github.com/artem/url-shortener/internal/geo"
	"github.com/artem/url-shortener/internal/model"
	"github.com/artem/url-shortener/internal/service"
)

func RegisterRoutes(
	r *gin.Engine,
	svc *service.ShortenerService,
	linkRepo model.LinkRepository,
	geoClient geo.GeoIPService,
) {
	cfg := config.Load()

	shorten := NewShortenHandler(svc, cfg, linkRepo)
	redirect := NewRedirectHandler(svc, linkRepo, geoClient)
	analytics := NewAnalyticsHandler(svc)
	dash := NewDashboardHandler(svc, linkRepo)

	r.POST("/api/shorten", shorten.Create)
	r.GET("/api/links", shorten.List)
	r.DELETE("/api/link", shorten.Delete)
	r.GET("/api/link/:id/stats", analytics.GetStats)
	r.GET("/:code", redirect.Redirect)

	dashGroup := r.Group("/dashboard")
	dashGroup.GET("/links", dash.LinksPage)
	dashGroup.GET("/link/:id", dash.StatsPage)
	dashGroup.GET("/", dash.IndexPage)
}
