package usecase

import (
	"context"
	"errors"
	"focusspot/focussessionservice/application/dto"
	"focusspot/focussessionservice/domain/entity"
	"focusspot/focussessionservice/domain/interfaces"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	ErrInvalidUserID              = errors.New("invalid user ID")
	ErrInvalidSessionID           = errors.New("invalid session ID")
	ErrNoSessionFound             = errors.New("no session found")
	ErrNoActiveSessionFound       = errors.New("no active session found")
	ErrNoSessionFoundAccessDenied = errors.New("no session found or access denied")
	ErrInvalidDateRange           = errors.New("invalid date range")
	ErrInvalidDuration            = errors.New("invalid duration")
	ErrAlreadyHaveActiveSession   = errors.New("you already have an active session")
)

type IFocusSessionUseCase interface {
	CreateSession(ctx context.Context, userID string, req dto.CreateSessionRequest) (*dto.FocusSessionResponse, error)
	GetSessionByID(ctx context.Context, id string, userID string) (*dto.FocusSessionResponse, error)
	GetUserSessions(ctx context.Context, userID string, req dto.GetSessionsRequest) (*dto.SessionsListResponse, error)
	GetActiveSession(ctx context.Context, userID string) (*dto.FocusSessionResponse, error)
	UpdateSession(ctx context.Context, id string, userID string, req dto.UpdateSessionRequest) (*dto.FocusSessionResponse, error)
	StartSession(ctx context.Context, id string, userID string) (*dto.FocusSessionResponse, error)
	EndSession(ctx context.Context, id string, userID string, req dto.EndSessionRequest) (*dto.FocusSessionResponse, error)
	CancelSession(ctx context.Context, id string, userID string) (*dto.FocusSessionResponse, error)
	DeleteSession(ctx context.Context, id string, userID string) error
	GetProductivityStats(ctx context.Context, userID string, req dto.GetProductivityStatsRequest) (*dto.ProductivityStatsResponse, error)
	GetProductivityTrends(ctx context.Context, userID string, req dto.GetProductivityTrendsRequest) (*dto.ProductivityTrendsResponse, error)
}

type focusSessionUseCase struct {
	sessionRepo interfaces.IFocusSessionRepository
}

func NewFocusSessionUseCase(sessionRepo interfaces.IFocusSessionRepository) IFocusSessionUseCase {
	return &focusSessionUseCase{
		sessionRepo: sessionRepo,
	}
}

func (uc *focusSessionUseCase) CreateSession(ctx context.Context, userID string, req dto.CreateSessionRequest) (*dto.FocusSessionResponse, error) {
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}

	var locationObjID *primitive.ObjectID
	if req.LocationID != "" {
		id, err := primitive.ObjectIDFromHex(req.LocationID)
		if err != nil {
			return nil, errors.New("invalid location ID")
		}

		locationObjID = &id
	}

	var locationDetails *entity.LocationDetails
	if req.LocationDetails != nil {
		locationDetails = &entity.LocationDetails{
			Name:      req.LocationDetails.Name,
			Address:   req.LocationDetails.Address,
			Latitude:  req.LocationDetails.Latitude,
			Longitude: req.LocationDetails.Longitude,
			Type:      req.LocationDetails.Type,
		}
	}

	now := time.Now()
	session := &entity.FocusSession{
		ID:              primitive.NewObjectID(),
		UserID:          userObjID,
		Title:           req.Title,
		Description:     req.Description,
		StartTime:       req.StartTime,
		Duration:        req.Duration,
		Status:          entity.StatusPlanned,
		LocationID:      locationObjID,
		LocationDetails: locationDetails,
		Tags:            req.Tags,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	// Check if start time is in the future
	if req.StartTime.Before(now) && req.StartTime.Add(time.Duration(req.Duration)*time.Minute).After(now) {
		// Session is happening now, set it to active
		session.Status = entity.StatusActive
	}

	if err := uc.sessionRepo.Create(ctx, session); err != nil {
		return nil, err
	}
	response := dto.ToFocusSessionResponse(session)
	return &response, nil

}

func (uc *focusSessionUseCase) GetSessionByID(ctx context.Context, id string, userID string) (*dto.FocusSessionResponse, error) {
	sessionID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrInvalidSessionID
	}

	session, err := uc.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	// Verify ownership
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil || session.UserID != userObjID {
		return nil, ErrNoSessionFoundAccessDenied
	}

	response := dto.ToFocusSessionResponse(session)
	return &response, nil
}

func (uc *focusSessionUseCase) GetUserSessions(ctx context.Context, userID string, req dto.GetSessionsRequest) (*dto.SessionsListResponse, error) {
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, ErrInvalidUserID
	}

	var sessions []*entity.FocusSession

	// Check if date range is provided in the request
	if req.StartDate != "" && req.EndDate != "" {
		startDate, err := time.Parse(time.RFC3339, req.StartDate)
		if err != nil {
			return nil, ErrInvalidDateRange
		}
		endDate, err := time.Parse(time.RFC3339, req.EndDate)
		if err != nil {
			return nil, ErrInvalidDateRange
		}

		// Make endDate inclusive by setting it to the end of the day
		endDate = endDate.Add(24 * time.Hour).Add(-1 * time.Second)

		sessions, err = uc.sessionRepo.GetSessionsByDateRange(ctx, userObjID, startDate, endDate)
		if err != nil {
			return nil, err
		}
	} else {
		// Otherwise, get paginated sessions with GetByUserID using limit and offset
		sessions, err = uc.sessionRepo.GetByUserID(ctx, userObjID, req.Limit, req.Offset)
		if err != nil {
			return nil, err
		}
	}

	// Convert each session entity to SessionResponse
	var sessionResponseList []dto.FocusSessionResponse
	for _, session := range sessions {
		sessionResponse := dto.ToFocusSessionResponse(session)
		sessionResponseList = append(sessionResponseList, sessionResponse)
	}

	// Populate the SessionsListResponse with the converted sessions and metadata
	response := dto.SessionsListResponse{
		Sessions: sessionResponseList,
		Total:    len(sessionResponseList),
		Limit:    req.Limit,
		Offset:   req.Offset,
	}
	return &response, nil
}

func (uc *focusSessionUseCase) GetActiveByUserID(ctx context.Context, userID string) (*dto.FocusSessionResponse, error) {
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, ErrInvalidUserID
	}

	session, err := uc.sessionRepo.GetActiveByUserID(ctx, userObjID)
	if err != nil {
		return nil, ErrNoActiveSessionFound
	}

	response := dto.ToFocusSessionResponse(session)

	return &response, nil
}

func (uc *focusSessionUseCase) GetActiveSession(ctx context.Context, userID string) (*dto.FocusSessionResponse, error) {
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}

	session, err := uc.sessionRepo.GetActiveByUserID(ctx, userObjID)
	if err != nil {
		return nil, err
	}

	if session == nil {
		return nil, errors.New("no active session found")
	}
	response := dto.ToFocusSessionResponse(session)
	return &response, nil
}

func (uc *focusSessionUseCase) UpdateSession(
	ctx context.Context,
	id string,
	userID string,
	req dto.UpdateSessionRequest,
) (*dto.FocusSessionResponse, error) {
	sessionID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrInvalidSessionID
	}

	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, ErrInvalidUserID
	}

	session, err := uc.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	// Verify ownership
	if session.UserID != userObjID {
		return nil, ErrNoSessionFoundAccessDenied
	}

	// Update only provided fields
	if req.Title != "" {
		session.Title = req.Title
	}

	if req.Description != "" {
		session.Description = req.Description
	}

	if req.StartTime != nil {
		session.StartTime = *req.StartTime
	}

	if req.Duration != nil {
		session.Duration = *req.Duration
	}

	if req.LocationID != "" {
		locID, err := primitive.ObjectIDFromHex(req.LocationID)
		if err != nil {
			return nil, errors.New("invalid location ID")
		}
		session.LocationID = &locID
	}

	if req.LocationDetails != nil {
		session.LocationDetails = &entity.LocationDetails{
			Name:      req.LocationDetails.Name,
			Address:   req.LocationDetails.Address,
			Latitude:  req.LocationDetails.Latitude,
			Longitude: req.LocationDetails.Longitude,
			Type:      req.LocationDetails.Type,
		}
	}

	if req.Tags != nil {
		session.Tags = req.Tags
	}

	if req.Status != "" {
		session.Status = entity.SessionStatus(req.Status)
	}

	session.UpdatedAt = time.Now()

	if err := uc.sessionRepo.Update(ctx, session); err != nil {
		return nil, err
	}

	response := dto.ToFocusSessionResponse(session)
	return &response, nil
}

func (uc *focusSessionUseCase) StartSession(ctx context.Context, id string, userID string) (
	*dto.FocusSessionResponse,
	error,
) {
	sessionID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrInvalidSessionID
	}

	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, ErrInvalidUserID
	}

	// Check if user already has an active session
	activeSession, _ := uc.sessionRepo.GetActiveByUserID(ctx, userObjID)
	if activeSession != nil {
		return nil, ErrAlreadyHaveActiveSession
	}

	session, err := uc.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	// Verify Ownership
	if session.UserID != userObjID {
		return nil, ErrNoSessionFoundAccessDenied
	}

	err = uc.sessionRepo.UpdateStatus(ctx, sessionID, entity.StatusActive)
	if err != nil {
		return nil, err
	}

	session.Status = entity.StatusActive
	session.UpdatedAt = time.Now()
	session.StartTime = time.Now()

	response := dto.ToFocusSessionResponse(session)
	return &response, nil
}

func (uc *focusSessionUseCase) EndSession(
	ctx context.Context,
	id string,
	userID string,
	req dto.EndSessionRequest,
) (*dto.FocusSessionResponse, error) {
	sessionID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrInvalidSessionID
	}

	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, ErrInvalidUserID
	}

	session, err := uc.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	// Verify ownership
	if session.UserID != userObjID {
		return nil, ErrNoSessionFoundAccessDenied
	}

	// Only active session can be ended
	if session.Status != entity.StatusActive {
		return nil, errors.New("only active session can be ended")
	}

	endTime := time.Now()
	err = uc.sessionRepo.EndSession(
		ctx,
		sessionID,
		endTime,
		req.Notes,
		req.Rating,
		req.Focus,
		req.Energy,
		req.Mood,
		req.Distractions,
	)
	if err != nil {
		return nil, err
	}

	session.Status = entity.StatusCompleted
	session.UpdatedAt = time.Now()
	session.EndTime = &endTime
	session.Notes = req.Notes
	session.Rating = req.Rating
	session.Focus = req.Focus
	session.Energy = req.Energy
	session.Mood = req.Mood
	session.Distractions = req.Distractions

	// Calculate actual duration in minutes
	duration := int(endTime.Sub(session.StartTime).Minutes())
	session.ActualDuration = &duration

	response := dto.ToFocusSessionResponse(session)
	return &response, nil
}

func (uc *focusSessionUseCase) CancelSession(
	ctx context.Context,
	id string,
	userID string,
) (*dto.FocusSessionResponse, error) {
	sessionID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrInvalidSessionID
	}

	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, ErrInvalidUserID
	}

	session, err := uc.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	// Verify ownership
	if session.UserID != userObjID {
		return nil, ErrNoSessionFoundAccessDenied
	}

	// Only active session can be canceled
	if session.Status != entity.StatusCancelled {
		return nil, errors.New("only active sessions can be cancelled")
	}

	err = uc.sessionRepo.UpdateStatus(
		ctx,
		sessionID,
		entity.StatusCancelled,
	)

	if err != nil {
		return nil, err
	}

	session.Status = entity.StatusCancelled
	session.UpdatedAt = time.Now()

	response := dto.ToFocusSessionResponse(session)

	return &response, nil
}

func (uc *focusSessionUseCase) DeleteSession(ctx context.Context, id string, userID string) error {
	sessionID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return ErrInvalidSessionID
	}

	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return ErrInvalidUserID
	}

	session, err := uc.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return err
	}

	// Verify ownership
	if session.UserID != userObjID {
		return ErrNoSessionFoundAccessDenied
	}

	return uc.sessionRepo.Delete(ctx, sessionID)
}

func (uc *focusSessionUseCase) GetProductivityStats(
	ctx context.Context,
	userID string,
	req dto.GetProductivityStatsRequest,
) (*dto.ProductivityStatsResponse, error) {
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, ErrInvalidUserID
	}

	var startDate, endDate time.Time

	// Default to last 30 days if not specified
	if req.StartDate == "" || req.EndDate == "" {
		endDate = time.Now()
		startDate = endDate.AddDate(0, 0, -30)
	} else {
		startDate, err = time.Parse(time.RFC3339, req.StartDate)
		if err != nil {
			return nil, errors.New("invalid start date format (use YYYY-MM-DD)")
		}

		endDate, err = time.Parse(time.RFC3339, req.EndDate)
		if err != nil {
			return nil, errors.New("invalid end date format (use YYYY-MM-DD)")
		}

		// Make endDate inclusive by setting it to the end of the day
		endDate = endDate.Add(24 * time.Hour).Add(-1 * time.Second)
	}

	stats, err := uc.sessionRepo.GetProductivityStats(ctx, userObjID, startDate, endDate)

	if err != nil {
		return nil, err
	}

	dateRange := dto.DateRange{
		StartDate: startDate,
		EndDate:   endDate,
	}

	response := dto.ToProductivityStatsResponse(stats, dateRange)
	return &response, nil
}

func (uc *focusSessionUseCase) GetProductivityTrends(ctx context.Context, userID string, req dto.GetProductivityTrendsRequest) (*dto.ProductivityTrendsResponse, error) {
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}

	if req.Period == "" {
		req.Period = entity.Weekly
	}

	if req.Limit <= 0 {
		req.Limit = 12
	}

	trends, err := uc.sessionRepo.GetProductivityTrends(ctx, userObjID, req.Period)
	if err != nil {
		return nil, err
	}

	return &dto.ProductivityTrendsResponse{
		Period:       trends.Period,
		Dates:        trends.Dates,
		Durations:    trends.Durations,
		Ratings:      trends.Ratings,
		Focus:        trends.Focus,
		Energy:       trends.Energy,
		Mood:         trends.Mood,
		Productivity: trends.Productivity,
	}, nil
}
