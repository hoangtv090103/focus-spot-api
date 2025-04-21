package entity

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID             primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Email          string             `json:"email" bson:"email"`
	Username       string             `json:"username" bson:"username"`
	HashedPassword string             `json:"-" bson:"password"`
	FullName       string             `json:"fullname" bson:"fullname"`
	CreatedAt      time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedAt      time.Time          `json:"updatedAt" bson:"updatedAt"`
	LastLogin      *time.Time         `json:"lastLogin,omitempty" bson:"lastLogin,omitempty"`
	Preferences    UserPreferences    `json:"preferences" bson:"preferences"`
	Active         bool               `json:"active" bson:"active"`
}

type UserPreferences struct {
	ThemeMode            string   `json:"themeMode" bson:"themeMode"`
	FocusSessionDuration int      `json:"focusSessionDuration" bson:"focusSessionDuration"`
	PreferredLocations   []string `json:"preferredLocations" bson:"preferredLocations"`
	NotificationsEnabled bool     `json:"notificationsEnabled" bson:"notificationsEnabled"`
}
