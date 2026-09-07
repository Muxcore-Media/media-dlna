package module

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/Muxcore-Media/core/pkg/contracts"
)

// Config bootstraps a MuxCore sidecar module.
type Config struct {
	Module   contracts.Module
	Insecure bool
}

// Run executes the standard module lifecycle until interrupted.
func Run(cfg Config) error {
	if cfg.Module == nil {
		return errModuleRequired
	}
	ctx := context.Background()
	info := cfg.Module.Info()
	if id := os.Getenv("MUXCORE_MODULE_ID"); id != "" {
		info.ID = id
	}
	slog.Info("module starting",
		"id", info.ID,
		"name", info.Name,
		"version", info.Version,
		"capabilities", info.Capabilities,
	)
	if err := cfg.Module.Init(ctx); err != nil {
		return err
	}
	if err := cfg.Module.Start(ctx); err != nil {
		_ = cfg.Module.Stop(ctx)
		return err
	}
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	return cfg.Module.Stop(ctx)
}

var errModuleRequired = &moduleError{"module is required"}

type moduleError struct{ msg string }

func (e *moduleError) Error() string { return e.msg }
