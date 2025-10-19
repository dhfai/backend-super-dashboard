package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User represents the user table in database
type User struct {
	ID              uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Username        string         `gorm:"type:varchar(50);unique;not null" json:"username" validate:"required,min=3,max=50"`
	Email           string         `gorm:"type:varchar(100);unique;not null" json:"email" validate:"required,email"`
	Password        string         `gorm:"type:varchar(255);not null" json:"-"`
	IsActive        bool           `gorm:"default:true" json:"is_active"`
	EmailVerified   bool           `gorm:"default:false" json:"email_verified"`
	EmailVerifiedAt *time.Time     `gorm:"type:timestamp" json:"email_verified_at,omitempty"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationship
	Profile *Profile `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"profile,omitempty"`
	OTPs    []OTP    `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
}

// Profile represents the user profile table in database
type Profile struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID      uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	FullName    string    `gorm:"type:varchar(100)" json:"full_name" validate:"max=100"`
	Address     string    `gorm:"type:text" json:"address"`
	PhoneNumber string    `gorm:"type:varchar(20)" json:"phone_number" validate:"max=20"`
	Country     string    `gorm:"type:varchar(50)" json:"country" validate:"max=50"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Relationship
	User User `gorm:"foreignKey:UserID;references:ID" json:"-"`
}

// OTP represents the OTP table for password reset
type OTP struct {
	ID                 uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID             *uuid.UUID `gorm:"type:uuid;index" json:"user_id,omitempty"`
	TempRegistrationID *uuid.UUID `gorm:"type:uuid;index" json:"temp_registration_id,omitempty"`
	Code               string     `gorm:"type:varchar(6);not null" json:"code"`
	Type               string     `gorm:"type:varchar(30);not null" json:"type"` // reset_password, verify_email, delete_account, register_verify
	ExpiresAt          time.Time  `gorm:"not null" json:"expires_at"`
	Used               bool       `gorm:"default:false" json:"used"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`

	// Relationships
	User             *User             `gorm:"foreignKey:UserID;references:ID" json:"-"`
	TempRegistration *TempRegistration `gorm:"foreignKey:TempRegistrationID;references:ID" json:"-"`
}

// TempRegistration represents temporary registration data before email verification
type TempRegistration struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Username  string    `gorm:"type:varchar(50);not null" json:"username"`
	Email     string    `gorm:"type:varchar(100);not null" json:"email"`
	Password  string    `gorm:"type:varchar(255);not null" json:"-"`
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relationship
	OTPs []OTP `gorm:"foreignKey:TempRegistrationID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
}

// TokenBlacklist represents blacklisted JWT tokens
type TokenBlacklist struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Token     string    `gorm:"type:text;not null;unique" json:"token"`
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

// BeforeCreate hook for User
func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return
}

// BeforeCreate hook for Profile
func (p *Profile) BeforeCreate(tx *gorm.DB) (err error) {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return
}

// BeforeCreate hook for OTP
func (o *OTP) BeforeCreate(tx *gorm.DB) (err error) {
	if o.ID == uuid.Nil {
		o.ID = uuid.New()
	}
	return
}

// BeforeCreate hook for TempRegistration
func (tr *TempRegistration) BeforeCreate(tx *gorm.DB) (err error) {
	if tr.ID == uuid.Nil {
		tr.ID = uuid.New()
	}
	return
}

// BeforeCreate hook for TokenBlacklist
func (t *TokenBlacklist) BeforeCreate(tx *gorm.DB) (err error) {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return
}

// TableName for User
func (User) TableName() string {
	return "users"
}

// TableName for Profile
func (Profile) TableName() string {
	return "profiles"
}

// TableName for OTP
func (OTP) TableName() string {
	return "otps"
}

// TableName for TempRegistration
func (TempRegistration) TableName() string {
	return "temp_registrations"
}

// TableName for TokenBlacklist
func (TokenBlacklist) TableName() string {
	return "token_blacklists"
}
