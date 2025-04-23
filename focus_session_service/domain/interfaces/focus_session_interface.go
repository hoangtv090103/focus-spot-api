package interfaces

import (
	"context"
	"focusspot/focussessionservice/domain/entity"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type IFocusSessionRepository interface {
	Create(ctx context.Context, session *entity.FocusSession) error
	GetByID(ctx context.Context, id primitive.ObjectID) (*entity.FocusSession, error)
	GetByUserID(ctx context.Context, userID primitive.ObjectID, limit, offset int) ([]*entity.FocusSession, error)
	GetActiveByUserID(ctx context.Context, userID primitive.ObjectID) (*entity.FocusSession, error)
	GetSessionsByDateRange(ctx context.Context, userID primitive.ObjectID, startDate, endDate time.Time) ([]*entity.FocusSession, error)
	Update(ctx context.Context, sesion *entity.FocusSession) error
	UpdateStatus(ctx context.Context, id primitive.ObjectID, status entity.SessionStatus) error
	EndSession(ctx context.Context, id primitive.ObjectID, endTime time.Time, notes string, rating, focus, energy, mood, distractions *int) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	GetProductivityStats(ctx context.Context, userID primitive.ObjectID, startDate, endDate time.Time) (*entity.ProductivityStats, error)
	GetProductivityTrends(ctx context.Context, userID primitive.ObjectID, period string) (*entity.ProductivityTrends, error)
}
