package users

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"studentos/backend/internal/database"
	"studentos/backend/internal/middleware"
)

type Handler struct {
	db *database.DB
}

func NewHandler(db *database.DB) *Handler {
	return &Handler{db: db}
}

// GetProfile returns the current student's full profile including skills and completion percentage
func (h *Handler) GetProfile(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	query := `
		SELECT id, user_id, full_name, COALESCE(avatar_url, ''), COALESCE(college, ''), COALESCE(degree, ''),
		       COALESCE(branch, ''), COALESCE(current_year, 1), graduation_year, COALESCE(cgpa, 0.0),
		       COALESCE(target_roles, '{}'), COALESCE(preferred_locations, '{}'), COALESCE(work_preference, 'ANY'),
		       COALESCE(bio, ''), COALESCE(github_url, ''), COALESCE(linkedin_url, ''),
		       COALESCE(portfolio_url, ''), COALESCE(resume_url, ''), profile_strength, created_at, updated_at
		FROM student_profiles
		WHERE user_id = $1
	`
	var p StudentProfile
	err := h.db.Pool.QueryRow(ctx, query, userID).Scan(
		&p.ID, &p.UserID, &p.FullName, &p.AvatarURL, &p.College, &p.Degree,
		&p.Branch, &p.CurrentYear, &p.GraduationYear, &p.CGPA,
		&p.TargetRoles, &p.PreferredLocations, &p.WorkPreference,
		&p.Bio, &p.GithubURL, &p.LinkedinURL,
		&p.PortfolioURL, &p.ResumeURL, &p.ProfileStrength, &p.CreatedAt, &p.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Profile not found"})
		return
	}
	if err != nil {
		log.Printf("[ERROR] get profile: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load profile"})
		return
	}

	// Fetch skills
	skillsQuery := `
		SELECT s.id, s.name, COALESCE(s.category, ''), COALESCE(ss.proficiency, '')
		FROM student_skills ss
		JOIN skills s ON ss.skill_id = s.id
		WHERE ss.student_id = $1
	`
	rows, err := h.db.Pool.Query(ctx, skillsQuery, p.ID)
	if err != nil {
		log.Printf("[ERROR] get profile skills: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load profile"})
		return
	}
	defer rows.Close()
	p.Skills = make([]Skill, 0)
	for rows.Next() {
		var sk Skill
		if err := rows.Scan(&sk.ID, &sk.Name, &sk.Category, &sk.Proficiency); err != nil {
			log.Printf("[ERROR] get profile skills: scan: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load profile"})
			return
		}
		p.Skills = append(p.Skills, sk)
	}
	if err := rows.Err(); err != nil {
		log.Printf("[ERROR] get profile skills: rows: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"profile": p})
}

// UpdateProfile updates candidate details and re-calculates profile strength score
func (h *Handler) UpdateProfile(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid update profile payload"})
		return
	}

	if req.WorkPreference == "" {
		req.WorkPreference = "ANY"
	}
	if req.TargetRoles == nil {
		req.TargetRoles = []string{}
	}
	if req.PreferredLocations == nil {
		req.PreferredLocations = []string{}
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// Strength from request fields; resume adds 10 more in SQL since it may be omitted (kept from DB).
	strength := 20
	if req.College != "" && req.Degree != "" {
		strength += 20
	}
	if req.CGPA > 0 {
		strength += 10
	}
	if len(req.TargetRoles) > 0 {
		strength += 15
	}
	if len(req.PreferredLocations) > 0 {
		strength += 10
	}
	if req.GithubURL != "" {
		strength += 10
	}
	if req.LinkedinURL != "" {
		strength += 5
	}

	tx, err := h.db.Pool.Begin(ctx)
	if err != nil {
		log.Printf("[ERROR] update profile: begin: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}
	defer tx.Rollback(ctx)

	query := `
		UPDATE student_profiles SET
			full_name = COALESCE(NULLIF($1, ''), full_name),
			college = $2,
			degree = $3,
			branch = $4,
			current_year = $5,
			graduation_year = $6,
			cgpa = $7,
			target_roles = $8,
			preferred_locations = $9,
			work_preference = $10,
			bio = $11,
			github_url = $12,
			linkedin_url = $13,
			portfolio_url = COALESCE($14, portfolio_url),
			resume_url = COALESCE($15, resume_url),
			profile_strength = LEAST(100, $16 + CASE WHEN COALESCE($15, resume_url, '') <> '' THEN 10 ELSE 0 END),
			updated_at = NOW()
		WHERE user_id = $17
		RETURNING id, profile_strength
	`
	var profileID string
	err = tx.QueryRow(ctx, query,
		req.FullName, req.College, req.Degree, req.Branch, req.CurrentYear,
		req.GraduationYear, req.CGPA, req.TargetRoles, req.PreferredLocations,
		req.WorkPreference, req.Bio, req.GithubURL, req.LinkedinURL,
		req.PortfolioURL, req.ResumeURL, strength, userID,
	).Scan(&profileID, &strength)
	if errors.Is(err, pgx.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Profile not found"})
		return
	}
	if err != nil {
		log.Printf("[ERROR] update profile: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	if req.Skills != nil {
		if err := replaceSkills(ctx, tx, profileID, *req.Skills); err != nil {
			log.Printf("[ERROR] update profile skills: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
			return
		}
	}

	if err := tx.Commit(ctx); err != nil {
		log.Printf("[ERROR] update profile: commit: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Profile updated successfully", "profile_strength": strength})
}

func replaceSkills(ctx context.Context, tx pgx.Tx, profileID string, skills []string) error {
	if _, err := tx.Exec(ctx, `DELETE FROM student_skills WHERE student_id = $1`, profileID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `INSERT INTO skills (name) SELECT unnest($1::text[]) ON CONFLICT (name) DO NOTHING`, skills); err != nil {
		return err
	}
	_, err := tx.Exec(ctx, `
		INSERT INTO student_skills (student_id, skill_id)
		SELECT $1, id FROM skills WHERE name = ANY($2) ON CONFLICT DO NOTHING`, profileID, skills)
	return err
}
