package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/artem/url-shortener/internal/model"
	"github.com/artem/url-shortener/internal/service"
)

type RedirectHandler struct {
	svc      *service.ShortenerService
	linkRepo model.LinkRepository
}

func NewRedirectHandler(svc *service.ShortenerService, linkRepo model.LinkRepository) *RedirectHandler {
	return &RedirectHandler{svc: svc, linkRepo: linkRepo}
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

	c.Redirect(http.StatusMovedPermanently, originalURL)
}

func (h *RedirectHandler) recordClick(c *gin.Context, shortCode string) {
	link, err := h.linkRepo.FindByShortCode(shortCode)
	if err != nil || link == nil {
		return
	}

	click := &model.Click{
		LinkID:    link.ID,
		IPAddress: c.ClientIP(),
		UserAgent: c.GetHeader("User-Agent"),
		Referer:   c.GetHeader("Referer"),
	}

	h.linkRepo.InsertClick(click)
}
