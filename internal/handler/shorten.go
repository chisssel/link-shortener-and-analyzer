package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/artem/url-shortener/internal/config"
	"github.com/artem/url-shortener/internal/model"
	"github.com/artem/url-shortener/internal/service"
)

type ShortenHandler struct {
	svc  *service.ShortenerService
	cfg  *config.Config
	repo model.LinkRepository
}

func NewShortenHandler(svc *service.ShortenerService, cfg *config.Config, repo model.LinkRepository) *ShortenHandler {
	return &ShortenHandler{svc: svc, cfg: cfg, repo: repo}
}

func (h *ShortenHandler) Create(c *gin.Context) {
	var req model.CreateLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.svc.Create(req.OriginalURL, req.OwnerID, h.cfg.BaseURL)
	if err != nil {
		if errors.Is(err, service.ErrInvalidURL) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid URL"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

func (h *ShortenHandler) List(c *gin.Context) {
	ownerID := c.Query("owner_id")
	if ownerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "owner_id is required"})
		return
	}

	links, err := h.repo.FindByOwner(ownerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, links)
}

func (h *ShortenHandler) Delete(c *gin.Context) {
	var req struct {
		LinkID int64 `json:"link_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.repo.Delete(req.LinkID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}
