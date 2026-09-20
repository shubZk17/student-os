package projects

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

// ListProjects returns all projects belonging to the student
func (h *Handler) ListProjects(c *gin.Context) {
	userID := middleware.GetUserID(c)
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	query := `
		SELECT p.id, p.student_id, p.title, COALESCE(p.tagline, ''), p.description,
		       COALESCE(p.role, 'Developer'), COALESCE(p.github_url, ''), COALESCE(p.demo_url, ''),
		       COALESCE(p.metrics, ''),
		       COALESCE((SELECT array_agg(s.name ORDER BY s.name) FROM project_skills ps
		                 JOIN skills s ON s.id = ps.skill_id WHERE ps.project_id = p.id), '{}'),
		       p.created_at, p.updated_at
		FROM projects p
		JOIN student_profiles sp ON p.student_id = sp.id
		WHERE sp.user_id = $1
		ORDER BY p.created_at DESC
	`
	rows, err := h.db.Pool.Query(ctx, query, userID)
	if err != nil {
		log.Printf("[ERROR] list projects: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve projects"})
		return
	}
	defer rows.Close()

	projects := make([]Project, 0)
	for rows.Next() {
		var p Project
		if err := rows.Scan(
			&p.ID, &p.StudentID, &p.Title, &p.Tagline, &p.Description,
			&p.Role, &p.GithubURL, &p.DemoURL, &p.Metrics, &p.Technologies, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			log.Printf("[ERROR] list projects: scan: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve projects"})
			return
		}
		projects = append(projects, p)
	}
	if err := rows.Err(); err != nil {
		log.Printf("[ERROR] list projects: rows: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve projects"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"projects": projects})
}

// CreateProject adds a new portfolio project along with its technologies
func (h *Handler) CreateProject(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project payload"})
		return
	}
	if req.Technologies == nil {
		req.Technologies = []string{}
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	tx, err := h.db.Pool.Begin(ctx)
	if err != nil {
		log.Printf("[ERROR] create project: begin: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create project"})
		return
	}
	defer tx.Rollback(ctx)

	p := Project{
		Title: req.Title, Tagline: req.Tagline, Description: req.Description, Role: req.Role,
		GithubURL: req.GithubURL, DemoURL: req.DemoURL, Metrics: req.Metrics, Technologies: req.Technologies,
	}
	err = tx.QueryRow(ctx, `
		INSERT INTO projects (student_id, title, tagline, description, role, github_url, demo_url, metrics)
		SELECT id, $2, $3, $4, COALESCE(NULLIF($5, ''), 'Developer'), $6, $7, $8 FROM student_profiles WHERE user_id = $1
		RETURNING id, student_id, role, created_at, updated_at`,
		userID, req.Title, req.Tagline, req.Description, req.Role, req.GithubURL, req.DemoURL, req.Metrics,
	).Scan(&p.ID, &p.StudentID, &p.Role, &p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Student profile not found"})
		return
	}
	if err != nil {
		log.Printf("[ERROR] create project: insert: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create project"})
		return
	}

	if len(req.Technologies) > 0 {
		if _, err := tx.Exec(ctx, `INSERT INTO skills (name) SELECT unnest($1::text[]) ON CONFLICT (name) DO NOTHING`, req.Technologies); err != nil {
			log.Printf("[ERROR] create project: skills: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create project"})
			return
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO project_skills (project_id, skill_id)
			SELECT $1, id FROM skills WHERE name = ANY($2) ON CONFLICT DO NOTHING`, p.ID, req.Technologies); err != nil {
			log.Printf("[ERROR] create project: project skills: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create project"})
			return
		}
	}

	if err := tx.Commit(ctx); err != nil {
		log.Printf("[ERROR] create project: commit: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create project"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Project created successfully", "id": p.ID, "project": p})
}

// DeleteProject deletes a project owned by the student
func (h *Handler) DeleteProject(c *gin.Context) {
	userID := middleware.GetUserID(c)
	projID := c.Param("id")
	if uuid.Validate(projID) != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	res, err := h.db.Pool.Exec(ctx, `
		DELETE FROM projects WHERE id = $1 AND student_id IN (
			SELECT id FROM student_profiles WHERE user_id = $2
		)
	`, projID, userID)
	if err != nil {
		log.Printf("[ERROR] delete project: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete project"})
		return
	}
	if res.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Project deleted successfully"})
}
