package dto

import (
	"focusspot/userservice/domain/entity"
	"time"
)

type UserResponse struct {
	ID          string                  `json:"id"`
	Email       string                  `json:"email"`
	Username    string                  `json:"username"`
	FullName    string                  `json:"fullName"`
	CreatedAt   time.Time               `json:"createdAt"`
	LastLogin   *time.Time              `json:"lastLogin,omitempty"`
	Preferences UserPreferencesResponse `json:"preferences"`
	Active      bool                    `json:"active"`
}

type UserPreferencesResponse struct {
	ThemeMode            string   `json:"themeMode"`
	FocusSessionDuration int      `json:"focusSessionDuration"`
	PreferredLocations   []string `json:"preferredLocations"`
	NotificationsEnabled bool     `json:"notificationsEnabled"`
}

type LoginResponse struct {
	AccessToken  string       `json:"accessToken"`
	RefreshToken string       `json:"refreshToken"`
	User         UserResponse `json:"user"`
}

func ToUserResponse(user *entity.User) UserResponse {
	return UserResponse{
		ID:        user.ID.Hex(),
		Email:     user.Email,
		Username:  user.Username,
		FullName:  user.FullName,
		CreatedAt: user.CreatedAt,
		LastLogin: user.LastLogin,
		Preferences: UserPreferencesResponse{
			ThemeMode:            user.Preferences.ThemeMode,
			FocusSessionDuration: user.Preferences.FocusSessionDuration,
			PreferredLocations:   user.Preferences.PreferredLocations,
			NotificationsEnabled: user.Preferences.NotificationsEnabled,
		},
		Active: user.Active,
	}
}

