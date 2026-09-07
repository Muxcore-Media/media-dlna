# AGENTS.md — media-dlna

MuxCore DLNA/UPnP sidecar (`media-dlna`). MVP ports and compose profile: [`../mvp/PORTS.md`](../mvp/PORTS.md) (when checked out in workspace).

## Module identity

| Field | Value |
|-------|-------|
| Directory | `media-dlna` |
| Capabilities | `media.dlna`, `settings` (see `muxcore.json`) |
| Contracts | none yet |

## Agent rules

- Modules run as gRPC sidecars; capabilities are the security boundary.
- TLS required in production (`MUXCORE_INSECURE_DISABLE_TLS` is dev-only).
- Honor the env contract exactly — defaults must stay `:9750` / `:9751` / `:8751`.
- Match existing Go patterns; run `gofmt` and `go test ./...` before finishing.
- Fixture/offline tests only; do not require LAN SSDP or physical renderers in CI.

## Env contract

| Variable | Default | Notes |
|----------|---------|-------|
| `DLNA_HTTP_ADDR` | `:9750` | DLNA/UPnP HTTP |
| `DLNA_GRPC_ADDR` | `:9751` | Module gRPC |
| `DLNA_HEALTH_HTTP_ADDR` | `:8751` | Dedicated health HTTP |
| `DLNA_MEDIA_PATH` | — | Library root; see soft-empty below |
| `DLNA_FRIENDLY_NAME` | `MuxCore DLNA` | UPnP friendly name |
| `DLNA_PROBE_CACHE_PATH` | — | ffprobe JSON cache path |

Shared mesh: `MUXCORE_GRPC_ADDR`, `MUXCORE_MODULE_ID`, `MUXCORE_INSECURE_DISABLE_TLS`.

## Missing media path (soft-empty)

If `DLNA_MEDIA_PATH` is unset or not a directory:

1. Module **starts** (gRPC + health HTTP).
2. DLNA/SSDP **does not** start.
3. Health JSON reports `dlna: inactive` with reason `media path missing or not a directory`.
4. `Health(ctx)` for the module SDK remains OK (process is healthy; library is simply not configured).

Do **not** fail `Init`/`Start` solely because the media path is missing — operators enable DLNA before the library volume exists.

## Build

```bash
cd media-dlna
go test ./...
go build -o media-dlna ./cmd/module
```

## Related

- Umbrella [#139](https://github.com/Muxcore-Media/umbrella/issues/139) — implement module
- Umbrella [#137](https://github.com/Muxcore-Media/umbrella/issues/137) / mvp#86 — compose `dlna` profile wiring
