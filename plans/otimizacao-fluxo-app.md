# Plano de Otimização do Fluxo do App - Schumacher Turismo

## Análise da Situação Atual

### Estrutura Atual
- **16 páginas** divididas em operacional (10) e financeiro (6)
- **Fluxo de reserva em 3 etapas** separadas (Viagem → Passageiro → Pagamento)
- **Dual payment entry**: Configuração de pagamento na reserva + página de pagamento separada
- **15-20 cliques** necessários para completar uma reserva com pagamento

### Principais Problemas Identificados

1. **Redundância no Fluxo de Pagamento**
   - Step 3 da reserva configura valores (total, sinal, restante)
   - Depois precisa ir para página Payments para processar pagamento real
   - Confuso: usuário não sabe se já pagou ou não após criar reserva

2. **Formulário de Reserva Fragmentado**
   - 3 steps sendo que Step 1 + 2 poderiam ser um só
   - Informações do passageiro são mínimas (só nome é obrigatório)

3. **Módulo Financeiro Disperso**
   - 6 páginas separadas para operações financeiras relacionadas
   - Todas usam mesmo padrão CRUD
   - Navegação complexa entre funcionalidades relacionadas

4. **Falta de Caching**
   - Cada página busca dados independentemente
   - Sem React Query ou state management global
   - Múltiplas chamadas API duplicadas

5. **CTAs Duplicados**
   - "Nova reserva" aparece em 3 lugares (Dashboard, Topbar, Sidebar)

## Objetivos da Otimização

1. ✅ Reduzir de **15-20 cliques para 8-10 cliques** no fluxo completo
2. ✅ Eliminar confusão entre "configurar pagamento" e "processar pagamento"
3. ✅ Consolidar telas relacionadas
4. ✅ Manter todas as funcionalidades existentes
5. ✅ Melhorar performance com caching

## Estratégia de Otimização

### Fase 1: Otimização do Fluxo de Reserva + Pagamento (PRIORITÁRIO)

#### 1.1 Consolidar Reserva em 2 Steps

**Arquivo:** [apps/app/src/pages/Bookings/index.tsx](apps/app/src/pages/Bookings/index.tsx)

**Mudança:**
```
ANTES: Step 1 (Viagem) → Step 2 (Passageiro) → Step 3 (Config Pagamento)
DEPOIS: Step 1 (Viagem + Passageiro) → Step 2 (Pagamento)
```

**Implementação:**
- Manter Step 1 com seleção de trip/stops/seat
- Adicionar campos do passageiro no final do Step 1 (não criar step separado)
- Transformar Step 3 atual em Step 2, MAS com processamento real de pagamento

#### 1.2 Integrar Processamento de Pagamento no Step 2

**Componente Novo:** `BookingPaymentStep.tsx`

**Funcionalidade:**
- Mostrar resumo da reserva (viagem, passageiro, assento)
- Oferecer 2 opções de pagamento:
  1. **"Pagar Total"** (PIX/Cartão) → Gera cobrança pelo valor total
  2. **"Pagar Sinal/Entrada"** → Processa pagamento parcial (mínimo configurável)

**Fluxo integrado:**
- SEMPRE processa pagamento (total ou parcial) no mesmo formulário
- Mostra resultado (QR code PIX, link checkout) inline
- Atualiza status da reserva automaticamente
- Botão "Concluir" só aparece após pagamento processado
- Métodos disponíveis: PIX, Cartão, Dinheiro, Transferência

**Regra de Negócio:**
- Sinal mínimo: 30% do valor total (configurável)
- Não permite criar reserva sem pagamento inicial
- Pagamento manual (dinheiro/transferência) marca como PAID imediatamente

**Benefício:** Elimina ida para página Payments em 90% dos casos + garante comprometimento

#### 1.3 Simplificar Página Payments

**Arquivo:** [apps/app/src/pages/Payments/index.tsx](apps/app/src/pages/Payments/index.tsx)

**Nova finalidade:**
- Processar pagamentos de reservas já existentes (saldo remanescente)
- Consultar status de pagamentos
- Histórico de transações

**Remover:** Não precisa mais ser entry point principal de pagamento

### Fase 2: Consolidar Módulo Financeiro

#### 2.1 Criar Página Unificada "Financeiro"

**Arquivo Novo:** `apps/app/src/pages/Financial/index.tsx`

**Estrutura:** Tabs horizontais (Ant Design)
```
[Adiantamentos] [Despesas] [Acertos] [Cartões] [Validações] [Documentos]
```

**Implementação:**
- Cada tab carrega componente correspondente
- Componentes existentes viram subcomponentes:
  - `TripAdvancesTab.tsx` (conteúdo atual de TripAdvances/index.tsx)
  - `TripExpensesTab.tsx` (conteúdo atual de TripExpenses/index.tsx)
  - `TripSettlementsTab.tsx` (conteúdo atual de TripSettlements/index.tsx)
  - `DriverCardsTab.tsx` (conteúdo atual de DriverCards/index.tsx)
  - `TripValidationsTab.tsx` (conteúdo atual de TripValidations/index.tsx)
  - `FiscalDocumentsTab.tsx` (conteúdo atual de FiscalDocuments/index.tsx)

**Benefício:**
- Reduz de 6 páginas para 1
- Navegação mais rápida entre funcionalidades relacionadas
- Contexto compartilhado entre tabs

#### 2.2 Atualizar Rotas

**Arquivo:** [apps/app/src/app/App.tsx](apps/app/src/app/App.tsx)

**Mudanças:**
```tsx
// Remover rotas individuais
- /trip-advances
- /trip-expenses
- /trip-settlements
- /driver-cards
- /trip-validations
- /fiscal-documents

// Adicionar rota única com tab navigation
+ /financial?tab=advances
+ /financial?tab=expenses
+ /financial?tab=settlements
+ /financial?tab=cards
+ /financial?tab=validations
+ /financial?tab=documents
```

#### 2.3 Atualizar Sidebar

**Arquivo:** [apps/app/src/app/Layout.tsx](apps/app/src/app/Layout.tsx)

**Mudanças:**
```tsx
// Substituir 6 links por 1 com submenu
Financial Section:
  └─ Financeiro [dropdown]
      ├─ Adiantamentos
      ├─ Despesas
      ├─ Acertos
      ├─ Cartões
      ├─ Validações
      └─ Documentos Fiscais
```

### Fase 3: Implementar Caching Global

#### 3.1 Setup React Query

**Arquivos Novos:**
- `apps/app/src/lib/queryClient.ts` - Configuração do QueryClient
- `apps/app/src/hooks/useTrips.ts` - Hook para trips
- `apps/app/src/hooks/useRoutes.ts` - Hook para routes
- `apps/app/src/hooks/useBuses.ts` - Hook para buses
- `apps/app/src/hooks/useDrivers.ts` - Hook para drivers
- `apps/app/src/hooks/useBookings.ts` - Hook para bookings

**Arquivo Modificado:** [apps/app/src/main.tsx](apps/app/src/main.tsx)

**Implementação:**
```tsx
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 5 * 60 * 1000, // 5 minutos
      cacheTime: 10 * 60 * 1000, // 10 minutos
      refetchOnWindowFocus: false,
    },
  },
})

<QueryClientProvider client={queryClient}>
  <BrowserRouter>
    <AuthGate>
      <App />
    </AuthGate>
  </BrowserRouter>
</QueryClientProvider>
```

**Benefício:**
- Reduz chamadas API duplicadas
- Cache automático de dados compartilhados
- Invalidação inteligente após mutations

#### 3.2 Refatorar Páginas para Usar Hooks

**Páginas a refatorar:**
- [Bookings](apps/app/src/pages/Bookings/index.tsx) - usar `useTrips()`, `useBookings()`
- [Trips](apps/app/src/pages/Trips/index.tsx) - usar `useTrips()`, `useRoutes()`, `useBuses()`, `useDrivers()`
- [Reports](apps/app/src/pages/Reports/index.tsx) - usar `useTrips()`, `useBookings()`
- [Payments](apps/app/src/pages/Payments/index.tsx) - usar `useBookings()`

### Fase 4: Melhorias UX Adicionais

#### 4.1 Smart Defaults

**Dashboard** ([apps/app/src/pages/Dashboard/index.tsx](apps/app/src/pages/Dashboard/index.tsx)):
- Mostrar "Viagens de Hoje" automaticamente
- Lista de "Últimas Reservas" com ação rápida de pagamento
- Remover botões duplicados de "Nova Reserva"

**Reports** ([apps/app/src/pages/Reports/index.tsx](apps/app/src/pages/Reports/index.tsx)):
- Pré-selecionar viagem mais recente
- Mostrar dados imediatamente
- Botão "Limpar seleção" para escolher outra viagem

#### 4.2 Consolidar CTAs

**Remover:**
- Botões de ação rápida no Dashboard (ou tornar contextuais)

**Manter:**
- Topbar: "Nova Reserva" principal
- Sidebar: Links de navegação

#### 4.3 Adicionar Shortcuts

**Bookings** ([apps/app/src/pages/Bookings/index.tsx](apps/app/src/pages/Bookings/index.tsx)):
- Botão "Repetir Última Reserva" (mesmo passageiro/viagem)
- Auto-complete de passageiros cadastrados
- Sugestão de assentos preferenciais

## Arquivos Críticos a Modificar

### Prioridade Alta (Fase 1)
1. [apps/app/src/pages/Bookings/index.tsx](apps/app/src/pages/Bookings/index.tsx) - Consolidar steps
2. [apps/app/src/pages/Bookings/BookingForm.tsx](apps/app/src/pages/Bookings/BookingForm.tsx) - Reformular wizard
3. [apps/app/src/pages/Payments/index.tsx](apps/app/src/pages/Payments/index.tsx) - Simplificar purpose
4. [apps/app/src/pages/Payments/PaymentForm.tsx](apps/app/src/pages/Payments/PaymentForm.tsx) - Adaptar para uso inline

### Prioridade Média (Fase 2)
5. `apps/app/src/pages/Financial/index.tsx` (CRIAR) - Página unificada
6. [apps/app/src/app/App.tsx](apps/app/src/app/App.tsx) - Atualizar rotas
7. [apps/app/src/app/Layout.tsx](apps/app/src/app/Layout.tsx) - Atualizar sidebar

### Prioridade Baixa (Fases 3-4)
8. [apps/app/src/main.tsx](apps/app/src/main.tsx) - Setup React Query
9. `apps/app/src/lib/queryClient.ts` (CRIAR) - Configuração
10. `apps/app/src/hooks/` (CRIAR hooks) - useTrips, useRoutes, etc.
11. [apps/app/src/pages/Dashboard/index.tsx](apps/app/src/pages/Dashboard/index.tsx) - Smart defaults
12. [apps/app/src/pages/Reports/index.tsx](apps/app/src/pages/Reports/index.tsx) - Pré-seleção

## Impacto Esperado

### Redução de Telas
- **Antes:** 16 páginas
- **Depois:** 11 páginas (-31%)
  - Financeiro consolidado: 6 → 1 (-5 páginas)

### Redução de Cliques (Fluxo Completo)
- **Antes:** 15-20 cliques (Reserva 3 steps + navegar para Payments + processar)
- **Depois:** 8-10 cliques (Reserva 2 steps com pagamento integrado)
- **Economia:** ~45% de redução

### Performance
- Redução de ~60% nas chamadas API duplicadas (com React Query)
- Cache inteligente melhora percepção de velocidade

### Manutenibilidade
- Código mais DRY (hooks compartilhados)
- Menos arquivos de rota para gerenciar
- Componentes reutilizáveis

## Ordem de Implementação (APROVADA)

✅ **Implementar todas as fases em sequência**

1. **Fase 1** (2-3 dias): Otimizar fluxo reserva + pagamento
   - Maior impacto para usuários
   - Elimina principal ponto de confusão
   - **REGRA:** Exigir pagamento mínimo (sinal ou total) - não permitir reserva sem pagamento

2. **Fase 3** (1-2 dias): Implementar React Query
   - Base para melhorias futuras
   - Beneficia todas as páginas
   - Reduz chamadas API duplicadas

3. **Fase 2** (1-2 dias): Consolidar módulo financeiro
   - Melhora navegação
   - Reduz 6 páginas em 1
   - Interface com tabs

4. **Fase 4** (1 dia): Melhorias UX adicionais
   - Polish final
   - Smart defaults
   - Shortcuts úteis

**Total estimado:** 5-8 dias de desenvolvimento

## Validação e Testes

### Testes Manuais Necessários

1. **Fluxo de Reserva Completo:**
   - Criar reserva com pagamento total PIX → Verificar QR code gerado
   - Criar reserva com pagamento total cartão → Verificar link checkout
   - Criar reserva com sinal (30%) → Verificar valores corretos e QR/link
   - Criar reserva com pagamento manual (dinheiro) → Verificar marca PAID imediato
   - **VALIDAR:** Não permitir criar reserva sem pagamento inicial (botão desabilitado)

2. **Navegação Financeiro:**
   - Acessar cada tab do módulo financeiro
   - Criar/editar/deletar registros em cada tab
   - Verificar que dados persistem ao trocar tabs

3. **Performance:**
   - Abrir Dashboard → verificar cache de trips/bookings
   - Navegar entre páginas → verificar que não re-fetcha desnecessariamente
   - Criar reserva → verificar que lista atualiza automaticamente

4. **Compatibilidade:**
   - Testar em Chrome/Edge/Firefox
   - Verificar responsividade (se aplicável)

### Testes Automatizados

Arquivos de teste a atualizar:
- `apps/app/src/pages/Bookings/__tests__/` - Novos steps
- `apps/app/src/pages/Payments/__tests__/` - Nova finalidade
- `apps/app/src/hooks/__tests__/` - Novos hooks React Query

## Riscos e Mitigações

### Risco 1: Quebrar fluxo existente
**Mitigação:** Implementar feature flags para rollback rápido se necessário

### Risco 2: Perda de funcionalidade
**Mitigação:** Checklist completo de todas as features atuais antes de começar

### Risco 3: Resistance do usuário
**Mitigação:** Documentar mudanças, criar guide de migração para operadores

## Próximos Passos

Após aprovação do plano:
1. Criar branch `feature/optimize-app-flow`
2. Começar por Fase 1 (maior impacto)
3. Fazer commits incrementais com testes
4. Review de cada fase antes de próxima
5. Deploy em staging para validação com usuários
