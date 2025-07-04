package models

import (
	"encoding/json"
	"gorm.io/gorm"
)

// TargetResponse is the response structure with parsed JSON fields
type TargetResponse struct {
	ID            uint                   `json:"id"`
	LinkID        uint                   `json:"link_id"`
	URL           string                 `json:"url"`
	Weight        int                    `json:"weight"`
	Cap           int                    `json:"cap"`
	CurrentHits   int                    `json:"current_hits"`
	Countries     []string               `json:"countries"`
	ParamMapping  map[string]string      `json:"param_mapping"`
	StaticParams  map[string]string      `json:"static_params"`
	IsActive      bool                   `json:"is_active"`
	CreatedAt     string                 `json:"created_at"`
	UpdatedAt     string                 `json:"updated_at"`
}

// AfterFind hook to parse JSON fields
func (t *Target) AfterFind(tx *gorm.DB) error {
	// Countries is already a string in the database
	// No need to parse here, we'll handle it in the API response
	return nil
}

// ToResponse converts Target model to response format with parsed JSON
func (t *Target) ToResponse() TargetResponse {
	resp := TargetResponse{
		ID:          t.ID,
		LinkID:      t.LinkID,
		URL:         t.URL,
		Weight:      t.Weight,
		Cap:         t.Cap,
		CurrentHits: t.CurrentHits,
		IsActive:    t.IsActive,
		CreatedAt:   t.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   t.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
	
	// Parse countries
	if t.Countries != "" {
		json.Unmarshal([]byte(t.Countries), &resp.Countries)
	}
	if resp.Countries == nil {
		resp.Countries = []string{}
	}
	
	// Parse param mapping
	if t.ParamMapping != "" {
		json.Unmarshal([]byte(t.ParamMapping), &resp.ParamMapping)
	}
	if resp.ParamMapping == nil {
		resp.ParamMapping = make(map[string]string)
	}
	
	// Parse static params
	if t.StaticParams != "" {
		json.Unmarshal([]byte(t.StaticParams), &resp.StaticParams)
	}
	if resp.StaticParams == nil {
		resp.StaticParams = make(map[string]string)
	}
	
	return resp
}