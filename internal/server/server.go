package server

import (
	"context"
	"crypto/tls"
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
	certFile   string
	keyFile    string
}

// New builds the HTTP server. playgroundHandler is optional: when non-nil it is
// served unauthenticated at GET /playground for local exploration; pass nil to
// not expose it at all. pictures, when non-nil, serves pet pictures over the same
// origin and TLS at GET /pictures/{objectKey...}.
func New(cfg config.Config, logger *slog.Logger, authenticator *auth.Authenticator, graphqlHandler, playgroundHandler http.Handler, pictures pictureReader) *Server {
	if authenticator == nil {
		panic("server: authenticator is required")
	}
	if graphqlHandler == nil {
		panic("server: graphql handler is required")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", handleHealth)

	requireAuth := auth.BasicAuth(authenticator)
	mux.Handle("/graphql", requireAuth(graphqlHandler))
	if playgroundHandler != nil {
		mux.Handle("GET /playground", playgroundHandler)
	}
	if pictures != nil {
		mux.Handle("GET /pictures/{objectKey...}", newPictureHandler(pictures, logger))
	}

	httpServer := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
		TLSConfig:         &tls.Config{MinVersion: tls.VersionTLS12},
	}

	return &Server{httpServer: httpServer, logger: logger, certFile: cfg.TLSCertFile, keyFile: cfg.TLSKeyFile}
}

func (s *Server) Run(ctx context.Context) error {
	listenErrors := make(chan error, 1)

	go func() {
		s.logger.Info("server listening", "addr", s.httpServer.Addr)
		listenErrors <- s.httpServer.ListenAndServeTLS(s.certFile, s.keyFile)
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
