package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/artem/url-shortener/internal/service"
)

type AnalyticsHandler struct {
	svc *service.ShortenerService
}

func NewAnalyticsHandler(svc *service.ShortenerService) *AnalyticsHandler {
	return &AnalyticsHandler{svc: svc}
}

func (h *AnalyticsHandler) GetStats(c *gin.Context) {
	idStr := c.Param("id")
	linkID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid link id"})
		return
	}

	stats, err := h.svc.GetStats(linkID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}
