package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/iagomaia/go-foodji/internal/config"
	"github.com/iagomaia/go-foodji/internal/handler"
	"github.com/iagomaia/go-foodji/internal/middleware"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

type Server struct {
	cfg            *config.Config
	log            *slog.Logger
	sessionHandler *handler.SessionHandler
	voteHandler    *handler.VoteHandler
}

func New(
	cfg *config.Config,
	log *slog.Logger,
	sessionHandler *handler.SessionHandler,
	voteHandler *handler.VoteHandler,
) *Server {
	return &Server{
		cfg:            cfg,
		log:            log,
		sessionHandler: sessionHandler,
		voteHandler:    voteHandler,
	}
}

func (s *Server) Run() error {
	if s.cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(otelgin.Middleware("go-foodji"))
	r.Use(middleware.Logger(s.log))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	v1 := r.Group("/api/v1")
	s.sessionHandler.RegisterRoutes(v1)
	s.voteHandler.RegisterRoutes(v1)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", s.cfg.AppPort),
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		s.log.Info("server starting", slog.String("port", s.cfg.AppPort))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.log.Error("server error", slog.Any("error", err))
			os.Exit(1)
		}
	}()

	<-quit
	s.log.Info("shutting down server")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return srv.Shutdown(ctx)
}
