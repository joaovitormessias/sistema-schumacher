-- 0014_routes_operational_v2.sql
-- Route workflow support: draft/publish and duplication traceability.

alter table routes
  alter column is_active set default false;

alter table routes
  add column if not exists duplicated_from_route_id uuid references routes(id) on delete set null;

create index if not exists idx_routes_duplicated_from
  on routes(duplicated_from_route_id);
