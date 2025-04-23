package dto

type CreateUserRequest struct {
	Email       string                 `json:"email" validate:"required,email"`
	Username    string                 `json:"username" validate:"required,min=3,max=30"`
	Password    string                 `json:"password" validate:"required,min=8"`
	FullName    string                 `json:"fullname" validate:"required"`
	Preferences UserPreferencesRequest `json:"preferences"`
}

type UpdateUserRequest struct {
	FullName    string                  `json:"fullname,omitempty"`
	Preferences *UserPreferencesRequest `json:"preferences,omitempty"`
}

type UserPreferencesRequest struct {
	ThemeMode            string   `json:"themeMode"`
	FocusSessionDuration int      `json:"focusSessionDuration"`
	PreferredLocations   []string `json:"preferredLocations"`
	NotificationsEnabled bool     `json:"notificationsEnabled"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}
