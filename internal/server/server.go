package server

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"roboticCrewChallenge/internal/auth"
	"roboticCrewChallenge/internal/config"
)

const shutdownTimeout = 10 * time.Second

type Server struct {
	httpServer *http.Server
	logger     *slog.Logger
}

func New(cfg config.Config, logger *slog.Logger, authenticator *auth.Authenticator) *Server {
	if authenticator == nil {
		panic("server: authenticator is required")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", handleHealth)

	requireAuth := auth.BasicAuth(authenticator)
	mux.Handle("/graphql", requireAuth(http.HandlerFunc(handleGraphQLNotImplemented)))

	httpServer := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	return &Server{httpServer: httpServer, logger: logger}
}

func (s *Server) Run(ctx context.Context) error {
	listenErrors := make(chan error, 1)

	go func() {
		s.logger.Info("server listening", "addr", s.httpServer.Addr)
		listenErrors <- s.httpServer.ListenAndServe()
	}()

	select {
	case err := <-listenErrors:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		return s.shutdown()
	}
}

func (s *Server) shutdown() error {
	s.logger.Info("shutdown signal received, draining connections")

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		return err
	}

	s.logger.Info("server stopped cleanly")
	return nil
}
