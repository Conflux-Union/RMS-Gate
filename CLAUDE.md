# Repository Guidelines

## Project Overview

RMS-Gate is a Gate-based Minecraft proxy plugin written in Go. The root module contains RMS-specific features such as whitelist checks, dynamic backend lifecycle control, permission handling, and backend load balancing. The repo also vendors an embedded Gate fork under `pkg/gate/`; that nested module is where proxy-core behavior is patched when upstream handling is wrong. Current local proxy guidance: keep fixes narrow, especially around plugin-message forwarding. The Axiom compatibility fix lives in `pkg/gate/pkg/edition/java/proxy/` and only changes backend play-state `minecraft:register` / `minecraft:unregister` forwarding so those packets reach the client.

## Project Structure

- `main.go`: plugin entrypoint, event wiring, commands
- `internal/config/`: JSON config models and loading
- `internal/whitelist/`: remote whitelist API checks
- `internal/permission/`: permission caching and lookup
- `internal/mcsmanager/`: MCSManager API client
- `internal/dynamicserver/`: auto-start and auto-stop backend control
- `internal/loadbalancer/`: backend selection, health checks, server metadata
- `pkg/gate/`: embedded Gate fork as a separate Go module
- `CLAUDE.md`: repository guidance for future changes

## Build & Test

- `go build -o rms-gate .`: build the root binary
- `go run .`: run locally
- `go vet ./...`: quick static sanity check
- `cd pkg/gate && GOCACHE=/tmp/go-cache go test ./pkg/edition/java/proxy`: validate embedded proxy changes
- `GOCACHE=/tmp/go-cache go build .`: confirm the root module still compiles against the local Gate fork

Manual verification still needs a live whitelist API, configured backends, and a real Minecraft client.

## Coding Style

- Keep changes surgical; fix root causes instead of stacking special cases
- Use `logr.Logger` for structured logs
- Error logs: `log.Error(err, "message", "key", value)`
- Info logs: `log.Info("message", "key", value)`
- Keep config in `rms-gate-config.json`
- Prefer small functions with obvious ownership and minimal branching

## Testing Guidelines

- If only RMS plugin code changes, start with the smallest relevant package or build
- If `pkg/gate/` changes, test inside the nested module first
- For Java proxy compatibility issues, inspect plugin-message routing before touching login flow; mods like Axiom key off backend-declared play channels and will fail if `minecraft:register` never reaches the client
- Do not leave temporary debug logging or speculative bridge code in the final patch unless it is the actual fix

## Commit Convention

Use conventional commits: `fix:`, `feat:`, `refactor:`, `docs:`
