create table if not exists chat_sessions (
  id uuid primary key default gen_random_uuid(),
  channel text not null default 'WHATSAPP',
  contact_key text not null,
  customer_phone text,
  customer_name text,
  status text not null default 'ACTIVE',
  handoff_status text not null default 'BOT',
  current_owner_user_id uuid references user_profiles(id) on delete set null,
  last_message_at timestamptz,
  last_inbound_at timestamptz,
  last_outbound_at timestamptz,
  metadata jsonb not null default '{}'::jsonb,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  unique (channel, contact_key)
);

create table if not exists chat_messages (
  id uuid primary key default gen_random_uuid(),
  session_id uuid not null references chat_sessions(id) on delete cascade,
  direction text not null,
  kind text not null default 'TEXT',
  provider_message_id text,
  idempotency_key text,
  sender_name text,
  sender_phone text,
  body text,
  payload jsonb not null default '{}'::jsonb,
  normalized_payload jsonb not null default '{}'::jsonb,
  processing_status text not null default 'RECEIVED',
  received_at timestamptz not null default now(),
  sent_at timestamptz,
  created_at timestamptz not null default now()
);

create table if not exists chat_tool_calls (
  id uuid primary key default gen_random_uuid(),
  session_id uuid not null references chat_sessions(id) on delete cascade,
  message_id uuid references chat_messages(id) on delete set null,
  tool_name text not null,
  request_payload jsonb not null default '{}'::jsonb,
  response_payload jsonb not null default '{}'::jsonb,
  status text not null default 'PENDING',
  error_code text,
  error_message text,
  started_at timestamptz not null default now(),
  finished_at timestamptz,
  created_at timestamptz not null default now()
);

create table if not exists chat_handoffs (
  id uuid primary key default gen_random_uuid(),
  session_id uuid not null references chat_sessions(id) on delete cascade,
  requested_by text not null default 'SYSTEM',
  reason text,
  status text not null default 'REQUESTED',
  assigned_user_id uuid references user_profiles(id) on delete set null,
  requested_at timestamptz not null default now(),
  resolved_at timestamptz,
  metadata jsonb not null default '{}'::jsonb,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create table if not exists automation_job_runs (
  id uuid primary key default gen_random_uuid(),
  job_name text not null,
  trigger_source text not null default 'SYSTEM',
  requested_by_user_id uuid references user_profiles(id) on delete set null,
  status text not null default 'PENDING',
  input_payload jsonb not null default '{}'::jsonb,
  result_payload jsonb not null default '{}'::jsonb,
  error_text text,
  started_at timestamptz not null default now(),
  finished_at timestamptz,
  created_at timestamptz not null default now()
);

create table if not exists outbound_messages (
  id uuid primary key default gen_random_uuid(),
  session_id uuid references chat_sessions(id) on delete set null,
  job_run_id uuid references automation_job_runs(id) on delete set null,
  channel text not null default 'WHATSAPP',
  recipient text not null,
  template_name text,
  payload jsonb not null default '{}'::jsonb,
  provider text not null default 'EVOLUTION',
  provider_message_id text,
  idempotency_key text not null,
  status text not null default 'PENDING',
  error_text text,
  scheduled_at timestamptz,
  sent_at timestamptz,
  delivered_at timestamptz,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  unique (idempotency_key)
);

create unique index if not exists idx_chat_messages_provider_message_id
  on chat_messages(provider_message_id)
  where provider_message_id is not null;

create unique index if not exists idx_chat_messages_idempotency_key
  on chat_messages(idempotency_key)
  where idempotency_key is not null;

create unique index if not exists idx_outbound_messages_provider_message_id
  on outbound_messages(provider, provider_message_id)
  where provider_message_id is not null;

create index if not exists idx_chat_sessions_status_last_message_at
  on chat_sessions(status, last_message_at desc nulls last);

create index if not exists idx_chat_sessions_owner_updated_at
  on chat_sessions(current_owner_user_id, updated_at desc);

create index if not exists idx_chat_messages_session_created_at
  on chat_messages(session_id, created_at desc);

create index if not exists idx_chat_messages_status_received_at
  on chat_messages(processing_status, received_at desc);

create index if not exists idx_chat_tool_calls_session_started_at
  on chat_tool_calls(session_id, started_at desc);

create index if not exists idx_chat_tool_calls_status_started_at
  on chat_tool_calls(status, started_at desc);

create index if not exists idx_chat_handoffs_session_requested_at
  on chat_handoffs(session_id, requested_at desc);

create index if not exists idx_chat_handoffs_status_requested_at
  on chat_handoffs(status, requested_at desc);

create index if not exists idx_automation_job_runs_name_started_at
  on automation_job_runs(job_name, started_at desc);

create index if not exists idx_automation_job_runs_status_started_at
  on automation_job_runs(status, started_at desc);

create index if not exists idx_outbound_messages_status_created_at
  on outbound_messages(status, created_at desc);

create index if not exists idx_outbound_messages_session_created_at
  on outbound_messages(session_id, created_at desc nulls last);
