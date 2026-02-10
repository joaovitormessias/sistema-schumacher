-- 0011_payment_events_lookup_indexes.sql
-- Lookup indexes for payment event audit and provider synchronization.

create index if not exists idx_payment_events_payment_created
  on public.payment_events (payment_id, created_at desc);

create index if not exists idx_payments_provider_ref
  on public.payments (provider_ref)
  where provider_ref is not null;
