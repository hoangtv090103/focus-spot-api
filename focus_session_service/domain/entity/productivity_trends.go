package entity

type Period string

const (
	Daily   Period = "daily"
	Weekly  Period = "weekly"
	Monthly Period = "monthly"
)

type ProductivityTrends struct {
	Period       Period    `json:"period"` // daily, weekly, monthly
	Dates        []string  `json:"dates"`
	Durations    []int     `json:"durations"` // in minutes
	Ratings      []float64 `json:"ratings"`
	Focus        []float64 `json:"focus"`
	Energy       []float64 `json:"energy"`
	Mood         []float64 `json:"mood"`
	Productivity []float64 `json:"productivity"` // overall productivity score
}

// Có thể thêm các helper methods
func (t *ProductivityTrends) GetAverageProductivity() float64 {
	// TODO: Tính trung bình của mảng Productivity
	// ...
	return 0
}

func (t *ProductivityTrends) IsImproving() bool {
	// Phân tích xu hướng để xác định nếu năng suất đang tăng
	// ...
	return false
}
