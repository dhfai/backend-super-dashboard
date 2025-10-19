package models

import "time"

type RegisterRequest struct {
	Username       string `json:"username" validate:"required,min=3,max=50"`
	Email          string `json:"email" validate:"required,email"`
	Password       string `json:"password" validate:"required,min=8"`
	RetypePassword string `json:"retype_password" validate:"required,eqfield=Password"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type ForgetPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type ResetPasswordRequest struct {
	Email       string `json:"email" validate:"required,email"`
	OTPCode     string `json:"otp_code" validate:"required,len=6"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
}

type VerifyEmailRequest struct {
	Email   string `json:"email" validate:"required,email"`
	OTPCode string `json:"otp_code" validate:"required,len=6"`
}

type VerifyRegistrationRequest struct {
	Email   string `json:"email" validate:"required,email"`
	OTPCode string `json:"otp_code" validate:"required,len=6"`
}

type ResendVerificationRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type DeleteAccountRequest struct {
	OTPCode string `json:"otp_code" validate:"required,len=6"`
}

type RequestDeleteAccountRequest struct {
	Password string `json:"password" validate:"required"`
}

type UpdateProfileRequest struct {
	FullName    string `json:"full_name" validate:"max=100"`
	Address     string `json:"address"`
	PhoneNumber string `json:"phone_number" validate:"max=20"`
	Country     string `json:"country" validate:"max=50"`
}

type LoginResponse struct {
	Token string        `json:"token"`
	User  *UserResponse `json:"user"`
}

type UserResponse struct {
	ID              string     `json:"id"`
	Username        string     `json:"username"`
	Email           string     `json:"email"`
	IsActive        bool       `json:"is_active"`
	EmailVerified   bool       `json:"email_verified"`
	EmailVerifiedAt *time.Time `json:"email_verified_at,omitempty"`
	Profile         *Profile   `json:"profile,omitempty"`
}

type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (u *User) ToUserResponse() *UserResponse {
	return &UserResponse{
		ID:              u.ID.String(),
		Username:        u.Username,
		Email:           u.Email,
		IsActive:        u.IsActive,
		EmailVerified:   u.EmailVerified,
		EmailVerifiedAt: u.EmailVerifiedAt,
		Profile:         u.Profile,
	}
}

// Note DTOs
type CreateNoteRequest struct {
	Title      string `json:"title" validate:"required,min=1,max=255"`
	Content    string `json:"content" validate:"required,min=1"`
	Tags       string `json:"tags" validate:"max=500"`
	IsFavorite bool   `json:"is_favorite"`
}

type UpdateNoteRequest struct {
	Title      string `json:"title" validate:"omitempty,min=1,max=255"`
	Content    string `json:"content" validate:"omitempty,min=1"`
	Tags       string `json:"tags" validate:"max=500"`
	IsFavorite *bool  `json:"is_favorite"`
}

type NoteResponse struct {
	ID         uint      `json:"id"`
	UserID     string    `json:"user_id"`
	Title      string    `json:"title"`
	Content    string    `json:"content"`
	Tags       string    `json:"tags"`
	IsFavorite bool      `json:"is_favorite"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type NotesListResponse struct {
	Notes      []NoteResponse `json:"notes"`
	Total      int64          `json:"total"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
	TotalPages int            `json:"total_pages"`
}

type NoteFilterRequest struct {
	Search     string `json:"search" form:"search"`
	Tags       string `json:"tags" form:"tags"`
	IsFavorite *bool  `json:"is_favorite" form:"is_favorite"`
	Page       int    `json:"page" form:"page"`
	PageSize   int    `json:"page_size" form:"page_size"`
	SortBy     string `json:"sort_by" form:"sort_by"`       // created_at, updated_at, title
	SortOrder  string `json:"sort_order" form:"sort_order"` // asc, desc
}

func (n *Note) ToNoteResponse() *NoteResponse {
	return &NoteResponse{
		ID:         n.ID,
		UserID:     n.UserID.String(),
		Title:      n.Title,
		Content:    n.Content,
		Tags:       n.Tags,
		IsFavorite: n.IsFavorite,
		CreatedAt:  n.CreatedAt,
		UpdatedAt:  n.UpdatedAt,
	}
}
