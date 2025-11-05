package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// User represents a registered user in the system
type User struct {
	bun.BaseModel `bun:"table:users,alias:u"`

	ID                uuid.UUID `bun:"id,pk,type:uuid,default:gen_random_uuid()" json:"id"`
	Email             string    `bun:"email,unique,notnull" json:"email"`
	EmailVerified     bool      `bun:"email_verified,default:false" json:"emailVerified"`
	PasswordHash      string    `bun:"password_hash,notnull" json:"-"` // Never expose password hash
	FullName          *string   `bun:"full_name" json:"fullName"`
	PreferredLanguage string    `bun:"preferred_language,default:'lt'" json:"preferredLanguage"`
	IsActive          bool      `bun:"is_active,default:true" json:"isActive"`

	// OAuth fields for future expansion
	OAuthProvider *string `bun:"oauth_provider" json:"oauthProvider"`
	OAuthID       *string `bun:"oauth_id" json:"oauthId"`
	AvatarURL     *string `bun:"avatar_url" json:"avatarUrl"`

	// Metadata as JSONB
	MetadataJSON json.RawMessage `bun:"metadata,type:jsonb,default:'{}'" json:"-"`
	Metadata     UserMetadata    `bun:"-" json:"metadata"`

	// Timestamps
	CreatedAt   time.Time  `bun:"created_at,default:now()" json:"createdAt"`
	UpdatedAt   time.Time  `bun:"updated_at,default:now()" json:"updatedAt"`
	LastLoginAt *time.Time `bun:"last_login_at" json:"lastLoginAt"`

	// Relations
	Sessions []*UserSession `bun:"rel:has-many,join:id=user_id" json:"sessions,omitempty"`
}

// UserMetadata represents flexible user metadata
type UserMetadata struct {
	Preferences   UserPreferences   `json:"preferences"`
	Profile       UserProfile       `json:"profile"`
	Notifications UserNotifications `json:"notifications"`
	Privacy       UserPrivacy       `json:"privacy"`
}

// UserPreferences represents user preferences
type UserPreferences struct {
	Theme              string   `json:"theme"` // 'light', 'dark', 'auto'
	Currency           string   `json:"currency"`
	DefaultStores      []int    `json:"defaultStores"`
	SearchFilters      []string `json:"searchFilters"`
	EmailNotifications bool     `json:"emailNotifications"`
	PushNotifications  bool     `json:"pushNotifications"`
}

// UserProfile represents user profile information
type UserProfile struct {
	Bio           string `json:"bio"`
	Location      string `json:"location"`
	Website       string `json:"website"`
	DateOfBirth   string `json:"dateOfBirth"`
	Gender        string `json:"gender"`
	PhoneNumber   string `json:"phoneNumber"`
	ProfilePublic bool   `json:"profilePublic"`
}

// UserNotifications represents notification settings
type UserNotifications struct {
	EmailEnabled      bool `json:"emailEnabled"`
	PushEnabled       bool `json:"pushEnabled"`
	WeeklyDigest      bool `json:"weeklyDigest"`
	PriceAlerts       bool `json:"priceAlerts"`
	NewFlyerAlerts    bool `json:"newFlyerAlerts"`
	ProductMatches    bool `json:"productMatches"`
	SecurityAlerts    bool `json:"securityAlerts"`
	MarketingEmails   bool `json:"marketingEmails"`
}

// UserPrivacy represents privacy settings
type UserPrivacy struct {
	ProfileVisibility string `json:"profileVisibility"` // 'public', 'private', 'friends'
	ShowEmail         bool   `json:"showEmail"`
	ShowFullName      bool   `json:"showFullName"`
	ShowLastLogin     bool   `json:"showLastLogin"`
	ShareUsageData    bool   `json:"shareUsageData"`
	AllowMarketing    bool   `json:"allowMarketing"`
}

// BeforeAppendModel implements bun.BeforeAppendModelHook
func (u *User) BeforeAppendModel(query bun.Query) error {
	switch query.(type) {
	case *bun.InsertQuery:
		u.CreatedAt = time.Now()
		u.UpdatedAt = time.Now()
	case *bun.UpdateQuery:
		u.UpdatedAt = time.Now()
	}

	// Marshal metadata if it has non-zero values
	if u.hasNonZeroMetadata() {
		data, err := json.Marshal(u.Metadata)
		if err != nil {
			return err
		}
		u.MetadataJSON = data
	}

	return nil
}

// AfterSelectModel implements bun.AfterSelectModelHook
func (u *User) AfterSelectModel() error {
	// Unmarshal metadata
	if len(u.MetadataJSON) > 0 {
		if err := json.Unmarshal(u.MetadataJSON, &u.Metadata); err != nil {
			// If unmarshal fails, use default metadata
			u.Metadata = DefaultUserMetadata()
		}
	} else {
		u.Metadata = DefaultUserMetadata()
	}

	return nil
}

// SetMetadata sets the user metadata and marshals it to JSON
func (u *User) SetMetadata(metadata UserMetadata) error {
	u.Metadata = metadata
	data, err := json.Marshal(metadata)
	if err != nil {
		return err
	}
	u.MetadataJSON = data
	return nil
}

// UpdateLastLogin updates the last login timestamp
func (u *User) UpdateLastLogin() {
	now := time.Now()
	u.LastLoginAt = &now
	u.UpdatedAt = now
}

// IsOAuthUser returns true if the user was created via OAuth
func (u *User) IsOAuthUser() bool {
	return u.OAuthProvider != nil && u.OAuthID != nil
}

// GetDisplayName returns the display name for the user
func (u *User) GetDisplayName() string {
	if u.FullName != nil && *u.FullName != "" {
		return *u.FullName
	}
	return u.Email
}

// hasNonZeroMetadata checks if metadata has any non-zero values
func (u *User) hasNonZeroMetadata() bool {
	return len(u.MetadataJSON) > 0 ||
		   u.Metadata.Preferences.Theme != "" ||
		   u.Metadata.Profile.Bio != "" ||
		   u.Metadata.Notifications.EmailEnabled != false ||
		   u.Metadata.Privacy.ProfileVisibility != ""
}

// DefaultUserMetadata returns default metadata for new users
func DefaultUserMetadata() UserMetadata {
	return UserMetadata{
		Preferences: UserPreferences{
			Theme:              "light",
			Currency:           "EUR",
			DefaultStores:      []int{},
			SearchFilters:      []string{},
			EmailNotifications: true,
			PushNotifications:  false,
		},
		Profile: UserProfile{
			ProfilePublic: false,
		},
		Notifications: UserNotifications{
			EmailEnabled:    true,
			PushEnabled:     false,
			WeeklyDigest:    true,
			PriceAlerts:     true,
			NewFlyerAlerts:  true,
			ProductMatches:  true,
			SecurityAlerts:  true,
			MarketingEmails: false,
		},
		Privacy: UserPrivacy{
			ProfileVisibility: "private",
			ShowEmail:         false,
			ShowFullName:      false,
			ShowLastLogin:     false,
			ShareUsageData:    false,
			AllowMarketing:    false,
		},
	}
}

// UserInput represents input for creating/updating users
type UserInput struct {
	Email             string        `json:"email" validate:"required,email"`
	Password          string        `json:"password" validate:"required,min=8"`
	FullName          *string       `json:"fullName" validate:"omitempty,min=2,max=255"`
	PreferredLanguage string        `json:"preferredLanguage" validate:"omitempty,oneof=lt en ru"`
	Metadata          *UserMetadata `json:"metadata"`
}

// UserUpdateInput represents input for updating users
type UserUpdateInput struct {
	FullName          *string       `json:"fullName" validate:"omitempty,min=2,max=255"`
	PreferredLanguage *string       `json:"preferredLanguage" validate:"omitempty,oneof=lt en ru"`
	Metadata          *UserMetadata `json:"metadata"`
}

// UserPasswordChangeInput represents input for changing passwords
type UserPasswordChangeInput struct {
	CurrentPassword string `json:"currentPassword" validate:"required"`
	NewPassword     string `json:"newPassword" validate:"required,min=8"`
	ConfirmPassword string `json:"confirmPassword" validate:"required,eqfield=NewPassword"`
}

// UserPasswordResetInput represents input for password reset
type UserPasswordResetInput struct {
	Email string `json:"email" validate:"required,email"`
}

// UserPasswordResetConfirmInput represents input for confirming password reset
type UserPasswordResetConfirmInput struct {
	Token           string `json:"token" validate:"required"`
	NewPassword     string `json:"newPassword" validate:"required,min=8"`
	ConfirmPassword string `json:"confirmPassword" validate:"required,eqfield=NewPassword"`
}