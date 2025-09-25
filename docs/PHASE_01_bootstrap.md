# Phase 01 — Bootstrap & Scaffolding

Goal: runnable Go skeleton with config, logging, HTTP server, health/metrics, graceful shutdown, plus a standard CLI entry in cmd/.

## Tasks
- [ ] Create go.mod (module name TBD) and minimal main.go (package main)
- [ ] Create CLI entry: ./golang/cmd/gtfsrt-to-siri/main.go
  - [ ] Flags: `-mode=server|oneshot`, `-format=json|xml`, `-call=vm|sm`, param flags for quick testing
  - [ ] In server mode call startServer(); in oneshot mode load feeds once and print response to stdout
- [ ] Implement logging init (structured logger)
- [ ] Implement config.go
  - [ ] Define Config struct (server, gtfs, gtfsrt, converter)
  - [ ] Load ./golang/config.yml
  - [ ] Validate with go-playground/validator
- [ ] Implement server bootstrap in main.go
  - [ ] startServer(): registers /api/health and /metrics placeholders
  - [ ] handleGracefulShutdown(): SIGINT/SIGTERM, context timeouts, wg
- [ ] Health handler
  - [ ] Returns JSON with service status and last GTFS-RT timestamp (placeholder null)
- [ ] Metrics stub (Prometheus)
- [ ] Write minimal README for running the service and CLI

## References
- guidelines.md: Project Layout, Entry Point, Config, Graceful Shutdown, Monitoring
- go-requirements.md: Layout section (flat + cmd/gtfsrt-to-siri)
