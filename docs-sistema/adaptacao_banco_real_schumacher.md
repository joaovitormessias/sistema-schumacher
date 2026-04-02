# Adequacao ao Banco Real `schumacher` (supabase-schumacher)

Data: 2026-04-01  
Projeto real: `balloihmdzatwynlblad`  
Status: banco de producao focado em venda de passagens

## Escopo confirmado do banco real

Tabelas `public` existentes:

- `routes`
- `stops`
- `trips`
- `trip_stops`
- `available_segments`
- `bookings`
- `booking_payment_details`
- `passengers`
- `manifest_data`
- `sheet_sync_queue`

## Gap principal com o sistema atual

O backend/admin atual foi construido sobre um modelo operacional mais amplo (frota, financeiro, compras, etc.) com dezenas de tabelas adicionais (`buses`, `payments`, `trip_expenses`, `route_stops`, etc).

No banco real atual:

- IDs de negocio sao `text` (`route_id`, `trip_id`, `booking_id`, etc), nao `uuid id`.
- Nomes de campos divergem do dominio atual.
- Fluxo de pagamento esta em `booking_payment_details` (nao `payments` + `payment_events`).
- Estrutura de parada usa `stops`/`trip_stops`, sem `route_stops`.

## Mapeamento inicial (essencial para passagens)

### Rotas

- Atual (codigo): `routes.id`, `routes.origin_city`, `routes.destination_city`
- Real: `routes.route_id`, `routes.route_name`, sem `origin_city/destination_city` persistidos

### Viagens

- Atual: `trips.id`, `departure_at`, `arrival_at`, `fare_id`, `driver_id`, `bus_id`
- Real: `trips.trip_id`, `trip_date`, `default_price`, `seats_total`, `seats_available`, `status`, `package_name`

### Paradas/segmentos

- Atual: `route_stops` + `trip_stops(route_stop_id, stop_order...)`
- Real: `stops` + `trip_stops(stop_id, stop_sequence...)` + `available_segments` pronto para venda

### Reservas

- Atual: `bookings` + `booking_passengers`
- Real: `bookings` + `passengers` + `manifest_data`

### Pagamentos

- Atual: `payments` + `payment_events`
- Real: `booking_payment_details` (campos de PIX/stripe e valores consolidados)

## Plano de adequacao (ordem recomendada)

1. Congelar escopo para "venda de passagens" no backend.
2. Criar camada de repositorio compativel com schema real para:
   - `routes` (listagem)
   - `trips` (listagem)
   - `trip_stops`/`available_segments` (cotacao e escolha de trecho)
   - `bookings` + `passengers` (criacao e consulta)
   - `booking_payment_details` (status de pagamento)
3. Ajustar handlers/endpoints para o novo payload sem depender de tabelas inexistentes.
4. Aplicar feature flag para endpoints operacionais nao suportados no banco real (retorno `501` com mensagem clara), evitando erro 500.
5. Ajustar frontend admin para consumir apenas endpoints suportados nesse modo.
6. Validar fluxo ponta a ponta:
   - buscar trechos
   - cotar valor
   - criar reserva
   - registrar/consultar pagamento
   - listar reservas

## Riscos tecnicos

- Tentar "forcar" schema antigo no banco real gera alto risco de regressao.
- Recomendado: adaptar codigo ao schema real, nao o contrario.
- Campos de origem/destino na rota podem precisar ser derivados de `trip_stops`/`stops`.

## Bloqueador de ambiente para API local

Para subir `apps/api` contra esse banco real, ainda falta a `DATABASE_URL` do projeto `schumacher` (pooler + senha do banco).  
`SUPABASE_URL` e chave publishable/anon ja foram identificadas, mas o backend usa conexao Postgres direta.
