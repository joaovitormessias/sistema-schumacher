**Component Library**

**ToastProvider + useToast**
Provider global para notificacoes.
```
import useToast from "../hooks/useToast";
const toast = useToast();
toast.success("Mensagem");
```

**PaginationControls**
```
<PaginationControls
  page={page}
  pageSize={50}
  itemCount={items.length}
  onPageChange={setPage}
/>
```

**SearchToolbar**
```
<SearchToolbar
  value={query}
  onChange={setQuery}
  placeholder="Buscar"
  filters={<select className="input" />}
  actions={<button className="button">Novo</button>}
/>
```

**FormFieldGroup**
```
<FormFieldGroup legend="Dados" hint="Campos principais" layout="grid">
  ...
</FormFieldGroup>
```

**CRUDListPage**
Template para listas CRUD com criacao, edicao e soft delete.
Principais props:
```
<CRUDListPage
  title="Motoristas"
  formTitle="Cadastro rapido"
  listTitle="Equipe cadastrada"
  formFields={fields}
  columns={columns}
  fetchItems={fetchItems}
  createItem={createItem}
  updateItem={updateItem}
  softDeleteItem={softDeleteItem}
  restoreItem={restoreItem}
  searchFilter={searchFilter}
/>
```

**TableActionButtons**
Botao de editar e desativar/reativar em tabelas.

**BookingStepForm + Steps**
Multi-step form com `TripStep`, `PassengerStep`, `PaymentStep`.

**BookingFormContext**
Contexto compartilhado para estado do formulario e etapa ativa.

**PricingRuleForm**
Formulario de regras com modo avancado e editor visual para sazonalidade.

**PricingRuleCard**
Card de lista com resumo de regra e botao de edicao.

**PaymentTabs**
Toggle entre pagamento automatico e manual.

**PaymentForm**
Formulario unificado para os dois modos.

**PaymentResult**
Resumo da cobranca/pagamento gerado.
