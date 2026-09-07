# media-dlna

MuxCore sidecar that serves household library media over **DLNA/UPnP** for smart TVs and LAN renderers. Wired into the MVP stack via compose profile `dlna` and `MVP_ENABLE_MEDIA_DLNA=1` on run-host.

Implements umbrella [#139](https://github.com/Muxcore-Media/umbrella/issues/139).

## Ports

| Port | Env | Purpose |
|------|-----|---------|
| 9750 | `DLNA_HTTP_ADDR` | DLNA/UPnP HTTP (SSDP + ContentDirectory) |
| 9751 | `DLNA_GRPC_ADDR` | Module gRPC (mesh registration, health service) |
| 8751 | `DLNA_HEALTH_HTTP_ADDR` | Dedicated JSON health HTTP (`/health`, `/healthz`) |

Host mappings in mvp: `MEDIA_DLNA_HTTP_PORT`, `MEDIA_DLNA_GRPC_PORT`, `MEDIA_DLNA_HEALTH_PORT`.

## Configure

| Env | Default | Purpose |
|-----|---------|---------|
| `DLNA_HTTP_ADDR` | `:9750` | DLNA HTTP listen address |
| `DLNA_GRPC_ADDR` | `:9751` | gRPC listen address |
| `DLNA_HEALTH_HTTP_ADDR` | `:8751` | Health HTTP listen address |
| `DLNA_MEDIA_PATH` | _(none)_ | Library root to browse and serve |
| `DLNA_FRIENDLY_NAME` | `MuxCore DLNA` | UPnP device friendly name |
| `DLNA_PROBE_CACHE_PATH` | _(none)_ | JSON ffprobe cache file (mvp uses `/data/dlna/ffprobe-cache.json`) |
| `MUXCORE_MODULE_ID` | `media-dlna` | Module ID override |
| `MUXCORE_GRPC_ADDR` | _(required for mesh)_ | Core mesh gRPC address |
| `MUXCORE_INSECURE_DISABLE_TLS` | — | Dev-only plaintext mesh gRPC |

See `.env.example` for a local template.

## Missing media path (soft-empty)

When `DLNA_MEDIA_PATH` is **unset**, empty, or **not an existing directory**, the module still starts and registers on the mesh:

- gRPC listens on `DLNA_GRPC_ADDR`
- Health HTTP returns **200** with `{"status":"ok","dlna":"inactive","reason":"media path missing or not a directory"}`
- SSDP/DLNA browsing is **not** started (no LAN announcements)

Set `DLNA_MEDIA_PATH` to a readable library directory to activate DLNA serving. This matches the soft-empty pattern used by other optional MuxCore modules (e.g. unconfigured indexer peers).

## Build

```bash
go test ./...
go build -o media-dlna ./cmd/module
make docker
```

## MVP integration

```bash
# Compose (profile dlna)
docker compose --profile dlna up -d media-dlna

# run-host
MVP_ENABLE_MEDIA_DLNA=1 ./run-host.sh
```

Library path defaults: compose mounts `library-data` at `/library`; run-host uses `$LIBRARY_ROOT` when `DLNA_MEDIA_PATH` is unset.

## Capabilities

- `media.dlna` — DLNA/UPnP media server
- `settings` — reserved for future module settings surface

## Out of scope

- Apple TV / living-room hardware validation (umbrella#2)
- DLNA device management UI
- mvp compose wiring (already on tip via umbrella#137 / mvp#86)
