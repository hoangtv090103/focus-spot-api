package dto

import (
	"focusspot/focussessionservice/domain/entity"
	"time"
)

type FocusSessionResponse struct {
	ID                string                   `json:"id"`
	UserID            string                   `json:"userId"`
	Title             string                   `json:"title"`
	Description       string                   `json:"description,omitempty"`
	StartTime         time.Time                `json:"startTime"`
	EndTime           *time.Time               `json:"endTime,omitempty"`
	Duration          int                      `json:"duration"`
	ActualDuration    *int                     `json:"actualDuration,omitempty"`
	Status            string                   `json:"status"`
	LocationID        string                   `json:"locationId,omitempty"`
	LocationDetails   *LocationDetailsResponse `json:"locationDetails,omitempty"`
	Tags              []string                 `json:"tags,omitempty"`
	Notes             string                   `json:"notes,omitempty"`
	Rating            *int                     `json:"rating,omitempty"`
	Focus             *int                     `json:"focus,omitempty"`
	Energy            *int                     `json:"energy,omitempty"`
	Mood              *int                     `json:"mood,omitempty"`
	Distractions      *int                     `json:"distractions,omitempty"`
	ProductivityScore *float64                 `json:"productivityScore,omitempty"`
	CreatedAt         time.Time                `json:"createdAt"`
	UpdatedAt         time.Time                `json:"updatedAt"`
}

type LocationDetailsResponse struct {
	Name      string  `json:"name"`
	Address   string  `json:"address,omitempty"`
	Latitude  float64 `json:"latitude,omitempty"`
	Longitude float64 `json:"longitude,omitempty"`
	Type      string  `json:"type,omitempty"`
}

type SessionsListResponse struct {
	Sessions []FocusSessionResponse `json:"sessions"`
	Total    int                    `json:"total"`
	Limit    int                    `json:"limit"`
	Offset   int                    `json:"offset"`
}

type DateRange struct {
	StartDate time.Time `json:"startDate"`
	EndDate   time.Time `json:"endDate"`
}

// ProductivityStatsResponse contains detailed productivity analytics for a user
type ProductivityStatsResponse struct {
	TotalSessions     int     `json:"totalSessions"`
	CompletedSessions int     `json:"completedSessions"`
	CancelledSessions int     `json:"cancelledSessions"`
	TotalDuration     int     `json:"totalDuration"` // in minutes
	AverageDuration   float64 `json:"averageDuration"`

	// Metrics averages
	AverageRating       float64 `json:"averageRating"`
	AverageFocus        float64 `json:"averageFocus"`
	AverageEnergy       float64 `json:"averageEnergy"`
	AverageMood         float64 `json:"averageMood"`
	AverageDistractions float64 `json:"averageDistractions"`

	// Productivity by day of week
	ProductivityByDay map[string]float64 `json:"productivityByDay"`
	MostProductiveDay string             `json:"mostProductiveDay"`

	// Productivity by time of day
	ProductivityByTime map[string]float64 `json:"productivityByTime"`
	MostProductiveTime string             `json:"mostProductiveTime"`

	// Productivity by location
	ProductivityByLocation map[string]float64 `json:"productivityByLocation"`
	MostProductiveLocation string             `json:"mostProductiveLocation"`

	// Productivity by location type
	ProductivityByLocationType map[string]float64 `json:"productivityByLocationType"`
	MostProductiveLocationType string             `json:"mostProductiveLocationType"`

	// Date range for the stats
	DateRange DateRange `json:"dateRange"`
}

// ProductivityTrendsResponse contains productivity data over time
type ProductivityTrendsResponse struct {
	Period       entity.Period `json:"period"` // daily, weekly, monthly
	Dates        []string      `json:"dates"`
	Durations    []int         `json:"durations"` // in minutes
	Ratings      []float64     `json:"ratings"`
	Focus        []float64     `json:"focus"`
	Energy       []float64     `json:"energy"`
	Mood         []float64     `json:"mood"`
	Productivity []float64     `json:"productivity"` // overall productivity score
}

// ToFocusSessionResponse converts a FocusSession entity to a FocusSessionResponse DTO
func ToFocusSessionResponse(session *entity.FocusSession) FocusSessionResponse {
	response := FocusSessionResponse{
		ID:             session.ID.Hex(),
		UserID:         session.UserID.Hex(),
		Title:          session.Title,
		Description:    session.Description,
		StartTime:      session.StartTime,
		EndTime:        session.EndTime,
		Duration:       session.Duration,
		ActualDuration: session.ActualDuration,
		Status:         string(session.Status),
		Tags:           session.Tags,
		Notes:          session.Notes,
		Rating:         session.Rating,
		Focus:          session.Focus,
		Energy:         session.Energy,
		Mood:           session.Mood,
		Distractions:   session.Distractions,
		CreatedAt:      session.CreatedAt,
		UpdatedAt:      session.UpdatedAt,
	}

	if session.Status == entity.StatusCompleted && session != nil {
		score := session.CalculateProductivityScore()
		response.ProductivityScore = &score
	}

	if session.LocationID != nil {
		response.LocationID = session.LocationID.Hex()
	}

	if session.LocationDetails != nil {
		response.LocationDetails = &LocationDetailsResponse{
			Name:      session.LocationDetails.Name,
			Address:   session.LocationDetails.Address,
			Latitude:  session.LocationDetails.Latitude,
			Longitude: session.LocationDetails.Longitude,
			Type:      session.LocationDetails.Type,
		}
	}
	return response
}

// ToProductivityStatsResponse converts domain stats to response DTO
func ToProductivityStatsResponse(stats *entity.ProductivityStats, dateRange DateRange) ProductivityStatsResponse {
	// Map weekday indices to names
	dayNames := map[time.Weekday]string{
		time.Sunday:    "Sunday",
		time.Monday:    "Monday",
		time.Tuesday:   "Tuesday",
		time.Wednesday: "Wednesday",
		time.Thursday:  "Thursday",
		time.Friday:    "Friday",
		time.Saturday:  "Saturday",
	}

	// Map TimeOfDay to readable names
	timeNames := map[entity.TimeOfDay]string{
		entity.EarlyMorning: "Early Morning (5:00-8:59)",
		entity.LateMorning:  "Late Morning (9:00-11:59)",
		entity.Afternoon:    "Afternoon (12:00-16:59)",
		entity.Evening:      "Evening (17:00-20:59)",
		entity.Night:        "Night (21:00-4:59)",
	}

	// Convert map with weekday keys to map with string keys
	productivityByDay := make(map[string]float64)
	for day, score := range stats.ProductivityByDay {
		productivityByDay[dayNames[day]] = score
	}

	// Convert map with TimeOfDay keys to map with string keys
	productivityByTime := make(map[string]float64)
	for tod, score := range stats.ProductivityByTime {
		productivityByTime[timeNames[tod]] = score
	}

	// Calculate average duration
	var avgDuration float64
	if stats.CompletedSessions > 0 {
		avgDuration = float64(stats.TotalDuration) / float64(stats.CompletedSessions)
	}

	return ProductivityStatsResponse{
		TotalSessions:              stats.TotalSessions,
		CompletedSessions:          stats.CompletedSessions,
		CancelledSessions:          stats.CancelledSessions,
		TotalDuration:              stats.TotalDuration,
		
		AverageDuration:            avgDuration,
		AverageRating:              stats.AverageRating,
		AverageFocus:               stats.AverageFocus,
		AverageEnergy:              stats.AverageEnergy,
		AverageMood:                stats.AverageMood,
		AverageDistractions:        stats.AverageDistractions,
		
		ProductivityByDay:          productivityByDay,
		MostProductiveDay:          dayNames[stats.MostProductiveDay],
		ProductivityByTime:         productivityByTime,
		MostProductiveTime:         timeNames[stats.MostProductiveTime],
		
		ProductivityByLocation:     stats.ProductivityByLocation,
		MostProductiveLocation:     stats.MostProductiveLocation,
		ProductivityByLocationType: stats.ProductivityByLocationType,
		MostProductiveLocationType: stats.MostProductiveLocationType,
		
		DateRange:                  dateRange,
	}
}
