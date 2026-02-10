-- 0010_payments_abacatepay_provider_ref_uq.sql
-- Idempotency guard for AbacatePay webhooks/sync:
-- provider_ref (billing id) must be unique for ABACATEPAY payments.

create unique index if not exists uq_payments_abacatepay_provider_ref
  on public.payments (provider, provider_ref)
  where provider = 'ABACATEPAY' and provider_ref is not null;
