package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/relay/backend/internal/auth"
	"github.com/relay/backend/internal/config"
	"github.com/relay/backend/internal/db"
	"github.com/relay/backend/internal/logger"
	"github.com/relay/backend/internal/user"
)

type Server struct {
	engine      *gin.Engine
	cfg         *config.Config
	log         *logger.Logger
	db          *db.DB
	authHandler *auth.Handler
	srv         *http.Server
}

func New(cfg *config.Config, log *logger.Logger, dbConn *db.DB) *Server {
	if cfg.IsProd() {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	engine := gin.New()

	userRepo := user.NewPostgresRepository(dbConn.Conn(), log)
	authService := auth.NewService(userRepo, log, cfg.Auth)
	authHandler := auth.NewHandler(authService)

	server := &Server{
		engine:      engine,
		cfg:         cfg,
		log:         log,
		db:          dbConn,
		authHandler: authHandler,
	}

	server.registerMiddleware()
	server.registerRoutes()

	return server
}

func (s *Server) registerMiddleware() {
	s.engine.Use(gin.Recovery())
	s.engine.Use(s.loggingMiddleware())
	s.engine.Use(corsMiddleware())
}

func (s *Server) loggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()
		c.Next()

		duration := time.Since(startTime)
		statusCode := c.Writer.Status()
		method := c.Request.Method
		path := c.Request.URL.Path
		level := slog.LevelInfo
		if statusCode >= 500 {
			level = slog.LevelError
		} else if statusCode >= 400 {
			level = slog.LevelWarn
		}

		s.log.Logger.Log(c.Request.Context(), level, "http_request",
			slog.String("method", method),
			slog.String("path", path),
			slog.Int("status_code", statusCode),
			slog.Duration("duration_ms", duration),
			slog.String("remote_addr", c.ClientIP()),
		)
	}
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func (s *Server) registerRoutes() {
	s.engine.GET("/health", s.healthHandler)
	s.engine.GET("/ready", s.readyHandler)

	v1 := s.engine.Group("/api/v1")
	{
		authGroup := v1.Group("/auth")
		{
			authGroup.POST("/register", s.authHandler.Register)
			authGroup.POST("/login", s.authHandler.Login)
		}

		v1.GET("/ping", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "pong"})
		})

		protected := v1.Group("")
		protected.Use(auth.Middleware(s.authHandler.Service()))
		{
			protected.GET("/me", s.authHandler.Me)
		}
	}
}

func (s *Server) healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"service": "relay-api",
		"time":    time.Now().UTC(),
	})
}

func (s *Server) readyHandler(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	if s.db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unready", "reason": "database unavailable"})
		return
	}

	if err := s.db.Ping(ctx); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unready", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "ready",
		"service": "relay-api",
		"time":    time.Now().UTC(),
	})
}

func (s *Server) Run() error {
	addr := fmt.Sprintf("%s:%d", s.cfg.Server.Host, s.cfg.Server.Port)

	s.srv = &http.Server{
		Addr:         addr,
		Handler:      s.engine,
		ReadTimeout:  time.Duration(s.cfg.Server.Timeout) * time.Second,
		WriteTimeout: time.Duration(s.cfg.Server.Timeout) * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	s.log.Info("starting http server", slog.String("addr", addr))

	if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server error: %w", err)
	}

	return nil
}

func (s *Server) Shutdown(timeout time.Duration) error {
	if s.srv == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return s.srv.Shutdown(ctx)
}
