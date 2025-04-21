package interfaces

import (
	"context"
	"focusspot/userservice/domain/entity"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type IUserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	GetByID(ctx context.Context, id primitive.ObjectID) (*entity.User, error)
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	GetByUsername(ctx context.Context, username string) (*entity.User, error)
	Update(ctx context.Context, user *entity.User) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	UpdateLastLogin(ctx context.Context, id primitive.ObjectID)
	UpdatePreferences(ctx context.Context, id primitive.ObjectID, preferences entity.UserPreferences) error
}
