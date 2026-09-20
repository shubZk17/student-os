package auth

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
	"studentos/backend/internal/config"
	"studentos/backend/internal/database"
)

type Handler struct {
	db  *database.DB
	cfg *config.Config
}

func NewHandler(db *database.DB, cfg *config.Config) *Handler {
	return &Handler{db: db, cfg: cfg}
}

// Register creates a new student user and initial profile
func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid registration payload"})
		return
	}
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	if req.TargetRoles == nil {
		req.TargetRoles = []string{}
	}

	hashedPassword, err := HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to securely hash password"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	tx, err := h.db.Pool.Begin(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database transaction failure"})
		return
	}
	defer tx.Rollback(ctx)

	var userID string
	userQuery := `INSERT INTO users (email, password_hash, role) VALUES ($1, $2, 'student') RETURNING id`
	err = tx.QueryRow(ctx, userQuery, req.Email, hashedPassword).Scan(&userID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			c.JSON(http.StatusConflict, gin.H{"error": "An account with this email already exists"})
			return
		}
		log.Printf("[ERROR] register: insert user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create account"})
		return
	}

	gradYear := req.GraduationYear
	if gradYear == 0 {
		gradYear = time.Now().Year() + 2
	}
	currentYear := req.CurrentYear
	if currentYear == 0 {
		currentYear = 3
	}

	profileQuery := `
		INSERT INTO student_profiles (
			user_id, full_name, college, degree, branch, current_year, graduation_year, target_roles, profile_strength
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 35)
	`
	_, err = tx.Exec(ctx, profileQuery, userID, req.FullName, req.College, req.Degree, req.Branch, currentYear, gradYear, req.TargetRoles)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initialize student profile"})
		return
	}

	tokens, err := GenerateTokenPair(userID, req.Email, "student", h.cfg.JWTSecret, h.cfg.JWTExpirationMinutes, h.cfg.RefreshExpirationDays)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate authentication tokens"})
		return
	}

	refreshExpiry := time.Now().Add(time.Duration(h.cfg.RefreshExpirationDays) * 24 * time.Hour)
	tokenHash := HashRefreshToken(tokens.RefreshToken)
	refreshQuery := `INSERT INTO refresh_tokens (user_id, token_hash, expires_at) VALUES ($1, $2, $3)`
	_, err = tx.Exec(ctx, refreshQuery, userID, tokenHash, refreshExpiry)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record session"})
		return
	}

	if err := tx.Commit(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Commit failure"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"user": gin.H{
			"id":        userID,
			"email":     req.Email,
			"full_name": req.FullName,
			"role":      "student",
		},
		"tokens": tokens,
	})
}

// Login verifies credentials and issues a new token pair
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid login payload"})
		return
	}
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var user User
	var fullName string
	query := `
		SELECT u.id, u.email, u.password_hash, u.role, COALESCE(sp.full_name, 'Student')
		FROM users u
		LEFT JOIN student_profiles sp ON u.id = sp.user_id
		WHERE lower(u.email) = $1
	`
	err := h.db.Pool.QueryRow(ctx, query, req.Email).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Role, &fullName)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	valid, err := CheckPasswordHash(req.Password, user.PasswordHash)
	if err != nil || !valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	tokens, err := GenerateTokenPair(user.ID, user.Email, user.Role, h.cfg.JWTSecret, h.cfg.JWTExpirationMinutes, h.cfg.RefreshExpirationDays)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
		return
	}

	tokenHash := HashRefreshToken(tokens.RefreshToken)
	refreshExpiry := time.Now().Add(time.Duration(h.cfg.RefreshExpirationDays) * 24 * time.Hour)
	_, err = h.db.Pool.Exec(ctx, `INSERT INTO refresh_tokens (user_id, token_hash, expires_at) VALUES ($1, $2, $3)`, user.ID, tokenHash, refreshExpiry)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to persist refresh token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":        user.ID,
			"email":     user.Email,
			"full_name": fullName,
			"role":      user.Role,
		},
		"tokens": tokens,
	})
}

// Refresh rotates the refresh token and yields a fresh access token
func (h *Handler) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Refresh token is required"})
		return
	}

	tokenHash := HashRefreshToken(req.RefreshToken)
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	tx, err := h.db.Pool.Begin(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Token rotation failed"})
		return
	}
	defer tx.Rollback(ctx)

	// Revoke-and-read in one statement so a token can only ever be redeemed once.
	var userID, email, role string
	query := `
		UPDATE refresh_tokens rt SET revoked = TRUE
		FROM users u
		WHERE rt.token_hash = $1 AND rt.user_id = u.id AND NOT rt.revoked AND rt.expires_at > NOW()
		RETURNING rt.user_id, u.email, u.role
	`
	if err := tx.QueryRow(ctx, query, tokenHash).Scan(&userID, &email, &role); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired refresh token"})
		return
	}

	newTokens, err := GenerateTokenPair(userID, email, role, h.cfg.JWTSecret, h.cfg.JWTExpirationMinutes, h.cfg.RefreshExpirationDays)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Token rotation failed"})
		return
	}

	newExpiry := time.Now().Add(time.Duration(h.cfg.RefreshExpirationDays) * 24 * time.Hour)
	if _, err := tx.Exec(ctx, `INSERT INTO refresh_tokens (user_id, token_hash, expires_at) VALUES ($1, $2, $3)`, userID, HashRefreshToken(newTokens.RefreshToken), newExpiry); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Token rotation failed"})
		return
	}
	if err := tx.Commit(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Token rotation failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"tokens": newTokens})
}

// Logout revokes the provided refresh token
func (h *Handler) Logout(c *gin.Context) {
	var req LogoutRequest
	if err := c.ShouldBindJSON(&req); err == nil && req.RefreshToken != "" {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
		defer cancel()
		if _, err := h.db.Pool.Exec(ctx, `UPDATE refresh_tokens SET revoked = TRUE WHERE token_hash = $1`, HashRefreshToken(req.RefreshToken)); err != nil {
			log.Printf("[ERROR] logout: revoke token: %v", err)
		}
	}
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}
