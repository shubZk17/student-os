package auth

import "time"

type RegisterRequest struct {
	Email          string   `json:"email" binding:"required,email,max=255"`
	Password       string   `json:"password" binding:"required,min=8,max=128"`
	FullName       string   `json:"full_name" binding:"required,max=255"`
	College        string   `json:"college" binding:"max=255"`
	Degree         string   `json:"degree" binding:"max=128"`
	Branch         string   `json:"branch" binding:"max=128"`
	GraduationYear int      `json:"graduation_year" binding:"omitempty,min=1950,max=2100"`
	CurrentYear    int      `json:"current_year" binding:"omitempty,min=1,max=5"`
	TargetRoles    []string `json:"target_roles" binding:"max=20,dive,max=128"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email,max=255"`
	Password string `json:"password" binding:"required,max=128"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type AuthResponse struct {
	User   UserResponse `json:"user"`
	Tokens TokenPair    `json:"tokens"`
}

type UserResponse struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	FullName string `json:"full_name"`
	Role     string `json:"role"`
}

type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"`
	IsVerified   bool      `json:"is_verified"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
