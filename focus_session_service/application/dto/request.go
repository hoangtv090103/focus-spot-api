package dto

import (
	"focusspot/focussessionservice/domain/entity"
	"time"
)

type CreateSessionRequest struct {
	Title           string                  `json:"title" validate:"required"`
	Description     string                  `json:"description"`
	StartTime       time.Time               `json:"startTime" validate:"required"`
	Duration        int                     `json:"duration" validate:"required,min=1"`
	LocationID      string                  `json:"locationId,omitempty"`
	LocationDetails *LocationDetailsRequest `json:"locationDetails,omitempty"`
	Tags            []string                `json:"tags,omitempty"`
}

type LocationDetailsRequest struct {
	Name      string  `json:"name" validate:"required"`
	Address   string  `json:"address"`
	Latitude  float64 `json:"latitude,omitempty"`
	Longitude float64 `json:"longitude,omitempty"`
	Type      string  `json:"type,omitempty"`
}

type UpdateSessionRequest struct {
	Title           string                  `json:"title,omitempty"`
	Description     string                  `json:"description,omitempty"`
	StartTime       *time.Time              `json:"startTime,omitempty"`
	Duration        *int                    `json:"duration,omitempty" validate:"omitempty,min=1"`
	LocationID      string                  `json:"locationId,omitempty"`
	LocationDetails *LocationDetailsRequest `json:"locationDetails,omitempty"`
	Tags            []string                `json:"tags,omitempty"`
	Status          string                  `json:"status,omitempty" validate:"omitempty,oneof=planned active completed cancelled"`
}

type EndSessionRequest struct {
	Notes        string `json:"notes,omitempty"`
	Rating       *int   `json:"rating,omitempty" validate:"omitempty,min=1,max=5"`
	Focus        *int   `json:"focus,omitempty" validate:"omitempty,min=1,max=10"`
	Energy       *int   `json:"energy,omitempty" validate:"omitempty,min=1,max=10"`
	Mood         *int   `json:"mood,omitempty" validate:"omitempty,min=1,max=10"`
	Distractions *int   `json:"distractions,omitempty" validate:"omitempty,min=0"`
}

type GetSessionsRequest struct {
	StartDate string `query:"startDate" validate:"omitempty,datetime=2006-01-02"`
	EndDate   string `query:"endDate" validate:"omitempty,datetime=2006-01-02"`
	Limit     int    `query:"limit,default=20" validate:"omitempty,min=1,max=100"`
	Offset    int    `query:"offset,default=0" validate:"omitempty,min=0"`
}

type GetProductivityStatsRequest struct {
	StartDate string `query:"startDate" validate:"omitempty,datetime=2006-01-02"`
	EndDate   string `query:"endDate" validate:"omitempty,datetime=2006-01-02"`
}

type GetProductivityTrendsRequest struct {
	Period entity.Period `query:"period,default=weekly" validate:"omitempty,oneof=daily weekly monthly"`
	Limit  int           `query:"limit,default=12" validate:"omitempty,min=1,max=52"`
}
