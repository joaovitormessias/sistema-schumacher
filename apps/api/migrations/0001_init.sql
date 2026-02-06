-- 0001_init.sql
-- Core schema for Schumacher Turismo

create extension if not exists pgcrypto;

-- Enums
create type booking_status as enum ('PENDING','CONFIRMED','CANCELLED','EXPIRED');
create type payment_status as enum ('PENDING','PAID','FAILED','REFUNDED','CANCELLED');
create type payment_method as enum ('PIX','CARD','CASH','TRANSFER','OTHER');
create type trip_status as enum ('SCHEDULED','IN_PROGRESS','COMPLETED','CANCELLED');
create type passenger_status as enum ('RESERVED','BOARDED','NO_SHOW','CANCELLED');
create type booking_source as enum ('WHATSAPP','APP','MANUAL');
create type deposit_type as enum ('PERCENT','FIXED');

-- Roles and users
create table if not exists roles (
  id uuid primary key default gen_random_uuid(),
  name text not null unique,
  created_at timestamptz not null default now()
);

create table if not exists user_profiles (
  id uuid primary key,
  email text,
  full_name text,
  created_at timestamptz not null default now()
);

create table if not exists user_roles (
  user_id uuid not null,
  role_id uuid not null references roles(id) on delete cascade,
  created_at timestamptz not null default now(),
  primary key (user_id, role_id)
);

-- Fleet
create table if not exists buses (
  id uuid primary key default gen_random_uuid(),
  name text not null,
  plate text,
  capacity int not null,
  seat_map_name text,
  is_active boolean not null default true,
  created_at timestamptz not null default now()
);

create table if not exists seat_types (
  id uuid primary key default gen_random_uuid(),
  name text not null,
  extra_price numeric(10,2) not null default 0,
  is_active boolean not null default true,
  created_at timestamptz not null default now()
);

create table if not exists bus_seats (
  id uuid primary key default gen_random_uuid(),
  bus_id uuid not null references buses(id) on delete cascade,
  seat_number int not null,
  seat_type_id uuid references seat_types(id),
  is_active boolean not null default true,
  created_at timestamptz not null default now(),
  unique (bus_id, seat_number)
);

create table if not exists drivers (
  id uuid primary key default gen_random_uuid(),
  name text not null,
  document text,
  phone text,
  is_active boolean not null default true,
  created_at timestamptz not null default now()
);

-- Routes and trips
create table if not exists routes (
  id uuid primary key default gen_random_uuid(),
  name text not null,
  origin_city text not null,
  destination_city text not null,
  is_active boolean not null default true,
  created_at timestamptz not null default now()
);

create table if not exists route_stops (
  id uuid primary key default gen_random_uuid(),
  route_id uuid not null references routes(id) on delete cascade,
  city text not null,
  stop_order int not null,
  eta_offset_minutes int,
  notes text,
  created_at timestamptz not null default now(),
  unique (route_id, stop_order)
);

create table if not exists fares (
  id uuid primary key default gen_random_uuid(),
  name text not null,
  currency text not null default 'BRL',
  total_amount numeric(10,2) not null,
  deposit_type deposit_type not null default 'PERCENT',
  deposit_value numeric(10,2) not null,
  remainder_due_hours int,
  cancellation_policy jsonb,
  created_at timestamptz not null default now()
);

create table if not exists trips (
  id uuid primary key default gen_random_uuid(),
  route_id uuid not null references routes(id),
  bus_id uuid not null references buses(id),
  driver_id uuid references drivers(id),
  fare_id uuid references fares(id),
  departure_at timestamptz not null,
  arrival_at timestamptz,
  status trip_status not null default 'SCHEDULED',
  pair_trip_id uuid,
  notes text,
  created_at timestamptz not null default now()
);

-- Bookings and passengers
create table if not exists bookings (
  id uuid primary key default gen_random_uuid(),
  trip_id uuid not null references trips(id) on delete cascade,
  status booking_status not null default 'PENDING',
  source booking_source not null default 'WHATSAPP',
  total_amount numeric(10,2) not null default 0,
  deposit_amount numeric(10,2) not null default 0,
  remainder_amount numeric(10,2) not null default 0,
  expires_at timestamptz,
  created_by uuid,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create table if not exists booking_passengers (
  id uuid primary key default gen_random_uuid(),
  booking_id uuid not null references bookings(id) on delete cascade,
  trip_id uuid not null references trips(id) on delete cascade,
  name text not null,
  document text,
  phone text,
  seat_id uuid not null references bus_seats(id),
  status passenger_status not null default 'RESERVED',
  created_at timestamptz not null default now(),
  unique (trip_id, seat_id)
);

-- Payments
create table if not exists payments (
  id uuid primary key default gen_random_uuid(),
  booking_id uuid not null references bookings(id) on delete cascade,
  amount numeric(10,2) not null,
  method payment_method not null,
  status payment_status not null default 'PENDING',
  provider text,
  provider_ref text,
  paid_at timestamptz,
  created_by uuid,
  metadata jsonb,
  created_at timestamptz not null default now()
);

create table if not exists payment_events (
  id uuid primary key default gen_random_uuid(),
  payment_id uuid references payments(id) on delete cascade,
  event text not null,
  payload jsonb,
  created_at timestamptz not null default now()
);

create table if not exists checkins (
  id uuid primary key default gen_random_uuid(),
  trip_id uuid not null references trips(id) on delete cascade,
  booking_passenger_id uuid not null references booking_passengers(id) on delete cascade,
  checked_in_at timestamptz not null default now(),
  method text,
  notes text
);

-- Indexes
create index if not exists idx_bookings_trip on bookings(trip_id);
create index if not exists idx_booking_passengers_booking on booking_passengers(booking_id);
create index if not exists idx_payments_booking on payments(booking_id);
create index if not exists idx_routes_origin on routes(origin_city);
create index if not exists idx_routes_destination on routes(destination_city);
