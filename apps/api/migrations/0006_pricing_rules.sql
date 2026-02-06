-- 0006_pricing_rules.sql
-- Configurable pricing multipliers and rules.

do $$ begin
  create type pricing_scope as enum ('GLOBAL','ROUTE','TRIP');
exception
  when duplicate_object then null;
end $$;

do $$ begin
  create type pricing_rule_type as enum ('OCCUPANCY','LEAD_TIME','DOW','SEASON');
exception
  when duplicate_object then null;
end $$;

create table if not exists pricing_rules (
  id uuid primary key default gen_random_uuid(),
  name text not null,
  scope pricing_scope not null default 'GLOBAL',
  scope_id uuid,
  rule_type pricing_rule_type not null,
  priority int not null default 100,
  is_active boolean not null default true,
  params jsonb not null default '{}'::jsonb,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create index if not exists pricing_rules_scope_idx
  on pricing_rules (scope, scope_id, rule_type)
  where is_active = true;

create index if not exists pricing_rules_priority_idx
  on pricing_rules (priority, created_at desc);
