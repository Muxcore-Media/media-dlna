package dlna

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"path/filepath"

	"github.com/anacrolix/dms/dlna/dms"
)

// Server wraps anacrolix/dms for MuxCore library paths.
type Server struct {
	listener     net.Listener
	friendlyName string
	mediaPath    string
	probeCache   dms.Cache
	server       *dms.Server
}

// NewServer prepares a DLNA server for the given media root.
func NewServer(addr, friendlyName, mediaPath string, probeCache dms.Cache) (*Server, error) {
	abs, err := filepath.Abs(mediaPath)
	if err != nil {
		return nil, err
	}
	var lc net.ListenConfig
	lis, err := lc.Listen(context.Background(), "tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("listen DLNA HTTP %s: %w", addr, err)
	}
	return &Server{
		listener:     lis,
		friendlyName: friendlyName,
		mediaPath:    abs,
		probeCache:   probeCache,
	}, nil
}

// Start initializes SSDP and begins serving DLNA HTTP.
func (s *Server) Start() error {
	s.server = &dms.Server{
		HTTPConn:       s.listener,
		FriendlyName:   s.friendlyName,
		RootObjectPath: s.mediaPath,
		FFProbeCache:   s.probeCache,
		NoTranscode:    true,
		Logger:         slog.Default(),
	}
	if err := s.server.Init(); err != nil {
		return err
	}
	go func() {
		slog.Info("media-dlna DLNA HTTP started",
			"addr", s.listener.Addr().String(),
			"media_path", s.mediaPath,
		)
		if err := s.server.Run(); err != nil {
			slog.Error("media-dlna DLNA serve error", "error", err)
		}
	}()
	return nil
}

// Stop closes the DLNA server and SSDP announcements.
func (s *Server) Stop() error {
	if s.server == nil {
		return nil
	}
	return s.server.Close()
}

// MediaPathReady reports whether the configured library path exists.
func MediaPathReady(path string) bool {
	path = filepath.Clean(path)
	if path == "" || path == "." {
		return false
	}
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
