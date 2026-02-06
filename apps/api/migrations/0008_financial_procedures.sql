-- 0008_financial_procedures.sql
-- Financial procedures: advances, expenses, settlements, cards, validations

-- ============================================================================
-- ENUMS
-- ============================================================================

do $$ begin
  create type advance_status as enum ('PENDING','DELIVERED','SETTLED','CANCELLED');
exception
  when duplicate_object then null;
end $$;

do $$ begin
  create type expense_type as enum ('FUEL','FOOD','LODGING','TOLL','MAINTENANCE','OTHER');
exception
  when duplicate_object then null;
end $$;

do $$ begin
  create type expense_payment_method as enum ('ADVANCE','CARD','PERSONAL','COMPANY');
exception
  when duplicate_object then null;
end $$;

do $$ begin
  create type settlement_status as enum ('DRAFT','UNDER_REVIEW','APPROVED','REJECTED','COMPLETED');
exception
  when duplicate_object then null;
end $$;

do $$ begin
  create type card_type as enum ('FUEL','MULTIPURPOSE','FOOD');
exception
  when duplicate_object then null;
end $$;

do $$ begin
  create type card_transaction_type as enum ('CREDIT','DEBIT','ADJUSTMENT','REFUND');
exception
  when duplicate_object then null;
end $$;

-- ============================================================================
-- TABLES
-- ============================================================================

-- 1. Driver Cards (Cartões de Motorista)
create table if not exists driver_cards (
  id uuid primary key default gen_random_uuid(),
  driver_id uuid not null references drivers(id) on delete cascade,
  card_number text not null unique,
  card_type card_type not null,
  current_balance numeric(10,2) not null default 0,
  is_active boolean not null default true,
  is_blocked boolean not null default false,
  issued_at timestamptz not null default now(),
  blocked_at timestamptz,
  blocked_by uuid,
  block_reason text,
  notes text,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

-- 2. Trip Advances (Adiantamentos de Viagem)
create table if not exists trip_advances (
  id uuid primary key default gen_random_uuid(),
  trip_id uuid not null references trips(id) on delete cascade,
  driver_id uuid not null references drivers(id) on delete restrict,
  amount numeric(10,2) not null check (amount >= 0),
  status advance_status not null default 'PENDING',
  purpose text,
  delivered_at timestamptz,
  delivered_by uuid,
  settled_at timestamptz,
  notes text,
  created_by uuid,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

-- 3. Trip Expenses (Despesas de Viagem)
create table if not exists trip_expenses (
  id uuid primary key default gen_random_uuid(),
  trip_id uuid not null references trips(id) on delete cascade,
  driver_id uuid not null references drivers(id) on delete restrict,
  expense_type expense_type not null,
  amount numeric(10,2) not null check (amount >= 0),
  description text not null,
  expense_date timestamptz not null,
  payment_method expense_payment_method not null default 'ADVANCE',
  driver_card_id uuid references driver_cards(id) on delete set null,
  receipt_number text,
  is_approved boolean not null default false,
  approved_by uuid,
  approved_at timestamptz,
  notes text,
  created_by uuid,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

-- 4. Trip Settlements (Acertos de Viagem)
create table if not exists trip_settlements (
  id uuid primary key default gen_random_uuid(),
  trip_id uuid not null references trips(id) on delete cascade unique,
  driver_id uuid not null references drivers(id) on delete restrict,
  status settlement_status not null default 'DRAFT',
  advance_amount numeric(10,2) not null default 0 check (advance_amount >= 0),
  expenses_total numeric(10,2) not null default 0 check (expenses_total >= 0),
  balance numeric(10,2) not null default 0,
  amount_to_return numeric(10,2) not null default 0 check (amount_to_return >= 0),
  amount_to_reimburse numeric(10,2) not null default 0 check (amount_to_reimburse >= 0),
  reviewed_by uuid,
  reviewed_at timestamptz,
  approved_by uuid,
  approved_at timestamptz,
  completed_at timestamptz,
  notes text,
  created_by uuid,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

-- 5. Advance Returns (Devoluções de Adiantamento)
create table if not exists advance_returns (
  id uuid primary key default gen_random_uuid(),
  trip_advance_id uuid not null references trip_advances(id) on delete cascade,
  trip_settlement_id uuid references trip_settlements(id) on delete set null,
  amount numeric(10,2) not null check (amount >= 0),
  return_date timestamptz not null default now(),
  payment_method text not null,
  received_by uuid,
  notes text,
  created_at timestamptz not null default now()
);

-- 6. Driver Card Transactions (Transações de Cartão)
create table if not exists driver_card_transactions (
  id uuid primary key default gen_random_uuid(),
  card_id uuid not null references driver_cards(id) on delete cascade,
  transaction_type card_transaction_type not null,
  amount numeric(10,2) not null check (amount >= 0),
  balance_before numeric(10,2) not null,
  balance_after numeric(10,2) not null,
  description text,
  trip_expense_id uuid references trip_expenses(id) on delete set null,
  performed_by uuid,
  created_at timestamptz not null default now()
);

-- 7. Trip Validations (Validações de Viagem)
create table if not exists trip_validations (
  id uuid primary key default gen_random_uuid(),
  trip_id uuid not null references trips(id) on delete cascade unique,
  odometer_initial int check (odometer_initial >= 0),
  odometer_final int check (odometer_final >= 0),
  distance_km int generated always as (odometer_final - odometer_initial) stored,
  passengers_expected int not null default 0,
  passengers_boarded int not null default 0,
  passengers_no_show int not null default 0,
  validation_notes text,
  validated_by uuid,
  validated_at timestamptz,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  check (odometer_final is null or odometer_initial is null or odometer_final >= odometer_initial)
);

-- 8. Fiscal Documents (Documentos Fiscais - Estrutura básica)
create table if not exists fiscal_documents (
  id uuid primary key default gen_random_uuid(),
  trip_id uuid not null references trips(id) on delete cascade,
  document_type text not null,
  document_number text,
  issue_date timestamptz not null default now(),
  amount numeric(10,2) not null check (amount >= 0),
  recipient_name text,
  recipient_document text,
  status text not null default 'PENDING',
  external_id text,
  metadata jsonb,
  created_by uuid,
  created_at timestamptz not null default now()
);

-- 9. Expense Categories (Planos de Conta)
create table if not exists expense_categories (
  id uuid primary key default gen_random_uuid(),
  name text not null unique,
  code text not null unique,
  expense_type expense_type,
  is_active boolean not null default true,
  created_at timestamptz not null default now()
);

-- ============================================================================
-- INDEXES
-- ============================================================================

create index if not exists idx_trip_advances_trip on trip_advances(trip_id);
create index if not exists idx_trip_advances_driver on trip_advances(driver_id);
create index if not exists idx_trip_advances_status on trip_advances(status) where status != 'CANCELLED';

create index if not exists idx_trip_expenses_trip on trip_expenses(trip_id);
create index if not exists idx_trip_expenses_driver on trip_expenses(driver_id);
create index if not exists idx_trip_expenses_type on trip_expenses(expense_type);
create index if not exists idx_trip_expenses_approved on trip_expenses(is_approved, approved_at);

create index if not exists idx_trip_settlements_trip on trip_settlements(trip_id);
create index if not exists idx_trip_settlements_status on trip_settlements(status);

create index if not exists idx_driver_cards_driver on driver_cards(driver_id);
create index if not exists idx_driver_cards_active on driver_cards(is_active, is_blocked) where is_active = true;

create index if not exists idx_card_transactions_card on driver_card_transactions(card_id);
create index if not exists idx_card_transactions_expense on driver_card_transactions(trip_expense_id) where trip_expense_id is not null;

create index if not exists idx_trip_validations_trip on trip_validations(trip_id);

create index if not exists idx_fiscal_documents_trip on fiscal_documents(trip_id);
create index if not exists idx_fiscal_documents_type on fiscal_documents(document_type, status);

-- ============================================================================
-- SEED DATA
-- ============================================================================

insert into expense_categories (name, code, expense_type) values
  ('Combustível', '106', 'FUEL'),
  ('Alimentação', '201', 'FOOD'),
  ('Hospedagem', '202', 'LODGING'),
  ('Pedágio', '203', 'TOLL'),
  ('Manutenção', '301', 'MAINTENANCE'),
  ('Outras Despesas', '999', 'OTHER')
on conflict (code) do nothing;
