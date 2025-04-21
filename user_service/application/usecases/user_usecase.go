package usecase

import (
	"context"
	"errors"
	"focusspot/userservice/application/dto"
	"focusspot/userservice/domain/entity"
	"focusspot/userservice/domain/interfaces"
	"focusspot/userservice/utils/hash"
	"focusspot/userservice/utils/token"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type IUserUseCase interface {
	Register(ctx context.Context, req dto.CreateUserRequest) (*dto.UserResponse, error)
	Login(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error)
	GetUserByID(ctx context.Context, id string) (*dto.UserResponse, error)
	UpdateUser(ctx context.Context, id string, req dto.UpdateUserRequest) (*dto.UserResponse, error)
	UpdatePreferences(ctx context.Context, id string, req dto.UserPreferencesRequest) (*dto.UserResponse, error)
}

type userUseCase struct {
	userRepo   interfaces.IUserRepository
	tokenMaker token.Maker
}

func NewUserUseCase(userRepo interfaces.IUserRepository, tokenMaker token.Maker) IUserUseCase {
	return &userUseCase{
		userRepo:   userRepo,
		tokenMaker: tokenMaker,
	}
}

func (uc *userUseCase) Register(ctx context.Context, req dto.CreateUserRequest) (*dto.UserResponse, error) {
	// Check if email or username already exists
	existingUserEmail, _ := uc.userRepo.GetByEmail(ctx, req.Email)
	if existingUserEmail != nil {
		return nil, errors.New("email already exists")
	}

	existingUserName, _ := uc.userRepo.GetByUsername(ctx, req.Username)
	if existingUserName != nil {
		return nil, errors.New("username already exists")
	}

	// Hash the password
	hashedPassword, err := hash.GenerateHash(req.Password)
	if err != nil {
		return nil, err
	}

	now := time.Now()

	// Create User entity
	user := &entity.User{
		Email:          req.Email,
		Username:       req.Username,
		HashedPassword: hashedPassword,
		FullName:       req.FullName,
		CreatedAt:      now,
		UpdatedAt:      now,
		Preferences: entity.UserPreferences{
			ThemeMode:            req.Preferences.ThemeMode,
			FocusSessionDuration: req.Preferences.FocusSessionDuration,
			PreferredLocations:   req.Preferences.PreferredLocations,
			NotificationsEnabled: req.Preferences.NotificationsEnabled,
		},
		Active: true,
	}

	// Save user to repository
	err = uc.userRepo.Create(ctx, user)
	if err != nil {
		return nil, err
	}

	// Convert to response DTO
	response := dto.ToUserResponse(user)
	return &response, nil
}

func (uc *userUseCase) Login(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error) {
	// Get user by email
	user, err := uc.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	// Verify password
	if !hash.VerifyHash(req.Password, user.HashedPassword) {
		return nil, errors.New("invalid email or password")
	}

	// Update last login time
	uc.userRepo.UpdateLastLogin(ctx, user.ID)

	// Generate tokens
	accessToken, err := uc.tokenMaker.CreateToken(user.ID.Hex(), user.Email, 24*time.Hour)
	if err != nil {
		return nil, err
	}

	refreshToken, err := uc.tokenMaker.CreateToken(user.ID.Hex(), user.Email, 7*24*time.Hour)
	if err != nil {
		return nil, err
	}

	userResponse := dto.ToUserResponse(user)
	return &dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         userResponse,
	}, nil
}

func (uc *userUseCase) GetUserByID(ctx context.Context, id string) (*dto.UserResponse, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}

	user, err := uc.userRepo.GetByID(ctx, objectID)
	if err != nil {
		return nil, err
	}

	response := dto.ToUserResponse(user)
	return &response, nil
}

func (uc *userUseCase) UpdateUser(ctx context.Context, id string, req dto.UpdateUserRequest) (*dto.UserResponse, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid user id")
	}

	// Get user by id
	user, err := uc.userRepo.GetByID(ctx, objectID)
	if err != nil {
		return nil, err
	}

	if req.FullName != "" {
		user.FullName = req.FullName
	}

	if req.Preferences != nil {
		user.Preferences = entity.UserPreferences{
			ThemeMode:            req.Preferences.ThemeMode,
			FocusSessionDuration: req.Preferences.FocusSessionDuration,
			PreferredLocations:   req.Preferences.PreferredLocations,
			NotificationsEnabled: req.Preferences.NotificationsEnabled,
		}
	}

	user.UpdatedAt = time.Now()
	err = uc.userRepo.Update(ctx, user)
	if err != nil {
		return nil, err
	}

	response := dto.ToUserResponse(user)
	return &response, nil
}

func (uc *userUseCase) UpdatePreferences(ctx context.Context, id string, req dto.UserPreferencesRequest) (*dto.UserResponse, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}

	user, err := uc.userRepo.GetByID(ctx, objectID)
	if err != nil {
		return nil, err
	}

	preferences := entity.UserPreferences{
		ThemeMode:            req.ThemeMode,
		FocusSessionDuration: req.FocusSessionDuration,
		PreferredLocations:   req.PreferredLocations,
		NotificationsEnabled: req.NotificationsEnabled,
	}

	err = uc.userRepo.UpdatePreferences(ctx, objectID, preferences)
	if err != nil {
		return nil, err
	}

	user.Preferences = preferences
	response := dto.ToUserResponse(user)
	return &response, nil
}
