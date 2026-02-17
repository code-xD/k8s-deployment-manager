package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/code-xd/k8s-deployment-manager/pkg/dto"
	"github.com/gin-gonic/gin"
)

type Server struct {
	router          *gin.Engine
	srv             *http.Server
	shutdownTimeout time.Duration
}

// NewServer creates a new Server with the given config and Gin router.
// The router should already have ginzap middleware and routes registered.
func NewServer(cfg *dto.APIConfig, router *gin.Engine) *Server {
	srv := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: router,
	}

	return &Server{
		router:          router,
		srv:             srv,
		shutdownTimeout: cfg.Server.ShutdownTimeout,
	}
}

func (s *Server) Run() error {
	errChan := make(chan error)
	go func(errChan chan error) {
		if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- fmt.Errorf("failed to start server: %w", err)
			return
		}

		errChan <- nil
	}(errChan)

	return <-errChan
}

func (s *Server) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
	defer cancel()

	if err := s.srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown server: %w", err)
	}

	return nil
}
