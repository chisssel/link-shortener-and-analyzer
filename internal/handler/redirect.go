package handler

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/artem/url-shortener/internal/geo"
	"github.com/artem/url-shortener/internal/model"
	"github.com/artem/url-shortener/internal/service"
)

type RedirectHandler struct {
	svc      *service.ShortenerService
	linkRepo model.LinkRepository
	geo      geo.GeoIPService
}

func NewRedirectHandler(svc *service.ShortenerService, linkRepo model.LinkRepository, geo geo.GeoIPService) *RedirectHandler {
	return &RedirectHandler{svc: svc, linkRepo: linkRepo, geo: geo}
}

func (h *RedirectHandler) Redirect(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "code is required"})
		return
	}

	originalURL, err := h.svc.Resolve(code)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "link not found"})
		return
	}

	go h.recordClick(c, code)

	c.Header("Cache-Control", "no-store, no-cache, must-revalidate")
	c.Redirect(http.StatusFound, originalURL)
}

func (h *RedirectHandler) recordClick(c *gin.Context, shortCode string) {
	link, err := h.linkRepo.FindByShortCode(shortCode)
	if err != nil || link == nil {
		return
	}

	ip := c.ClientIP()
	country, city, geoErr := h.geo.Lookup(ip)
	if geoErr != nil {
		log.Printf("geo lookup failed for %s: %v", ip, geoErr)
	}

	click := &model.Click{
		LinkID:    link.ID,
		IPAddress: ip,
		UserAgent: c.GetHeader("User-Agent"),
		Referer:   c.GetHeader("Referer"),
		Country:   country,
		City:      city,
	}

	h.linkRepo.InsertClick(click)
}
