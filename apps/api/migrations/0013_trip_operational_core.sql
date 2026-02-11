-- 0013_trip_operational_core.sql
-- Core operational workflow for charter trips with strict stage gating.

-- ============================================================================
-- ENUMS
-- ============================================================================

do $$ begin
  create type trip_request_source as enum ('EMAIL','SYSTEM');
exception
  when duplicate_object then null;
end $$;

do $$ begin
  create type trip_request_status as enum ('OPEN','IN_REVIEW','APPROVED','REJECTED');
exception
  when duplicate_object then null;
end $$;

do $$ begin
  create type trip_operational_status as enum (
    'REQUESTED',
    'PASSENGERS_READY',
    'ITINERARY_READY',
    'DISPATCH_VALIDATED',
    'AUTHORIZED',
    'IN_PROGRESS',
    'RETURNED',
    'RETURN_CHECKED',
    'SETTLED',
    'CLOSED'
  );
exception
  when duplicate_object then null;
end $$;

do $$ begin
  create type manifest_entry_source as enum ('BOOKING','MANUAL');
exception
  when duplicate_object then null;
end $$;

do $$ begin
  create type manifest_passenger_status as enum ('EXPECTED','BOARDED','NO_SHOW','CANCELLED');
exception
  when duplicate_object then null;
end $$;

do $$ begin
  create type authorization_authority as enum ('ANTT','DETER','EXCEPTIONAL');
exception
  when duplicate_object then null;
end $$;

do $$ begin
  create type authorization_status as enum ('PENDING','ISSUED','REJECTED','EXPIRED');
exception
  when duplicate_object then null;
end $$;

do $$ begin
  create type checklist_stage as enum ('PRE_DEPARTURE','RETURN');
exception
  when duplicate_object then null;
end $$;

do $$ begin
  create type trip_attachment_type as enum (
    'TRIP_REQUEST',
    'AUTHORIZATION',
    'INSURANCE',
    'CHECKLIST',
    'TACHOGRAPH',
    'RECEIPT',
    'FISCAL',
    'DRIVER_REPORT',
    'OTHER'
  );
exception
  when duplicate_object then null;
end $$;

-- ============================================================================
-- TABLES
-- ============================================================================

create table if not exists trip_requests (
  id uuid primary key default gen_random_uuid(),
  route_id uuid references routes(id) on delete set null,
  source trip_request_source not null,
  status trip_request_status not null default 'OPEN',
  requester_name text,
  requester_contact text,
  requested_departure_at timestamptz,
  notes text,
  created_by uuid,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create table if not exists trip_attachments (
  id uuid primary key default gen_random_uuid(),
  trip_id uuid not null references trips(id) on delete cascade,
  attachment_type trip_attachment_type not null,
  storage_bucket text not null default 'trip-documents',
  storage_path text not null,
  file_name text not null,
  mime_type text,
  file_size bigint,
  metadata jsonb,
  uploaded_by uuid,
  uploaded_at timestamptz not null default now(),
  created_at timestamptz not null default now(),
  unique (storage_bucket, storage_path)
);

create table if not exists trip_manifest_entries (
  id uuid primary key default gen_random_uuid(),
  trip_id uuid not null references trips(id) on delete cascade,
  booking_passenger_id uuid references booking_passengers(id) on delete set null,
  passenger_name text not null,
  passenger_document text,
  passenger_phone text,
  source manifest_entry_source not null default 'MANUAL',
  status manifest_passenger_status not null default 'EXPECTED',
  seat_number int,
  is_active boolean not null default true,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create table if not exists trip_authorizations (
  id uuid primary key default gen_random_uuid(),
  trip_id uuid not null references trips(id) on delete cascade,
  authority authorization_authority not null,
  status authorization_status not null default 'PENDING',
  protocol_number text,
  license_number text,
  issued_at timestamptz,
  valid_until timestamptz,
  src_policy_number text,
  src_valid_until date,
  exceptional_deadline_ok boolean not null default true,
  attachment_id uuid references trip_attachments(id) on delete set null,
  notes text,
  created_by uuid,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create table if not exists trip_checklists (
  id uuid primary key default gen_random_uuid(),
  trip_id uuid not null references trips(id) on delete cascade,
  stage checklist_stage not null,
  checklist_data jsonb not null default '{}'::jsonb,
  is_complete boolean not null default false,
  documents_checked boolean not null default false,
  tachograph_checked boolean not null default false,
  receipts_checked boolean not null default false,
  rest_compliance_ok boolean not null default true,
  completed_by uuid,
  completed_at timestamptz,
  notes text,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  unique (trip_id, stage)
);

create table if not exists trip_driver_reports (
  id uuid primary key default gen_random_uuid(),
  trip_id uuid not null references trips(id) on delete cascade unique,
  driver_id uuid references drivers(id) on delete set null,
  odometer_start int,
  odometer_end int,
  fuel_used_liters numeric(10,2),
  incidents text,
  delays text,
  rest_hours numeric(6,2),
  notes text,
  submitted_by uuid,
  submitted_at timestamptz not null default now(),
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  check (odometer_start is null or odometer_start >= 0),
  check (odometer_end is null or odometer_end >= 0),
  check (
    odometer_start is null
    or odometer_end is null
    or odometer_end >= odometer_start
  ),
  check (fuel_used_liters is null or fuel_used_liters >= 0),
  check (rest_hours is null or rest_hours >= 0)
);

create table if not exists trip_receipt_reconciliations (
  id uuid primary key default gen_random_uuid(),
  trip_id uuid not null references trips(id) on delete cascade unique,
  total_receipts_amount numeric(10,2) not null default 0,
  total_approved_expenses numeric(10,2) not null default 0,
  difference numeric(10,2) generated always as (total_approved_expenses - total_receipts_amount) stored,
  receipts_validated boolean not null default false,
  verified_expense_ids uuid[] not null default '{}',
  notes text,
  reconciled_by uuid,
  reconciled_at timestamptz,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  check (total_receipts_amount >= 0),
  check (total_approved_expenses >= 0)
);

-- ============================================================================
-- ALTER EXISTING TABLES
-- ============================================================================

alter table trips
  add column if not exists request_id uuid references trip_requests(id) on delete set null,
  add column if not exists operational_status trip_operational_status not null default 'REQUESTED',
  add column if not exists estimated_km numeric(10,2) not null default 0,
  add column if not exists dispatch_validated_at timestamptz,
  add column if not exists dispatch_validated_by uuid,
  add column if not exists updated_at timestamptz not null default now();

alter table trip_stops
  add column if not exists leg_distance_km numeric(10,2) not null default 0,
  add column if not exists cumulative_distance_km numeric(10,2) not null default 0,
  add column if not exists updated_at timestamptz not null default now();

alter table trip_expenses
  add column if not exists receipt_verified boolean not null default false,
  add column if not exists receipt_verified_by uuid,
  add column if not exists receipt_verified_at timestamptz,
  add column if not exists receipt_verification_notes text;

-- ============================================================================
-- INDEXES
-- ============================================================================

create index if not exists idx_trip_requests_status on trip_requests(status, created_at desc);
create index if not exists idx_trip_requests_route on trip_requests(route_id);

do $$ begin
  alter table trip_manifest_entries
    add constraint trip_manifest_entries_trip_booking_unique unique (trip_id, booking_passenger_id);
exception
  when duplicate_object then null;
end $$;
create index if not exists idx_trip_manifest_trip_active on trip_manifest_entries(trip_id, is_active);

create index if not exists idx_trip_authorizations_trip on trip_authorizations(trip_id, status);
create index if not exists idx_trip_authorizations_authority on trip_authorizations(authority, status);

create index if not exists idx_trip_checklists_trip_stage on trip_checklists(trip_id, stage, is_complete);
create index if not exists idx_trip_driver_reports_trip on trip_driver_reports(trip_id);
create index if not exists idx_trip_reconciliations_trip on trip_receipt_reconciliations(trip_id, receipts_validated);

create index if not exists idx_trip_attachments_trip on trip_attachments(trip_id, attachment_type);
create index if not exists idx_trips_operational_status on trips(operational_status, departure_at desc);

-- ============================================================================
-- BACKFILL
-- ============================================================================

update trips
set operational_status = 'REQUESTED',
    updated_at = now()
where operational_status is null;

insert into trip_stops (
  trip_id,
  route_stop_id,
  stop_order,
  arrive_at,
  depart_at,
  leg_distance_km,
  cumulative_distance_km,
  created_at,
  updated_at
)
select
  t.id,
  rs.id,
  rs.stop_order,
  case
    when rs.eta_offset_minutes is not null
      then t.departure_at + make_interval(mins => rs.eta_offset_minutes)
    else null
  end,
  null,
  0,
  0,
  now(),
  now()
from trips t
join route_stops rs on rs.route_id = t.route_id
where not exists (
  select 1 from trip_stops ts where ts.trip_id = t.id
)
on conflict (trip_id, route_stop_id) do nothing;

insert into trip_manifest_entries (
  trip_id,
  booking_passenger_id,
  passenger_name,
  passenger_document,
  passenger_phone,
  source,
  status,
  seat_number,
  is_active,
  created_at,
  updated_at
)
select
  bp.trip_id,
  bp.id,
  bp.name,
  bp.document,
  bp.phone,
  'BOOKING'::manifest_entry_source,
  case bp.status
    when 'BOARDED' then 'BOARDED'::manifest_passenger_status
    when 'NO_SHOW' then 'NO_SHOW'::manifest_passenger_status
    when 'CANCELLED' then 'CANCELLED'::manifest_passenger_status
    else 'EXPECTED'::manifest_passenger_status
  end,
  bs.seat_number,
  bp.is_active,
  bp.created_at,
  now()
from booking_passengers bp
left join bus_seats bs on bs.id = bp.seat_id
where bp.is_active = true
on conflict (trip_id, booking_passenger_id)
do update
set
  passenger_name = excluded.passenger_name,
  passenger_document = excluded.passenger_document,
  passenger_phone = excluded.passenger_phone,
  status = excluded.status,
  seat_number = excluded.seat_number,
  is_active = excluded.is_active,
  updated_at = now();

-- Keep legacy trip status as source for active operations bootstrap.
update trips
set operational_status = case
  when status = 'IN_PROGRESS' then 'IN_PROGRESS'::trip_operational_status
  when status = 'COMPLETED' then 'SETTLED'::trip_operational_status
  when status = 'CANCELLED' then 'CLOSED'::trip_operational_status
  else operational_status
end,
updated_at = now()
where status in ('IN_PROGRESS','COMPLETED','CANCELLED');
