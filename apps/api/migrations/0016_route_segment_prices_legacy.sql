-- 0016_route_segment_prices_legacy.sql
-- Legacy ticketing schema: route-level segment price matrix with compatibility projection.

create table if not exists route_segment_prices (
  route_id text not null references routes(route_id) on delete cascade,
  origin_stop_id text not null references stops(stop_id) on delete cascade,
  destination_stop_id text not null references stops(stop_id) on delete cascade,
  price numeric(10,2) not null,
  status text not null default 'ACTIVE',
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  primary key (route_id, origin_stop_id, destination_stop_id),
  check (origin_stop_id <> destination_stop_id),
  check (price >= 0),
  check (upper(status) in ('ACTIVE','INACTIVE'))
);

create index if not exists route_segment_prices_lookup_idx
  on route_segment_prices (route_id, origin_stop_id, destination_stop_id, status);

create or replace function set_route_segment_prices_updated_at()
returns trigger
language plpgsql
as $$
begin
  new.updated_at := now();
  return new;
end;
$$;

drop trigger if exists trg_route_segment_prices_updated_at on route_segment_prices;
create trigger trg_route_segment_prices_updated_at
before update on route_segment_prices
for each row
execute function set_route_segment_prices_updated_at();

insert into route_segment_prices (
  route_id,
  origin_stop_id,
  destination_stop_id,
  price,
  status,
  created_at,
  updated_at
)
select distinct on (s.route_id, s.origin_stop_id, s.destination_stop_id)
  s.route_id,
  s.origin_stop_id,
  s.destination_stop_id,
  s.price,
  case
    when upper(coalesce(s.status, 'ATIVO')) in ('INATIVO', 'INACTIVE') then 'INACTIVE'
    else 'ACTIVE'
  end,
  coalesce(s.created_at, now()),
  now()
from available_segments s
where s.route_id is not null
order by s.route_id, s.origin_stop_id, s.destination_stop_id, s.created_at desc
on conflict (route_id, origin_stop_id, destination_stop_id) do update
set
  price = excluded.price,
  status = excluded.status,
  updated_at = now();

create or replace function refresh_available_segments_for_route(p_route_id text default null)
returns bigint
language plpgsql
as $$
declare
  v_rows bigint := 0;
begin
  if p_route_id is null then
    delete from available_segments;
  else
    delete from available_segments where route_id = p_route_id;
  end if;

  insert into available_segments (
    segment_id,
    trip_id,
    route_id,
    origin_stop_id,
    destination_stop_id,
    origin_display_snapshot,
    destination_display_snapshot,
    origin_depart_time,
    trip_date,
    price,
    seats_available,
    status,
    package_name,
    created_at
  )
  select
    concat(t.trip_id, '_', o.stop_id, '_', d.stop_id),
    t.trip_id,
    t.route_id,
    o.stop_id,
    d.stop_id,
    os.display_name,
    ds.display_name,
    o.depart_time,
    t.trip_date,
    rsp.price,
    t.seats_available,
    case
      when upper(rsp.status) = 'INACTIVE' then 'INATIVO'
      else 'ATIVO'
    end,
    t.package_name,
    now()
  from trips t
  join trip_stops o on o.trip_id = t.trip_id and o.is_active = true
  join trip_stops d on d.trip_id = t.trip_id and d.is_active = true and d.stop_sequence > o.stop_sequence
  join stops os on os.stop_id = o.stop_id
  join stops ds on ds.stop_id = d.stop_id
  join route_segment_prices rsp
    on rsp.route_id = t.route_id
   and rsp.origin_stop_id = o.stop_id
   and rsp.destination_stop_id = d.stop_id
  where p_route_id is null or t.route_id = p_route_id;

  get diagnostics v_rows = row_count;
  return v_rows;
end;
$$;

create or replace function public.refresh_manifest_data()
returns void
language plpgsql
as $function$
begin
  truncate table public.manifest_data;

  insert into public.manifest_data (
    passenger_id,
    booking_id,
    trip_id,
    trip_date,
    full_name,
    document,
    phone,
    origin,
    destination,
    seat_number,
    payment_summary,
    payment_status,
    amount_total,
    amount_paid,
    amount_remaining,
    updated_at
  )
  select
    p.passenger_id,
    p.booking_id,
    p.trip_id,
    t.trip_date,
    p.full_name,
    p.document,
    p.phone,
    coalesce(origin_stop.display_name, p.origin_stop_id) as origin,
    coalesce(destination_stop.display_name, p.destination_stop_id) as destination,
    p.seat_number,
    concat_ws(
      ' | ',
      'Pagador: ' || b.customer_name,
      case when pay.payment_type is not null then 'Tipo: ' || pay.payment_type end,
      case when pay.payment_status is not null then 'Status: ' || pay.payment_status end,
      'Total: R$ ' || to_char(coalesce(pay.amount_total, 0), 'FM999999990.00'),
      'Pago: R$ ' || to_char(coalesce(pay.amount_paid, 0), 'FM999999990.00'),
      'Falta: R$ ' || to_char(coalesce(pay.amount_due, 0), 'FM999999990.00'),
      'Grupo: ' || b.passenger_qty || ' pax',
      'Falta/pax: R$ ' || to_char(
        case
          when b.passenger_qty > 0 then coalesce(pay.amount_due, 0) / b.passenger_qty
          else 0
        end,
        'FM999999990.00'
      )
    ) as payment_summary,
    pay.payment_status,
    pay.amount_total,
    pay.amount_paid,
    pay.amount_due as amount_remaining,
    greatest(p.updated_at, b.updated_at) as updated_at
  from public.passengers p
  join public.bookings b on b.booking_id = p.booking_id
  join public.trips t on t.trip_id = p.trip_id
  left join public.booking_payment_details pay on pay.booking_id = b.booking_id
  left join public.stops origin_stop on origin_stop.stop_id = coalesce(p.origin_stop_id, b.origin_stop_id)
  left join public.stops destination_stop on destination_stop.stop_id = coalesce(p.destination_stop_id, b.destination_stop_id);
end;
$function$;

create or replace function public.refresh_manifest_data_for_booking(p_booking_id text)
returns void
language plpgsql
as $function$
begin
  delete from public.manifest_data
  where booking_id = p_booking_id;

  insert into public.manifest_data (
    passenger_id,
    booking_id,
    trip_id,
    trip_date,
    full_name,
    document,
    phone,
    origin,
    destination,
    seat_number,
    payment_summary,
    payment_status,
    amount_total,
    amount_paid,
    amount_remaining,
    updated_at
  )
  select
    p.passenger_id,
    p.booking_id,
    p.trip_id,
    t.trip_date,
    p.full_name,
    p.document,
    p.phone,
    coalesce(origin_stop.display_name, p.origin_stop_id) as origin,
    coalesce(destination_stop.display_name, p.destination_stop_id) as destination,
    p.seat_number,
    concat_ws(
      ' | ',
      'Pagador: ' || b.customer_name,
      case when pay.payment_type is not null then 'Tipo: ' || pay.payment_type end,
      case when pay.payment_status is not null then 'Status: ' || pay.payment_status end,
      'Total: R$ ' || to_char(coalesce(pay.amount_total, 0), 'FM999999990.00'),
      'Pago: R$ ' || to_char(coalesce(pay.amount_paid, 0), 'FM999999990.00'),
      'Falta: R$ ' || to_char(coalesce(pay.amount_due, 0), 'FM999999990.00'),
      'Grupo: ' || b.passenger_qty || ' pax',
      'Falta/pax: R$ ' || to_char(
        case
          when b.passenger_qty > 0 then coalesce(pay.amount_due, 0) / b.passenger_qty
          else 0
        end,
        'FM999999990.00'
      )
    ) as payment_summary,
    pay.payment_status,
    pay.amount_total,
    pay.amount_paid,
    pay.amount_due as amount_remaining,
    greatest(p.updated_at, b.updated_at) as updated_at
  from public.passengers p
  join public.bookings b on b.booking_id = p.booking_id
  join public.trips t on t.trip_id = p.trip_id
  left join public.booking_payment_details pay on pay.booking_id = b.booking_id
  left join public.stops origin_stop on origin_stop.stop_id = coalesce(p.origin_stop_id, b.origin_stop_id)
  left join public.stops destination_stop on destination_stop.stop_id = coalesce(p.destination_stop_id, b.destination_stop_id)
  where p.booking_id = p_booking_id;
end;
$function$;

select refresh_available_segments_for_route(null);
