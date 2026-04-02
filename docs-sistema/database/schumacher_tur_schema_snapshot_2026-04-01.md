# Snapshot de Estrutura - schumacher_tur

- Projeto Supabase: `schumacher_tur`
- Project ID: `mgdlxkqfttwbhpwslgxe`
- Capturado em: `2026-04-01`
- Tipo: **somente estrutura** (sem dados)
- Fonte: metadados via `information_schema` e `table_constraints`

## Totais por schema

| Schema | Tabelas | Colunas |
|---|---:|---:|
| `public` | 53 | 635 |
| `auth` | 23 | 239 |
| `storage` | 8 | 66 |
| `realtime` | 3 | 18 |

## Inventario de tabelas por schema

### public

- `advance_returns`
- `booking_passengers`
- `bookings`
- `bus_seats`
- `buses`
- `checkins`
- `driver_card_transactions`
- `driver_cards`
- `drivers`
- `expense_categories`
- `fares`
- `fiscal_documents`
- `invoice_items`
- `invoices`
- `payment_events`
- `payments`
- `pricing_rules`
- `products`
- `purchase_order_items`
- `purchase_orders`
- `roles`
- `route_stops`
- `routes`
- `seat_types`
- `segment_fares`
- `service_orders`
- `stg_xlsx_available_segments`
- `stg_xlsx_bookings`
- `stg_xlsx_manifest_data`
- `stg_xlsx_passengers`
- `stg_xlsx_routes`
- `stg_xlsx_stops`
- `stg_xlsx_trip_stops`
- `stg_xlsx_trips`
- `stock_movements`
- `suppliers`
- `trip_advances`
- `trip_attachments`
- `trip_authorizations`
- `trip_checklists`
- `trip_driver_reports`
- `trip_expenses`
- `trip_manifest_entries`
- `trip_receipt_reconciliations`
- `trip_requests`
- `trip_settlements`
- `trip_stops`
- `trip_validations`
- `trips`
- `user_profiles`
- `user_roles`
- `xlsx_import_batches`
- `xlsx_import_errors`

### auth

- `audit_log_entries`
- `custom_oauth_providers`
- `flow_state`
- `identities`
- `instances`
- `mfa_amr_claims`
- `mfa_challenges`
- `mfa_factors`
- `oauth_authorizations`
- `oauth_client_states`
- `oauth_clients`
- `oauth_consents`
- `one_time_tokens`
- `refresh_tokens`
- `saml_providers`
- `saml_relay_states`
- `schema_migrations`
- `sessions`
- `sso_domains`
- `sso_providers`
- `users`
- `webauthn_challenges`
- `webauthn_credentials`

### storage

- `buckets`
- `buckets_analytics`
- `buckets_vectors`
- `migrations`
- `objects`
- `s3_multipart_uploads`
- `s3_multipart_uploads_parts`
- `vector_indexes`

### realtime

- `messages`
- `schema_migrations`
- `subscription`

## Detalhamento estrutural (colunas, PK e FKs)

| Schema | Tabela | Colunas | PK (colunas) | FKs |
|---|---|---:|---:|---:|
| auth | audit_log_entries | 5 | 1 | 0 |
| auth | custom_oauth_providers | 24 | 1 | 0 |
| auth | flow_state | 17 | 1 | 0 |
| auth | identities | 9 | 1 | 1 |
| auth | instances | 5 | 1 | 0 |
| auth | mfa_amr_claims | 5 | 1 | 1 |
| auth | mfa_challenges | 7 | 1 | 1 |
| auth | mfa_factors | 13 | 1 | 1 |
| auth | oauth_authorizations | 17 | 1 | 2 |
| auth | oauth_client_states | 4 | 1 | 0 |
| auth | oauth_clients | 13 | 1 | 0 |
| auth | oauth_consents | 6 | 1 | 2 |
| auth | one_time_tokens | 7 | 1 | 1 |
| auth | refresh_tokens | 9 | 1 | 1 |
| auth | saml_providers | 9 | 1 | 1 |
| auth | saml_relay_states | 8 | 1 | 2 |
| auth | schema_migrations | 1 | 0 | 0 |
| auth | sessions | 15 | 1 | 2 |
| auth | sso_domains | 5 | 1 | 1 |
| auth | sso_providers | 5 | 1 | 0 |
| auth | users | 35 | 1 | 0 |
| auth | webauthn_challenges | 6 | 1 | 1 |
| auth | webauthn_credentials | 14 | 1 | 1 |
| public | advance_returns | 9 | 1 | 2 |
| public | booking_passengers | 20 | 1 | 5 |
| public | bookings | 12 | 1 | 1 |
| public | bus_seats | 6 | 1 | 2 |
| public | buses | 7 | 1 | 0 |
| public | checkins | 6 | 1 | 2 |
| public | driver_card_transactions | 10 | 1 | 2 |
| public | driver_cards | 14 | 1 | 1 |
| public | drivers | 6 | 1 | 0 |
| public | expense_categories | 6 | 1 | 0 |
| public | fares | 9 | 1 | 0 |
| public | fiscal_documents | 13 | 1 | 1 |
| public | invoice_items | 8 | 1 | 2 |
| public | invoices | 24 | 1 | 5 |
| public | payment_events | 5 | 1 | 1 |
| public | payments | 11 | 1 | 1 |
| public | pricing_rules | 10 | 1 | 0 |
| public | products | 11 | 1 | 0 |
| public | purchase_order_items | 9 | 1 | 2 |
| public | purchase_orders | 16 | 1 | 2 |
| public | roles | 3 | 1 | 0 |
| public | route_stops | 8 | 1 | 1 |
| public | routes | 8 | 1 | 1 |
| public | seat_types | 5 | 1 | 0 |
| public | segment_fares | 12 | 1 | 2 |
| public | service_orders | 18 | 1 | 2 |
| public | stg_xlsx_available_segments | 18 | 1 | 1 |
| public | stg_xlsx_bookings | 34 | 1 | 1 |
| public | stg_xlsx_manifest_data | 21 | 1 | 1 |
| public | stg_xlsx_passengers | 19 | 1 | 1 |
| public | stg_xlsx_routes | 8 | 1 | 1 |
| public | stg_xlsx_stops | 13 | 1 | 1 |
| public | stg_xlsx_trip_stops | 11 | 1 | 1 |
| public | stg_xlsx_trips | 15 | 1 | 1 |
| public | stock_movements | 12 | 1 | 1 |
| public | suppliers | 11 | 1 | 0 |
| public | trip_advances | 13 | 1 | 2 |
| public | trip_attachments | 12 | 1 | 1 |
| public | trip_authorizations | 16 | 1 | 2 |
| public | trip_checklists | 14 | 1 | 1 |
| public | trip_driver_reports | 14 | 1 | 2 |
| public | trip_expenses | 21 | 1 | 3 |
| public | trip_manifest_entries | 12 | 1 | 2 |
| public | trip_receipt_reconciliations | 12 | 1 | 1 |
| public | trip_requests | 11 | 1 | 1 |
| public | trip_settlements | 18 | 1 | 2 |
| public | trip_stops | 11 | 1 | 2 |
| public | trip_validations | 13 | 1 | 1 |
| public | trips | 18 | 1 | 5 |
| public | user_profiles | 4 | 1 | 0 |
| public | user_roles | 3 | 2 | 1 |
| public | xlsx_import_batches | 7 | 1 | 0 |
| public | xlsx_import_errors | 8 | 1 | 1 |
| realtime | messages | 8 | 2 | 0 |
| realtime | schema_migrations | 2 | 1 | 0 |
| realtime | subscription | 8 | 1 | 0 |
| storage | buckets | 11 | 1 | 0 |
| storage | buckets_analytics | 7 | 1 | 0 |
| storage | buckets_vectors | 4 | 0 | 0 |
| storage | migrations | 4 | 0 | 0 |
| storage | objects | 12 | 1 | 1 |
| storage | s3_multipart_uploads | 9 | 1 | 1 |
| storage | s3_multipart_uploads_parts | 10 | 1 | 2 |
| storage | vector_indexes | 9 | 0 | 0 |

## Relacionamentos FK (schema public)

- `advance_returns.trip_advance_id -> trip_advances.id`
- `advance_returns.trip_settlement_id -> trip_settlements.id`
- `booking_passengers.alight_stop_id -> trip_stops.id`
- `booking_passengers.board_stop_id -> trip_stops.id`
- `booking_passengers.booking_id -> bookings.id`
- `booking_passengers.seat_id -> bus_seats.id`
- `booking_passengers.trip_id -> trips.id`
- `bookings.trip_id -> trips.id`
- `bus_seats.bus_id -> buses.id`
- `bus_seats.seat_type_id -> seat_types.id`
- `checkins.booking_passenger_id -> booking_passengers.id`
- `checkins.trip_id -> trips.id`
- `driver_card_transactions.card_id -> driver_cards.id`
- `driver_card_transactions.trip_expense_id -> trip_expenses.id`
- `driver_cards.driver_id -> drivers.id`
- `fiscal_documents.trip_id -> trips.id`
- `invoice_items.invoice_id -> invoices.id`
- `invoice_items.product_id -> products.id`
- `invoices.bus_id -> buses.id`
- `invoices.driver_id -> drivers.id`
- `invoices.purchase_order_id -> purchase_orders.id`
- `invoices.service_order_id -> service_orders.id`
- `invoices.supplier_id -> suppliers.id`
- `payment_events.payment_id -> payments.id`
- `payments.booking_id -> bookings.id`
- `purchase_order_items.product_id -> products.id`
- `purchase_order_items.purchase_order_id -> purchase_orders.id`
- `purchase_orders.service_order_id -> service_orders.id`
- `purchase_orders.supplier_id -> suppliers.id`
- `route_stops.route_id -> routes.id`
- `routes.duplicated_from_route_id -> routes.id`
- `segment_fares.from_route_stop_id -> route_stops.id`
- `segment_fares.to_route_stop_id -> route_stops.id`
- `service_orders.bus_id -> buses.id`
- `service_orders.driver_id -> drivers.id`
- `stg_xlsx_available_segments.batch_id -> xlsx_import_batches.id`
- `stg_xlsx_bookings.batch_id -> xlsx_import_batches.id`
- `stg_xlsx_manifest_data.batch_id -> xlsx_import_batches.id`
- `stg_xlsx_passengers.batch_id -> xlsx_import_batches.id`
- `stg_xlsx_routes.batch_id -> xlsx_import_batches.id`
- `stg_xlsx_stops.batch_id -> xlsx_import_batches.id`
- `stg_xlsx_trip_stops.batch_id -> xlsx_import_batches.id`
- `stg_xlsx_trips.batch_id -> xlsx_import_batches.id`
- `stock_movements.product_id -> products.id`
- `trip_advances.driver_id -> drivers.id`
- `trip_advances.trip_id -> trips.id`
- `trip_attachments.trip_id -> trips.id`
- `trip_authorizations.attachment_id -> trip_attachments.id`
- `trip_authorizations.trip_id -> trips.id`
- `trip_checklists.trip_id -> trips.id`
- `trip_driver_reports.driver_id -> drivers.id`
- `trip_driver_reports.trip_id -> trips.id`
- `trip_expenses.driver_card_id -> driver_cards.id`
- `trip_expenses.driver_id -> drivers.id`
- `trip_expenses.trip_id -> trips.id`
- `trip_manifest_entries.booking_passenger_id -> booking_passengers.id`
- `trip_manifest_entries.trip_id -> trips.id`
- `trip_receipt_reconciliations.trip_id -> trips.id`
- `trip_requests.route_id -> routes.id`
- `trip_settlements.driver_id -> drivers.id`
- `trip_settlements.trip_id -> trips.id`
- `trip_stops.route_stop_id -> route_stops.id`
- `trip_stops.trip_id -> trips.id`
- `trip_validations.trip_id -> trips.id`
- `trips.bus_id -> buses.id`
- `trips.driver_id -> drivers.id`
- `trips.fare_id -> fares.id`
- `trips.request_id -> trip_requests.id`
- `trips.route_id -> routes.id`
- `user_roles.role_id -> roles.id`
- `xlsx_import_errors.batch_id -> xlsx_import_batches.id`
