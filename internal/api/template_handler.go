package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
	
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	
	"github.com/raoxb/smart_redirect/internal/models"
)

type TemplateHandler struct {
	db *gorm.DB
}

func NewTemplateHandler(db *gorm.DB) *TemplateHandler {
	return &TemplateHandler{db: db}
}

type LinkTemplate struct {
	ID           uint                   `gorm:"primaryKey" json:"id"`
	Name         string                 `gorm:"size:100" json:"name"`
	Description  string                 `json:"description"`
	BusinessUnit string                 `gorm:"size:10" json:"business_unit"`
	Network      string                 `gorm:"size:50" json:"network"`
	TotalCap     int                    `json:"total_cap"`
	BackupURL    string                 `json:"backup_url"`
	TargetConfig string                 `gorm:"type:jsonb" json:"target_config"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

type TemplateTargetConfig struct {
	URL          string            `json:"url"`
	Weight       int               `json:"weight"`
	Cap          int               `json:"cap"`
	Countries    []string          `json:"countries"`
	ParamMapping map[string]string `json:"param_mapping"`
	StaticParams map[string]string `json:"static_params"`
}

type CreateTemplateRequest struct {
	Name         string                 `json:"name" binding:"required"`
	Description  string                 `json:"description"`
	BusinessUnit string                 `json:"business_unit" binding:"required"`
	Network      string                 `json:"network" binding:"required"`
	TotalCap     int                    `json:"total_cap"`
	BackupURL    string                 `json:"backup_url"`
	Targets      []TemplateTargetConfig `json:"targets" binding:"required"`
}

func (h *TemplateHandler) CreateTemplate(c *gin.Context) {
	var req CreateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	targetsJSON, err := json.Marshal(req.Targets)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to serialize targets"})
		return
	}
	
	template := &LinkTemplate{
		Name:         req.Name,
		Description:  req.Description,
		BusinessUnit: req.BusinessUnit,
		Network:      req.Network,
		TotalCap:     req.TotalCap,
		BackupURL:    req.BackupURL,
		TargetConfig: string(targetsJSON),
	}
	
	if err := h.db.Create(template).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create template"})
		return
	}
	
	c.JSON(http.StatusCreated, template)
}

func (h *TemplateHandler) GetTemplate(c *gin.Context) {
	templateID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid template ID"})
		return
	}
	
	var template LinkTemplate
	if err := h.db.First(&template, templateID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Template not found"})
		return
	}
	
	c.JSON(http.StatusOK, template)
}

func (h *TemplateHandler) ListTemplates(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	
	var total int64
	var templates []LinkTemplate
	
	h.db.Model(&LinkTemplate{}).Count(&total)
	
	offset := (page - 1) * pageSize
	if err := h.db.Offset(offset).Limit(pageSize).Find(&templates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch templates"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"total": total,
		"page":  page,
		"size":  pageSize,
		"data":  templates,
	})
}

func (h *TemplateHandler) UpdateTemplate(c *gin.Context) {
	templateID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid template ID"})
		return
	}
	
	var template LinkTemplate
	if err := h.db.First(&template, templateID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Template not found"})
		return
	}
	
	var req CreateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	targetsJSON, err := json.Marshal(req.Targets)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to serialize targets"})
		return
	}
	
	template.Name = req.Name
	template.Description = req.Description
	template.BusinessUnit = req.BusinessUnit
	template.Network = req.Network
	template.TotalCap = req.TotalCap
	template.BackupURL = req.BackupURL
	template.TargetConfig = string(targetsJSON)
	
	if err := h.db.Save(&template).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update template"})
		return
	}
	
	c.JSON(http.StatusOK, template)
}

func (h *TemplateHandler) DeleteTemplate(c *gin.Context) {
	templateID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid template ID"})
		return
	}
	
	result := h.db.Delete(&LinkTemplate{}, templateID)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete template"})
		return
	}
	
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Template not found"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "Template deleted successfully"})
}

type CreateFromTemplateRequest struct {
	TemplateID uint                   `json:"template_id" binding:"required"`
	Count      int                    `json:"count" binding:"required,min=1,max=100"`
	Overrides  map[string]interface{} `json:"overrides"`
}

func (h *TemplateHandler) CreateLinksFromTemplate(c *gin.Context) {
	var req CreateFromTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	var template LinkTemplate
	if err := h.db.First(&template, req.TemplateID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Template not found"})
		return
	}
	
	var targetConfigs []TemplateTargetConfig
	if err := json.Unmarshal([]byte(template.TargetConfig), &targetConfigs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid template configuration"})
		return
	}
	
	response := struct {
		Success []BatchResult `json:"success"`
		Errors  []BatchError  `json:"errors"`
	}{
		Success: []BatchResult{},
		Errors:  []BatchError{},
	}
	
	for i := 0; i < req.Count; i++ {
		link := &models.Link{
			BusinessUnit: template.BusinessUnit,
			Network:      template.Network,
			TotalCap:     template.TotalCap,
			BackupURL:    template.BackupURL,
		}
		
		if overrides, ok := req.Overrides["business_unit"].(string); ok {
			link.BusinessUnit = overrides
		}
		if overrides, ok := req.Overrides["network"].(string); ok {
			link.Network = overrides
		}
		if overrides, ok := req.Overrides["total_cap"].(float64); ok {
			link.TotalCap = int(overrides)
		}
		if overrides, ok := req.Overrides["backup_url"].(string); ok {
			link.BackupURL = overrides
		}
		
		if err := h.db.Create(link).Error; err != nil {
			response.Errors = append(response.Errors, BatchError{
				Index:   i,
				Message: fmt.Sprintf("Failed to create link: %v", err),
			})
			continue
		}
		
		hasError := false
		for _, targetConfig := range targetConfigs {
			paramMapping, _ := json.Marshal(targetConfig.ParamMapping)
			staticParams, _ := json.Marshal(targetConfig.StaticParams)
			countries := ""
			if len(targetConfig.Countries) > 0 {
				countriesJSON, _ := json.Marshal(targetConfig.Countries)
				countries = string(countriesJSON)
			}
			
			target := &models.Target{
				LinkID:       link.ID,
				URL:          targetConfig.URL,
				Weight:       targetConfig.Weight,
				Cap:          targetConfig.Cap,
				Countries:    countries,
				ParamMapping: string(paramMapping),
				StaticParams: string(staticParams),
			}
			
			if err := h.db.Create(target).Error; err != nil {
				response.Errors = append(response.Errors, BatchError{
					Index:   i,
					Message: fmt.Sprintf("Failed to create target: %v", err),
				})
				hasError = true
				break
			}
		}
		
		if !hasError {
			linkURL := fmt.Sprintf("api.domain.com/v1/%s/%s?network=%s", 
				link.BusinessUnit, link.LinkID, link.Network)
			
			response.Success = append(response.Success, BatchResult{
				Index:   i,
				LinkID:  link.LinkID,
				LinkURL: linkURL,
			})
		}
	}
	
	c.JSON(http.StatusOK, response)
}