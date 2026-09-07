package internal

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestModuleSoftEmptyWithoutMediaPath(t *testing.T) {
	t.Setenv("DLNA_MEDIA_PATH", "")
	t.Setenv("DLNA_HTTP_ADDR", ":0")
	t.Setenv("DLNA_GRPC_ADDR", ":0")
	t.Setenv("DLNA_HEALTH_HTTP_ADDR", ":0")

	m := NewModule(Config{})
	ctx := context.Background()
	if err := m.Init(ctx); err != nil {
		t.Fatal(err)
	}
	if err := m.Start(ctx); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = m.Stop(ctx) })

	if m.dlnaActive() {
		t.Fatal("expected DLNA inactive without media path")
	}
	if m.dlnaSrv != nil {
		t.Fatal("expected no DLNA server when media path missing")
	}
	if err := m.Health(ctx); err != nil {
		t.Fatalf("health: %v", err)
	}
	status := m.healthStatus(ctx)
	if !status.OK || status.DLNA != "inactive" {
		t.Fatalf("unexpected status: %+v", status)
	}
}

func TestModuleStartsDLNAWithLibrary(t *testing.T) {
	lib := t.TempDir()
	if err := os.WriteFile(filepath.Join(lib, "movie.mkv"), []byte("fixture"), 0o644); err != nil {
		t.Fatal(err)
	}
	cachePath := filepath.Join(t.TempDir(), "ffprobe-cache.json")

	m := NewModule(Config{
		HTTPAddr:       ":0",
		GRPCAddr:       ":0",
		HealthHTTPAddr: ":0",
		MediaPath:      lib,
		FriendlyName:   "Test DLNA",
		ProbeCachePath: cachePath,
	})
	ctx := context.Background()
	if err := m.Init(ctx); err != nil {
		t.Fatal(err)
	}
	if err := m.Start(ctx); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = m.Stop(ctx) })

	if !m.dlnaActive() {
		t.Fatal("expected DLNA active with library path")
	}
	if m.dlnaSrv == nil {
		t.Fatal("expected DLNA server")
	}
}

func TestModuleInfo(t *testing.T) {
	m := NewModule(Config{})
	info := m.Info()
	if info.ID != "media-dlna" {
		t.Errorf("id: %s", info.ID)
	}
	if info.HTTPAddr != defaultHTTPAddr {
		t.Errorf("http addr: %s", info.HTTPAddr)
	}
	if len(info.Capabilities) == 0 || info.Capabilities[0] != "media.dlna" {
		t.Errorf("capabilities: %v", info.Capabilities)
	}
}

func TestLoadConfigDefaults(t *testing.T) {
	t.Setenv("DLNA_HTTP_ADDR", "")
	t.Setenv("DLNA_GRPC_ADDR", "")
	t.Setenv("DLNA_HEALTH_HTTP_ADDR", "")
	cfg := LoadConfig()
	if cfg.HTTPAddr != defaultHTTPAddr || cfg.GRPCAddr != defaultGRPCAddr || cfg.HealthHTTPAddr != defaultHealthHTTPAddr {
		t.Fatalf("defaults: %+v", cfg)
	}
	if cfg.FriendlyName != defaultFriendlyName {
		t.Fatalf("friendly name: %s", cfg.FriendlyName)
	}
}
