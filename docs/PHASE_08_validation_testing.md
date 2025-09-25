# Phase 8 — Validation, Testing, and Parity

Goal: validate inputs/outputs (optional), and ensure parity with Node implementation through tests.

## Validators (optional)
- [ ] GTFS-RT CI validation using MobilityData validator (Java CLI) on recorded feeds
- [ ] Runtime sanity checks (dev): monotonic timestamps, required fields
- [ ] SIRI XML: XSD validation in CI; JSON: JSON schema validation in dev

## Unit tests (_test.go)
- [ ] Builders: MVJ fields correctness given fixtures
- [ ] Calls: inclusion/exclusion rules and limits
- [ ] Query selection: StopMonitoring and VehicleMonitoring paths; edge cases
- [ ] Response assembly: computed length = buffer length; timestamps offsets
- [ ] Tracker: states and geometry transitions

## Integration tests
- [ ] Replay sample GTFS + GTFS-RT messages and compare against known-good golden buffers
- [ ] Golden files: store response buffers and compare byte-for-byte

## Load/Concurrency
- [ ] ensure no data races (go test -race)

## Parity Checklist
- [ ] Field-by-field parity (JSON & XML) vs Node for representative trips/stops
- [ ] ErrorCondition payloads match structure/messages
- [ ] Memoization behavior
- [ ] ValidUntil computation parity
