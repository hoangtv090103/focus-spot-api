package entity

import "time"

type TimeOfDay string

const (
    EarlyMorning TimeOfDay = "early_morning" // 5:00-8:59
    LateMorning TimeOfDay = "late_morning" // 9:00-11:59
    Afternoon TimeOfDay = "afternon" // 12:00-16:59
    Evening TimeOfDay = "evening" // 17:00-20:59
    Night TimeOfDay = "night" // 21:00-4:59
)

// GetTimeOfDay returns the time of day for a given time
func GetTimeOfDay(t time.Time) TimeOfDay {
    hour := t.Hour()

    switch {
        case hour >= 5 && hour < 9:
            return EarlyMorning
        case hour >= 9 && hour < 12:
            return LateMorning
        case hour >= 12 && hour < 17:
            return Afternoon
        case hour >= 17 && hour < 21:
            return Evening
        default:
            return Night
    }
}