# Phase 7 — HTTP API, Auth, Rate Limiting

Goal: expose SIRI endpoints with auth and proper response streaming.

## handler_monitoring.go
- [ ] Routes: /api/siri/vehicle-monitoring.(json|xml), /api/siri/stop-monitoring.(json|xml)
- [ ] Parse query params to lower case; store originals under _keys
- [ ] Content-Type: application/json or application/xml
- [ ] Call converter service; stream []byte directly; set Content-Length
- [ ] Error path: return SIRI ErrorCondition payloads with appropriate HTTP status

## auth.go
- [ ] Modes: openAccess or API key list
- [ ] 403 handling via ErrorCondition message

## rate limiting
- [ ] Apply lightweight limiter; on overload return 503 with ErrorCondition text "Service Unavailable: Back-end server is at capacity."

## health.go
- [ ] /api/health: report last GTFS-RT timestamp, GTFS indices loaded, OK
- [ ] /metrics: promhttp handler

## References
- Server repo: routes/monitoringCallController.js, src/services/AuthorizationService.js
