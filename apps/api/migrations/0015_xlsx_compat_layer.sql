-- 0015_xlsx_compat_layer.sql
-- XLSX compatibility layer with staging + reconciliation + promotion helpers.

create extension if not exists pgcrypto;

-- Legacy keys on canonical tables
alter table routes add column if not exists legacy_route_code text;
alter table route_stops add column if not exists legacy_stop_code text;
alter table trips add column if not exists legacy_trip_code text;
alter table trip_stops add column if not exists legacy_trip_stop_code text;
alter table bookings add column if not exists legacy_booking_code text;
alter table booking_passengers add column if not exists legacy_passenger_code text;

create unique index if not exists uq_routes_legacy_route_code on routes(legacy_route_code) where legacy_route_code is not null;
create unique index if not exists uq_route_stops_legacy_stop_code on route_stops(legacy_stop_code) where legacy_stop_code is not null;
create unique index if not exists uq_trips_legacy_trip_code on trips(legacy_trip_code) where legacy_trip_code is not null;
create unique index if not exists uq_trip_stops_legacy_trip_stop_code on trip_stops(legacy_trip_stop_code) where legacy_trip_stop_code is not null;
create unique index if not exists uq_bookings_legacy_booking_code on bookings(legacy_booking_code) where legacy_booking_code is not null;
create unique index if not exists uq_booking_passengers_legacy_passenger_code on booking_passengers(legacy_passenger_code) where legacy_passenger_code is not null;

create index if not exists idx_routes_legacy_route_code on routes(legacy_route_code);
create index if not exists idx_route_stops_legacy_stop_code on route_stops(legacy_stop_code);
create index if not exists idx_trips_legacy_trip_code on trips(legacy_trip_code);
create index if not exists idx_trip_stops_legacy_trip_stop_code on trip_stops(legacy_trip_stop_code);
create index if not exists idx_bookings_legacy_booking_code on bookings(legacy_booking_code);
create index if not exists idx_booking_passengers_legacy_passenger_code on booking_passengers(legacy_passenger_code);

-- Batch metadata and error quarantine
create table if not exists xlsx_import_batches (
  id uuid primary key default gen_random_uuid(),
  source_file_name text,
  status text not null default 'UPLOADED' check (status in ('UPLOADED','VALIDATED','PROMOTED','FAILED')),
  uploaded_at timestamptz not null default now(),
  validated_at timestamptz,
  promoted_at timestamptz,
  report jsonb not null default '{}'::jsonb
);

create table if not exists xlsx_import_errors (
  id uuid primary key default gen_random_uuid(),
  batch_id uuid not null references xlsx_import_batches(id) on delete cascade,
  sheet_name text not null,
  row_number integer,
  error_code text not null,
  error_message text not null,
  row_data jsonb,
  created_at timestamptz not null default now()
);

create index if not exists idx_xlsx_import_errors_batch on xlsx_import_errors(batch_id, sheet_name);

-- Staging tables
create table if not exists stg_xlsx_routes (
  id bigserial primary key,
  batch_id uuid not null references xlsx_import_batches(id) on delete cascade,
  row_number integer not null,
  route_id text,
  route_name text,
  active boolean,
  raw_row jsonb not null default '{}'::jsonb,
  created_at timestamptz not null default now()
);
create index if not exists idx_stg_xlsx_routes_batch on stg_xlsx_routes(batch_id, row_number);

create table if not exists stg_xlsx_stops (
  id bigserial primary key,
  batch_id uuid not null references xlsx_import_batches(id) on delete cascade,
  row_number integer not null,
  stop_id text,
  stop_name text,
  state text,
  display_name text,
  active boolean,
  route_id text,
  stop_order integer,
  eta_offset_minutes integer,
  raw_row jsonb not null default '{}'::jsonb,
  created_at timestamptz not null default now()
);
create index if not exists idx_stg_xlsx_stops_batch on stg_xlsx_stops(batch_id, row_number);

create table if not exists stg_xlsx_trips (
  id bigserial primary key,
  batch_id uuid not null references xlsx_import_batches(id) on delete cascade,
  row_number integer not null,
  trip_id text,
  route_id text,
  trip_date date,
  bus_id text,
  price_default numeric,
  seats_total integer,
  seats_available integer,
  duration_hours numeric,
  status text,
  package text,
  raw_row jsonb not null default '{}'::jsonb,
  created_at timestamptz not null default now()
);
create index if not exists idx_stg_xlsx_trips_batch on stg_xlsx_trips(batch_id, row_number);

create table if not exists stg_xlsx_trip_stops (
  id bigserial primary key,
  batch_id uuid not null references xlsx_import_batches(id) on delete cascade,
  row_number integer not null,
  trip_stop_id text,
  trip_id text,
  stop_id text,
  seq integer,
  depart_time text,
  active boolean,
  raw_row jsonb not null default '{}'::jsonb,
  created_at timestamptz not null default now()
);
create index if not exists idx_stg_xlsx_trip_stops_batch on stg_xlsx_trip_stops(batch_id, row_number);

create table if not exists stg_xlsx_available_segments (
  id bigserial primary key,
  batch_id uuid not null references xlsx_import_batches(id) on delete cascade,
  row_number integer not null,
  segment_id text,
  trip_id text,
  origin_stop_id text,
  origin_display text,
  origin_depart_time text,
  dest_stop_id text,
  dest_display text,
  trip_date date,
  price numeric,
  seats_available integer,
  status text,
  route_id text,
  package text,
  raw_row jsonb not null default '{}'::jsonb,
  created_at timestamptz not null default now()
);
create index if not exists idx_stg_xlsx_segments_batch on stg_xlsx_available_segments(batch_id, row_number);

create table if not exists stg_xlsx_bookings (
  id bigserial primary key,
  batch_id uuid not null references xlsx_import_batches(id) on delete cascade,
  row_number integer not null,
  booking_id text,
  trip_id text,
  origin_stop_id text,
  dest_stop_id text,
  qty integer,
  customer_name text,
  customer_phone text,
  payment_type text,
  payment_status text,
  amount_total numeric,
  amount_paid numeric,
  created_at timestamptz,
  status text,
  updated_at timestamptz,
  reservation_code text,
  reserved_until timestamptz,
  row_type text,
  idempotency_key text,
  payment_method text,
  amount_due numeric,
  abacate_brcode text,
  abacate_brcode_base64 text,
  abacate_expires_at timestamptz,
  remaining_at_boarding numeric,
  abacate_pix_id text,
  stripe_payment_intent_id text,
  stripe_pix_copy_paste text,
  stripe_pix_expires_at timestamptz,
  stripe_hosted_instructions_url text,
  raw_row jsonb not null default '{}'::jsonb,
  created_row_at timestamptz not null default now()
);
create index if not exists idx_stg_xlsx_bookings_batch on stg_xlsx_bookings(batch_id, row_number);

create table if not exists stg_xlsx_passengers (
  id bigserial primary key,
  batch_id uuid not null references xlsx_import_batches(id) on delete cascade,
  row_number integer not null,
  passenger_id text,
  booking_id text,
  trip_id text,
  full_name text,
  document text,
  seat_number integer,
  origin_stop_id text,
  dest_stop_id text,
  phone text,
  notes text,
  status text,
  created_at timestamptz,
  updated_at timestamptz,
  row_type text,
  raw_row jsonb not null default '{}'::jsonb,
  created_row_at timestamptz not null default now()
);
create index if not exists idx_stg_xlsx_passengers_batch on stg_xlsx_passengers(batch_id, row_number);

create table if not exists stg_xlsx_manifest_data (
  id bigserial primary key,
  batch_id uuid not null references xlsx_import_batches(id) on delete cascade,
  row_number integer not null,
  passenger_id text,
  booking_id text,
  trip_id text,
  trip_date date,
  full_name text,
  document text,
  phone text,
  origin text,
  destination text,
  seat_number integer,
  payment_summary text,
  payment_status text,
  amount_total numeric,
  amount_paid numeric,
  amount_remaining numeric,
  update_at timestamptz,
  raw_row jsonb not null default '{}'::jsonb,
  created_row_at timestamptz not null default now()
);
create index if not exists idx_stg_xlsx_manifest_batch on stg_xlsx_manifest_data(batch_id, row_number);

-- Compatibility views
create or replace view v_xlsx_routes as
select
  r.legacy_route_code as route_id,
  r.name as route_name,
  r.is_active as active
from routes r
where r.legacy_route_code is not null;

create or replace view v_xlsx_stops as
select
  rs.legacy_stop_code as stop_id,
  rs.city as stop_name,
  null::text as state,
  rs.city as display_name,
  true as active
from route_stops rs
where rs.legacy_stop_code is not null;

create or replace view v_xlsx_trips as
select
  t.legacy_trip_code as trip_id,
  r.legacy_route_code as route_id,
  (t.departure_at at time zone 'utc')::date as trip_date,
  coalesce(b.plate, b.name) as bus_id,
  f.total_amount as price_default,
  b.capacity as seats_total,
  greatest(b.capacity - coalesce(occ.occupied, 0), 0) as seats_available,
  case when t.arrival_at is null then null else extract(epoch from (t.arrival_at - t.departure_at))/3600 end as duration_hours,
  t.status::text as status,
  null::text as package
from trips t
join routes r on r.id = t.route_id
join buses b on b.id = t.bus_id
left join fares f on f.id = t.fare_id
left join (
  select bp.trip_id, count(*)::int as occupied
  from booking_passengers bp
  where bp.is_active = true
  group by bp.trip_id
) occ on occ.trip_id = t.id
where t.legacy_trip_code is not null and r.legacy_route_code is not null;

create or replace view v_xlsx_trip_stops as
select
  ts.legacy_trip_stop_code as trip_stop_id,
  t.legacy_trip_code as trip_id,
  rs.legacy_stop_code as stop_id,
  ts.stop_order as seq,
  to_char(ts.depart_at at time zone 'utc', 'HH24:MI') as depart_time,
  true as active
from trip_stops ts
join trips t on t.id = ts.trip_id
join route_stops rs on rs.id = ts.route_stop_id
where ts.legacy_trip_stop_code is not null
  and t.legacy_trip_code is not null
  and rs.legacy_stop_code is not null;

create or replace view v_xlsx_available_segments as
select
  concat(t.legacy_trip_code, '_', frs.legacy_stop_code, '_', trs.legacy_stop_code) as segment_id,
  t.legacy_trip_code as trip_id,
  frs.legacy_stop_code as origin_stop_id,
  frs.city as origin_display,
  to_char(from_ts.depart_at at time zone 'utc', 'HH24:MI') as origin_depart_time,
  trs.legacy_stop_code as dest_stop_id,
  trs.city as dest_display,
  (t.departure_at at time zone 'utc')::date as trip_date,
  sf.base_amount as price,
  greatest(b.capacity - coalesce(occ.occupied, 0), 0) as seats_available,
  case when sf.is_active then 'ATIVO' else 'INATIVO' end as status,
  r.legacy_route_code as route_id,
  null::text as package
from segment_fares sf
join route_stops frs on frs.id = sf.from_route_stop_id
join route_stops trs on trs.id = sf.to_route_stop_id
join routes r on r.id = frs.route_id and r.id = trs.route_id
join trips t on t.route_id = r.id
join trip_stops from_ts on from_ts.trip_id = t.id and from_ts.route_stop_id = frs.id
join buses b on b.id = t.bus_id
left join (
  select bp.trip_id, count(*)::int as occupied
  from booking_passengers bp
  where bp.is_active = true
  group by bp.trip_id
) occ on occ.trip_id = t.id
where t.legacy_trip_code is not null
  and r.legacy_route_code is not null
  and frs.legacy_stop_code is not null
  and trs.legacy_stop_code is not null;

create or replace view v_xlsx_bookings as
select
  b.legacy_booking_code as booking_id,
  t.legacy_trip_code as trip_id,
  broot.legacy_stop_code as origin_stop_id,
  aroot.legacy_stop_code as dest_stop_id,
  pcount.passenger_qty as qty,
  bp.name as customer_name,
  bp.phone as customer_phone,
  pmt.method::text as payment_type,
  pmt.status::text as payment_status,
  b.total_amount as amount_total,
  coalesce(paid.amount_paid, 0) as amount_paid,
  b.created_at,
  b.status::text as status,
  b.updated_at,
  null::text as reservation_code,
  b.expires_at as reserved_until,
  null::text as row_type,
  null::text as idempotency_key,
  pmt.method::text as payment_method,
  greatest(b.total_amount - coalesce(paid.amount_paid, 0), 0) as amount_due,
  null::text as abacate_brcode,
  null::text as abacate_brcode_base64,
  null::timestamptz as abacate_expires_at,
  greatest(b.total_amount - coalesce(paid.amount_paid, 0), 0) as remaining_at_boarding,
  null::text as abacate_pix_id,
  null::text as stripe_payment_intent_id,
  null::text as stripe_pix_copy_paste,
  null::timestamptz as stripe_pix_expires_at,
  null::text as stripe_hosted_instructions_url
from bookings b
join trips t on t.id = b.trip_id
left join lateral (
  select bp1.*
  from booking_passengers bp1
  where bp1.booking_id = b.id
  order by bp1.created_at asc
  limit 1
) bp on true
left join trip_stops bts on bts.id = bp.board_stop_id
left join route_stops broot on broot.id = bts.route_stop_id
left join trip_stops ats on ats.id = bp.alight_stop_id
left join route_stops aroot on aroot.id = ats.route_stop_id
left join lateral (
  select count(*)::int as passenger_qty
  from booking_passengers bp2
  where bp2.booking_id = b.id and bp2.is_active = true
) pcount on true
left join lateral (
  select p.method, p.status
  from payments p
  where p.booking_id = b.id
  order by p.created_at desc
  limit 1
) pmt on true
left join lateral (
  select sum(case when p.status = 'PAID' then p.amount else 0 end)::numeric as amount_paid
  from payments p
  where p.booking_id = b.id
) paid on true
where b.legacy_booking_code is not null and t.legacy_trip_code is not null;

create or replace view v_xlsx_passengers as
select
  bp.legacy_passenger_code as passenger_id,
  b.legacy_booking_code as booking_id,
  t.legacy_trip_code as trip_id,
  bp.name as full_name,
  bp.document,
  s.seat_number,
  broot.legacy_stop_code as origin_stop_id,
  aroot.legacy_stop_code as dest_stop_id,
  bp.phone,
  null::text as notes,
  bp.status::text as status,
  bp.created_at,
  b.updated_at,
  null::text as row_type
from booking_passengers bp
join bookings b on b.id = bp.booking_id
join trips t on t.id = bp.trip_id
left join bus_seats s on s.id = bp.seat_id
left join trip_stops bts on bts.id = bp.board_stop_id
left join route_stops broot on broot.id = bts.route_stop_id
left join trip_stops ats on ats.id = bp.alight_stop_id
left join route_stops aroot on aroot.id = ats.route_stop_id
where bp.legacy_passenger_code is not null
  and b.legacy_booking_code is not null
  and t.legacy_trip_code is not null;

create or replace view v_xlsx_manifest_data as
select
  bp.legacy_passenger_code as passenger_id,
  b.legacy_booking_code as booking_id,
  t.legacy_trip_code as trip_id,
  (t.departure_at at time zone 'utc')::date as trip_date,
  bp.name as full_name,
  bp.document,
  bp.phone,
  broot.city as origin,
  aroot.city as destination,
  s.seat_number,
  concat('TOTAL=', b.total_amount::text, ';PAID=', coalesce(paid.amount_paid, 0)::text) as payment_summary,
  b.status::text as payment_status,
  b.total_amount as amount_total,
  coalesce(paid.amount_paid, 0) as amount_paid,
  greatest(b.total_amount - coalesce(paid.amount_paid, 0), 0) as amount_remaining,
  b.updated_at as update_at
from booking_passengers bp
join bookings b on b.id = bp.booking_id
join trips t on t.id = bp.trip_id
left join bus_seats s on s.id = bp.seat_id
left join trip_stops bts on bts.id = bp.board_stop_id
left join route_stops broot on broot.id = bts.route_stop_id
left join trip_stops ats on ats.id = bp.alight_stop_id
left join route_stops aroot on aroot.id = ats.route_stop_id
left join lateral (
  select sum(case when p.status = 'PAID' then p.amount else 0 end)::numeric as amount_paid
  from payments p
  where p.booking_id = b.id
) paid on true
where bp.legacy_passenger_code is not null
  and b.legacy_booking_code is not null
  and t.legacy_trip_code is not null;

-- Reconciliation helper: count + deterministic hash
create or replace function fn_compare_xlsx_vs_db(p_sheet_name text)
returns table(metric text, value text)
language plpgsql
as $$
declare
  v_count bigint;
  v_hash text;
begin
  if p_sheet_name = 'routes' then
    select count(*), coalesce(md5(string_agg(row_to_json(v)::text, '|' order by v.route_id)), md5(''))
      into v_count, v_hash
      from v_xlsx_routes v;
  elsif p_sheet_name = 'stops' then
    select count(*), coalesce(md5(string_agg(row_to_json(v)::text, '|' order by v.stop_id)), md5(''))
      into v_count, v_hash
      from v_xlsx_stops v;
  elsif p_sheet_name = 'trips' then
    select count(*), coalesce(md5(string_agg(row_to_json(v)::text, '|' order by v.trip_id)), md5(''))
      into v_count, v_hash
      from v_xlsx_trips v;
  elsif p_sheet_name = 'trip_stops' then
    select count(*), coalesce(md5(string_agg(row_to_json(v)::text, '|' order by v.trip_stop_id)), md5(''))
      into v_count, v_hash
      from v_xlsx_trip_stops v;
  elsif p_sheet_name = 'available_segments' then
    select count(*), coalesce(md5(string_agg(row_to_json(v)::text, '|' order by v.segment_id)), md5(''))
      into v_count, v_hash
      from v_xlsx_available_segments v;
  elsif p_sheet_name = 'bookings' then
    select count(*), coalesce(md5(string_agg(row_to_json(v)::text, '|' order by v.booking_id)), md5(''))
      into v_count, v_hash
      from v_xlsx_bookings v;
  elsif p_sheet_name = 'passengers' then
    select count(*), coalesce(md5(string_agg(row_to_json(v)::text, '|' order by v.passenger_id)), md5(''))
      into v_count, v_hash
      from v_xlsx_passengers v;
  elsif p_sheet_name = 'manifest_data' then
    select count(*), coalesce(md5(string_agg(row_to_json(v)::text, '|' order by v.passenger_id)), md5(''))
      into v_count, v_hash
      from v_xlsx_manifest_data v;
  else
    raise exception 'unsupported sheet_name: %', p_sheet_name;
  end if;

  return query
    select 'count'::text, v_count::text
    union all
    select 'hash_md5'::text, v_hash::text;
end;
$$;

create or replace function fn_validate_xlsx_batch(p_batch_id uuid)
returns table(sheet_name text, error_count bigint)
language plpgsql
as $$
begin
  delete from xlsx_import_errors where batch_id = p_batch_id;

  insert into xlsx_import_errors(batch_id, sheet_name, row_number, error_code, error_message, row_data)
  select p_batch_id, 'routes', row_number, 'MISSING_ROUTE_ID', 'route_id is required', raw_row
  from stg_xlsx_routes
  where batch_id = p_batch_id and coalesce(trim(route_id), '') = '';

  insert into xlsx_import_errors(batch_id, sheet_name, row_number, error_code, error_message, row_data)
  select p_batch_id, 'stops', row_number, 'MISSING_STOP_ID', 'stop_id is required', raw_row
  from stg_xlsx_stops
  where batch_id = p_batch_id and coalesce(trim(stop_id), '') = '';

  insert into xlsx_import_errors(batch_id, sheet_name, row_number, error_code, error_message, row_data)
  select p_batch_id, 'trips', row_number, 'MISSING_TRIP_ID', 'trip_id is required', raw_row
  from stg_xlsx_trips
  where batch_id = p_batch_id and coalesce(trim(trip_id), '') = '';

  insert into xlsx_import_errors(batch_id, sheet_name, row_number, error_code, error_message, row_data)
  select p_batch_id, 'trip_stops', row_number, 'MISSING_TRIP_STOP_ID', 'trip_stop_id is required', raw_row
  from stg_xlsx_trip_stops
  where batch_id = p_batch_id and coalesce(trim(trip_stop_id), '') = '';

  insert into xlsx_import_errors(batch_id, sheet_name, row_number, error_code, error_message, row_data)
  select p_batch_id, 'bookings', row_number, 'MISSING_BOOKING_ID', 'booking_id is required', raw_row
  from stg_xlsx_bookings
  where batch_id = p_batch_id and coalesce(trim(booking_id), '') = '';

  insert into xlsx_import_errors(batch_id, sheet_name, row_number, error_code, error_message, row_data)
  select p_batch_id, 'passengers', row_number, 'MISSING_PASSENGER_ID', 'passenger_id is required', raw_row
  from stg_xlsx_passengers
  where batch_id = p_batch_id and coalesce(trim(passenger_id), '') = '';

  update xlsx_import_batches
  set status = case when exists(select 1 from xlsx_import_errors e where e.batch_id = p_batch_id) then 'FAILED' else 'VALIDATED' end,
      validated_at = now()
  where id = p_batch_id;

  return query
  select x.sheet_name, count(*)::bigint
  from xlsx_import_errors x
  where x.batch_id = p_batch_id
  group by x.sheet_name
  order by x.sheet_name;
end;
$$;

create or replace function fn_promote_xlsx_batch(p_batch_id uuid)
returns jsonb
language plpgsql
as $$
declare
  v_errors bigint;
  v_result jsonb := '{}'::jsonb;
begin
  select count(*) into v_errors from xlsx_import_errors where batch_id = p_batch_id;
  if v_errors > 0 then
    raise exception 'batch % has validation errors', p_batch_id;
  end if;

  -- routes
  insert into routes(name, origin_city, destination_city, is_active, legacy_route_code)
  select
    coalesce(nullif(trim(route_name), ''), route_id),
    coalesce(split_part(route_name, ' to ', 1), route_id),
    coalesce(split_part(route_name, ' to ', 2), route_id),
    coalesce(active, false),
    route_id
  from stg_xlsx_routes s
  where s.batch_id = p_batch_id
    and coalesce(trim(route_id), '') <> ''
  on conflict (legacy_route_code) where legacy_route_code is not null
  do update set
    name = excluded.name,
    is_active = excluded.is_active;

  -- stops -> route_stops (requires route_id + stop_order in staging)
  insert into route_stops(route_id, city, stop_order, eta_offset_minutes, notes, legacy_stop_code)
  select
    r.id,
    coalesce(nullif(trim(s.stop_name), ''), s.display_name, s.stop_id),
    coalesce(s.stop_order, row_number() over (partition by s.batch_id, s.route_id order by s.row_number)),
    s.eta_offset_minutes,
    null,
    s.stop_id
  from stg_xlsx_stops s
  join routes r on r.legacy_route_code = s.route_id
  where s.batch_id = p_batch_id
    and coalesce(trim(s.stop_id), '') <> ''
  on conflict (legacy_stop_code) where legacy_stop_code is not null
  do update set
    city = excluded.city,
    stop_order = excluded.stop_order,
    eta_offset_minutes = excluded.eta_offset_minutes;

  -- trips (bus mapping by plate or name)
  insert into trips(route_id, bus_id, departure_at, status, operational_status, estimated_km, legacy_trip_code, updated_at)
  select
    r.id,
    b.id,
    coalesce(s.trip_date::timestamptz, now()),
    case when upper(coalesce(s.status, '')) in ('ATIVO','SCHEDULED','IN_PROGRESS','COMPLETED','CANCELLED')
      then case when upper(s.status) = 'ATIVO' then 'SCHEDULED' else upper(s.status) end::trip_status
      else 'SCHEDULED'::trip_status end,
    'REQUESTED'::trip_operational_status,
    coalesce(s.duration_hours, 0) * 60,
    s.trip_id,
    now()
  from stg_xlsx_trips s
  join routes r on r.legacy_route_code = s.route_id
  join buses b on lower(coalesce(b.plate, '')) = lower(coalesce(s.bus_id, ''))
            or lower(coalesce(b.name, '')) = lower(coalesce(s.bus_id, ''))
  where s.batch_id = p_batch_id
    and coalesce(trim(s.trip_id), '') <> ''
  on conflict (legacy_trip_code) where legacy_trip_code is not null
  do update set
    route_id = excluded.route_id,
    bus_id = excluded.bus_id,
    departure_at = excluded.departure_at,
    status = excluded.status,
    updated_at = now();

  -- trip_stops
  insert into trip_stops(trip_id, route_stop_id, stop_order, depart_at, arrive_at, leg_distance_km, cumulative_distance_km, legacy_trip_stop_code, updated_at)
  select
    t.id,
    rs.id,
    coalesce(s.seq, 1),
    case when coalesce(s.depart_time, '') <> '' then ((t.departure_at at time zone 'utc')::date + s.depart_time::time)::timestamptz else null end,
    null,
    0,
    0,
    s.trip_stop_id,
    now()
  from stg_xlsx_trip_stops s
  join trips t on t.legacy_trip_code = s.trip_id
  join route_stops rs on rs.legacy_stop_code = s.stop_id and rs.route_id = t.route_id
  where s.batch_id = p_batch_id
    and coalesce(trim(s.trip_stop_id), '') <> ''
  on conflict (legacy_trip_stop_code) where legacy_trip_stop_code is not null
  do update set
    stop_order = excluded.stop_order,
    depart_at = excluded.depart_at,
    updated_at = now();

  -- bookings
  insert into bookings(trip_id, status, source, total_amount, deposit_amount, remainder_amount, expires_at, created_at, updated_at, legacy_booking_code)
  select
    t.id,
    case when upper(coalesce(s.status, '')) in ('PENDING','CONFIRMED','CANCELLED','EXPIRED')
      then upper(s.status)::booking_status else 'PENDING'::booking_status end,
    'MANUAL'::booking_source,
    coalesce(s.amount_total, 0),
    coalesce(s.amount_paid, 0),
    greatest(coalesce(s.amount_total, 0) - coalesce(s.amount_paid, 0), 0),
    s.reserved_until,
    coalesce(s.created_at, now()),
    coalesce(s.updated_at, now()),
    s.booking_id
  from stg_xlsx_bookings s
  join trips t on t.legacy_trip_code = s.trip_id
  where s.batch_id = p_batch_id
    and coalesce(trim(s.booking_id), '') <> ''
  on conflict (legacy_booking_code) where legacy_booking_code is not null
  do update set
    status = excluded.status,
    total_amount = excluded.total_amount,
    deposit_amount = excluded.deposit_amount,
    remainder_amount = excluded.remainder_amount,
    updated_at = now();

  -- passengers
  insert into booking_passengers(
    booking_id, trip_id, name, document, phone, seat_id, status, created_at, is_active,
    board_stop_id, alight_stop_id, board_stop_order, alight_stop_order, fare_amount_calc, fare_amount_final,
    legacy_passenger_code
  )
  select
    b.id,
    t.id,
    coalesce(nullif(trim(s.full_name), ''), s.passenger_id),
    s.document,
    s.phone,
    bs.id,
    case when upper(coalesce(s.status, '')) in ('RESERVED','BOARDED','NO_SHOW','CANCELLED')
      then upper(s.status)::passenger_status else 'RESERVED'::passenger_status end,
    coalesce(s.created_at, now()),
    true,
    board_ts.id,
    alight_ts.id,
    board_ts.stop_order,
    alight_ts.stop_order,
    0,
    0,
    s.passenger_id
  from stg_xlsx_passengers s
  join bookings b on b.legacy_booking_code = s.booking_id
  join trips t on t.id = b.trip_id
  left join bus_seats bs on bs.bus_id = t.bus_id and bs.seat_number = s.seat_number
  left join route_stops board_rs on board_rs.legacy_stop_code = s.origin_stop_id and board_rs.route_id = t.route_id
  left join route_stops alight_rs on alight_rs.legacy_stop_code = s.dest_stop_id and alight_rs.route_id = t.route_id
  left join trip_stops board_ts on board_ts.trip_id = t.id and board_ts.route_stop_id = board_rs.id
  left join trip_stops alight_ts on alight_ts.trip_id = t.id and alight_ts.route_stop_id = alight_rs.id
  where s.batch_id = p_batch_id
    and coalesce(trim(s.passenger_id), '') <> ''
    and bs.id is not null
  on conflict (legacy_passenger_code) where legacy_passenger_code is not null
  do update set
    name = excluded.name,
    phone = excluded.phone,
    status = excluded.status,
    updated_at = now();

  -- payment projection from staging bookings
  insert into payments(booking_id, amount, method, status, provider, provider_ref, created_at)
  select
    b.id,
    coalesce(s.amount_paid, 0),
    case
      when upper(coalesce(s.payment_method, s.payment_type, '')) in ('PIX','CARD','CASH','TRANSFER','OTHER')
        then upper(coalesce(s.payment_method, s.payment_type))::payment_method
      else 'OTHER'::payment_method
    end,
    case
      when upper(coalesce(s.payment_status, '')) in ('PENDING','PAID','FAILED','REFUNDED','CANCELLED')
        then upper(s.payment_status)::payment_status
      else 'PENDING'::payment_status
    end,
    'xlsx_import',
    s.booking_id || ':' || s.row_number::text,
    now()
  from stg_xlsx_bookings s
  join bookings b on b.legacy_booking_code = s.booking_id
  where s.batch_id = p_batch_id
    and coalesce(s.amount_paid, 0) > 0
  on conflict do nothing;

  update xlsx_import_batches
    set status = 'PROMOTED', promoted_at = now(),
        report = jsonb_build_object('batch_id', p_batch_id, 'promoted_at', now())
  where id = p_batch_id;

  v_result := jsonb_build_object('batch_id', p_batch_id, 'status', 'PROMOTED');
  return v_result;
end;
$$;
