# XLSX Compatibility Mapping

## Goal
Dictionary for XLSX migration into canonical schema, keeping UUID internal keys and using `legacy_*` external IDs.

## Sheet mapping

| XLSX sheet | Canonical destination | Notes |
|---|---|---|
| routes | routes | `route_id -> routes.legacy_route_code` |
| stops | route_stops | `stop_id -> route_stops.legacy_stop_code` |
| trips | trips | `trip_id -> trips.legacy_trip_code` |
| trip_stops | trip_stops | `trip_stop_id -> trip_stops.legacy_trip_stop_code` |
| available_segments | segment_fares (derived) | Compatibility layer consumes/exports this shape |
| bookings | bookings + payments | `booking_id -> bookings.legacy_booking_code` |
| passengers | booking_passengers | `passenger_id -> booking_passengers.legacy_passenger_code` |
| manifest_data | derived view | Read model generated from bookings/passengers/payments |

## Status and enum rules

- `ATIVO` (XLSX) in trips maps to `SCHEDULED` (`trip_status`).
- Booking status allowed: `PENDING`, `CONFIRMED`, `CANCELLED`, `EXPIRED`.
- Passenger status allowed: `RESERVED`, `BOARDED`, `NO_SHOW`, `CANCELLED`.
- Payment method allowed: `PIX`, `CARD`, `CASH`, `TRANSFER`, `OTHER`.
- Payment status allowed: `PENDING`, `PAID`, `FAILED`, `REFUNDED`, `CANCELLED`.

## Date/time and numeric rules

- `trip_date` is normalized to UTC date for compatibility views.
- `depart_time` is treated as `HH:MM` and combined with trip departure date for `trip_stops.depart_at`.
- Monetary fields are numeric and must be non-negative.
- Empty XLSX columns are ignored/discarded by design.

## Operational assumptions

- Canonical DB remains operational source of truth.
- Compatibility views expose XLSX shape for reporting and migration checks.
- Migration pipeline uses batch lifecycle: `UPLOADED -> VALIDATED -> PROMOTED`.
