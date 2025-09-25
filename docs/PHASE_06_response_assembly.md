# Phase 6 — Response Assembly (Buffers)

Goal: assemble SIRI JSON/XML responses with pre-sized buffers, templates, and timestamp insertion.

## response.go
- [ ] Template chunks: define per-format chunk tables analogous to CachedMessageTemplateChunks.js
- [ ] Compute general piece lengths and per-journey lengths
- [ ] getLengthOfResponse(): sum chunks for requestedTripKeys; account for json commas; add SituationExchange length if included
- [ ] initResponseBuffer(): allocate `make([]byte, overallLen)` and write header chunks; leave placeholders for timestamps; copy ValidUntil; write start of trips data
- [ ] pipeTrip(): write journey start, copy journey bytes; write MonitoredCall (empty or non-empty); write OnwardCalls (empty/non-empty respecting limits); handle json comma trim
- [ ] pipeSituationExchange(): close trips array; append SituationExchange if included; write closing chunks
- [ ] applyTimestamps(): format two ResponseTimestamp values and copy into offsets
- [ ] Memoize success or error tuple and drain waiting callbacks

## Offsets and correctness
- [ ] Maintain firstTimestampOffset/secondTimestampOffset per deliveryType+format
- [ ] Verify final offset equals buffer length; log if mismatch

## References
- Node: lib/caching/ResponseBuilder.js, CachedMessageTemplateChunks.js
