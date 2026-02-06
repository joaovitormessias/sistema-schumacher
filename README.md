# Schumacher Turismo - Monorepo

## Apps
- `apps/web`: site institucional (schumacher.tu.br)
- `apps/app`: sistema interno (app.schumacher.tu.br)
- `apps/api`: API Go (api.schumacher.tu.br)

## Pacotes
- `packages/shared`: tipos e DTOs comuns
- `packages/ui`: componentes reutilizaveis
- `packages/config`: configs compartilhadas

## Scripts
- `pnpm dev`
- `pnpm build`
- `pnpm lint`

## Observacao
As chaves e variaveis sensiveis devem ficar em arquivos `.env.local` (ignorados por git).
