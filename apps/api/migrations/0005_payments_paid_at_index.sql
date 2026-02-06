-- 0005_payments_paid_at_index.sql
-- Speed up paid_at queries for dashboard filtering.

create index if not exists idx_payments_paid_at
  on payments (paid_at)
  where status = 'PAID';
