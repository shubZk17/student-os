package ingest

import (
	"context"
	"crypto/subtle"
	"log"
	"net/http"
	"sync/atomic"

	"github.com/gin-gonic/gin"
	"studentos/backend/internal/database"
)

// Handler returns the scheduled-ingest route handler, guarded by a shared secret sent
// in X-Ingest-Token (EventBridge Scheduler sets it). The caller must not register the
// route at all when token is empty.
//
// A run can take minutes, so the handler starts it in the background and replies 202
// immediately rather than holding the scheduler's connection open.
func Handler(db *database.DB, token string) gin.HandlerFunc {
	var running atomic.Bool

	return func(c *gin.Context) {
		given := c.GetHeader("X-Ingest-Token")
		// ConstantTimeCompare reports equal for two empty slices, so an empty token
		// would otherwise authorise a request that simply omits the header.
		if token == "" || subtle.ConstantTimeCompare([]byte(given), []byte(token)) != 1 {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		// Overlapping runs would fight over the same opportunity rows.
		if !running.CompareAndSwap(false, true) {
			c.JSON(http.StatusConflict, gin.H{"status": "already running"})
			return
		}
		// Detached from the request context, which ends as soon as we reply.
		go func() {
			defer running.Store(false)
			ctx, cancel := context.WithTimeout(context.Background(), Timeout)
			defer cancel()
			if _, err := Run(ctx, db); err != nil {
				log.Printf("[ERROR] scheduled ingestion: %v", err)
			}
		}()
		c.JSON(http.StatusAccepted, gin.H{"status": "started"})
	}
}
