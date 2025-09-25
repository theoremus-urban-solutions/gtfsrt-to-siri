# Phase 01 — Bootstrap & Scaffolding (Library-only)

Goal: runnable Go library with ZIP-only GTFS static ingestion, logging, config, and a minimal oneshot CLI in cmd/ for quick testing.

## Tasks
- [ ] Create go.mod (module name TBD) and minimal main.go (package main)
- [ ] Create CLI entry: ./golang/cmd/gtfsrt-to-siri/main.go
  - [ ] Flags: `-mode=oneshot`, `-format=json|xml`, `-call=vm|sm`, plus quick-testing params (lineref, directionref, etc.)
  - [ ] In oneshot mode load GTFS static from ZIP URL (config) and print response to stdout
- [ ] Implement logging init (structured logger)
- [ ] Implement config.go
  - [ ] Define Config struct (gtfs, gtfsrt, converter)
  - [ ] Load ./golang/config.yml
  - [ ] Validate with go-playground/validator
- [ ] Remove HTTP server and HTTP handlers (out of scope for library)
- [ ] Minimal README for using the library and running the oneshot CLI

## References
- guidelines.md: Project Layout, Entry Point
- go-requirements.md: Layout section (flat + cmd/gtfsrt-to-siri)
