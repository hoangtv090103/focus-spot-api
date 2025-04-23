package entity

import "time"

type SessionStats struct {
	TotalSessions              int                      `json:"totalSessions"`
	CompletedSessions          int                      `json:"completedSessions"`
	CancelledSessions          int                      `json:"cancelledSessions"`
	TotalDuration              int                      `json:"totalDuration"` // in minutes
	
	AverageDuration            float64                  `json:"averageDuration"`
	AverageRating              float64                  `json:"averageRating"`
	AverageFocus               float64                  `json:"averageFocus"`
	AverageEnergy              float64                  `json:"averageEnergy"`
	AverageDistractions        float64                  `json:"averageDistractions"`
	
	MostUsedLocation           string                   `json:"mostUsedLocation"`
	MostProductiveLocation     string                   `json:"mostProductiveLocation"`
	MostProductiveLocationType string                   `json:"mostProductiveLocationType"`
	
	MostProductiveDay          time.Weekday             `json:"mostProductivityDay"`
	MostProductiveTime         *time.Time               `json:"mostProductivityTime,omitempty"`
	
	ProductivityByDay          map[time.Weekday]float64 `json:"productivityByDay"`
	ProductivityByTime         map[TimeOfDay]float64    `json:"productivityByTime"`
	ProductivityByLocationType map[string]float64       `json:"productivityByLocationType"`    
}
