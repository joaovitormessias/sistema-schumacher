# Plano de Implementação: Sistema de Procedimentos Financeiros para Viagens

## 1. VISÃO GERAL

### 1.1 Resumo Executivo

Este plano detalha a implementação completa dos procedimentos financeiros para gestão de viagens no sistema Schumacher Tur, baseado no documento "Procedimento Operacional Financeiro". O sistema atual possui funcionalidades básicas de viagens e pagamentos de passagens, mas carece de todo o processo de acerto financeiro pós-viagem descrito nos procedimentos operacionais da empresa.

### 1.2 Objetivos Principais

1. Implementar sistema de **adiantamentos** para motoristas antes das viagens
2. Implementar controle de **despesas de viagem** (combustível, alimentação, hospedagem, pedágio)
3. Implementar **acerto de viagem** com cálculo automático de saldos
4. Implementar **devolução de adiantamentos** e reembolsos
5. Implementar gestão de **cartões de motorista** com controle de saldo
6. Implementar **validações operacionais** (quilometragem, conferência de passageiros)
7. Criar estrutura básica para **documentos fiscais** (NFS-e, CT-e)

### 1.3 Escopo

**O que SERÁ implementado:**
- 7 novos módulos backend (Go)
- 6 novas páginas frontend (React)
- 1 migration com 9 novas tabelas e 6 enums
- Integração completa entre módulos
- Validações de negócio
- Interface de usuário para todas as operações

**O que NÃO será implementado nesta fase:**
- Integração real com sistemas de NF-e/CT-e (apenas estrutura)
- Upload de comprovantes de despesas (v2)
- Mobile app para motoristas
- Dashboard financeiro avançado
- Relatórios em PDF/Excel

---

## 2. ARQUITETURA E TECNOLOGIAS

### 2.1 Stack Tecnológico

**Backend:**
- Go 1.22
- Chi router v5
- PostgreSQL (via pgxpool v5)
- JWT auth via Supabase

**Frontend:**
- React 19
- TypeScript
- Vite
- Tailwind CSS

### 2.2 Padrões de Implementação

**Backend (Go):**
```
module_name/
├── model.go       # DTOs, structs de domínio, inputs e outputs
├── repository.go  # Camada de acesso a dados (SQL)
├── service.go     # Lógica de negócios
├── handler.go     # Camada HTTP (controllers)
└── handler_test.go # Testes (opcional)
```

**Frontend (React):**
- Páginas CRUD usam componente `CRUDListPage` reutilizável
- Integração via `apiGet/apiPost/apiPatch` em `src/services/api.ts`
- Estados gerenciados com `useState` e `useEffect`
- Validações HTML5 + validações customizadas

---

## 3. BANCO DE DADOS - MIGRATION 0008

### 3.1 Arquivo da Migration

**Path:** `c:\Users\Geinfo\schumacher-tur\apps\api\migrations\0008_financial_procedures.sql`

**Conteúdo completo:**

```sql
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

-- 1. Trip Advances (Adiantamentos de Viagem)
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

-- 2. Trip Expenses (Despesas de Viagem)
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

-- 3. Trip Settlements (Acertos de Viagem)
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

-- 4. Advance Returns (Devoluções de Adiantamento)
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

-- 5. Driver Cards (Cartões de Motorista)
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
```

### 3.2 Explicação das Tabelas

**trip_advances:**
- Armazena adiantamentos dados aos motoristas antes da viagem
- Estados: PENDING → DELIVERED → SETTLED
- `delivered_at` marca quando o dinheiro foi efetivamente entregue ao motorista

**trip_expenses:**
- Registra todas as despesas durante a viagem
- `payment_method`: como foi paga (adiantamento, cartão, pessoal, empresa)
- `driver_card_id`: vincula à despesa paga com cartão
- Precisa de aprovação (`is_approved`) antes de entrar no acerto

**trip_settlements:**
- Acerto financeiro pós-viagem (um por viagem)
- Calcula automaticamente `balance = advance_amount - expenses_total`
- Se balance > 0: motorista deve devolver (`amount_to_return`)
- Se balance < 0: empresa deve reembolsar (`amount_to_reimburse`)
- Workflow: DRAFT → UNDER_REVIEW → APPROVED/REJECTED → COMPLETED

**advance_returns:**
- Registra devoluções de saldo não utilizado
- Vinculado ao adiantamento e opcionalmente ao acerto

**driver_cards:**
- Cartões de crédito/débito dos motoristas
- Controla saldo atual e permite bloqueio
- Tipos: FUEL (só combustível), MULTIPURPOSE (geral), FOOD (alimentação)

**driver_card_transactions:**
- Histórico de todas as operações no cartão
- CREDIT: adicionar saldo, DEBIT: gastar, ADJUSTMENT: correção, REFUND: estorno
- Registra saldo antes e depois para auditoria

**trip_validations:**
- Validações operacionais (quilometragem, passageiros)
- `distance_km` é calculado automaticamente (campo GENERATED)

**fiscal_documents:**
- Estrutura para NFS-e, CT-e (implementação básica)
- `metadata` JSONB para dados flexíveis do provedor

**expense_categories:**
- Planos de conta (alimentação, hospedagem, etc)
- Pré-populado com categorias padrão

---

## 4. BACKEND - IMPLEMENTAÇÃO DOS MÓDULOS

### 4.1 Módulo: trip_advances

**Path:** `c:\Users\Geinfo\schumacher-tur\apps\api\internal\trip_advances\`

#### 4.1.1 Model (`model.go`)

```go
package trip_advances

import "time"

type TripAdvance struct {
	ID          string     `json:"id"`
	TripID      string     `json:"trip_id"`
	DriverID    string     `json:"driver_id"`
	Amount      float64    `json:"amount"`
	Status      string     `json:"status"`
	Purpose     *string    `json:"purpose"`
	DeliveredAt *time.Time `json:"delivered_at"`
	DeliveredBy *string    `json:"delivered_by"`
	SettledAt   *time.Time `json:"settled_at"`
	Notes       *string    `json:"notes"`
	CreatedBy   *string    `json:"created_by"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type CreateTripAdvanceInput struct {
	TripID   string  `json:"trip_id"`
	DriverID string  `json:"driver_id"`
	Amount   float64 `json:"amount"`
	Purpose  *string `json:"purpose"`
	Notes    *string `json:"notes"`
}

type UpdateTripAdvanceInput struct {
	Amount  *float64 `json:"amount"`
	Purpose *string  `json:"purpose"`
	Notes   *string  `json:"notes"`
}

type DeliverAdvanceInput struct {
	DeliveredBy string `json:"delivered_by"`
}

type ListFilter struct {
	TripID   string
	DriverID string
	Status   string
	Limit    int
	Offset   int
}
```

#### 4.1.2 Repository (`repository.go`)

```go
package trip_advances

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) List(ctx context.Context, filter ListFilter) ([]TripAdvance, error) {
	query := `select id, trip_id, driver_id, amount, status, purpose, delivered_at, delivered_by, settled_at, notes, created_by, created_at, updated_at from trip_advances`
	args := []interface{}{}
	clauses := []string{}

	if filter.TripID != "" {
		args = append(args, filter.TripID)
		clauses = append(clauses, fmt.Sprintf("trip_id=$%d", len(args)))
	}
	if filter.DriverID != "" {
		args = append(args, filter.DriverID)
		clauses = append(clauses, fmt.Sprintf("driver_id=$%d", len(args)))
	}
	if filter.Status != "" {
		args = append(args, filter.Status)
		clauses = append(clauses, fmt.Sprintf("status=$%d", len(args)))
	}

	if len(clauses) > 0 {
		query += " where " + strings.Join(clauses, " and ")
	}

	query += " order by created_at desc"

	limit := filter.Limit
	if limit <= 0 || limit > 500 {
		limit = 200
	}
	args = append(args, limit)
	query += fmt.Sprintf(" limit $%d", len(args))

	if filter.Offset > 0 {
		args = append(args, filter.Offset)
		query += fmt.Sprintf(" offset $%d", len(args))
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []TripAdvance{}
	for rows.Next() {
		var item TripAdvance
		if err := rows.Scan(&item.ID, &item.TripID, &item.DriverID, &item.Amount, &item.Status, &item.Purpose, &item.DeliveredAt, &item.DeliveredBy, &item.SettledAt, &item.Notes, &item.CreatedBy, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) Get(ctx context.Context, id uuid.UUID) (TripAdvance, error) {
	var item TripAdvance
	row := r.pool.QueryRow(ctx, `select id, trip_id, driver_id, amount, status, purpose, delivered_at, delivered_by, settled_at, notes, created_by, created_at, updated_at from trip_advances where id=$1`, id)
	if err := row.Scan(&item.ID, &item.TripID, &item.DriverID, &item.Amount, &item.Status, &item.Purpose, &item.DeliveredAt, &item.DeliveredBy, &item.SettledAt, &item.Notes, &item.CreatedBy, &item.CreatedAt, &item.UpdatedAt); err != nil {
		return item, err
	}
	return item, nil
}

func (r *Repository) Create(ctx context.Context, input CreateTripAdvanceInput, createdBy *string) (TripAdvance, error) {
	var item TripAdvance
	row := r.pool.QueryRow(ctx,
		`insert into trip_advances (trip_id, driver_id, amount, purpose, notes, created_by)
		 values ($1,$2,$3,$4,$5,$6)
		 returning id, trip_id, driver_id, amount, status, purpose, delivered_at, delivered_by, settled_at, notes, created_by, created_at, updated_at`,
		input.TripID, input.DriverID, input.Amount, input.Purpose, input.Notes, createdBy,
	)
	if err := row.Scan(&item.ID, &item.TripID, &item.DriverID, &item.Amount, &item.Status, &item.Purpose, &item.DeliveredAt, &item.DeliveredBy, &item.SettledAt, &item.Notes, &item.CreatedBy, &item.CreatedAt, &item.UpdatedAt); err != nil {
		return item, err
	}
	return item, nil
}

func (r *Repository) Update(ctx context.Context, id uuid.UUID, input UpdateTripAdvanceInput) (TripAdvance, error) {
	sets := []string{}
	args := []interface{}{}
	idx := 1

	if input.Amount != nil {
		sets = append(sets, fmt.Sprintf("amount=$%d", idx))
		args = append(args, *input.Amount)
		idx++
	}
	if input.Purpose != nil {
		sets = append(sets, fmt.Sprintf("purpose=$%d", idx))
		args = append(args, *input.Purpose)
		idx++
	}
	if input.Notes != nil {
		sets = append(sets, fmt.Sprintf("notes=$%d", idx))
		args = append(args, *input.Notes)
		idx++
	}

	if len(sets) == 0 {
		return r.Get(ctx, id)
	}

	sets = append(sets, fmt.Sprintf("updated_at=now()"))

	args = append(args, id)
	query := fmt.Sprintf(`update trip_advances set %s where id=$%d returning id, trip_id, driver_id, amount, status, purpose, delivered_at, delivered_by, settled_at, notes, created_by, created_at, updated_at`, strings.Join(sets, ", "), idx)

	var item TripAdvance
	row := r.pool.QueryRow(ctx, query, args...)
	if err := row.Scan(&item.ID, &item.TripID, &item.DriverID, &item.Amount, &item.Status, &item.Purpose, &item.DeliveredAt, &item.DeliveredBy, &item.SettledAt, &item.Notes, &item.CreatedBy, &item.CreatedAt, &item.UpdatedAt); err != nil {
		return item, err
	}
	return item, nil
}

func (r *Repository) Deliver(ctx context.Context, id uuid.UUID, input DeliverAdvanceInput) (TripAdvance, error) {
	var item TripAdvance
	row := r.pool.QueryRow(ctx,
		`update trip_advances set status='DELIVERED', delivered_at=now(), delivered_by=$1, updated_at=now() where id=$2 and status='PENDING'
		 returning id, trip_id, driver_id, amount, status, purpose, delivered_at, delivered_by, settled_at, notes, created_by, created_at, updated_at`,
		input.DeliveredBy, id,
	)
	if err := row.Scan(&item.ID, &item.TripID, &item.DriverID, &item.Amount, &item.Status, &item.Purpose, &item.DeliveredAt, &item.DeliveredBy, &item.SettledAt, &item.Notes, &item.CreatedBy, &item.CreatedAt, &item.UpdatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return item, fmt.Errorf("advance not found or already delivered")
		}
		return item, err
	}
	return item, nil
}

func (r *Repository) MarkSettled(ctx context.Context, tx pgx.Tx, tripID string) error {
	_, err := tx.Exec(ctx, `update trip_advances set status='SETTLED', settled_at=now(), updated_at=now() where trip_id=$1 and status='DELIVERED'`, tripID)
	return err
}

func IsNotFound(err error) bool {
	return err == pgx.ErrNoRows
}
```

#### 4.1.3 Service (`service.go`)

```go
package trip_advances

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

var (
	ErrInvalidAmount   = errors.New("amount must be greater than zero")
	ErrMissingTripID   = errors.New("trip_id is required")
	ErrMissingDriverID = errors.New("driver_id is required")
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context, filter ListFilter) ([]TripAdvance, error) {
	return s.repo.List(ctx, filter)
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (TripAdvance, error) {
	return s.repo.Get(ctx, id)
}

func (s *Service) Create(ctx context.Context, input CreateTripAdvanceInput, createdBy *string) (TripAdvance, error) {
	if input.TripID == "" {
		return TripAdvance{}, ErrMissingTripID
	}
	if input.DriverID == "" {
		return TripAdvance{}, ErrMissingDriverID
	}
	if input.Amount <= 0 {
		return TripAdvance{}, ErrInvalidAmount
	}
	return s.repo.Create(ctx, input, createdBy)
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, input UpdateTripAdvanceInput) (TripAdvance, error) {
	if input.Amount != nil && *input.Amount <= 0 {
		return TripAdvance{}, ErrInvalidAmount
	}
	return s.repo.Update(ctx, id, input)
}

func (s *Service) Deliver(ctx context.Context, id uuid.UUID, input DeliverAdvanceInput) (TripAdvance, error) {
	if input.DeliveredBy == "" {
		return TripAdvance{}, errors.New("delivered_by is required")
	}
	return s.repo.Deliver(ctx, id, input)
}
```

#### 4.1.4 Handler (`handler.go`)

```go
package trip_advances

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	httpx "schumacher-tur/api/internal/shared/http"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/trip-advances", func(r chi.Router) {
		r.Get("/", h.list)
		r.Post("/", h.create)
		r.Get("/{advanceId}", h.get)
		r.Patch("/{advanceId}", h.update)
		r.Post("/{advanceId}/deliver", h.deliver)
	})
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	filter, err := parseListFilter(r)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_PAGINATION", "invalid pagination parameters", nil)
		return
	}
	items, err := h.svc.List(r.Context(), filter)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "ADVANCE_LIST_ERROR", "could not list advances", err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusOK, items)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.ParseUUIDParam(r, "advanceId")
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid advance id", nil)
		return
	}
	item, err := h.svc.Get(r.Context(), id)
	if err != nil {
		if IsNotFound(err) {
			httpx.WriteError(w, http.StatusNotFound, "NOT_FOUND", "advance not found", nil)
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "ADVANCE_GET_ERROR", "could not get advance", err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var input CreateTripAdvanceInput
	if err := httpx.DecodeJSON(r, &input); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_BODY", "invalid json", err.Error())
		return
	}

	item, err := h.svc.Create(r.Context(), input, nil)
	if err != nil {
		if errors.Is(err, ErrMissingTripID) || errors.Is(err, ErrMissingDriverID) || errors.Is(err, ErrInvalidAmount) {
			httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "ADVANCE_CREATE_ERROR", "could not create advance", err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, item)
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.ParseUUIDParam(r, "advanceId")
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid advance id", nil)
		return
	}

	var input UpdateTripAdvanceInput
	if err := httpx.DecodeJSON(r, &input); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_BODY", "invalid json", err.Error())
		return
	}

	item, err := h.svc.Update(r.Context(), id, input)
	if err != nil {
		if IsNotFound(err) {
			httpx.WriteError(w, http.StatusNotFound, "NOT_FOUND", "advance not found", nil)
			return
		}
		if errors.Is(err, ErrInvalidAmount) {
			httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "ADVANCE_UPDATE_ERROR", "could not update advance", err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) deliver(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.ParseUUIDParam(r, "advanceId")
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid advance id", nil)
		return
	}

	var input DeliverAdvanceInput
	if err := httpx.DecodeJSON(r, &input); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_BODY", "invalid json", err.Error())
		return
	}

	item, err := h.svc.Deliver(r.Context(), id, input)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "ADVANCE_DELIVER_ERROR", "could not deliver advance", err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

func parseListFilter(r *http.Request) (ListFilter, error) {
	filter := ListFilter{}
	q := r.URL.Query()

	filter.TripID = q.Get("trip_id")
	filter.DriverID = q.Get("driver_id")
	filter.Status = q.Get("status")

	if v := q.Get("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n <= 0 {
			return filter, errors.New("invalid limit")
		}
		filter.Limit = n
	}
	if v := q.Get("offset"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 0 {
			return filter, errors.New("invalid offset")
		}
		filter.Offset = n
	}
	return filter, nil
}
```

---

### 4.2 Módulo: trip_expenses

**Path:** `c:\Users\Geinfo\schumacher-tur\apps\api\internal\trip_expenses\`

**Implementar seguindo o mesmo padrão do trip_advances, com as seguintes diferenças:**

#### Model adicional:
```go
type ApproveExpenseInput struct {
	ApprovedBy string `json:"approved_by"`
}

type RejectExpenseInput struct {
	RejectedBy string `json:"rejected_by"`
	Reason     string `json:"reason"`
}
```

#### Repository adicional:
```go
func (r *Repository) Approve(ctx context.Context, id uuid.UUID, approvedBy string) (TripExpense, error) {
	var item TripExpense
	row := r.pool.QueryRow(ctx,
		`update trip_expenses set is_approved=true, approved_by=$1, approved_at=now(), updated_at=now() where id=$2
		 returning id, trip_id, driver_id, expense_type, amount, description, expense_date, payment_method, driver_card_id, receipt_number, is_approved, approved_by, approved_at, notes, created_by, created_at, updated_at`,
		approvedBy, id,
	)
	// ... scan
	return item, nil
}

func (r *Repository) CreateWithCardDebit(ctx context.Context, input CreateTripExpenseInput, createdBy *string) (TripExpense, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return TripExpense{}, err
	}
	defer tx.Rollback(ctx)

	// 1. Criar despesa
	var expense TripExpense
	row := tx.QueryRow(ctx,
		`insert into trip_expenses (trip_id, driver_id, expense_type, amount, description, expense_date, payment_method, driver_card_id, receipt_number, notes, created_by)
		 values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		 returning id, trip_id, driver_id, expense_type, amount, description, expense_date, payment_method, driver_card_id, receipt_number, is_approved, approved_by, approved_at, notes, created_by, created_at, updated_at`,
		input.TripID, input.DriverID, input.ExpenseType, input.Amount, input.Description, input.ExpenseDate, input.PaymentMethod, input.DriverCardID, input.ReceiptNumber, input.Notes, createdBy,
	)
	// ... scan

	// 2. Se pago com cartão, debitar
	if input.DriverCardID != nil && *input.DriverCardID != "" {
		var balanceBefore, balanceAfter float64
		err := tx.QueryRow(ctx, `select current_balance from driver_cards where id=$1`, *input.DriverCardID).Scan(&balanceBefore)
		if err != nil {
			return TripExpense{}, err
		}

		balanceAfter = balanceBefore - input.Amount
		_, err = tx.Exec(ctx, `update driver_cards set current_balance=$1, updated_at=now() where id=$2`, balanceAfter, *input.DriverCardID)
		if err != nil {
			return TripExpense{}, err
		}

		// Criar transação
		_, err = tx.Exec(ctx,
			`insert into driver_card_transactions (card_id, transaction_type, amount, balance_before, balance_after, description, trip_expense_id, performed_by)
			 values ($1, 'DEBIT', $2, $3, $4, $5, $6, $7)`,
			*input.DriverCardID, input.Amount, balanceBefore, balanceAfter, "Despesa: "+input.Description, expense.ID, createdBy,
		)
		if err != nil {
			return TripExpense{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return TripExpense{}, err
	}
	return expense, nil
}
```

#### Handler adicional:
```go
r.Post("/{expenseId}/approve", h.approve)
r.Post("/{expenseId}/reject", h.reject)
```

---

### 4.3 Módulo: trip_settlements (CRÍTICO)

Este é o módulo mais complexo, responsável pelo cálculo automático do acerto.

**Path:** `c:\Users\Geinfo\schumacher-tur\apps\api\internal\trip_settlements\`

#### Service (`service.go`) - LÓGICA CRÍTICA:

```go
package trip_settlements

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"schumacher-tur/api/internal/trip_advances"
	"schumacher-tur/api/internal/trip_expenses"
)

var (
	ErrSettlementExists  = errors.New("settlement already exists for this trip")
	ErrInvalidStatus     = errors.New("invalid status transition")
	ErrMissingTripID     = errors.New("trip_id is required")
	ErrNoAdvances        = errors.New("no delivered advances found for trip")
)

type Service struct {
	repo         *Repository
	advancesRepo *trip_advances.Repository
	expensesRepo *trip_expenses.Repository
	pool         *pgxpool.Pool
}

func NewService(repo *Repository, advancesRepo *trip_advances.Repository, expensesRepo *trip_expenses.Repository, pool *pgxpool.Pool) *Service {
	return &Service{
		repo:         repo,
		advancesRepo: advancesRepo,
		expensesRepo: expensesRepo,
		pool:         pool,
	}
}

func (s *Service) Create(ctx context.Context, input CreateSettlementInput, createdBy *string) (TripSettlement, error) {
	if input.TripID == "" {
		return TripSettlement{}, ErrMissingTripID
	}

	// Verificar se já existe acerto para essa viagem
	existing, err := s.repo.GetByTrip(ctx, input.TripID)
	if err == nil && existing.ID != "" {
		return TripSettlement{}, ErrSettlementExists
	}

	// 1. Buscar adiantamentos entregues
	advances, err := s.advancesRepo.List(ctx, trip_advances.ListFilter{
		TripID: input.TripID,
		Status: "DELIVERED",
	})
	if err != nil {
		return TripSettlement{}, err
	}
	if len(advances) == 0 {
		return TripSettlement{}, ErrNoAdvances
	}

	// 2. Calcular total de adiantamentos
	var advanceAmount float64
	for _, adv := range advances {
		advanceAmount += adv.Amount
	}

	// 3. Buscar despesas aprovadas
	expenses, err := s.expensesRepo.List(ctx, trip_expenses.ListFilter{
		TripID:     input.TripID,
		IsApproved: true,
	})
	if err != nil {
		return TripSettlement{}, err
	}

	// 4. Calcular total de despesas
	var expensesTotal float64
	for _, exp := range expenses {
		expensesTotal += exp.Amount
	}

	// 5. Calcular saldo (balance)
	balance := advanceAmount - expensesTotal

	// 6. Determinar valores a devolver ou reembolsar
	amountToReturn := 0.0
	amountToReimburse := 0.0
	if balance > 0 {
		amountToReturn = balance // Motorista deve devolver
	} else if balance < 0 {
		amountToReimburse = -balance // Empresa deve reembolsar
	}

	// 7. Criar settlement
	data := CreateSettlementData{
		TripID:            input.TripID,
		DriverID:          advances[0].DriverID,
		AdvanceAmount:     advanceAmount,
		ExpensesTotal:     expensesTotal,
		Balance:           balance,
		AmountToReturn:    amountToReturn,
		AmountToReimburse: amountToReimburse,
	}

	return s.repo.Create(ctx, data, createdBy)
}

func (s *Service) Review(ctx context.Context, id uuid.UUID, reviewedBy string) (TripSettlement, error) {
	return s.repo.UpdateStatus(ctx, id, "UNDER_REVIEW", &reviewedBy, nil)
}

func (s *Service) Approve(ctx context.Context, id uuid.UUID, approvedBy string) (TripSettlement, error) {
	return s.repo.UpdateStatus(ctx, id, "APPROVED", nil, &approvedBy)
}

func (s *Service) Reject(ctx context.Context, id uuid.UUID, input RejectSettlementInput) (TripSettlement, error) {
	// Rejeitar apenas atualiza status e adiciona notas
	return s.repo.Reject(ctx, id, input.RejectedBy, input.Reason)
}

func (s *Service) Complete(ctx context.Context, id uuid.UUID) (TripSettlement, error) {
	// Só pode completar se estiver APPROVED
	settlement, err := s.repo.Get(ctx, id)
	if err != nil {
		return TripSettlement{}, err
	}
	if settlement.Status != "APPROVED" {
		return TripSettlement{}, errors.New("only approved settlements can be completed")
	}

	// Marcar adiantamentos como SETTLED
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return TripSettlement{}, err
	}
	defer tx.Rollback(ctx)

	if err := s.advancesRepo.MarkSettled(ctx, tx, settlement.TripID); err != nil {
		return TripSettlement{}, err
	}

	var completed TripSettlement
	row := tx.QueryRow(ctx,
		`update trip_settlements set status='COMPLETED', completed_at=now(), updated_at=now() where id=$1
		 returning id, trip_id, driver_id, status, advance_amount, expenses_total, balance, amount_to_return, amount_to_reimburse, reviewed_by, reviewed_at, approved_by, approved_at, completed_at, notes, created_by, created_at, updated_at`,
		id,
	)
	// ... scan

	if err := tx.Commit(ctx); err != nil {
		return TripSettlement{}, err
	}

	return completed, nil
}
```

---

### 4.4 Módulos Restantes (Estrutura Similar)

**driver_cards:**
- CRUD básico de cartões
- Endpoint adicional: `POST /driver-cards/{cardId}/adjust-balance` para ajustes manuais
- Repository deve criar transação ao ajustar saldo

**trip_validations:**
- CRUD básico
- Validação: odometer_final >= odometer_initial

**advance_returns:**
- CRUD básico
- Vincula trip_advance_id e trip_settlement_id

**fiscal_documents:**
- CRUD básico (estrutura para futuro)

---

### 4.5 Registro no main.go

**Path:** `c:\Users\Geinfo\schumacher-tur\apps\api\cmd\api\main.go`

**Adicionar imports:**
```go
import (
	"schumacher-tur/api/internal/trip_advances"
	"schumacher-tur/api/internal/trip_expenses"
	"schumacher-tur/api/internal/trip_settlements"
	"schumacher-tur/api/internal/driver_cards"
	"schumacher-tur/api/internal/trip_validations"
	"schumacher-tur/api/internal/advance_returns"
	"schumacher-tur/api/internal/fiscal_documents"
)
```

**Dentro do `r.Group` autenticado, adicionar:**
```go
// Financial procedures
advancesRepo := trip_advances.NewRepository(pool)
advancesHandler := trip_advances.NewHandler(trip_advances.NewService(advancesRepo))
advancesHandler.RegisterRoutes(pr)

expensesRepo := trip_expenses.NewRepository(pool)
expensesHandler := trip_expenses.NewHandler(trip_expenses.NewService(expensesRepo))
expensesHandler.RegisterRoutes(pr)

settlementsRepo := trip_settlements.NewRepository(pool)
settlementsHandler := trip_settlements.NewHandler(trip_settlements.NewService(settlementsRepo, advancesRepo, expensesRepo, pool))
settlementsHandler.RegisterRoutes(pr)

cardsHandler := driver_cards.NewHandler(driver_cards.NewService(driver_cards.NewRepository(pool)))
cardsHandler.RegisterRoutes(pr)

validationsHandler := trip_validations.NewHandler(trip_validations.NewService(trip_validations.NewRepository(pool)))
validationsHandler.RegisterRoutes(pr)

returnsHandler := advance_returns.NewHandler(advance_returns.NewService(advance_returns.NewRepository(pool)))
returnsHandler.RegisterRoutes(pr)

fiscalHandler := fiscal_documents.NewHandler(fiscal_documents.NewService(fiscal_documents.NewRepository(pool)))
fiscalHandler.RegisterRoutes(pr)
```

---

## 5. FRONTEND - IMPLEMENTAÇÃO DAS PÁGINAS

### 5.1 Tipos Compartilhados

**Path:** `c:\Users\Geinfo\schumacher-tur\apps\app\src\types\financial.ts`

```typescript
export type TripAdvance = {
  id: string;
  trip_id: string;
  driver_id: string;
  amount: number;
  status: 'PENDING' | 'DELIVERED' | 'SETTLED' | 'CANCELLED';
  purpose?: string;
  delivered_at?: string;
  delivered_by?: string;
  settled_at?: string;
  notes?: string;
  created_by?: string;
  created_at: string;
  updated_at: string;
};

export type TripExpense = {
  id: string;
  trip_id: string;
  driver_id: string;
  expense_type: 'FUEL' | 'FOOD' | 'LODGING' | 'TOLL' | 'MAINTENANCE' | 'OTHER';
  amount: number;
  description: string;
  expense_date: string;
  payment_method: 'ADVANCE' | 'CARD' | 'PERSONAL' | 'COMPANY';
  driver_card_id?: string;
  receipt_number?: string;
  is_approved: boolean;
  approved_by?: string;
  approved_at?: string;
  notes?: string;
  created_by?: string;
  created_at: string;
  updated_at: string;
};

export type TripSettlement = {
  id: string;
  trip_id: string;
  driver_id: string;
  status: 'DRAFT' | 'UNDER_REVIEW' | 'APPROVED' | 'REJECTED' | 'COMPLETED';
  advance_amount: number;
  expenses_total: number;
  balance: number;
  amount_to_return: number;
  amount_to_reimburse: number;
  reviewed_by?: string;
  reviewed_at?: string;
  approved_by?: string;
  approved_at?: string;
  completed_at?: string;
  notes?: string;
  created_by?: string;
  created_at: string;
  updated_at: string;
};

export type DriverCard = {
  id: string;
  driver_id: string;
  card_number: string;
  card_type: 'FUEL' | 'MULTIPURPOSE' | 'FOOD';
  current_balance: number;
  is_active: boolean;
  is_blocked: boolean;
  issued_at: string;
  blocked_at?: string;
  blocked_by?: string;
  block_reason?: string;
  notes?: string;
  created_at: string;
  updated_at: string;
};

export type TripValidation = {
  id: string;
  trip_id: string;
  odometer_initial?: number;
  odometer_final?: number;
  distance_km?: number;
  passengers_expected: number;
  passengers_boarded: number;
  passengers_no_show: number;
  validation_notes?: string;
  validated_by?: string;
  validated_at?: string;
  created_at: string;
  updated_at: string;
};
```

### 5.2 Labels e Utilitários

**Path:** `c:\Users\Geinfo\schumacher-tur\apps\app\src\utils\financialLabels.ts`

```typescript
export const advanceStatusLabel: Record<string, string> = {
  PENDING: 'Pendente',
  DELIVERED: 'Entregue',
  SETTLED: 'Acertado',
  CANCELLED: 'Cancelado',
};

export const expenseTypeLabel: Record<string, string> = {
  FUEL: 'Combustível',
  FOOD: 'Alimentação',
  LODGING: 'Hospedagem',
  TOLL: 'Pedágio',
  MAINTENANCE: 'Manutenção',
  OTHER: 'Outros',
};

export const paymentMethodLabel: Record<string, string> = {
  ADVANCE: 'Adiantamento',
  CARD: 'Cartão',
  PERSONAL: 'Pessoal',
  COMPANY: 'Empresa',
};

export const settlementStatusLabel: Record<string, string> = {
  DRAFT: 'Rascunho',
  UNDER_REVIEW: 'Em Revisão',
  APPROVED: 'Aprovado',
  REJECTED: 'Rejeitado',
  COMPLETED: 'Concluído',
};

export const cardTypeLabel: Record<string, string> = {
  FUEL: 'Combustível',
  MULTIPURPOSE: 'Múltiplo Propósito',
  FOOD: 'Alimentação',
};

export function formatCurrency(value: number): string {
  return new Intl.NumberFormat('pt-BR', {
    style: 'currency',
    currency: 'BRL',
  }).format(value);
}
```

### 5.3 Página: TripAdvances

**Path:** `c:\Users\Geinfo\schumacher-tur\apps\app\src\pages\TripAdvances\index.tsx`

```typescript
import { useMemo, useState } from "react";
import CRUDListPage, {
  type ColumnConfig,
  type FormFieldConfig,
} from "../../components/layout/CRUDListPage";
import StatusBadge from "../../components/StatusBadge";
import { apiGet, apiPost, apiPatch } from "../../services/api";
import type { TripAdvance } from "../../types/financial";
import { advanceStatusLabel, formatCurrency } from "../../utils/financialLabels";
import { formatDateTime } from "../../utils/format";

type TripAdvanceForm = {
  trip_id: string;
  driver_id: string;
  amount: number;
  purpose?: string;
  notes?: string;
};

export default function TripAdvances() {
  const [trips, setTrips] = useState<{ id: string; route_name: string }[]>([]);
  const [drivers, setDrivers] = useState<{ id: string; name: string }[]>([]);

  const formFields: FormFieldConfig<TripAdvanceForm>[] = useMemo(() => [
    {
      key: "trip_id",
      label: "Viagem",
      type: "select",
      required: true,
      options: [
        { label: "Selecione a viagem", value: "" },
        ...trips.map((t) => ({ label: t.route_name, value: t.id })),
      ],
    },
    {
      key: "driver_id",
      label: "Motorista",
      type: "select",
      required: true,
      options: [
        { label: "Selecione o motorista", value: "" },
        ...drivers.map((d) => ({ label: d.name, value: d.id })),
      ],
    },
    {
      key: "amount",
      label: "Valor",
      type: "number",
      required: true,
      hint: "Valor do adiantamento em reais",
    },
    {
      key: "purpose",
      label: "Finalidade",
      type: "textarea",
      hint: "Descreva o propósito do adiantamento",
    },
    {
      key: "notes",
      label: "Observações",
      type: "textarea",
    },
  ], [trips, drivers]);

  const columns: ColumnConfig<TripAdvance>[] = [
    { label: "Viagem", accessor: (item) => item.trip_id },
    { label: "Motorista", accessor: (item) => item.driver_id },
    { label: "Valor", accessor: (item) => formatCurrency(item.amount) },
    {
      label: "Status",
      render: (item) => <StatusBadge tone={getStatusTone(item.status)}>{advanceStatusLabel[item.status]}</StatusBadge>,
    },
    { label: "Criado em", accessor: (item) => formatDateTime(item.created_at) },
  ];

  return (
    <CRUDListPage<TripAdvance, TripAdvanceForm>
      title="Adiantamentos de Viagem"
      subtitle="Gestão de adiantamentos para motoristas"
      formFields={formFields}
      columns={columns}
      fetchItems={async ({ page, pageSize }) => {
        const data = await apiGet<TripAdvance[]>(`/trip-advances?limit=${pageSize}&offset=${page * pageSize}`);
        // Carregar dados auxiliares
        const [tripsData, driversData] = await Promise.all([
          apiGet<any[]>("/trips?limit=500"),
          apiGet<any[]>("/drivers?limit=500"),
        ]);
        setTrips(tripsData);
        setDrivers(driversData);
        return data;
      }}
      createItem={(form) => apiPost("/trip-advances", form)}
      updateItem={(id, form) => apiPatch(`/trip-advances/${id}`, form)}
      getId={(item) => item.id}
    />
  );
}

function getStatusTone(status: string) {
  switch (status) {
    case "PENDING": return "warning";
    case "DELIVERED": return "info";
    case "SETTLED": return "success";
    case "CANCELLED": return "danger";
    default: return "neutral";
  }
}
```

### 5.4 Página: TripExpenses

**Seguir o mesmo padrão de TripAdvances, com:**
- Seletor de `expense_type` (FUEL, FOOD, etc)
- Seletor de `payment_method` (ADVANCE, CARD, etc)
- Botão "Aprovar" para despesas pendentes
- Filtro por status de aprovação

### 5.5 Página: TripSettlements

**Path:** `c:\Users\Geinfo\schumacher-tur\apps\app\src\pages\TripSettlements\index.tsx`

```typescript
import { useState } from "react";
import CRUDListPage, { type ColumnConfig } from "../../components/layout/CRUDListPage";
import StatusBadge from "../../components/StatusBadge";
import { apiGet, apiPost } from "../../services/api";
import type { TripSettlement } from "../../types/financial";
import { settlementStatusLabel, formatCurrency } from "../../utils/financialLabels";
import { formatDateTime } from "../../utils/format";

type SettlementForm = {
  trip_id: string;
};

export default function TripSettlements() {
  const [trips, setTrips] = useState<{ id: string; route_name: string }[]>([]);

  const formFields = [
    {
      key: "trip_id" as const,
      label: "Viagem",
      type: "select" as const,
      required: true,
      options: [
        { label: "Selecione a viagem", value: "" },
        ...trips.map((t) => ({ label: t.route_name, value: t.id })),
      ],
    },
  ];

  const columns: ColumnConfig<TripSettlement>[] = [
    { label: "Viagem", accessor: (item) => item.trip_id },
    { label: "Adiantamento", accessor: (item) => formatCurrency(item.advance_amount) },
    { label: "Despesas", accessor: (item) => formatCurrency(item.expenses_total) },
    {
      label: "Saldo",
      render: (item) => (
        <span style={{ color: item.balance >= 0 ? 'green' : 'red' }}>
          {formatCurrency(item.balance)}
        </span>
      ),
    },
    { label: "A Devolver", accessor: (item) => formatCurrency(item.amount_to_return) },
    { label: "A Reembolsar", accessor: (item) => formatCurrency(item.amount_to_reimburse) },
    {
      label: "Status",
      render: (item) => (
        <StatusBadge tone={getSettlementStatusTone(item.status)}>
          {settlementStatusLabel[item.status]}
        </StatusBadge>
      ),
    },
  ];

  return (
    <CRUDListPage<TripSettlement, SettlementForm>
      title="Acertos de Viagem"
      subtitle="Reconciliação financeira pós-viagem"
      formFields={formFields}
      columns={columns}
      fetchItems={async ({ page, pageSize }) => {
        const data = await apiGet<TripSettlement[]>(`/trip-settlements?limit=${pageSize}&offset=${page * pageSize}`);
        const tripsData = await apiGet<any[]>("/trips?limit=500");
        setTrips(tripsData);
        return data;
      }}
      createItem={(form) => apiPost("/trip-settlements", form)}
      updateItem={undefined} // Acertos não são editáveis manualmente
      getId={(item) => item.id}
    />
  );
}

function getSettlementStatusTone(status: string) {
  switch (status) {
    case "DRAFT": return "neutral";
    case "UNDER_REVIEW": return "info";
    case "APPROVED": return "success";
    case "REJECTED": return "danger";
    case "COMPLETED": return "success";
    default: return "neutral";
  }
}
```

### 5.6 Demais Páginas

**DriverCards, TripValidations, FiscalDocuments:**
- Seguir o mesmo padrão CRUDListPage
- Adaptar formFields e columns conforme os tipos
- DriverCards deve ter campo de saldo atual (somente leitura)

### 5.7 Registro de Rotas

**Path:** `c:\Users\Geinfo\schumacher-tur\apps\app\src\app\App.tsx`

**Adicionar imports:**
```typescript
import TripAdvances from "../pages/TripAdvances";
import TripExpenses from "../pages/TripExpenses";
import TripSettlements from "../pages/TripSettlements";
import DriverCards from "../pages/DriverCards";
import TripValidations from "../pages/TripValidations";
import FiscalDocuments from "../pages/FiscalDocuments";
```

**Dentro do `<Routes>`, adicionar:**
```typescript
<Route path="/trip-advances" element={<TripAdvances />} />
<Route path="/trip-expenses" element={<TripExpenses />} />
<Route path="/trip-settlements" element={<TripSettlements />} />
<Route path="/driver-cards" element={<DriverCards />} />
<Route path="/trip-validations" element={<TripValidations />} />
<Route path="/fiscal-documents" element={<FiscalDocuments />} />
```

### 5.8 Menu de Navegação

**Path:** `c:\Users\Geinfo\schumacher-tur\apps\app\src\app\Layout.tsx`

**Adicionar seção de Financeiro no menu:**
```typescript
<nav>
  {/* ... menu existente ... */}

  <div className="menu-section">
    <h3>Financeiro</h3>
    <Link to="/trip-advances">Adiantamentos</Link>
    <Link to="/trip-expenses">Despesas</Link>
    <Link to="/trip-settlements">Acertos</Link>
    <Link to="/driver-cards">Cartões</Link>
    <Link to="/trip-validations">Validações</Link>
    <Link to="/fiscal-documents">Documentos Fiscais</Link>
  </div>
</nav>
```

---

## 6. ORDEM DE IMPLEMENTAÇÃO

### Fase 1: Banco de Dados (1h)
1. Criar migration `0008_financial_procedures.sql`
2. Executar migration: `psql -f apps/api/migrations/0008_financial_procedures.sql`
3. Verificar criação das tabelas: `\dt` no psql

### Fase 2: Backend - Módulos Independentes (4-6h)
1. `driver_cards` (1h)
2. `trip_validations` (1h)
3. `trip_advances` (1.5h)
4. `trip_expenses` (1.5h)

### Fase 3: Backend - Módulos Dependentes (3-4h)
1. `trip_settlements` (2h) - depende de advances e expenses
2. `advance_returns` (1h)
3. `fiscal_documents` (0.5h - estrutura básica)
4. Registrar handlers em `main.go` (0.5h)

### Fase 4: Frontend - Tipos e Utilitários (1h)
1. Criar `types/financial.ts`
2. Criar `utils/financialLabels.ts`

### Fase 5: Frontend - Páginas (4-5h)
1. `TripAdvances` (1h)
2. `TripExpenses` (1h)
3. `DriverCards` (0.5h)
4. `TripValidations` (0.5h)
5. `TripSettlements` (1.5h)
6. `FiscalDocuments` (0.5h)
7. Registrar rotas e menu (0.5h)

### Fase 6: Testes e Ajustes (2-3h)
1. Testar fluxo completo: criar adiantamento → criar despesas → criar acerto
2. Testar cálculos do settlement
3. Testar débito de cartão ao criar despesa
4. Ajustar bugs encontrados

**TEMPO TOTAL ESTIMADO: 15-20 horas**

---

## 7. TESTES E VALIDAÇÃO

### 7.1 Checklist de Testes Backend

**Trip Advances:**
- [ ] Criar adiantamento com valores válidos
- [ ] Tentar criar com valor negativo (deve falhar)
- [ ] Marcar como entregue
- [ ] Listar por trip_id
- [ ] Listar por driver_id
- [ ] Filtrar por status

**Trip Expenses:**
- [ ] Criar despesa de cada tipo
- [ ] Criar despesa paga com cartão (verificar débito)
- [ ] Aprovar despesa
- [ ] Tentar aprovar despesa já aprovada
- [ ] Listar despesas aprovadas vs não aprovadas

**Trip Settlements:**
- [ ] Criar acerto com adiantamento > despesas
- [ ] Verificar amount_to_return calculado corretamente
- [ ] Criar acerto com despesas > adiantamento
- [ ] Verificar amount_to_reimburse calculado corretamente
- [ ] Tentar criar acerto duplicado para mesma viagem (deve falhar)
- [ ] Aprovar acerto
- [ ] Completar acerto aprovado
- [ ] Verificar adiantamentos marcados como SETTLED

**Driver Cards:**
- [ ] Criar cartão
- [ ] Ajustar saldo manualmente
- [ ] Bloquear cartão
- [ ] Desbloquear cartão
- [ ] Visualizar transações

### 7.2 Checklist de Testes Frontend

**Navegação:**
- [ ] Acessar todas as 6 novas páginas via menu
- [ ] Verificar títulos e subtítulos corretos

**CRUD Básico:**
- [ ] Criar registro em cada página
- [ ] Editar registro (onde aplicável)
- [ ] Buscar por texto
- [ ] Paginação funciona

**Funcionalidades Específicas:**
- [ ] TripAdvances: marcar como entregue
- [ ] TripExpenses: aprovar despesa
- [ ] TripSettlements: visualizar cálculos corretos
- [ ] DriverCards: visualizar saldo atual

### 7.3 Testes End-to-End

**Cenário 1: Acerto com saldo positivo**
1. Criar adiantamento de R$ 2000
2. Marcar como entregue
3. Criar despesas:
   - Combustível: R$ 800
   - Alimentação: R$ 300
   - Pedágio: R$ 100
4. Aprovar todas as despesas
5. Criar acerto
6. Verificar:
   - advance_amount = R$ 2000
   - expenses_total = R$ 1200
   - balance = R$ 800
   - amount_to_return = R$ 800
   - amount_to_reimburse = R$ 0
7. Aprovar acerto
8. Completar acerto
9. Verificar adiantamento marcado como SETTLED

**Cenário 2: Acerto com saldo negativo**
1. Criar adiantamento de R$ 1000
2. Marcar como entregue
3. Criar despesas totalizando R$ 1500
4. Aprovar despesas
5. Criar acerto
6. Verificar:
   - balance = -R$ 500
   - amount_to_return = R$ 0
   - amount_to_reimburse = R$ 500

**Cenário 3: Despesa paga com cartão**
1. Criar cartão com saldo R$ 500
2. Criar despesa de R$ 100 paga com cartão
3. Verificar:
   - Cartão tem saldo R$ 400
   - Existe transação de débito
   - Despesa vinculada ao cartão

---

## 8. ARQUIVOS CRÍTICOS PARA IMPLEMENTAÇÃO

### Backend:
1. `apps/api/migrations/0008_financial_procedures.sql` ⭐ COMEÇAR AQUI
2. `apps/api/internal/trip_settlements/service.go` ⭐ LÓGICA CRÍTICA DE CÁLCULO
3. `apps/api/cmd/api/main.go` ⭐ REGISTRO DE HANDLERS

### Frontend:
1. `apps/app/src/types/financial.ts` ⭐ TIPOS COMPARTILHADOS
2. `apps/app/src/utils/financialLabels.ts` ⭐ LABELS E FORMATAÇÃO
3. `apps/app/src/app/App.tsx` ⭐ ROTAS
4. `apps/app/src/app/Layout.tsx` ⭐ MENU

---

## 9. NOTAS FINAIS

### 9.1 Boas Práticas

1. **Sempre usar transações** quando modificar múltiplas tabelas
2. **Validar status transitions** (ex: não permitir PENDING → COMPLETED direto)
3. **Valores monetários** sempre `NUMERIC(10,2)` no banco e `float64` no Go
4. **Campos de auditoria** sempre preencher (created_by, updated_at)
5. **Soft deletes** preferir `is_active=false` a `DELETE`

### 9.2 Segurança

1. **Autenticação obrigatória** em todas as rotas
2. **Validar permissões** antes de aprovar/completar acertos
3. **Não permitir edição** de adiantamentos/despesas após SETTLED
4. **Log de auditoria** em operações críticas

### 9.3 Performance

1. **Índices criados** em todas as foreign keys
2. **Paginação obrigatória** nas listagens
3. **Limit máximo** de 500 registros
4. **Cache** em listas de trips/drivers (futuro)

### 9.4 Extensões Futuras

1. Upload de comprovantes de despesas
2. Dashboard financeiro
3. Relatórios em PDF
4. Integração real com NF-e/CT-e
5. Mobile app para motoristas
6. Aprovações multi-nível
7. Notificações por email/webhook

---

## 10. GLOSSÁRIO

| Termo | Significado |
|-------|-------------|
| Adiantamento | Valor entregue ao motorista antes da viagem |
| Acerto | Reconciliação financeira após a viagem |
| Despesa | Gasto realizado durante a viagem |
| Devolução | Retorno do saldo não utilizado |
| Reembolso | Pagamento ao motorista quando despesas > adiantamento |
| Validação | Conferência de dados operacionais (km, passageiros) |
| Settlement | Termo técnico para "acerto de contas" |

---

**FIM DO PLANO**

Este plano é completo e autocontido. Siga a ordem de implementação, teste cada módulo antes de avançar, e mantenha consistência com os padrões identificados. Boa implementação! 🚀

---

## 11. STATUS DE EXECUÇÃO (05/02/2026)

**Status geral:** Implementação aplicada no código.

**Concluído:**
- Migration `0008_financial_procedures.sql` criada.
- Módulos backend implementados e rotas registradas.
- Tipos e utilitários frontend adicionados.
- Páginas financeiras criadas e menu/rotas atualizados.

**Pendências:**
- Executar a migration no banco de dados.
- Rodar os testes de validação (backend/frontend).
- Validar fluxo completo em ambiente de execução.
