package notifications

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"studentos/backend/internal/database"
	"studentos/backend/internal/middleware"
)

type Notification struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Title     string    `json:"title"`
	Message   string    `json:"message"`
	Type      string    `json:"type"` // 'HIGH_MATCH', 'DEADLINE_SOON', 'INTERVIEW_ALERT', 'PROFILE_TIP'
	LinkURL   string    `json:"link_url"`
	IsRead    bool      `json:"is_read"`
	CreatedAt time.Time `json:"created_at"`
}

type Handler struct {
	db *database.DB
}

func NewHandler(db *database.DB) *Handler {
	return &Handler{db: db}
}

// ListNotifications returns the latest notifications with the total unread count
func (h *Handler) ListNotifications(c *gin.Context) {
	userID := middleware.GetUserID(c)
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	query := `
		SELECT id, user_id, title, message, type, COALESCE(link_url, ''), is_read, created_at,
		       COUNT(*) FILTER (WHERE NOT is_read) OVER ()
		FROM notifications
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT 50
	`
	rows, err := h.db.Pool.Query(ctx, query, userID)
	if err != nil {
		log.Printf("[ERROR] list notifications: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch notifications"})
		return
	}
	defer rows.Close()

	notifs := make([]Notification, 0)
	unread := 0
	for rows.Next() {
		var n Notification
		if err := rows.Scan(&n.ID, &n.UserID, &n.Title, &n.Message, &n.Type, &n.LinkURL, &n.IsRead, &n.CreatedAt, &unread); err != nil {
			log.Printf("[ERROR] list notifications: scan: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch notifications"})
			return
		}
		notifs = append(notifs, n)
	}
	if err := rows.Err(); err != nil {
		log.Printf("[ERROR] list notifications: rows: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch notifications"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"unread_count":  unread,
		"notifications": notifs,
	})
}

// MarkAsRead marks a notification as read
func (h *Handler) MarkAsRead(c *gin.Context) {
	userID := middleware.GetUserID(c)
	notifID := c.Param("id")
	if uuid.Validate(notifID) != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Notification not found"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	res, err := h.db.Pool.Exec(ctx, `UPDATE notifications SET is_read = TRUE WHERE id = $1 AND user_id = $2`, notifID, userID)
	if err != nil {
		log.Printf("[ERROR] mark notification read: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update notification"})
		return
	}
	if res.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Notification not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notification marked as read"})
}
