-- 0004_booking_passenger_active.sql
-- Allow freeing seats by deactivating passengers instead of deleting rows.

alter table booking_passengers
  add column if not exists is_active boolean not null default true;

alter table booking_passengers
  drop constraint if exists booking_passengers_trip_id_seat_id_key;

create unique index if not exists booking_passengers_trip_seat_active_uq
  on booking_passengers (trip_id, seat_id)
  where is_active = true;
