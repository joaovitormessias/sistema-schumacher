# Documentacao da API Backend (Schumacher Tur)

Atualizado em: 2026-04-02

## 1) Visao geral

- Servidor HTTP em Go (chi router).
- Rotas sem prefixo de versao (`/api/v1` nao existe no estado atual).
- Quase todas as rotas exigem JWT de usuario (Supabase), com excecao de health checks e webhook publico.

Base local padrao:

```txt
http://localhost:<PORT>
```

## 2) Convencoes de requisicao

### 2.1 Headers

- `Authorization: Bearer <jwt>` para rotas protegidas.
- `Content-Type: application/json` para rotas com body JSON.

### 2.2 Body JSON (importante)

O backend usa decoder estrito:

- Campos desconhecidos geram erro (`DisallowUnknownFields`).
- Body vazio em endpoints que esperam JSON gera erro.
- Dados extras depois do JSON valido tambem geram erro.

### 2.3 Datas e horas

- Campos `time.Time` no JSON devem ser enviados em RFC3339.
- Alguns filtros aceitam tambem data simples `YYYY-MM-DD` (ex.: pagamentos).

### 2.4 Paginacao e filtros

- Listagens usam `limit` e `offset`.
- `limit` deve ser `> 0`.
- `offset` deve ser `>= 0`.

### 2.5 IDs

- A maioria dos IDs eh `string` (geralmente UUID).
- Alguns endpoints validam UUID explicitamente no path/body.

## 3) Autenticacao

### 3.1 Publicas (sem JWT)

- `GET /health`
- `GET /ready`
- `POST /webhooks/abacatepay`

### 3.2 Protegidas (com JWT)

Todas as demais rotas.

Formato esperado:

```http
Authorization: Bearer <token>
```

## 4) Padrao de resposta de erro

Formato padrao:

```json
{
  "code": "VALIDATION_ERROR",
  "message": "descricao do erro",
  "details": "opcional"
}
```

Obs.: alguns endpoints retornam payload de erro customizado (ex.: bloqueios de workflow/publicacao).

## 5) Endpoints

## 5.1 Infra

- `GET /health`
- `GET /ready`

## 5.2 Usuarios

- `GET /users/me` (atualmente retorna `NOT_IMPLEMENTED`)

## 5.3 Rotas (`/routes`)

- `GET /routes` (query: `search`, `status`, `limit`, `offset`)
- `POST /routes`
- `GET /routes/cities/candidates` (query: `query`, `limit` ate 10)
- `GET /routes/{routeId}`
- `PATCH /routes/{routeId}`
- `POST /routes/{routeId}/publish`
- `POST /routes/{routeId}/duplicate`
- `GET /routes/{routeId}/stops`
- `POST /routes/{routeId}/stops`
- `PATCH /routes/{routeId}/stops/{stopId}`
- `DELETE /routes/{routeId}/stops/{stopId}`
- `GET /routes/{routeId}/segment-prices`
- `PUT /routes/{routeId}/segment-prices`

## 5.4 Viagens (`/trips`)

- `GET /trips` (query: `search`, `status`, `limit`, `offset`)
- `POST /trips`
- `GET /trips/{tripId}`
- `PATCH /trips/{tripId}`
- `GET /trips/{tripId}/seats` (query opcional: `board_stop_id` + `alight_stop_id` juntos)
- `GET /trips/{tripId}/stops`
- `POST /trips/{tripId}/stops`
- `GET /trips/{tripId}/segment-prices`
- `PUT /trips/{tripId}/segment-prices`

## 5.5 Operacao de viagem

### Trip Requests

- `GET /trip-requests` (query: `limit`, `offset`)
- `POST /trip-requests`

### Manifesto

- `GET /trips/{tripId}/manifest`
- `POST /trips/{tripId}/manifest`
- `POST /trips/{tripId}/manifest/sync`
- `PATCH /trips/{tripId}/manifest/{entryId}`

### Autorizacoes

- `GET /trips/{tripId}/authorizations`
- `POST /trips/{tripId}/authorizations`
- `PATCH /trips/{tripId}/authorizations/{authorizationId}`

### Checklist / Relatorio / Conciliacao / Anexos / Workflow

- `GET /trips/{tripId}/checklists/{stage}`
- `PUT /trips/{tripId}/checklists/{stage}`
- `GET /trips/{tripId}/driver-report`
- `PUT /trips/{tripId}/driver-report`
- `GET /trips/{tripId}/reconciliation`
- `PUT /trips/{tripId}/reconciliation`
- `GET /trips/{tripId}/attachments`
- `POST /trips/{tripId}/attachments`
- `POST /trips/{tripId}/workflow/advance`

## 5.6 Reservas e pagamentos

### Bookings (`/bookings`)

- `GET /bookings` (query: `limit`, `offset`)
- `POST /bookings`
- `POST /bookings/checkout`
- `GET /bookings/{bookingId}`
- `PATCH /bookings/{bookingId}`

### Payments (`/payments`)

- `GET /payments` (query: `booking_id`, `status`, `since`, `until`, `paid_since`, `paid_until`, `limit`, `offset`)
- `POST /payments`
- `POST /payments/manual`
- `GET /payments/{paymentId}/status`
- `POST /payments/{paymentId}/sync`

### Webhook publico

- `POST /webhooks/abacatepay`

Headers/query aceitos para seguranca:

- `X-Webhook-Secret: <secret>` ou
- `Authorization: Bearer <secret>` ou
- query `?webhookSecret=<secret>`
- assinatura: `X-AbacatePay-Signature` (ou `X-Webhook-Signature`)

## 5.7 Pricing

- `POST /pricing/quote`
- `GET /pricing/rules` (query: `limit`, `offset`, `scope`, `scope_id`, `rule_type`, `is_active`)
- `POST /pricing/rules`
- `GET /pricing/rules/{ruleId}`
- `PATCH /pricing/rules/{ruleId}`

## 5.8 Frota e motoristas

### Buses

- `GET /buses` (query: `search`, `limit`, `offset`)
- `POST /buses`
- `GET /buses/{busId}`
- `PATCH /buses/{busId}`

### Drivers

- `GET /drivers` (query: `search`, `limit`, `offset`)
- `POST /drivers`
- `GET /drivers/{driverId}`
- `PATCH /drivers/{driverId}`

### Driver Cards

- `GET /driver-cards` (query: `driver_id`, `is_active`, `is_blocked`, `limit`, `offset`)
- `POST /driver-cards`
- `GET /driver-cards/{cardId}`
- `PATCH /driver-cards/{cardId}`
- `POST /driver-cards/{cardId}/block` (body opcional)
- `POST /driver-cards/{cardId}/unblock`
- `GET /driver-cards/{cardId}/transactions` (query: `limit`, `offset`)
- `POST /driver-cards/{cardId}/transactions`

## 5.9 Financeiro de viagem

### Trip Advances

- `GET /trip-advances` (query: `trip_id`, `driver_id`, `status`, `limit`, `offset`)
- `POST /trip-advances`
- `GET /trip-advances/{advanceId}`
- `PATCH /trip-advances/{advanceId}`
- `POST /trip-advances/{advanceId}/deliver`

### Trip Expenses

- `GET /trip-expenses` (query: `trip_id`, `driver_id`, `expense_type`, `payment_method`, `approved`, `limit`, `offset`)
- `POST /trip-expenses`
- `GET /trip-expenses/{expenseId}`
- `PATCH /trip-expenses/{expenseId}`
- `POST /trip-expenses/{expenseId}/approve`

### Trip Settlements

- `GET /trip-settlements` (query: `trip_id`, `driver_id`, `status`, `limit`, `offset`)
- `POST /trip-settlements`
- `GET /trip-settlements/{settlementId}`
- `POST /trip-settlements/{settlementId}/review`
- `POST /trip-settlements/{settlementId}/approve`
- `POST /trip-settlements/{settlementId}/reject`
- `POST /trip-settlements/{settlementId}/complete`

### Trip Validations

- `GET /trip-validations` (query: `trip_id`, `limit`, `offset`)
- `POST /trip-validations`
- `GET /trip-validations/{validationId}`
- `PATCH /trip-validations/{validationId}`

### Advance Returns

- `GET /advance-returns` (query: `trip_advance_id`, `trip_settlement_id`, `limit`, `offset`)
- `POST /advance-returns`
- `GET /advance-returns/{returnId}`

### Fiscal Documents

- `GET /fiscal-documents` (query: `trip_id`, `document_type`, `status`, `limit`, `offset`)
- `POST /fiscal-documents`
- `GET /fiscal-documents/{documentId}`
- `PATCH /fiscal-documents/{documentId}`

## 5.10 Almoxarifado e compras

### Suppliers

- `GET /suppliers` (query: `limit`, `offset`, `active`)
- `POST /suppliers`
- `GET /suppliers/{supplierId}`
- `PATCH /suppliers/{supplierId}`
- `DELETE /suppliers/{supplierId}`

### Products

- `GET /products` (query: `limit`, `offset`, `active`, `category`, `search`)
- `POST /products`
- `GET /products/{productId}`
- `PATCH /products/{productId}`
- `DELETE /products/{productId}`

### Service Orders

- `GET /service-orders` (query: `limit`, `offset`, `status`, `order_type`, `bus_id`)
- `POST /service-orders`
- `GET /service-orders/{orderId}`
- `PATCH /service-orders/{orderId}`
- `POST /service-orders/{orderId}/start`
- `POST /service-orders/{orderId}/close`
- `POST /service-orders/{orderId}/cancel`
- `DELETE /service-orders/{orderId}`

### Purchase Orders

- `GET /purchase-orders` (query: `limit`, `offset`, `status`, `supplier_id`, `service_order_id`)
- `POST /purchase-orders`
- `GET /purchase-orders/{orderId}`
- `PATCH /purchase-orders/{orderId}`
- `POST /purchase-orders/{orderId}/items`
- `DELETE /purchase-orders/{orderId}/items/{itemId}`
- `POST /purchase-orders/{orderId}/send`
- `POST /purchase-orders/{orderId}/receive`
- `POST /purchase-orders/{orderId}/cancel`
- `DELETE /purchase-orders/{orderId}`

### Invoices

- `GET /invoices` (query: `limit`, `offset`, `status`, `supplier_id`, `service_order_id`, `purchase_order_id`, `bus_id`)
- `POST /invoices`
- `GET /invoices/{invoiceId}`
- `PATCH /invoices/{invoiceId}`
- `POST /invoices/{invoiceId}/process`
- `POST /invoices/{invoiceId}/cancel`
- `DELETE /invoices/{invoiceId}`

## 5.11 Importacao e relatorios

### Import XLSX

- `POST /imports/xlsx/upload`
- `POST /imports/xlsx/{batchId}/validate`
- `POST /imports/xlsx/{batchId}/promote`
- `GET /imports/xlsx/{batchId}/report`

### Reports

- `GET /reports/passengers` (query obrigatoria: `trip_id`; `format=json|csv`)

## 6) Schemas de request (JSON)

Os exemplos abaixo cobrem os payloads de escrita mais usados.

## 6.1 Rotas

`POST /routes`

```json
{
  "name": "Linha Serra - Capital",
  "origin_city": "Lages",
  "origin_latitude": -27.815,
  "origin_longitude": -50.325,
  "destination_city": "Florianopolis",
  "destination_latitude": -27.594,
  "destination_longitude": -48.548,
  "is_active": false
}
```

`PATCH /routes/{routeId}`

```json
{
  "name": "Linha Serra - Capital (Ajustada)",
  "origin_city": "Lages",
  "destination_city": "Florianopolis",
  "is_active": true
}
```

`POST /routes/{routeId}/stops`

```json
{
  "city": "Sao Joaquim",
  "latitude": -28.293,
  "longitude": -49.932,
  "stop_order": 2,
  "eta_offset_minutes": 80,
  "notes": "Parada principal"
}
```

`PUT /routes/{routeId}/segment-prices`

```json
{
  "items": [
    {
      "origin_stop_id": "uuid",
      "destination_stop_id": "uuid",
      "price": 89.9,
      "status": "ACTIVE"
    }
  ]
}
```

## 6.2 Viagens

`POST /trips`

```json
{
  "route_id": "uuid",
  "bus_id": "uuid",
  "driver_id": "uuid",
  "fare_id": "uuid",
  "request_id": "uuid",
  "departure_at": "2026-04-10T08:00:00Z",
  "arrival_at": "2026-04-10T14:00:00Z",
  "status": "PLANNED",
  "estimated_km": 280.5,
  "pair_trip_id": "uuid",
  "notes": "Viagem extra"
}
```

`PATCH /trips/{tripId}`

```json
{
  "bus_id": "uuid",
  "driver_id": "uuid",
  "departure_at": "2026-04-10T09:00:00Z",
  "estimated_km": 300,
  "notes": "Ajuste operacional"
}
```

`PUT /trips/{tripId}/segment-prices`

```json
{
  "items": [
    {
      "origin_stop_id": "uuid",
      "destination_stop_id": "uuid",
      "price": 99.9,
      "status": "ACTIVE"
    }
  ]
}
```

## 6.3 Bookings e checkout

`POST /bookings`

```json
{
  "trip_id": "uuid",
  "seat_id": "uuid",
  "board_stop_id": "uuid",
  "alight_stop_id": "uuid",
  "fare_mode": "AUTO",
  "fare_amount_final": 129.9,
  "passenger": {
    "name": "Maria Souza",
    "document": "12345678900",
    "phone": "48999999999",
    "email": "maria@email.com"
  },
  "source": "COUNTER",
  "total_amount": 129.9,
  "deposit_amount": 50.0,
  "remainder_amount": 79.9
}
```

`POST /bookings/checkout`

```json
{
  "trip_id": "uuid",
  "seat_id": "uuid",
  "board_stop_id": "uuid",
  "alight_stop_id": "uuid",
  "fare_mode": "AUTO",
  "fare_amount_final": 129.9,
  "passenger": {
    "name": "Maria Souza",
    "document": "12345678900",
    "phone": "48999999999",
    "email": "maria@email.com"
  },
  "source": "COUNTER",
  "total_amount": 129.9,
  "deposit_amount": 50.0,
  "remainder_amount": 79.9,
  "initial_payment": {
    "method": "PIX",
    "amount": 50.0,
    "description": "Sinal da reserva",
    "notes": "Pago no guiche",
    "customer": {
      "name": "Maria Souza",
      "email": "maria@email.com",
      "phone": "48999999999",
      "document": "12345678900"
    }
  }
}
```

`PATCH /bookings/{bookingId}`

```json
{
  "status": "CONFIRMED"
}
```

Valores aceitos para `status`: `PENDING`, `CONFIRMED`, `CANCELLED`, `EXPIRED`.

## 6.4 Pagamentos

`POST /payments`

```json
{
  "booking_id": "uuid",
  "amount": 79.9,
  "method": "PIX",
  "description": "Pagamento saldo",
  "customer": {
    "name": "Maria Souza",
    "email": "maria@email.com",
    "phone": "48999999999",
    "document": "12345678900"
  }
}
```

Valores aceitos para `method`:

- Provider: `PIX`, `CARD`

`POST /payments/manual`

```json
{
  "booking_id": "uuid",
  "amount": 79.9,
  "method": "CASH",
  "notes": "Recebido em especie"
}
```

Valores aceitos para `method` manual:

- `CASH`, `TRANSFER`, `OTHER`

Filtro `status` em `GET /payments`:

- `PENDING`, `PAID`, `FAILED`, `REFUNDED`, `CANCELLED`

## 6.5 Pricing

`POST /pricing/quote`

```json
{
  "trip_id": "uuid",
  "board_stop_id": "uuid",
  "alight_stop_id": "uuid",
  "fare_mode": "AUTO",
  "fare_amount_final": 129.9
}
```

`POST /pricing/rules`

```json
{
  "name": "Regra promocional",
  "scope": "TRIP",
  "scope_id": "uuid",
  "rule_type": "PERCENT_DISCOUNT",
  "priority": 10,
  "is_active": true,
  "params": {
    "percent": 5
  }
}
```

## 6.6 Operacao de viagem (payloads principais)

`POST /trip-requests`

```json
{
  "route_id": "uuid",
  "source": "SYSTEM",
  "status": "OPEN",
  "requester_name": "Comercial",
  "requester_contact": "ramal 201",
  "requested_departure_at": "2026-04-15T10:00:00Z",
  "notes": "Evento corporativo"
}
```

`POST /trips/{tripId}/manifest`

```json
{
  "booking_passenger_id": "uuid",
  "passenger_name": "Maria Souza",
  "passenger_document": "12345678900",
  "passenger_phone": "48999999999",
  "status": "EXPECTED",
  "seat_number": 12
}
```

`POST /trips/{tripId}/authorizations`

```json
{
  "authority": "DETER",
  "status": "PENDING",
  "protocol_number": "ABC123",
  "license_number": "LIC-01",
  "issued_at": "2026-04-09T12:00:00Z",
  "valid_until": "2026-04-20T23:59:59Z",
  "src_policy_number": "SRC-777",
  "src_valid_until": "2026-04-20T23:59:59Z",
  "exceptional_deadline_ok": false,
  "attachment_id": "uuid",
  "notes": "Aguardando emissao"
}
```

`PUT /trips/{tripId}/checklists/{stage}`

```json
{
  "checklist_data": {
    "items": [
      {
        "label": "Tacografo",
        "ok": true
      }
    ]
  },
  "is_complete": true,
  "documents_checked": true,
  "tachograph_checked": true,
  "receipts_checked": false,
  "rest_compliance_ok": true,
  "notes": "Checklist pre-partida"
}
```

`PUT /trips/{tripId}/driver-report`

```json
{
  "driver_id": "uuid",
  "odometer_start": 150000,
  "odometer_end": 150320,
  "fuel_used_liters": 72.5,
  "incidents": "Nenhum",
  "delays": "10 min de transito",
  "rest_hours": 8,
  "notes": "Viagem regular"
}
```

`PUT /trips/{tripId}/reconciliation`

```json
{
  "total_receipts_amount": 450.75,
  "receipts_validated": true,
  "verified_expense_ids": [
    "uuid"
  ],
  "notes": "Conferencia final"
}
```

`POST /trips/{tripId}/attachments`

```json
{
  "attachment_type": "PDF",
  "storage_bucket": "trip-docs",
  "storage_path": "trips/uuid/doc.pdf",
  "file_name": "doc.pdf",
  "mime_type": "application/pdf",
  "file_size": 123456,
  "metadata": {
    "source": "mobile"
  }
}
```

`POST /trips/{tripId}/workflow/advance`

```json
{
  "to_status": "PASSENGERS_READY"
}
```

Enums principais do modulo de operacao:

- `trip-requests.source`: `EMAIL`, `SYSTEM`
- `trip-requests.status`: `OPEN`, `IN_REVIEW`, `APPROVED`, `REJECTED`
- `manifest.status`: `EXPECTED`, `BOARDED`, `NO_SHOW`, `CANCELLED`
- `authorizations.authority`: `ANTT`, `DETER`, `EXCEPTIONAL`
- `authorizations.status`: `PENDING`, `ISSUED`, `REJECTED`, `EXPIRED`
- `checklists.stage`: `PRE_DEPARTURE`, `RETURN`
- `workflow.to_status`: `REQUESTED`, `PASSENGERS_READY`, `ITINERARY_READY`, `DISPATCH_VALIDATED`, `AUTHORIZED`, `IN_PROGRESS`, `RETURNED`, `RETURN_CHECKED`, `SETTLED`, `CLOSED`

## 6.7 Cadastros e financeiro (esquema resumido)

### Buses

`POST /buses`

```json
{
  "name": "Bus 101",
  "plate": "ABC1D23",
  "capacity": 46,
  "seat_map_name": "2x2",
  "is_active": true,
  "create_seats": true
}
```

### Drivers

`POST /drivers`

```json
{
  "name": "Joao Silva",
  "document": "12345678900",
  "phone": "48999999999",
  "is_active": true
}
```

### Driver Cards

`POST /driver-cards`

```json
{
  "driver_id": "uuid",
  "card_number": "9999888877776666",
  "card_type": "FUEL",
  "current_balance": 300,
  "notes": "Cartao principal"
}
```

`POST /driver-cards/{cardId}/transactions`

```json
{
  "transaction_type": "DEBIT",
  "amount": 120.5,
  "description": "Abastecimento"
}
```

Tipos de cartao: `FUEL`, `MULTIPURPOSE`, `FOOD`.

### Trip Advances

`POST /trip-advances`

```json
{
  "trip_id": "uuid",
  "driver_id": "uuid",
  "amount": 500,
  "purpose": "Pedagios e despesas",
  "notes": "Entrega antecipada"
}
```

`POST /trip-advances/{advanceId}/deliver`

```json
{
  "delivered_by": "uuid"
}
```

### Trip Expenses

`POST /trip-expenses`

```json
{
  "trip_id": "uuid",
  "driver_id": "uuid",
  "expense_type": "FUEL",
  "amount": 180.5,
  "description": "Abastecimento",
  "expense_date": "2026-04-02T16:00:00Z",
  "payment_method": "CARD",
  "driver_card_id": "uuid",
  "receipt_number": "REC-123",
  "notes": "Posto central"
}
```

### Trip Settlements

`POST /trip-settlements`

```json
{
  "trip_id": "uuid",
  "notes": "Fechamento da viagem"
}
```

### Trip Validations

`POST /trip-validations`

```json
{
  "trip_id": "uuid",
  "odometer_initial": 150000,
  "odometer_final": 150320,
  "passengers_expected": 40,
  "passengers_boarded": 38,
  "passengers_no_show": 2,
  "validation_notes": "Tudo conforme"
}
```

### Advance Returns

`POST /advance-returns`

```json
{
  "trip_advance_id": "uuid",
  "trip_settlement_id": "uuid",
  "amount": 50,
  "return_date": "2026-04-03T10:00:00Z",
  "payment_method": "CASH",
  "notes": "Troco devolvido"
}
```

### Fiscal Documents

`POST /fiscal-documents`

```json
{
  "trip_id": "uuid",
  "document_type": "NFE",
  "document_number": "12345",
  "issue_date": "2026-04-02T12:00:00Z",
  "amount": 1000,
  "recipient_name": "Empresa X",
  "recipient_document": "00111222000199",
  "status": "ISSUED",
  "external_id": "ext-01",
  "metadata": {
    "serie": "1"
  }
}
```

## 6.8 Almoxarifado e compras (esquema resumido)

### Suppliers

`POST /suppliers`

```json
{
  "name": "Fornecedor A",
  "document": "00111222000199",
  "phone": "4833334444",
  "email": "contato@fornecedor.com",
  "payment_terms": "30 dias",
  "billing_day": 10,
  "is_active": true,
  "notes": "Atende regiao sul"
}
```

### Products

`POST /products`

```json
{
  "code": "OLEO-15W40",
  "name": "Oleo 15W40",
  "category": "LUBRIFICANTE",
  "unit": "L",
  "min_stock": 20,
  "is_active": true
}
```

### Service Orders

`POST /service-orders`

```json
{
  "bus_id": "uuid",
  "driver_id": "uuid",
  "order_type": "CORRECTIVE",
  "description": "Troca de pastilha de freio",
  "odometer_km": 150120,
  "scheduled_date": "2026-04-05T08:00:00Z",
  "location": "Garagem A",
  "notes": "Prioridade alta"
}
```

### Purchase Orders

`POST /purchase-orders`

```json
{
  "service_order_id": "uuid",
  "supplier_id": "uuid",
  "expected_delivery": "2026-04-08T10:00:00Z",
  "own_delivery": false,
  "discount": 10,
  "freight": 25,
  "notes": "Entrega parcial permitida",
  "items": [
    {
      "product_id": "uuid",
      "quantity": 4,
      "unit_price": 85.5,
      "discount": 0
    }
  ]
}
```

`POST /purchase-orders/{orderId}/items`

```json
{
  "product_id": "uuid",
  "quantity": 2,
  "unit_price": 90,
  "discount": 5
}
```

### Invoices

`POST /invoices`

```json
{
  "invoice_number": "NF-2026-001",
  "barcode": "123456789",
  "supplier_id": "uuid",
  "purchase_order_id": "uuid",
  "service_order_id": "uuid",
  "bus_id": "uuid",
  "issue_date": "2026-04-02T00:00:00Z",
  "issue_time": "14:30",
  "cfop": "5102",
  "payment_type": "PIX",
  "due_date": "2026-05-02T00:00:00Z",
  "discount": 0,
  "freight": 0,
  "notes": "Compra mensal",
  "driver_id": "uuid",
  "odometer_km": 150120,
  "items": [
    {
      "product_id": "uuid",
      "quantity": 3,
      "unit_price": 120,
      "discount": 0
    }
  ]
}
```

## 6.9 Importacao XLSX

`POST /imports/xlsx/upload`

```json
{
  "source_file_name": "arquivo.xlsx",
  "sheets": {
    "aba1": [
      {
        "campo": "valor"
      }
    ]
  }
}
```

## 7) Exemplo rapido (cURL)

```bash
curl -X POST "http://localhost:8080/trips" \
  -H "Authorization: Bearer <jwt>" \
  -H "Content-Type: application/json" \
  -d '{
    "route_id":"<uuid>",
    "bus_id":"<uuid>",
    "departure_at":"2026-04-10T08:00:00Z"
  }'
```

## 8) Fontes no codigo

- `apps/api/cmd/api/main.go`
- `apps/api/internal/*/handler.go`
- `apps/api/internal/*/model.go`
- `apps/api/internal/shared/http/decoder.go`
- `apps/api/internal/shared/http/response.go`
