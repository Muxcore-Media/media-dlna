package internal

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"strings"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	"github.com/Muxcore-Media/core/pkg/contracts"
	"github.com/Muxcore-Media/core/sdk/go/client"
	modulesdk "github.com/Muxcore-Media/core/sdk/go/module"
	manifest "github.com/Muxcore-Media/media-dlna"
	"github.com/Muxcore-Media/media-dlna/internal/dlna"
	dlnahealth "github.com/Muxcore-Media/media-dlna/internal/health"
	"github.com/Muxcore-Media/media-dlna/internal/probe"
)

// Module implements the MuxCore media-dlna sidecar.
type Module struct { //nolint:govet // lifecycle fields grouped for readability
	mu sync.RWMutex

	cfg Config

	grpcAddr string
	grpcSrv  *grpc.Server
	lis      net.Listener

	healthSrv *dlnahealth.Server
	dlnaSrv   *dlna.Server
	probe     *probe.Cache

	mc *client.Client

	started bool
}

// NewModule builds a module from config, reading env when fields are empty.
func NewModule(cfg Config) *Module {
	env := LoadConfig()
	if cfg.ID == "" {
		cfg.ID = env.ID
	}
	if cfg.HTTPAddr == "" {
		cfg.HTTPAddr = env.HTTPAddr
	}
	if cfg.GRPCAddr == "" {
		cfg.GRPCAddr = env.GRPCAddr
	}
	if cfg.HealthHTTPAddr == "" {
		cfg.HealthHTTPAddr = env.HealthHTTPAddr
	}
	if cfg.MediaPath == "" {
		cfg.MediaPath = env.MediaPath
	}
	if cfg.FriendlyName == "" {
		cfg.FriendlyName = env.FriendlyName
	}
	if cfg.ProbeCachePath == "" {
		cfg.ProbeCachePath = env.ProbeCachePath
	}
	return &Module{cfg: cfg}
}

func (m *Module) Info() contracts.ModuleInfo {
	return contracts.ModuleInfo{
		ID:             m.cfg.ID,
		Name:           "MuxCore DLNA",
		Version:        modulesdk.ManifestVersion(manifest.ManifestJSON),
		Roles:          []string{"media"},
		Description:    "DLNA/UPnP media server for household library paths",
		Author:         "MuxCore",
		Capabilities:   []string{"media.dlna", "settings"},
		MinCoreVersion: "0.5.8",
		HTTPAddr:       m.cfg.HTTPAddr,
	}
}

func (m *Module) Init(ctx context.Context) error {
	cache, err := probe.NewCache(m.cfg.ProbeCachePath)
	if err != nil {
		return fmt.Errorf("ffprobe cache: %w", err)
	}
	m.probe = cache

	var lc net.ListenConfig
	lis, err := lc.Listen(ctx, "tcp", m.cfg.GRPCAddr)
	if err != nil {
		return fmt.Errorf("listen gRPC %s: %w", m.cfg.GRPCAddr, err)
	}
	m.lis = lis
	if ta, ok := lis.Addr().(*net.TCPAddr); ok && strings.HasPrefix(m.cfg.GRPCAddr, ":") {
		m.grpcAddr = fmt.Sprintf(":%d", ta.Port)
	} else {
		m.grpcAddr = m.cfg.GRPCAddr
	}

	m.healthSrv = dlnahealth.NewServer(m.cfg.HealthHTTPAddr, m.healthStatus)
	slog.Info("media-dlna initialized",
		"grpc", m.grpcAddr,
		"dlna_http", m.cfg.HTTPAddr,
		"health_http", m.cfg.HealthHTTPAddr,
		"media_path", m.cfg.MediaPath,
		"dlna_active", m.dlnaActive(),
	)
	return nil
}

func (m *Module) Start(ctx context.Context) error {
	m.grpcSrv = grpc.NewServer()
	healthpb.RegisterHealthServer(m.grpcSrv, health.NewServer())
	reflection.Register(m.grpcSrv)
	go func() {
		slog.Info("media-dlna gRPC started", "addr", m.grpcAddr)
		if err := m.grpcSrv.Serve(m.lis); err != nil {
			slog.Error("media-dlna gRPC serve error", "error", err)
		}
	}()

	if err := m.healthSrv.Start(); err != nil {
		return err
	}

	if m.dlnaActive() {
		srv, err := dlna.NewServer(m.cfg.HTTPAddr, m.cfg.FriendlyName, m.cfg.MediaPath, m.probe)
		if err != nil {
			return err
		}
		if err := srv.Start(); err != nil {
			return err
		}
		m.dlnaSrv = srv
	} else {
		slog.Warn("media-dlna: DLNA inactive (media path missing or not a directory)",
			"path", m.cfg.MediaPath,
		)
	}

	go m.dialCore(context.WithoutCancel(ctx))
	m.started = true
	return nil
}

func (m *Module) Stop(ctx context.Context) error {
	if m.dlnaSrv != nil {
		_ = m.dlnaSrv.Stop()
	}
	if m.healthSrv != nil {
		_ = m.healthSrv.Stop(ctx)
	}
	if m.grpcSrv != nil {
		m.grpcSrv.GracefulStop()
	}
	if m.probe != nil {
		if err := m.probe.Save(); err != nil {
			slog.Warn("media-dlna: save ffprobe cache", "error", err)
		}
	}
	m.mu.Lock()
	if m.mc != nil {
		_ = m.mc.Close()
		m.mc = nil
	}
	m.mu.Unlock()
	slog.Info("media-dlna stopped")
	return nil
}

func (m *Module) Health(ctx context.Context) error {
	status := m.healthStatus(ctx)
	if !status.OK {
		return fmt.Errorf("%s", status.Reason)
	}
	return nil
}

func (m *Module) healthStatus(ctx context.Context) dlnahealth.Status {
	_ = ctx
	if !m.started {
		return dlnahealth.Status{OK: false, Status: "starting", DLNA: "inactive", Reason: "not started"}
	}
	if !m.dlnaActive() {
		return dlnahealth.Status{
			OK:        true,
			Status:    "ok",
			DLNA:      "inactive",
			Reason:    "media path missing or not a directory",
			MediaPath: m.cfg.MediaPath,
		}
	}
	return dlnahealth.Status{OK: true, Status: "ok", DLNA: "active", MediaPath: m.cfg.MediaPath}
}

func (m *Module) dlnaActive() bool {
	return dlna.MediaPathReady(m.cfg.MediaPath)
}

func (m *Module) dialCore(ctx context.Context) {
	meshAddr := strings.TrimSpace(os.Getenv("MUXCORE_GRPC_ADDR"))
	if meshAddr == "" {
		return
	}
	insecure := os.Getenv("MUXCORE_INSECURE_DISABLE_TLS") == "true" || os.Getenv("MUXCORE_GRPC_INSECURE") == "true"
	var opts []client.Option
	if insecure {
		opts = append(opts, client.WithInsecure())
	}
	c, err := client.Dial(meshAddr, opts...)
	if err != nil {
		slog.Warn("media-dlna: dial core mesh", "error", err)
		return
	}
	m.mu.Lock()
	if m.mc != nil {
		_ = m.mc.Close()
	}
	m.mc = c
	m.mu.Unlock()
	slog.Info("media-dlna: mesh client ready", "addr", meshAddr)
	_ = ctx
}
