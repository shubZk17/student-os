package applications

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"studentos/backend/internal/database"
	"studentos/backend/internal/middleware"
)

type Handler struct {
	db *database.DB
}

func NewHandler(db *database.DB) *Handler {
	return &Handler{db: db}
}

// ListApplications returns all applications for the authenticated student
func (h *Handler) ListApplications(c *gin.Context) {
	userID := middleware.GetUserID(c)
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	query := `
		SELECT a.id, a.user_id, a.opportunity_id, o.company_name, o.title, a.stage,
		       a.applied_at, a.interview_date, a.next_follow_up, COALESCE(a.notes, ''),
		       o.location, COALESCE(o.company_logo_url, ''), a.created_at, a.updated_at
		FROM applications a
		JOIN opportunities o ON a.opportunity_id = o.id
		WHERE a.user_id = $1
		ORDER BY a.updated_at DESC
	`
	rows, err := h.db.Pool.Query(ctx, query, userID)
	if err != nil {
		log.Printf("[ERROR] list applications: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch applications"})
		return
	}
	defer rows.Close()

	apps := make([]Application, 0)
	for rows.Next() {
		var a Application
		if err := rows.Scan(
			&a.ID, &a.UserID, &a.OpportunityID, &a.Company, &a.Title, &a.Stage,
			&a.AppliedAt, &a.InterviewDate, &a.NextFollowUp, &a.Notes,
			&a.Location, &a.CompanyLogoURL, &a.CreatedAt, &a.UpdatedAt,
		); err != nil {
			log.Printf("[ERROR] list applications: scan: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch applications"})
			return
		}
		apps = append(apps, a)
	}
	if err := rows.Err(); err != nil {
		log.Printf("[ERROR] list applications: rows: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch applications"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"applications": apps})
}

// CreateApplication adds an opportunity to the student's tracker
func (h *Handler) CreateApplication(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req CreateApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid application payload"})
		return
	}

	stage := req.Stage
	if stage == "" {
		stage = "SAVED"
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	query := `
		INSERT INTO applications (user_id, opportunity_id, stage, notes, interview_date, applied_at)
		VALUES ($1, $2, $3, $4, $5, CASE WHEN $6 THEN NOW() END)
		ON CONFLICT (user_id, opportunity_id)
		DO UPDATE SET stage = EXCLUDED.stage, notes = EXCLUDED.notes,
		              applied_at = COALESCE(applications.applied_at, EXCLUDED.applied_at), updated_at = NOW()
		RETURNING id
	`
	var appID string
	err := h.db.Pool.QueryRow(ctx, query, userID, req.OpportunityID, stage, req.Notes, req.InterviewDate, stage != "SAVED").Scan(&appID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" { // foreign_key_violation
			c.JSON(http.StatusNotFound, gin.H{"error": "Opportunity not found"})
			return
		}
		log.Printf("[ERROR] create application: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to track application"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Application tracked successfully", "id": appID})
}

// UpdateStage updates the Kanban column and, when provided, interview date, follow-up or notes
func (h *Handler) UpdateStage(c *gin.Context) {
	userID := middleware.GetUserID(c)
	appID := c.Param("id")
	if uuid.Validate(appID) != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Application not found"})
		return
	}

	var req UpdateStageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid update stage payload"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	query := `
		UPDATE applications SET
			stage = $1,
			notes = COALESCE($2, notes),
			interview_date = COALESCE($3, interview_date),
			next_follow_up = COALESCE($4, next_follow_up),
			applied_at = CASE WHEN $7 THEN COALESCE(applied_at, NOW()) ELSE applied_at END,
			updated_at = NOW()
		WHERE id = $5 AND user_id = $6
	`
	res, err := h.db.Pool.Exec(ctx, query, req.Stage, req.Notes, req.InterviewDate, req.NextFollowUp, appID, userID, req.Stage != "SAVED")
	if err != nil {
		log.Printf("[ERROR] update application: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update application stage"})
		return
	}
	if res.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Application not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Application stage updated successfully"})
}

// DeleteApplication removes an application from the student's tracker
func (h *Handler) DeleteApplication(c *gin.Context) {
	userID := middleware.GetUserID(c)
	appID := c.Param("id")
	if uuid.Validate(appID) != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Application not found"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	res, err := h.db.Pool.Exec(ctx, `DELETE FROM applications WHERE id = $1 AND user_id = $2`, appID, userID)
	if err != nil {
		log.Printf("[ERROR] delete application: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove application"})
		return
	}
	if res.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Application not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Application removed from tracker"})
}
