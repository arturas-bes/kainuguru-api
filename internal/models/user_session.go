package models

import (
	"encoding/json"
	"net"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// UserSession represents an active user session
type UserSession struct {
	bun.BaseModel `bun:"table:user_sessions,alias:us"`

	ID       uuid.UUID `bun:"id,pk,type:uuid,default:gen_random_uuid()" json:"id"`
	UserID   uuid.UUID `bun:"user_id,notnull" json:"userId"`
	IsActive bool      `bun:"is_active,default:true" json:"isActive"`

	// Token information
	TokenHash        string     `bun:"token_hash,unique,notnull" json:"-"` // Never expose token hashes
	ExpiresAt        time.Time  `bun:"expires_at,notnull" json:"expiresAt"`
	RefreshTokenHash *string    `bun:"refresh_token_hash" json:"-"`
	RefreshExpiresAt *time.Time `bun:"refresh_expires_at" json:"refreshExpiresAt"`

	// Request metadata
	IPAddress  *net.IP `bun:"ip_address,type:inet" json:"ipAddress"`
	UserAgent  *string `bun:"user_agent" json:"userAgent"`
	DeviceType string  `bun:"device_type,default:'web'" json:"deviceType"`

	// Browser and location info as JSONB
	BrowserInfoJSON  json.RawMessage `bun:"browser_info,type:jsonb,default:'{}'" json:"-"`
	LocationInfoJSON json.RawMessage `bun:"location_info,type:jsonb,default:'{}'" json:"-"`
	BrowserInfo      BrowserInfo     `bun:"-" json:"browserInfo"`
	LocationInfo     LocationInfo    `bun:"-" json:"locationInfo"`

	// Timestamps
	CreatedAt  time.Time `bun:"created_at,default:now()" json:"createdAt"`
	LastUsedAt time.Time `bun:"last_used_at,default:now()" json:"lastUsedAt"`

	// Relations
	User *User `bun:"rel:belongs-to,join:user_id=id" json:"user,omitempty"`
}

// BrowserInfo represents browser-specific information
type BrowserInfo struct {
	Name       string `json:"name"`
	Version    string `json:"version"`
	Platform   string `json:"platform"`
	Language   string `json:"language"`
	TimeZone   string `json:"timeZone"`
	ScreenSize string `json:"screenSize"`
	IsMobile   bool   `json:"isMobile"`
	IsTablet   bool   `json:"isTablet"`
	IsDesktop  bool   `json:"isDesktop"`
}

// LocationInfo represents location-specific information
type LocationInfo struct {
	Country   string  `json:"country"`
	Region    string  `json:"region"`
	City      string  `json:"city"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	ISP       string  `json:"isp"`
	Timezone  string  `json:"timezone"`
}

// BeforeAppendModel implements bun.BeforeAppendModelHook
func (us *UserSession) BeforeAppendModel(query bun.Query) error {
	switch query.(type) {
	case *bun.InsertQuery:
		us.CreatedAt = time.Now()
		us.LastUsedAt = time.Now()
	case *bun.UpdateQuery:
		us.LastUsedAt = time.Now()
	}

	// Marshal browser info if it has data
	if us.hasNonZeroBrowserInfo() {
		data, err := json.Marshal(us.BrowserInfo)
		if err != nil {
			return err
		}
		us.BrowserInfoJSON = data
	}

	// Marshal location info if it has data
	if us.hasNonZeroLocationInfo() {
		data, err := json.Marshal(us.LocationInfo)
		if err != nil {
			return err
		}
		us.LocationInfoJSON = data
	}

	return nil
}

// AfterSelectModel implements bun.AfterSelectModelHook
func (us *UserSession) AfterSelectModel() error {
	// Unmarshal browser info
	if len(us.BrowserInfoJSON) > 0 {
		if err := json.Unmarshal(us.BrowserInfoJSON, &us.BrowserInfo); err != nil {
			// If unmarshal fails, use default
			us.BrowserInfo = BrowserInfo{}
		}
	}

	// Unmarshal location info
	if len(us.LocationInfoJSON) > 0 {
		if err := json.Unmarshal(us.LocationInfoJSON, &us.LocationInfo); err != nil {
			// If unmarshal fails, use default
			us.LocationInfo = LocationInfo{}
		}
	}

	return nil
}

// IsExpired returns true if the session has expired
func (us *UserSession) IsExpired() bool {
	return time.Now().After(us.ExpiresAt)
}

// IsRefreshExpired returns true if the refresh token has expired
func (us *UserSession) IsRefreshExpired() bool {
	if us.RefreshExpiresAt == nil {
		return true
	}
	return time.Now().After(*us.RefreshExpiresAt)
}

// IsValid returns true if the session is active and not expired
func (us *UserSession) IsValid() bool {
	return us.IsActive && !us.IsExpired()
}

// CanRefresh returns true if the session can be refreshed
func (us *UserSession) CanRefresh() bool {
	return us.RefreshTokenHash != nil && !us.IsRefreshExpired()
}

// UpdateLastUsed updates the last used timestamp
func (us *UserSession) UpdateLastUsed() {
	us.LastUsedAt = time.Now()
}

// Expire marks the session as expired and inactive
func (us *UserSession) Expire() {
	us.IsActive = false
	us.ExpiresAt = time.Now()
}

// SetBrowserInfo sets the browser information and marshals it to JSON
func (us *UserSession) SetBrowserInfo(info BrowserInfo) error {
	us.BrowserInfo = info
	data, err := json.Marshal(info)
	if err != nil {
		return err
	}
	us.BrowserInfoJSON = data
	return nil
}

// SetLocationInfo sets the location information and marshals it to JSON
func (us *UserSession) SetLocationInfo(info LocationInfo) error {
	us.LocationInfo = info
	data, err := json.Marshal(info)
	if err != nil {
		return err
	}
	us.LocationInfoJSON = data
	return nil
}

// GetDeviceDescription returns a human-readable device description
func (us *UserSession) GetDeviceDescription() string {
	if us.BrowserInfo.Name != "" {
		if us.BrowserInfo.IsMobile {
			return us.BrowserInfo.Name + " Mobile"
		} else if us.BrowserInfo.IsTablet {
			return us.BrowserInfo.Name + " Tablet"
		} else {
			return us.BrowserInfo.Name + " Desktop"
		}
	}

	switch us.DeviceType {
	case "mobile":
		return "Mobile Device"
	case "tablet":
		return "Tablet"
	case "api":
		return "API Client"
	default:
		return "Web Browser"
	}
}

// hasNonZeroBrowserInfo checks if browser info has any non-zero values
func (us *UserSession) hasNonZeroBrowserInfo() bool {
	return us.BrowserInfo.Name != "" || us.BrowserInfo.Version != "" || us.BrowserInfo.Platform != ""
}

// hasNonZeroLocationInfo checks if location info has any non-zero values
func (us *UserSession) hasNonZeroLocationInfo() bool {
	return us.LocationInfo.Country != "" || us.LocationInfo.City != "" || us.LocationInfo.ISP != ""
}

// GetLocationDescription returns a human-readable location description
func (us *UserSession) GetLocationDescription() string {
	if us.LocationInfo.City != "" && us.LocationInfo.Country != "" {
		return us.LocationInfo.City + ", " + us.LocationInfo.Country
	} else if us.LocationInfo.Country != "" {
		return us.LocationInfo.Country
	} else if us.IPAddress != nil {
		return us.IPAddress.String()
	}
	return "Unknown Location"
}

// SessionCreateInput represents input for creating a new session
type SessionCreateInput struct {
	UserID           uuid.UUID     `json:"userId" validate:"required"`
	TokenHash        string        `json:"-"` // Set internally, not from input
	ExpiresAt        time.Time     `json:"expiresAt" validate:"required"`
	RefreshTokenHash *string       `json:"-"` // Set internally, not from input
	RefreshExpiresAt *time.Time    `json:"refreshExpiresAt"`
	IPAddress        *net.IP       `json:"ipAddress"`
	UserAgent        *string       `json:"userAgent"`
	DeviceType       string        `json:"deviceType" validate:"omitempty,oneof=web mobile api unknown"`
	BrowserInfo      *BrowserInfo  `json:"browserInfo"`
	LocationInfo     *LocationInfo `json:"locationInfo"`
}

// SessionFilters represents filters for querying sessions
type SessionFilters struct {
	UserID        *uuid.UUID `json:"userId"`
	IsActive      *bool      `json:"isActive"`
	DeviceType    *string    `json:"deviceType"`
	IsExpired     *bool      `json:"isExpired"`
	IPAddress     *net.IP    `json:"ipAddress"`
	CreatedAfter  *time.Time `json:"createdAfter"`
	CreatedBefore *time.Time `json:"createdBefore"`
	Limit         int        `json:"limit"`
	Offset        int        `json:"offset"`
	OrderBy       string     `json:"orderBy"`
	OrderDir      string     `json:"orderDir"`
}
