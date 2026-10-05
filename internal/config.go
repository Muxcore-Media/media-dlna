package internal

import (
	"os"
	"strings"
)

const (
	defaultHTTPAddr       = ":9750"
	defaultGRPCAddr       = ":9751"
	defaultHealthHTTPAddr = ":8751"
	defaultFriendlyName   = "MuxCore DLNA"
)

// Config holds runtime settings for the media-dlna module.
type Config struct {
	ID             string
	HTTPAddr       string
	GRPCAddr       string
	HealthHTTPAddr string
	MediaPath      string
	FriendlyName   string
	ProbeCachePath string
}

// LoadConfig reads module settings from the environment with MuxCore defaults.
func LoadConfig() Config {
	cfg := Config{
		ID:             strings.TrimSpace(os.Getenv("MUXCORE_MODULE_ID")),
		HTTPAddr:       envOr("DLNA_HTTP_ADDR", defaultHTTPAddr),
		GRPCAddr:       envOr("DLNA_GRPC_ADDR", defaultGRPCAddr),
		HealthHTTPAddr: envOr("DLNA_HEALTH_HTTP_ADDR", defaultHealthHTTPAddr),
		MediaPath:      strings.TrimSpace(os.Getenv("DLNA_MEDIA_PATH")),
		FriendlyName:   envOr("DLNA_FRIENDLY_NAME", defaultFriendlyName),
		ProbeCachePath: strings.TrimSpace(os.Getenv("DLNA_PROBE_CACHE_PATH")),
	}
	if cfg.ID == "" {
		cfg.ID = "media-dlna"
	}
	return cfg
}

func envOr(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}
