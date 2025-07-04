package api

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
	
	"github.com/raoxb/smart_redirect/internal/models"
	"github.com/raoxb/smart_redirect/internal/services"
)

type BatchHandler struct {
	linkService *services.LinkService
	db          *gorm.DB
}

func NewBatchHandler(db *gorm.DB, redis *redis.Client) *BatchHandler {
	return &BatchHandler{
		linkService: services.NewLinkService(db, redis),
		db:          db,
	}
}

type BatchCreateLinksRequest struct {
	Links []BatchLinkItem `json:"links" binding:"required"`
}

type BatchLinkItem struct {
	BusinessUnit string              `json:"business_unit" binding:"required"`
	Network      string              `json:"network" binding:"required"`
	TotalCap     int                 `json:"total_cap"`
	BackupURL    string              `json:"backup_url"`
	Targets      []BatchTargetItem   `json:"targets" binding:"required"`
}

type BatchTargetItem struct {
	URL          string            `json:"url" binding:"required"`
	Weight       int               `json:"weight" binding:"required,min=1"`
	Cap          int               `json:"cap"`
	Countries    []string          `json:"countries"`
	ParamMapping map[string]string `json:"param_mapping"`
	StaticParams map[string]string `json:"static_params"`
}

type BatchResponse struct {
	Success []BatchResult `json:"success"`
	Errors  []BatchError  `json:"errors"`
}

type BatchResult struct {
	Index    int    `json:"index"`
	LinkID   string `json:"link_id"`
	LinkURL  string `json:"link_url"`
}

type BatchError struct {
	Index   int    `json:"index"`
	Message string `json:"message"`
}

func (h *BatchHandler) BatchCreateLinks(c *gin.Context) {
	var req BatchCreateLinksRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	response := BatchResponse{
		Success: []BatchResult{},
		Errors:  []BatchError{},
	}
	
	for i, linkItem := range req.Links {
		link := &models.Link{
			BusinessUnit: linkItem.BusinessUnit,
			Network:      linkItem.Network,
			TotalCap:     linkItem.TotalCap,
			BackupURL:    linkItem.BackupURL,
		}
		
		if err := h.linkService.CreateLink(link); err != nil {
			response.Errors = append(response.Errors, BatchError{
				Index:   i,
				Message: fmt.Sprintf("Failed to create link: %v", err),
			})
			continue
		}
		
		hasError := false
		for _, targetItem := range linkItem.Targets {
			paramMapping, _ := json.Marshal(targetItem.ParamMapping)
			staticParams, _ := json.Marshal(targetItem.StaticParams)
			countries := ""
			if len(targetItem.Countries) > 0 {
				countriesJSON, _ := json.Marshal(targetItem.Countries)
				countries = string(countriesJSON)
			}
			
			target := &models.Target{
				LinkID:       link.ID,
				URL:          targetItem.URL,
				Weight:       targetItem.Weight,
				Cap:          targetItem.Cap,
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

func (h *BatchHandler) BatchUpdateLinks(c *gin.Context) {
	type BatchUpdateRequest struct {
		Updates []struct {
			LinkID       string `json:"link_id" binding:"required"`
			BusinessUnit string `json:"business_unit"`
			Network      string `json:"network"`
			TotalCap     *int   `json:"total_cap"`
			BackupURL    string `json:"backup_url"`
			IsActive     *bool  `json:"is_active"`
		} `json:"updates" binding:"required"`
	}
	
	var req BatchUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	response := BatchResponse{
		Success: []BatchResult{},
		Errors:  []BatchError{},
	}
	
	for i, update := range req.Updates {
		var link models.Link
		if err := h.db.Where("link_id = ?", update.LinkID).First(&link).Error; err != nil {
			response.Errors = append(response.Errors, BatchError{
				Index:   i,
				Message: fmt.Sprintf("Link not found: %s", update.LinkID),
			})
			continue
		}
		
		if update.BusinessUnit != "" {
			link.BusinessUnit = update.BusinessUnit
		}
		if update.Network != "" {
			link.Network = update.Network
		}
		if update.TotalCap != nil {
			link.TotalCap = *update.TotalCap
		}
		if update.BackupURL != "" {
			link.BackupURL = update.BackupURL
		}
		if update.IsActive != nil {
			link.IsActive = *update.IsActive
		}
		
		if err := h.db.Save(&link).Error; err != nil {
			response.Errors = append(response.Errors, BatchError{
				Index:   i,
				Message: fmt.Sprintf("Failed to update link: %v", err),
			})
			continue
		}
		
		response.Success = append(response.Success, BatchResult{
			Index:  i,
			LinkID: link.LinkID,
		})
	}
	
	c.JSON(http.StatusOK, response)
}

func (h *BatchHandler) BatchDeleteLinks(c *gin.Context) {
	type BatchDeleteRequest struct {
		LinkIDs []string `json:"link_ids" binding:"required"`
	}
	
	var req BatchDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	response := BatchResponse{
		Success: []BatchResult{},
		Errors:  []BatchError{},
	}
	
	for i, linkID := range req.LinkIDs {
		result := h.db.Where("link_id = ?", linkID).Delete(&models.Link{})
		if result.Error != nil {
			response.Errors = append(response.Errors, BatchError{
				Index:   i,
				Message: fmt.Sprintf("Failed to delete link: %v", result.Error),
			})
			continue
		}
		
		if result.RowsAffected == 0 {
			response.Errors = append(response.Errors, BatchError{
				Index:   i,
				Message: fmt.Sprintf("Link not found: %s", linkID),
			})
			continue
		}
		
		response.Success = append(response.Success, BatchResult{
			Index:  i,
			LinkID: linkID,
		})
	}
	
	c.JSON(http.StatusOK, response)
}

func (h *BatchHandler) ImportLinksFromCSV(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}
	defer file.Close()
	
	if !strings.HasSuffix(header.Filename, ".csv") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File must be CSV format"})
		return
	}
	
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse CSV"})
		return
	}
	
	if len(records) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "CSV must have header and at least one data row"})
		return
	}
	
	response := BatchResponse{
		Success: []BatchResult{},
		Errors:  []BatchError{},
	}
	
	headers := records[0]
	expectedHeaders := []string{"business_unit", "network", "total_cap", "backup_url", "target_url", "weight", "cap", "countries"}
	for _, expected := range expectedHeaders {
		found := false
		for _, header := range headers {
			if strings.TrimSpace(strings.ToLower(header)) == expected {
				found = true
				break
			}
		}
		if !found {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Missing required header: %s", expected)})
			return
		}
	}
	
	for i, record := range records[1:] {
		if len(record) < len(expectedHeaders) {
			response.Errors = append(response.Errors, BatchError{
				Index:   i,
				Message: "Insufficient columns in row",
			})
			continue
		}
		
		totalCap, _ := strconv.Atoi(strings.TrimSpace(record[2]))
		weight, _ := strconv.Atoi(strings.TrimSpace(record[5]))
		cap, _ := strconv.Atoi(strings.TrimSpace(record[6]))
		
		link := &models.Link{
			BusinessUnit: strings.TrimSpace(record[0]),
			Network:      strings.TrimSpace(record[1]),
			TotalCap:     totalCap,
			BackupURL:    strings.TrimSpace(record[3]),
		}
		
		if err := h.linkService.CreateLink(link); err != nil {
			response.Errors = append(response.Errors, BatchError{
				Index:   i,
				Message: fmt.Sprintf("Failed to create link: %v", err),
			})
			continue
		}
		
		countries := ""
		if strings.TrimSpace(record[7]) != "" {
			countryList := strings.Split(strings.TrimSpace(record[7]), ";")
			countriesJSON, _ := json.Marshal(countryList)
			countries = string(countriesJSON)
		}
		
		target := &models.Target{
			LinkID:    link.ID,
			URL:       strings.TrimSpace(record[4]),
			Weight:    weight,
			Cap:       cap,
			Countries: countries,
		}
		
		if err := h.db.Create(target).Error; err != nil {
			response.Errors = append(response.Errors, BatchError{
				Index:   i,
				Message: fmt.Sprintf("Failed to create target: %v", err),
			})
			continue
		}
		
		linkURL := fmt.Sprintf("api.domain.com/v1/%s/%s?network=%s", 
			link.BusinessUnit, link.LinkID, link.Network)
		
		response.Success = append(response.Success, BatchResult{
			Index:   i,
			LinkID:  link.LinkID,
			LinkURL: linkURL,
		})
	}
	
	c.JSON(http.StatusOK, response)
}

func (h *BatchHandler) ExportLinksToCSV(c *gin.Context) {
	var links []models.Link
	if err := h.db.Preload("Targets").Find(&links).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch links"})
		return
	}
	
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment; filename=links_export.csv")
	
	writer := csv.NewWriter(c.Writer)
	defer writer.Flush()
	
	headers := []string{"link_id", "business_unit", "network", "total_cap", "backup_url", "target_url", "weight", "cap", "countries", "current_hits"}
	writer.Write(headers)
	
	for _, link := range links {
		for _, target := range link.Targets {
			countries := ""
			if target.Countries != "" {
				var countryList []string
				if err := json.Unmarshal([]byte(target.Countries), &countryList); err == nil {
					countries = strings.Join(countryList, ";")
				}
			}
			
			record := []string{
				link.LinkID,
				link.BusinessUnit,
				link.Network,
				strconv.Itoa(link.TotalCap),
				link.BackupURL,
				target.URL,
				strconv.Itoa(target.Weight),
				strconv.Itoa(target.Cap),
				countries,
				strconv.Itoa(link.CurrentHits),
			}
			writer.Write(record)
		}
	}
}