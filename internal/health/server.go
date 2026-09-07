package health

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"sync"
)

// Status describes module readiness for the dedicated health listener.
type Status struct {
	OK       bool   `json:"ok"`
	Status   string `json:"status"`
	DLNA     string `json:"dlna"`
	Reason   string `json:"reason,omitempty"`
	MediaPath string `json:"media_path,omitempty"`
}

// Server serves JSON health on a dedicated HTTP port.
type Server struct {
	addr   string
	check  func(context.Context) Status
	http   *http.Server
	mu     sync.Mutex
	closed bool
}

// NewServer creates a health HTTP server bound to addr.
func NewServer(addr string, check func(context.Context) Status) *Server {
	return &Server{addr: addr, check: check}
}

// Start listens and serves /health until Stop is called.
func (s *Server) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/healthz", s.handleHealth)
	var lc net.ListenConfig
	lis, err := lc.Listen(context.Background(), "tcp", s.addr)
	if err != nil {
		return fmt.Errorf("listen health %s: %w", s.addr, err)
	}
	s.http = &http.Server{Handler: mux}
	go func() {
		slog.Info("media-dlna health HTTP started", "addr", lis.Addr().String())
		if err := s.http.Serve(lis); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("media-dlna health HTTP error", "error", err)
		}
	}()
	return nil
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	status := s.check(r.Context())
	w.Header().Set("Content-Type", "application/json")
	if !status.OK {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
	_ = json.NewEncoder(w).Encode(status)
}

// Stop shuts down the health HTTP server.
func (s *Server) Stop(ctx context.Context) error {
	s.mu.Lock()
	if s.closed || s.http == nil {
		s.mu.Unlock()
		return nil
	}
	s.closed = true
	srv := s.http
	s.mu.Unlock()
	return srv.Shutdown(ctx)
}
