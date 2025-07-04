package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
	
	"github.com/raoxb/smart_redirect/internal/models"
	"github.com/raoxb/smart_redirect/internal/services"
)

type LinkHandler struct {
	linkService *services.LinkService
	db          *gorm.DB
}

func NewLinkHandler(db *gorm.DB, redis *redis.Client) *LinkHandler {
	return &LinkHandler{
		linkService: services.NewLinkService(db, redis),
		db:          db,
	}
}

type CreateLinkRequest struct {
	BusinessUnit string `json:"business_unit" binding:"required"`
	Network      string `json:"network" binding:"required"`
	TotalCap     int    `json:"total_cap"`
	BackupURL    string `json:"backup_url"`
}

type CreateTargetRequest struct {
	URL          string            `json:"url" binding:"required,url"`
	Weight       int               `json:"weight" binding:"required,min=1"`
	Cap          int               `json:"cap"`
	Countries    []string          `json:"countries"`
	ParamMapping map[string]string `json:"param_mapping"`
	StaticParams map[string]string `json:"static_params"`
}

func (h *LinkHandler) CreateLink(c *gin.Context) {
	var req CreateLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	link := &models.Link{
		BusinessUnit: req.BusinessUnit,
		Network:      req.Network,
		TotalCap:     req.TotalCap,
		BackupURL:    req.BackupURL,
	}
	
	if err := h.linkService.CreateLink(link); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create link"})
		return
	}
	
	c.JSON(http.StatusCreated, link)
}

func (h *LinkHandler) GetLink(c *gin.Context) {
	linkID := c.Param("link_id")
	
	link, err := h.linkService.GetLinkByID(linkID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	
	if link == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "link not found"})
		return
	}
	
	c.JSON(http.StatusOK, link)
}

func (h *LinkHandler) UpdateLink(c *gin.Context) {
	linkID := c.Param("link_id")
	
	var link models.Link
	if err := h.db.Where("link_id = ?", linkID).First(&link).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "link not found"})
		return
	}
	
	var req CreateLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	link.BusinessUnit = req.BusinessUnit
	link.Network = req.Network
	link.TotalCap = req.TotalCap
	link.BackupURL = req.BackupURL
	
	if err := h.db.Save(&link).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update link"})
		return
	}
	
	_ = h.linkService.CreateLink(&link)
	
	c.JSON(http.StatusOK, link)
}

func (h *LinkHandler) DeleteLink(c *gin.Context) {
	linkID := c.Param("link_id")
	
	result := h.db.Where("link_id = ?", linkID).Delete(&models.Link{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete link"})
		return
	}
	
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "link not found"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "link deleted successfully"})
}

func (h *LinkHandler) ListLinks(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	
	var total int64
	var links []models.Link
	
	h.db.Model(&models.Link{}).Count(&total)
	
	offset := (page - 1) * pageSize
	if err := h.db.Offset(offset).Limit(pageSize).Find(&links).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch links"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"total": total,
		"page":  page,
		"size":  pageSize,
		"data":  links,
	})
}

func (h *LinkHandler) CreateTarget(c *gin.Context) {
	linkID := c.Param("link_id")
	
	var link models.Link
	if err := h.db.Where("link_id = ?", linkID).First(&link).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "link not found"})
		return
	}
	
	var req CreateTargetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	paramMapping, _ := json.Marshal(req.ParamMapping)
	staticParams, _ := json.Marshal(req.StaticParams)
	countries := ""
	if len(req.Countries) > 0 {
		countriesJSON, _ := json.Marshal(req.Countries)
		countries = string(countriesJSON)
	}
	
	target := &models.Target{
		LinkID:       link.ID,
		URL:          req.URL,
		Weight:       req.Weight,
		Cap:          req.Cap,
		Countries:    countries,
		ParamMapping: string(paramMapping),
		StaticParams: string(staticParams),
	}
	
	if err := h.db.Create(target).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create target"})
		return
	}
	
	_ = h.linkService.CreateLink(&link)
	
	c.JSON(http.StatusCreated, target)
}

func (h *LinkHandler) GetTargets(c *gin.Context) {
	linkID := c.Param("link_id")
	
	var link models.Link
	if err := h.db.Where("link_id = ?", linkID).First(&link).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "link not found"})
		return
	}
	
	var targets []models.Target
	if err := h.db.Where("link_id = ?", link.ID).Find(&targets).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch targets"})
		return
	}
	
	c.JSON(http.StatusOK, targets)
}

func (h *LinkHandler) UpdateTarget(c *gin.Context) {
	targetID := c.Param("target_id")
	
	var target models.Target
	if err := h.db.First(&target, targetID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "target not found"})
		return
	}
	
	var req CreateTargetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	paramMapping, _ := json.Marshal(req.ParamMapping)
	staticParams, _ := json.Marshal(req.StaticParams)
	countries := ""
	if len(req.Countries) > 0 {
		countriesJSON, _ := json.Marshal(req.Countries)
		countries = string(countriesJSON)
	}
	
	target.URL = req.URL
	target.Weight = req.Weight
	target.Cap = req.Cap
	target.Countries = countries
	target.ParamMapping = string(paramMapping)
	target.StaticParams = string(staticParams)
	
	if err := h.db.Save(&target).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update target"})
		return
	}
	
	c.JSON(http.StatusOK, target)
}

func (h *LinkHandler) DeleteTarget(c *gin.Context) {
	targetID := c.Param("target_id")
	
	result := h.db.Delete(&models.Target{}, targetID)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete target"})
		return
	}
	
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "target not found"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "target deleted successfully"})
}