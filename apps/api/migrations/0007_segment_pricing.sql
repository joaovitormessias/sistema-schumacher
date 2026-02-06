-- 0007_segment_pricing.sql
-- Segment-based pricing and trip stops.

create extension if not exists btree_gist;

create table if not exists trip_stops (
  id uuid primary key default gen_random_uuid(),
  trip_id uuid not null references trips(id) on delete cascade,
  route_stop_id uuid not null references route_stops(id) on delete cascade,
  stop_order int not null,
  arrive_at timestamptz,
  depart_at timestamptz,
  created_at timestamptz not null default now(),
  unique (trip_id, stop_order),
  unique (trip_id, route_stop_id)
);

create table if not exists segment_fares (
  id uuid primary key default gen_random_uuid(),
  scope pricing_scope not null default 'ROUTE',
  scope_id uuid not null,
  from_route_stop_id uuid not null references route_stops(id) on delete cascade,
  to_route_stop_id uuid not null references route_stops(id) on delete cascade,
  base_amount numeric(10,2) not null,
  currency text not null default 'BRL',
  valid_from timestamptz,
  valid_to timestamptz,
  priority int not null default 100,
  is_active boolean not null default true,
  created_at timestamptz not null default now(),
  check (from_route_stop_id <> to_route_stop_id)
);

create index if not exists segment_fares_lookup_idx
  on segment_fares (scope, scope_id, from_route_stop_id, to_route_stop_id, is_active, priority);

do $$ begin
  create type booking_fare_mode as enum ('AUTO','FIXED','MANUAL');
exception
  when duplicate_object then null;
end $$;

alter table booking_passengers
  add column if not exists board_stop_id uuid references trip_stops(id) on delete restrict,
  add column if not exists alight_stop_id uuid references trip_stops(id) on delete restrict,
  add column if not exists board_stop_order int,
  add column if not exists alight_stop_order int,
  add column if not exists fare_mode booking_fare_mode,
  add column if not exists fare_amount_calc numeric(10,2) not null default 0,
  add column if not exists fare_amount_final numeric(10,2) not null default 0,
  add column if not exists fare_snapshot jsonb;

alter table booking_passengers
  add constraint booking_passengers_stop_order_check
  check (board_stop_order is null or alight_stop_order is null or board_stop_order < alight_stop_order);

drop index if exists booking_passengers_trip_seat_active_uq;

alter table booking_passengers
  add constraint booking_passengers_no_overlap
  exclude using gist (
    trip_id with =,
    seat_id with =,
    int4range(board_stop_order, alight_stop_order, '[)') with &&
  )
  where (is_active = true and board_stop_order is not null and alight_stop_order is not null);
