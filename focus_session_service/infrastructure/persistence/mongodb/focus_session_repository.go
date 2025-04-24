package mongodb

import (
	"context"
	"errors"
	"fmt"
	"focusspot/focussessionservice/domain/entity"
	"focusspot/focussessionservice/domain/interfaces"
	"sort"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type mongoFocusSessionRepository struct {
	collection *mongo.Collection
}

func NewMongoFocusSessionRepository(db *mongo.Database) interfaces.IFocusSessionRepository {
	collection := db.Collection("focus_sessions")

	// Create indexes
	_, err := collection.Indexes().CreateMany(
		context.Background(),
		[]mongo.IndexModel{
			{
				Keys: bson.D{
					{Key: "userId", Value: 1},
					{Key: "startTime", Value: -1},
				},
			},
			{
				Keys: bson.D{
					{Key: "userId", Value: 1}, {Key: "status", Value: 1},
				},
			},
		})

	if err != nil {
		// TODO: In production, handle this error properly
		panic(err)
	}

	return &mongoFocusSessionRepository{
		collection: collection,
	}
}

func (r *mongoFocusSessionRepository) Create(ctx context.Context, session *entity.FocusSession) error {
	if session.ID.IsZero() {
		session.ID = primitive.NewObjectID()
	}

	_, err := r.collection.InsertOne(ctx, session)

	return err
}

func (r *mongoFocusSessionRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*entity.FocusSession, error) {
	var session entity.FocusSession

	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&session)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("session not found")
		}
		return nil, err
	}

	return &session, nil
}

func (r *mongoFocusSessionRepository) GetByUserID(ctx context.Context, userID primitive.ObjectID, limit, offset int) ([]*entity.FocusSession, error) {
	findOptions := options.Find()
	findOptions.SetLimit(int64(limit))
	findOptions.SetSkip(int64(offset))
	findOptions.SetSort(bson.D{{Key: "startTime", Value: -1}})

	// Add filter for active sessions only
	filter := bson.M{
		"userId": userID,
		"active": true,
	}

	cursor, err := r.collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var sessions []*entity.FocusSession
	if err := cursor.All(ctx, &sessions); err != nil {
		return nil, err
	}

	return sessions, nil
}

func (r *mongoFocusSessionRepository) GetActiveByUserID(ctx context.Context, userID primitive.ObjectID) (*entity.FocusSession, error) {
	var session entity.FocusSession

	err := r.collection.FindOne(ctx, bson.M{
		"userId": userID,
		"status": entity.StatusActive,
		"active": true,
	}).Decode(&session)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil // No active session found, not an error
		}
		return nil, err
	}

	return &session, nil
}

func (r *mongoFocusSessionRepository) GetSessionsByDateRange(ctx context.Context, userID primitive.ObjectID, startDate, endDate time.Time) ([]*entity.FocusSession, error) {
	filter := bson.M{
		"userId": userID,
		"startTime": bson.M{
			"$gte": startDate,
			"$lte": endDate,
		},
		"active": true,
	}

	findOptions := options.Find()
	findOptions.SetSort(bson.D{{Key: "startTime", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var sessions []*entity.FocusSession
	if err := cursor.All(ctx, &sessions); err != nil {
		return nil, err
	}

	return sessions, nil
}

func (r *mongoFocusSessionRepository) Update(ctx context.Context, session *entity.FocusSession) error {
	session.UpdatedAt = time.Now()

	_, err := r.collection.ReplaceOne(
		ctx,
		bson.M{"_id": session.ID},
		session,
	)

	return err
}

func (r *mongoFocusSessionRepository) UpdateStatus(ctx context.Context, id primitive.ObjectID, status entity.SessionStatus) error {
	now := time.Now()

	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{
			"$set": bson.M{
				"status":    status,
				"updatedAt": now,
			},
		},
	)

	return err
}

func (r *mongoFocusSessionRepository) EndSession(ctx context.Context, id primitive.ObjectID, endTime time.Time, notes string, rating, focus, energy, mood, distractions *int) error {
	now := time.Now()

	// Calculate actual duration
	var session entity.FocusSession
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&session)
	if err != nil {
		return err
	}

	actualDuration := int(endTime.Sub(session.StartTime).Minutes())

	update := bson.M{
		"$set": bson.M{
			"status":         entity.StatusCompleted,
			"endTime":        endTime,
			"actualDuration": actualDuration,
			"updatedAt":      now,
		},
	}

	if notes != "" {
		update["$set"].(bson.M)["notes"] = notes
	}

	if rating != nil {
		update["$set"].(bson.M)["rating"] = rating
	}

	if focus != nil {
		update["$set"].(bson.M)["focus"] = focus
	}

	if energy != nil {
		update["$set"].(bson.M)["energy"] = energy
	}

	if mood != nil {
		update["$set"].(bson.M)["mood"] = mood
	}

	if distractions != nil {
		update["$set"].(bson.M)["distractions"] = distractions
	}

	_, err = r.collection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		update,
	)

	return err
}

func (r *mongoFocusSessionRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{
			"$set": bson.M{
				"active": false,
			},
		})
	return err
}

func (r *mongoFocusSessionRepository) GetProductivityStats(
	ctx context.Context,
	userID primitive.ObjectID,
	startDate, endDate time.Time,
) (*entity.ProductivityStats, error) {
	// Get completed sessions in date range
	filter := bson.M{
		"userId": userID,
		"startTime": bson.M{
			"$gte": startDate,
			"$lte": endDate,
		},
		"status": entity.StatusCompleted,
		"active": true,
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var sessions []*entity.FocusSession
	if err := cursor.All(ctx, &sessions); err != nil {
		return nil, err
	}

	stats := &entity.ProductivityStats{
		TotalSessions:              len(sessions),
		ProductivityByDay:          make(map[time.Weekday]float64),
		ProductivityByTime:         make(map[entity.TimeOfDay]float64),
		ProductivityByLocation:     make(map[string]float64),
		ProductivityByLocationType: make(map[string]float64),
	}

	if len(sessions) == 0 {
		return stats, nil
	}

	// Counter structures for analysis
	type MetricCounter struct {
		Count         int
		TotalScore    float64
		TotalRating   int
		TotalFocus    int
		TotalEnergy   int
		TotalMood     int
		TotalDistract int
		TotalDuration int
	}

	dayCounter := make(map[time.Weekday]*MetricCounter)
	timeCounter := make(map[entity.TimeOfDay]*MetricCounter)
	locationCounter := make(map[string]*MetricCounter)
	locationTypeCounter := make(map[string]*MetricCounter)

	// Initialize counters for all possible values
	for day := time.Sunday; day <= time.Saturday; day++ {
		dayCounter[day] = &MetricCounter{}
	}

	for _, tod := range []entity.TimeOfDay{entity.EarlyMorning, entity.LateMorning, entity.Afternoon, entity.Evening, entity.Night} {
		timeCounter[tod] = &MetricCounter{}
	}

	var completedCount, cancelledCount int
	var totalRating, ratingCount int
	var totalFocus, focusCount int
	var totalEnergy, energyCount int
	var totalMood, moodCount int
	var totalDistract, distractCount int
	var totalDuration int

	// Process each session
	for _, session := range sessions {
		// Skip sessions with no actual duration
		if session.ActualDuration != nil {
			completedCount++
			totalDuration += *session.ActualDuration

			// Calculate productivity score
			prodScore := session.CalculateProductivityScore()

			// Process by day of week
			day := session.StartTime.Weekday()
			counter := dayCounter[day]
			counter.Count++
			counter.TotalScore += prodScore
			counter.TotalDuration += *session.ActualDuration

			// Process by time of day
			tod := entity.GetTimeOfDay(session.StartTime)
			timeCounter[tod].Count++
			timeCounter[tod].TotalScore += prodScore
			timeCounter[tod].TotalDuration += *session.ActualDuration

			// Process by location
			if session.LocationDetails != nil {
				// By location name
				locName := session.LocationDetails.Name
				if _, exists := locationCounter[locName]; !exists {
					locationCounter[locName] = &MetricCounter{}
				}
				locationCounter[locName].Count++
				locationCounter[locName].TotalScore += prodScore
				locationCounter[locName].TotalDuration += *session.ActualDuration

				// By location type
				if session.LocationDetails.Type != "" {
					locType := session.LocationDetails.Type
					if _, exists := locationTypeCounter[locType]; !exists {
						locationTypeCounter[locType] = &MetricCounter{}
					}
					locationTypeCounter[locType].Count++
					locationTypeCounter[locType].TotalScore += prodScore
					locationTypeCounter[locType].TotalDuration += *session.ActualDuration
				}
			}

			// Collect metrics
			if session.Rating != nil {
				totalRating += *session.Rating
				ratingCount++

				// Add to counters
				dayCounter[day].TotalRating += *session.Rating
				timeCounter[tod].TotalRating += *session.Rating

				if session.LocationDetails != nil {
					locationCounter[session.LocationDetails.Name].TotalRating += *session.Rating
					if session.LocationDetails.Type != "" {
						locationTypeCounter[session.LocationDetails.Type].TotalRating += *session.Rating
					}
				}
			}

			if session.Focus != nil {
				totalFocus += *session.Focus
				focusCount++

				// Add to counters
				dayCounter[day].TotalFocus += *session.Focus
				timeCounter[tod].TotalFocus += *session.Focus

				if session.LocationDetails != nil {
					locationCounter[session.LocationDetails.Name].TotalFocus += *session.Focus
					if session.LocationDetails.Type != "" {
						locationTypeCounter[session.LocationDetails.Type].TotalFocus += *session.Focus
					}
				}
			}

			if session.Energy != nil {
				totalEnergy += *session.Energy
				energyCount++

				// Add to counters
				dayCounter[day].TotalEnergy += *session.Energy
				timeCounter[tod].TotalEnergy += *session.Energy

				if session.LocationDetails != nil {
					locationCounter[session.LocationDetails.Name].TotalEnergy += *session.Energy
					if session.LocationDetails.Type != "" {
						locationTypeCounter[session.LocationDetails.Type].TotalEnergy += *session.Energy
					}
				}
			}

			if session.Mood != nil {
				totalMood += *session.Mood
				moodCount++

				// Add to counters
				dayCounter[day].TotalMood += *session.Mood
				timeCounter[tod].TotalMood += *session.Mood

				if session.LocationDetails != nil {
					locationCounter[session.LocationDetails.Name].TotalMood += *session.Mood
					if session.LocationDetails.Type != "" {
						locationTypeCounter[session.LocationDetails.Type].TotalMood += *session.Mood
					}
				}
			}

			if session.Distractions != nil {
				totalDistract += *session.Distractions
				distractCount++

				// Add to counters
				dayCounter[day].TotalDistract += *session.Distractions
				timeCounter[tod].TotalDistract += *session.Distractions

				if session.LocationDetails != nil {
					locationCounter[session.LocationDetails.Name].TotalDistract += *session.Distractions
					if session.LocationDetails.Type != "" {
						locationTypeCounter[session.LocationDetails.Type].TotalDistract += *session.Distractions
					}
				}
			}
		} else if session.Status == entity.StatusCancelled {
			cancelledCount++
		}
	}

	// Set overall stats
	stats.CompletedSessions = completedCount
	stats.CancelledSessions = cancelledCount
	stats.TotalDuration = totalDuration

	// Calculate averages
	if ratingCount > 0 {
		stats.AverageRating = float64(totalRating) / float64(ratingCount)
	}

	if focusCount > 0 {
		stats.AverageFocus = float64(totalFocus) / float64(focusCount)
	}

	if energyCount > 0 {
		stats.AverageEnergy = float64(totalEnergy) / float64(energyCount)
	}

	if moodCount > 0 {
		stats.AverageMood = float64(totalMood) / float64(moodCount)
	}

	if distractCount > 0 {
		stats.AverageDistractions = float64(totalDistract) / float64(distractCount)
	}

	// Calculate productivity by day
	var maxDayScore float64
	for day, counter := range dayCounter {
		if counter.Count > 0 {
			// Average productivity score for this day
			avgScore := counter.TotalScore / float64(counter.Count)
			stats.ProductivityByDay[day] = avgScore

			if avgScore > maxDayScore {
				maxDayScore = avgScore
				stats.MostProductiveDay = day
			}
		}
	}

	// Calculate productivity by time of day
	var maxTimeScore float64
	for tod, counter := range timeCounter {
		if counter.Count > 0 {
			// Average productivity score for this time of day
			avgScore := counter.TotalScore / float64(counter.Count)
			stats.ProductivityByTime[tod] = avgScore

			if avgScore > maxTimeScore {
				maxTimeScore = avgScore
				stats.MostProductiveTime = tod
			}
		}
	}

	// Calculate productivity by location
	var maxLocationScore float64
	for loc, counter := range locationCounter {
		if counter.Count > 0 {
			// Average productivity score for this location
			avgScore := counter.TotalScore / float64(counter.Count)
			stats.ProductivityByLocation[loc] = avgScore

			if avgScore > maxLocationScore {
				maxLocationScore = avgScore
				stats.MostProductiveLocation = loc
			}
		}
	}

	// Calculate productivity by location type
	var maxLocationTypeScore float64
	for locType, counter := range locationTypeCounter {
		if counter.Count > 0 {
			// Average productivity score for this location type
			avgScore := counter.TotalScore / float64(counter.Count)
			stats.ProductivityByLocationType[locType] = avgScore

			if avgScore > maxLocationTypeScore {
				maxLocationTypeScore = avgScore
				stats.MostProductiveLocationType = locType
			}
		}
	}

	return stats, nil
}

func (r *mongoFocusSessionRepository) GetProductivityTrends(ctx context.Context, userID primitive.ObjectID, period entity.Period) (*entity.ProductivityTrends, error) {
	// Determine date range based on period
	endDate := time.Now()
	var startDate time.Time

	switch period {
	case entity.Daily:
		// Last 30 days
		startDate = endDate.AddDate(0, 0, -30)
	case entity.Weekly:
		// Last 12 weeks
		startDate = endDate.AddDate(0, 0, -12*7)
	case entity.Monthly:
		// Last 12 months
		startDate = endDate.AddDate(-1, 0, 0)
	default:
		// Default to weekly
		startDate = endDate.AddDate(0, 0, -12*7)
		period = entity.Weekly
	}

	// Get all sessions in date range
	sessions, err := r.GetSessionsByDateRange(ctx, userID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// Group sessions by period
	type PeriodData struct {
		StartDate         time.Time
		Sessions          []*entity.FocusSession
		TotalDuration     int
		TotalRating       int
		RatingCount       int
		TotalFocus        int
		FocusCount        int
		TotalEnergy       int
		EnergyCount       int
		TotalMood         int
		MoodCount         int
		TotalProductivity float64
		ProductivityCount int
	}

	// Create a map to group sessions
	periodMap := make(map[string]*PeriodData)

	// Define format strings for each period type
	var formatStr string
	var truncateFunc func(time.Time) time.Time

	switch period {
	case entity.Daily:
		formatStr = "2006-01-02"
		truncateFunc = func(t time.Time) time.Time {
			return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
		}
	case entity.Weekly:
		formatStr = "2006-W%02d" // ISO week format
		truncateFunc = func(t time.Time) time.Time {
			year, _ := t.ISOWeek()
			// Start of ISO week (Monday)
			// Go to the Monday of this week
			daysToSubtract := int(t.Weekday())
			if daysToSubtract == 0 { // Sunday
				daysToSubtract = 6
			} else {
				daysToSubtract--
			}
			return time.Date(year, t.Month(), t.Day()-daysToSubtract, 0, 0, 0, 0, t.Location())
		}
	case entity.Monthly:
		formatStr = "2006-01"
		truncateFunc = func(t time.Time) time.Time {
			return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
		}
	}

	// Create entries for all periods in range
	currentDate := truncateFunc(startDate)
	endTruncated := truncateFunc(endDate)

	for currentDate.Before(endTruncated) || currentDate.Equal(endTruncated) {
		var periodKey string

		if period == entity.Weekly {
			year, week := currentDate.ISOWeek()
			periodKey = fmt.Sprintf(formatStr, year, week)
		} else {
			periodKey = currentDate.Format(formatStr)
		}

		periodMap[periodKey] = &PeriodData{
			StartDate: currentDate,
			Sessions:  make([]*entity.FocusSession, 0),
		}

		// Advance to next period
		switch period {
		case entity.Daily:
			currentDate = currentDate.AddDate(0, 0, 1)
		case entity.Weekly:
			currentDate = currentDate.AddDate(0, 0, 7)
		case "monthly":
			currentDate = currentDate.AddDate(0, 1, 0)
		}
	}

	// Group sessions into periods
	for _, session := range sessions {
		if session.Status != entity.StatusCompleted || session.ActualDuration == nil {
			continue
		}

		// Get period key
		var periodKey string
		periodStart := truncateFunc(session.StartTime)

		if period == entity.Weekly {
			year, week := periodStart.ISOWeek()
			periodKey = fmt.Sprintf(formatStr, year, week)
		} else {
			periodKey = periodStart.Format(formatStr)
		}

		// Add session to period
		if data, exists := periodMap[periodKey]; exists {
			data.Sessions = append(data.Sessions, session)

			// Accumulate metrics
			data.TotalDuration += *session.ActualDuration

			if session.Rating != nil {
				data.TotalRating += *session.Rating
				data.RatingCount++
			}

			if session.Focus != nil {
				data.TotalFocus += *session.Focus
				data.FocusCount++
			}

			if session.Energy != nil {
				data.TotalEnergy += *session.Energy
				data.EnergyCount++
			}

			if session.Mood != nil {
				data.TotalMood += *session.Mood
				data.MoodCount++
			}

			// Calculate productivity score
			prodScore := session.CalculateProductivityScore()
			if prodScore > 0 {
				data.TotalProductivity += prodScore
				data.ProductivityCount++
			}
		}
	}

	// Sort periods by date
	periodKeys := make([]string, 0, len(periodMap))
	for key := range periodMap {
		periodKeys = append(periodKeys, key)
	}

	// Sort keys
	sort.Slice(periodKeys, func(i, j int) bool {
		return periodMap[periodKeys[i]].StartDate.Before(periodMap[periodKeys[j]].StartDate)
	})

	// Build result
	trends := &entity.ProductivityTrends{
		Period:       period,
		Dates:        make([]string, 0, len(periodKeys)),
		Durations:    make([]int, 0, len(periodKeys)),
		Ratings:      make([]float64, 0, len(periodKeys)),
		Focus:        make([]float64, 0, len(periodKeys)),
		Energy:       make([]float64, 0, len(periodKeys)),
		Mood:         make([]float64, 0, len(periodKeys)),
		Productivity: make([]float64, 0, len(periodKeys)),
	}

	for _, key := range periodKeys {
		data := periodMap[key]

		// Format date display
		var displayDate string
		switch period {
		case entity.Daily:
			displayDate = data.StartDate.Format("Jan 02")
		case entity.Weekly:
			// Format as "Jan 02-08" (start and end of week)
			weekEnd := data.StartDate.AddDate(0, 0, 6)
			if data.StartDate.Month() == weekEnd.Month() {
				displayDate = fmt.Sprintf("%s %d-%d", data.StartDate.Format("Jan"), data.StartDate.Day(), weekEnd.Day())
			} else {
				displayDate = fmt.Sprintf("%s-%s", data.StartDate.Format("Jan 02"), weekEnd.Format("Jan 02"))
			}
		case entity.Monthly:
			displayDate = data.StartDate.Format("Jan 2006")
		}

		trends.Dates = append(trends.Dates, displayDate)

		// Calculate averages
		if len(data.Sessions) > 0 {
			trends.Durations = append(trends.Durations, data.TotalDuration)
		} else {
			trends.Durations = append(trends.Durations, 0)
		}

		if data.RatingCount > 0 {
			trends.Ratings = append(trends.Ratings, float64(data.TotalRating)/float64(data.RatingCount))
		} else {
			trends.Ratings = append(trends.Ratings, 0)
		}

		if data.FocusCount > 0 {
			trends.Focus = append(trends.Focus, float64(data.TotalFocus)/float64(data.FocusCount))
		} else {
			trends.Focus = append(trends.Focus, 0)
		}

		if data.EnergyCount > 0 {
			trends.Energy = append(trends.Energy, float64(data.TotalEnergy)/float64(data.EnergyCount))
		} else {
			trends.Energy = append(trends.Energy, 0)
		}

		if data.MoodCount > 0 {
			trends.Mood = append(trends.Mood, float64(data.TotalMood)/float64(data.MoodCount))
		} else {
			trends.Mood = append(trends.Mood, 0)
		}

		if data.ProductivityCount > 0 {
			trends.Productivity = append(trends.Productivity, data.TotalProductivity/float64(data.ProductivityCount))
		} else {
			trends.Productivity = append(trends.Productivity, 0)
		}
	}

	return trends, nil
}
