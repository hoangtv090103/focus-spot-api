package entity

import "time"

type ProductivityStats struct {
	TotalSessions       int     `json:"totalSessions"`
	CompletedSessions   int     `json:"completedSessions"`
	CancelledSessions   int     `json:"cancelledSessions"`
	TotalDuration       int     `json:"totalDuration"` // in minutes
	AverageRating       float64 `json:"averageRating"`
	AverageFocus        float64 `json:"averageFocus"`
	AverageEnergy       float64 `json:"averageEnergy"`
	AverageMood         float64 `json:"averageMood"`
	AverageDistractions float64 `json:"averageDistractions"`

	// Productivity by day of week (0=Sunday, 6=Saturday)
	ProductivityByDay map[time.Weekday]float64 `json:"productivityByDay"`
	MostProductiveDay time.Weekday             `json:"mostProductiveDay"`

	// Productivity by time of day
	ProductivityByTime map[TimeOfDay]float64 `json:"productivityByTime"`
	MostProductiveTime TimeOfDay             `json:"mostProductiveTime"`

	// Productivity by location
	ProductivityByLocation map[string]float64 `json:"productivityByLocation"`
	MostProductiveLocation string             `json:"mostProductiveLocation"`

	// Productivity by location type
	ProductivityByLocationType map[string]float64 `json:"productivityByLocationType"`
	MostProductiveLocationType string             `json:"mostProductiveLocationType"`
}

func (s *ProductivityStats) GetAverageDuration() float64 {
    if s.CompletedSessions == 0 {
        return 0
    }
    return float64(s.TotalDuration) / float64(s.CompletedSessions)
}

func (s *ProductivityStats) GetOverallProductivityScore() float64 {
    // TODO: Tính điểm năng suất tổng thể dựa trên nhiều yếu tố
    return 0
}