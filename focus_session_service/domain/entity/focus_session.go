package entity

import (
	"math"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type FocusSession struct {
	ID              primitive.ObjectID  `json:"id" bson:"_id,omitempty"`
	UserID          primitive.ObjectID  `json:"userId" bson:"userId"`
	Title           string              `json:"title" bson:"title"`
	Description     string              `json:"description" bson:"description"`
	StartTime       time.Time           `json:"startTime" bson:"startTime"`
	EndTime         *time.Time          `json:"endTime" bson:"endTime"`
	Duration        int                 `json:"duration" bson:"duration"` // minutes
	ActualDuration  *int                `json:"actualDuration" bson:"actualDuration"`
	Status          SessionStatus       `json:"status" bson:"status"`
	LocationID      *primitive.ObjectID `json:"locationId,omitempty" bson:"locationId,omitempty"`
	LocationDetails *LocationDetails    `json:"locationDetails,omitempty" bson:"locationDetails,omitempty"`
	Tags            []string            `json:"tags,omitempty" bson:"tags"`
	Notes           string              `json:"notes,omitempty" bson:"notes,omitempty"`
	Rating          *int                `json:"rating,omitempty"`
	Distractions    *int                `json:"distractions,omitempty" bson:"distractions,omitempty"`
	Focus           *int                `json:"focus,omitempty" bson:"focus,omitempty"`   // Focus level (1-10)
	Energy          *int                `json:"energy,omitempty" bson:"energy,omitempty"` // Energy level (1-10)
	Mood            *int                `json:"mood,omitempty" bson:"mood,omitempty"`     // Mood
	CreatedAt       time.Time           `json:"createdAt" bson:"createdAt"`
	UpdatedAt       time.Time           `json:"updatedAt" bson:"updatedAt"`
	Active          bool                `json:"active" bson:"active"` // default: true
	DeletedAt       *time.Time          `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
}

type SessionStatus string

const (
	StatusPlanned   SessionStatus = "planned"
	StatusActive    SessionStatus = "active"
	StatusCompleted SessionStatus = "completed"
	StatusCancelled SessionStatus = "cancelled"
)

type LocationDetails struct {
	Name      string  `json:"name" bson:"name"`
	Address   string  `json:"address,omitempty" bson:"address,omitempty"`
	Latitude  float64 `json:"latitude,omitempty" bson:"latitude,omitempty"`
	Longitude float64 `json:"longitude,omitempty" bson:"longitude,omitempty"`
	Type      string  `json:"type,omitempty" bson:"type,omitempty"` // coffee shop, library, etc.
}

// GetCalculateProductivityScore calculates the productivity score for a session
func (s *FocusSession) CalculateProductivityScore() float64 {
	if s.Status != StatusCompleted || s.ActualDuration == nil || s.Rating == nil {
		return 0.0
	}

	// Base score is the rating (1-5)
	score := float64(*s.Rating)

	// Adjust score based on actual vs planned Duration
	if s.Duration > 0 {
		completionRatio := float64(*s.ActualDuration) / float64(s.Duration)
		if completionRatio > 1.2 {
			// Bonus for working longer than planned
			score += 1.0
		} else {
			// Penalty for working less than planned
			score -= 1.0
		}
	}

	// Adjust for focus if available
	if s.Focus != nil {
		// 0-2 points based on focus
		focusBonus := float64(*s.Focus) / 10.0 * 2.0
		score += focusBonus
	}

	// Penalize for Distractions
	if s.Distractions != nil && *s.Distractions > 0 {
		// Convert to 0-1 penalty (more distractions = higher penalty)
		distractionPenalty := math.Min(float64(*s.Distractions)/10.0, 1.0)
		score -= distractionPenalty
	}

	return math.Max(0, math.Min(10, score))
}
