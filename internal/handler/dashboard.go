package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/artem/url-shortener/internal/model"
	"github.com/artem/url-shortener/internal/service"
)

type DashboardHandler struct {
	svc      *service.ShortenerService
	linkRepo model.LinkRepository
}

func NewDashboardHandler(svc *service.ShortenerService, linkRepo model.LinkRepository) *DashboardHandler {
	return &DashboardHandler{svc: svc, linkRepo: linkRepo}
}

func (h *DashboardHandler) IndexPage(c *gin.Context) {
	c.HTML(http.StatusOK, "layout.html", gin.H{
		"Title":   "URL Shortener",
		"Content": "index",
	})
}

func (h *DashboardHandler) LinksPage(c *gin.Context) {
	ownerID := c.Query("owner_id")
	if ownerID == "" {
		ownerID = "demo"
	}

	links, err := h.linkRepo.FindByOwner(ownerID)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "layout.html", gin.H{
			"Title": "Error",
			"Error": err.Error(),
		})
		return
	}

	c.HTML(http.StatusOK, "layout.html", gin.H{
		"Title":   "My Links",
		"Content": "links",
		"Links":   links,
		"OwnerID": ownerID,
	})
}

func (h *DashboardHandler) StatsPage(c *gin.Context) {
	idStr := c.Param("id")
	linkID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.HTML(http.StatusBadRequest, "layout.html", gin.H{
			"Title": "Error",
			"Error": "invalid link id",
		})
		return
	}

	link, err := h.linkRepo.FindByID(linkID)
	if err != nil || link == nil {
		c.HTML(http.StatusNotFound, "layout.html", gin.H{
			"Title": "Error",
			"Error": "link not found",
		})
		return
	}

	stats, err := h.svc.GetStats(linkID)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "layout.html", gin.H{
			"Title": "Error",
			"Error": err.Error(),
		})
		return
	}

	c.HTML(http.StatusOK, "layout.html", gin.H{
		"Title":   "Link Stats",
		"Content": "stats",
		"Link":    link,
		"Stats":   stats,
	})
}
