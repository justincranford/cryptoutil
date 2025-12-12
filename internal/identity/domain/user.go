// Copyright (c) 2025 Justin Cranford
//
//

package domain

import (
	"time"

	googleUuid "github.com/google/uuid"
	"gorm.io/gorm"
)

// User represents a user entity with OIDC standard claims.
type User struct {
	// Primary identifier.
	ID googleUuid.UUID `gorm:"type:text;primaryKey" json:"id"`

	// OIDC standard claims.
	Sub               string   `gorm:"uniqueIndex;not null" json:"sub"`                                                            // Subject identifier.
	Name              string   `json:"name,omitempty"`                                                                             // Full name.
	GivenName         string   `json:"given_name,omitempty"`                                                                       // Given name(s) or first name(s).
	FamilyName        string   `json:"family_name,omitempty"`                                                                      // Surname(s) or last name(s).
	MiddleName        string   `json:"middle_name,omitempty"`                                                                      // Middle name(s).
	Nickname          string   `json:"nickname,omitempty"`                                                                         // Casual name.
	PreferredUsername string   `gorm:"uniqueIndex" json:"preferred_username,omitempty"`                                            // Shorthand name.
	Profile           string   `json:"profile,omitempty"`                                                                          // Profile page URL.
	Picture           string   `json:"picture,omitempty"`                                                                          // Profile picture URL.
	Website           string   `json:"website,omitempty"`                                                                          // Web page or blog URL.
	Email             string   `gorm:"uniqueIndex" json:"email,omitempty"`                                                         // Email address.
	EmailVerified     IntBool  `gorm:"type:integer;default:0" json:"email_verified,omitempty"`                                     // Email verification status (INTEGER for cross-DB compatibility).
	Gender            string   `json:"gender,omitempty"`                                                                           // Gender.
	Birthdate         string   `json:"birthdate,omitempty"`                                                                        // Birthday (YYYY-MM-DD).
	Zoneinfo          string   `json:"zoneinfo,omitempty"`                                                                         // Time zone.
	Locale            string   `json:"locale,omitempty"`                                                                           // Locale.
	PhoneNumber       string   `json:"phone_number,omitempty"`                                                                     // Phone number.
	PhoneVerified     IntBool  `gorm:"column:phone_number_verified;type:integer;default:0" json:"phone_number_verified,omitempty"` // Phone verification status (INTEGER for cross-DB compatibility).
	Address           *Address `gorm:"embedded;embeddedPrefix:address_" json:"address,omitempty"`                                  // Physical address.

	// MFA device tokens.
	PushDeviceToken string `json:"push_device_token,omitempty"` // Push notification device token (FCM, APNs).

	// Authentication credentials.
	PasswordHash string `gorm:"not null" json:"-"` // Bcrypt password hash.

	// Account status.
	Enabled   bool       `gorm:"default:true" json:"enabled"`       // Account enabled status.
	Locked    bool       `gorm:"default:false" json:"locked"`       // Account locked status.
	CreatedAt time.Time  `json:"created_at"`                        // Account creation timestamp.
	UpdatedAt time.Time  `json:"updated_at"`                        // Last update timestamp (OIDC claim + GORM timestamp).
	DeletedAt *time.Time `gorm:"index" json:"deleted_at,omitempty"` // Soft delete timestamp.
}

// Address represents an OIDC address claim.
type Address struct {
	Formatted     string `json:"formatted,omitempty"`      // Full mailing address.
	StreetAddress string `json:"street_address,omitempty"` // Street address.
	Locality      string `json:"locality,omitempty"`       // City or locality.
	Region        string `json:"region,omitempty"`         // State, province, or region.
	PostalCode    string `json:"postal_code,omitempty"`    // ZIP or postal code.
	Country       string `json:"country,omitempty"`        // Country name.
}

// BeforeCreate generates UUID for new users.
func (u *User) BeforeCreate(_ *gorm.DB) error {
	if u.ID == googleUuid.Nil {
		u.ID = googleUuid.Must(googleUuid.NewV7())
	}

	return nil
}

// TableName returns the table name for User entities.
func (User) TableName() string {
	return "users"
}
