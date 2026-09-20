package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"studentos/backend/internal/applications"
	"studentos/backend/internal/auth"
	"studentos/backend/internal/config"
	"studentos/backend/internal/dashboard"
	"studentos/backend/internal/database"
	"studentos/backend/internal/ingest"
	"studentos/backend/internal/jobs"
	"studentos/backend/internal/middleware"
	"studentos/backend/internal/notifications"
	"studentos/backend/internal/projects"
	"studentos/backend/internal/users"
	"studentos/backend/migrations"
)

func main() {
	cfg, err := config.Load()
	if err == nil {
		err = cfg.ValidateJWT()
	}
	if err != nil {
		log.Fatalf("[FATAL] Invalid configuration: %v", err)
	}

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("[FATAL] Database: %v", err)
	}
	defer db.Close()

	if err := migrations.Up(context.Background(), db.Pool); err != nil {
		log.Fatalf("[FATAL] Migrations: %v", err)
	}

	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	// Client IP comes from the TCP peer unless TRUSTED_PLATFORM names a header set by the
	// edge proxy (Render: CF-Connecting-IP). Never trust X-Forwarded-For: clients can forge it.
	_ = r.SetTrustedProxies(nil)
	r.TrustedPlatform = cfg.TrustedPlatform
	r.Use(gin.Recovery())
	r.Use(middleware.StructuredLogger())
	r.Use(middleware.SetupCORS(cfg.AllowedOrigins))
	r.Use(func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1<<20) // 1 MiB
		c.Next()
	})

	// Health check
	r.GET("/health", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()
		if err := db.Pool.Ping(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unhealthy", "database": "unreachable"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "healthy", "database": "connected"})
	})

	// Scheduled ingestion (EventBridge Scheduler -> this route). Not registered at all
	// unless INGEST_TOKEN is set, so it can never be reachable with an empty secret.
	if cfg.IngestToken != "" {
		r.POST("/internal/ingest", ingest.Handler(db, cfg.IngestToken))
		log.Println("[INFO] Scheduled ingestion route enabled at POST /internal/ingest")
	}

	// Handlers
	authHandler := auth.NewHandler(db, cfg)
	userHandler := users.NewHandler(db)
	jobHandler := jobs.NewHandler(db)
	appHandler := applications.NewHandler(db)
	projHandler := projects.NewHandler(db)
	notifHandler := notifications.NewHandler(db)
	dashHandler := dashboard.NewHandler(db)

	api := r.Group("/api/v1")
	{
		// Public Auth Routes
		authGroup := api.Group("/auth")
		// Many students can share one campus NAT IP, so keep this generous; it exists to stop
		// password guessing, not normal use.
		credentialLimit := middleware.NewRateLimiter(30, 5*time.Minute).Middleware()
		{
			authGroup.POST("/register", credentialLimit, authHandler.Register)
			authGroup.POST("/login", credentialLimit, authHandler.Login)
			authGroup.POST("/refresh", authHandler.Refresh)
			authGroup.POST("/logout", authHandler.Logout)
		}

		// Public Opportunities Browsing
		api.GET("/opportunities", jobHandler.ListOpportunities)
		api.GET("/opportunities/:id", jobHandler.GetOpportunityByID)

		// Protected Student Routes
		protected := api.Group("")
		protected.Use(middleware.RequireAuth(cfg.JWTSecret))
		{
			// Student Profile
			protected.GET("/profile", userHandler.GetProfile)
			protected.PUT("/profile", userHandler.UpdateProfile)

			// Personalized Recommendations
			protected.GET("/recommendations", jobHandler.GetRecommendations)

			// Application Tracking (Kanban)
			protected.GET("/applications", appHandler.ListApplications)
			protected.POST("/applications", appHandler.CreateApplication)
			protected.PATCH("/applications/:id", appHandler.UpdateStage)
			protected.DELETE("/applications/:id", appHandler.DeleteApplication)

			// Projects & Portfolio
			protected.GET("/projects", projHandler.ListProjects)
			protected.POST("/projects", projHandler.CreateProject)
			protected.DELETE("/projects/:id", projHandler.DeleteProject)

			// Notifications
			protected.GET("/notifications", notifHandler.ListNotifications)
			protected.PATCH("/notifications/:id/read", notifHandler.MarkAsRead)

			// Dashboard Command Center
			protected.GET("/dashboard/summary", dashHandler.GetSummary)
		}
	}

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("[INFO] StudentOS Backend starting on port %s (%s)", cfg.Port, cfg.Env)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[FATAL] Server error: %v", err)
		}
	}()

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("[INFO] Shutting down server gracefully...")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("[ERROR] Server forced to shutdown: %v", err)
	}
	log.Println("[INFO] Server exited cleanly.")
}
