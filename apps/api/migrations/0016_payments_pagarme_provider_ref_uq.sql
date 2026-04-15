-- 0016_payments_pagarme_provider_ref_uq.sql
-- Idempotency guard for Pagar.me webhooks/sync:
-- provider_ref (charge id) must be unique for PAGARME payments.

create unique index if not exists uq_payments_pagarme_provider_ref
  on public.payments (provider, provider_ref)
  where provider = 'PAGARME' and provider_ref is not null;
