package mongodb

import (
	"context"
	"errors"
	"focusspot/userservice/domain/entity"
	"focusspot/userservice/domain/interfaces"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type mongoUserRepository struct {
	collection *mongo.Collection
}

func NewMongoUserRepository(db *mongo.Database) interfaces.IUserRepository {
	collection := db.Collection("users")

	// Create indexes for email and username (unique)
	_, err := collection.Indexes().CreateMany(
		context.Background(),
		[]mongo.IndexModel{
			{
				Keys:    bson.D{{Key: "email", Value: 1}},
				Options: options.Index().SetUnique(true),
			},
			{
				Keys:    bson.D{{Key: "username", Value: 1}},
				Options: options.Index().SetUnique(true),
			},
		})
	if err != nil {
		//TODO: In production, handle this error properly
		panic(err)
	}

	return &mongoUserRepository{
		collection: collection,
	}
}

func (r *mongoUserRepository) Create(ctx context.Context, user *entity.User) error {
	if user.ID.IsZero() {
		user.ID = primitive.NewObjectID()
	}

	_, err := r.collection.InsertOne(ctx, user)
	return err
}

func (r *mongoUserRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*entity.User, error) {
	var user entity.User

	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&user)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return &user, err
}

func (r *mongoUserRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	var user entity.User

	err := r.collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return &user, nil
}

func (r *mongoUserRepository) GetByUsername(ctx context.Context, username string) (*entity.User, error) {
	var user entity.User

	err := r.collection.FindOne(ctx, bson.M{"username": username}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return &user, nil
}

func (r *mongoUserRepository) Update(ctx context.Context, user *entity.User) error {
	user.UpdatedAt = time.Now()

	_, err := r.collection.ReplaceOne(
		ctx,
		bson.M{"_id": user.ID},
		user,
	)

	return err
}

func (r *mongoUserRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	now := time.Now()

	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{
			"$set": bson.M{
				"deletedAt": now,
				"updatedAt": now,
				"active":    false,
			},
		},
	)

	return err
}

func (r *mongoUserRepository) UpdateLastLogin(ctx context.Context, id primitive.ObjectID) error {
	now := time.Now()

	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{
			"$set": bson.M{
				"lastLogin": now,
				"updatedAt": now,
			},
		},
	)

	return err
}

func (r *mongoUserRepository) UpdatePreferences(ctx context.Context, id primitive.ObjectID, preferences entity.UserPreferences) error {
	now := time.Now()

	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{
			"$set": bson.M{
				"preferences": preferences,
				"updatedAt":   now,
			},
		},
	)

	return err
}
