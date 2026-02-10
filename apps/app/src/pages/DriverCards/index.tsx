import { useMemo, useState, type CSSProperties, type FormEvent } from "react";
import FormField from "../../components/FormField";
import InlineAlert from "../../components/InlineAlert";
import CRUDListPage, { type ColumnConfig, type FormFieldConfig } from "../../components/layout/CRUDListPage";
import StatusBadge from "../../components/StatusBadge";
import useToast from "../../hooks/useToast";
import { apiGet, apiPatch, apiPost } from "../../services/api";
import type { DriverCard, DriverCardTransaction } from "../../types/financial";
import { cardTransactionLabel, cardTypeLabel, formatCurrency } from "../../utils/financialLabels";
import { formatDateTime, formatShortId } from "../../utils/format";

type DriverCardForm = {
  driver_id: string;
  card_number: string;
  card_type: string;
  initial_balance: number;
  current_balance: number;
  is_active: boolean;
  notes: string;
};

type DriverItem = { id: string; name: string };

type AdjustmentFormState = {
  transaction_type: "CREDIT" | "DEBIT" | "ADJUSTMENT" | "REFUND";
  amount: string;
  description: string;
};

const defaultAdjustmentState: AdjustmentFormState = {
  transaction_type: "CREDIT",
  amount: "",
  description: "Ajuste manual",
};

type DriverCardsProps = {
  embedded?: boolean;
};

export default function DriverCards({ embedded = false }: DriverCardsProps) {
  const toast = useToast();
  const [drivers, setDrivers] = useState<DriverItem[]>([]);
  const [reloadKey, setReloadKey] = useState(0);
  const [selectedCard, setSelectedCard] = useState<DriverCard | null>(null);
  const [transactions, setTransactions] = useState<DriverCardTransaction[]>([]);
  const [transactionsLoading, setTransactionsLoading] = useState(false);
  const [transactionsError, setTransactionsError] = useState<string | null>(null);
  const [confirmCard, setConfirmCard] = useState<DriverCard | null>(null);
  const [confirmLoading, setConfirmLoading] = useState(false);
  const [adjustCard, setAdjustCard] = useState<DriverCard | null>(null);
  const [adjustLoading, setAdjustLoading] = useState(false);
  const [adjustForm, setAdjustForm] = useState<AdjustmentFormState>(defaultAdjustmentState);

  const driverMap = useMemo(() => new Map(drivers.map((driver) => [driver.id, driver.name])), [drivers]);

  const formFields: FormFieldConfig<DriverCardForm>[] = [
    {
      key: "driver_id",
      label: "Motorista",
      type: "select",
      required: true,
      options: [
        { label: "Selecione o motorista", value: "" },
        ...drivers.map((driver) => ({ label: driver.name, value: driver.id })),
      ],
    },
    {
      key: "card_number",
      label: "Numero do cartao",
      required: true,
    },
    {
      key: "card_type",
      label: "Tipo",
      type: "select",
      required: true,
      options: [
        { label: "Selecione o tipo", value: "" },
        { label: "Combustivel", value: "FUEL" },
        { label: "Multiplo proposito", value: "MULTIPURPOSE" },
        { label: "Alimentacao", value: "FOOD" },
      ],
    },
    {
      key: "initial_balance",
      label: "Saldo inicial",
      type: "number",
      showWhen: "create",
      inputProps: { min: 0, step: 0.01 },
      hint: "Opcional",
    },
    {
      key: "current_balance",
      label: "Saldo atual",
      type: "number",
      showWhen: "edit",
      disabled: true,
      inputProps: { readOnly: true },
    },
    {
      key: "is_active",
      label: "Status",
      type: "checkbox",
      hint: "Cartao ativo",
    },
    {
      key: "notes",
      label: "Observacoes",
      type: "textarea",
      colSpan: "full",
    },
  ];

  const columns: ColumnConfig<DriverCard>[] = [
    {
      label: "Motorista",
      accessor: (item) => driverMap.get(item.driver_id) ?? formatShortId(item.driver_id),
    },
    { label: "Cartao", accessor: (item) => item.card_number },
    {
      label: "Tipo",
      accessor: (item) => cardTypeLabel[item.card_type] ?? item.card_type,
    },
    { label: "Saldo", accessor: (item) => formatCurrency(item.current_balance) },
    {
      label: "Status",
      render: (item) => (
        <StatusBadge tone={item.is_blocked ? "danger" : item.is_active ? "success" : "warning"}>
          {item.is_blocked ? "Bloqueado" : item.is_active ? "Ativo" : "Inativo"}
        </StatusBadge>
      ),
    },
  ];

  const refreshTransactions = async (card: DriverCard) => {
    setTransactionsLoading(true);
    setTransactionsError(null);
    try {
      const data = await apiGet<DriverCardTransaction[]>(`/driver-cards/${card.id}/transactions?limit=50&offset=0`);
      setTransactions(data);
      setSelectedCard(card);
    } catch (err: any) {
      setTransactionsError(err.message || "Erro ao carregar transacoes");
    } finally {
      setTransactionsLoading(false);
    }
  };

  const openBlockToggle = (item: DriverCard) => {
    setConfirmCard(item);
  };

  const handleBlockToggle = async () => {
    if (!confirmCard) return;
    const action = confirmCard.is_blocked ? "unblock" : "block";
    try {
      setConfirmLoading(true);
      await apiPost(`/driver-cards/${confirmCard.id}/${action}`, {});
      toast.success(confirmCard.is_blocked ? "Cartao desbloqueado." : "Cartao bloqueado.");
      setReloadKey((value) => value + 1);
      setConfirmCard(null);
    } catch (err: any) {
      toast.error(err.message || "Erro ao atualizar cartao");
    } finally {
      setConfirmLoading(false);
    }
  };

  const openAdjustBalance = (item: DriverCard) => {
    setAdjustCard(item);
    setAdjustForm(defaultAdjustmentState);
  };

  const handleAdjustBalance = async (event: FormEvent) => {
    event.preventDefault();
    if (!adjustCard) return;
    const amount = Number(adjustForm.amount);
    if (!Number.isFinite(amount) || amount <= 0) {
      toast.error("Informe um valor valido.");
      return;
    }

    try {
      setAdjustLoading(true);
      await apiPost(`/driver-cards/${adjustCard.id}/transactions`, {
        transaction_type: adjustForm.transaction_type,
        amount,
        description: adjustForm.description || undefined,
      });
      toast.success("Saldo ajustado com sucesso.");
      setReloadKey((value) => value + 1);
      if (selectedCard?.id === adjustCard.id) {
        await refreshTransactions(adjustCard);
      }
      setAdjustCard(null);
      setAdjustForm(defaultAdjustmentState);
    } catch (err: any) {
      toast.error(err.message || "Erro ao ajustar saldo");
    } finally {
      setAdjustLoading(false);
    }
  };

  return (
    <>
      <CRUDListPage<DriverCard, DriverCardForm>
        key={reloadKey}
        hidePageHeader={embedded}
        title="Cartoes de Motorista"
        subtitle="Controle de cartoes e saldos para despesas."
        formTitle="Novo cartao"
        listTitle="Cartoes cadastrados"
        createLabel="Criar cartao"
        updateLabel="Salvar cartao"
        emptyState={{
          title: "Nenhum cartao encontrado",
          description: "Cadastre um cartao para comecar.",
        }}
        formFields={formFields}
        columns={columns}
        initialForm={{
          driver_id: "",
          card_number: "",
          card_type: "",
          initial_balance: 0,
          current_balance: 0,
          is_active: true,
          notes: "",
        }}
        mapItemToForm={(item) => ({
          driver_id: item.driver_id,
          card_number: item.card_number,
          card_type: item.card_type,
          initial_balance: 0,
          current_balance: item.current_balance,
          is_active: item.is_active,
          notes: item.notes ?? "",
        })}
        getId={(item) => item.id}
        fetchItems={async ({ page, pageSize }) => {
          const data = await apiGet<DriverCard[]>(`/driver-cards?limit=${pageSize}&offset=${page * pageSize}`);
          const driversData = await apiGet<DriverItem[]>("/drivers?limit=500&offset=0");
          setDrivers(driversData);
          return data;
        }}
        createItem={(form) =>
          apiPost("/driver-cards", {
            driver_id: form.driver_id,
            card_number: form.card_number,
            card_type: form.card_type,
            current_balance: form.initial_balance || undefined,
            notes: form.notes || undefined,
          })
        }
        updateItem={(id, form) =>
          apiPatch(`/driver-cards/${id}`, {
            card_number: form.card_number,
            card_type: form.card_type,
            is_active: form.is_active,
            notes: form.notes || undefined,
          })
        }
        softDeleteItem={(item) => apiPatch(`/driver-cards/${item.id}`, { is_active: false })}
        restoreItem={(item) => apiPatch(`/driver-cards/${item.id}`, { is_active: true })}
        getIsActive={(item) => item.is_active}
        searchFilter={(item, term) => {
          const driver = (driverMap.get(item.driver_id) ?? "").toLowerCase();
          const card = item.card_number.toLowerCase();
          return driver.includes(term) || card.includes(term);
        }}
        rowActions={(item) => (
          <>
            <button
              className={item.is_blocked ? "button success sm" : "button ghost sm"}
              type="button"
              onClick={() => openBlockToggle(item)}
            >
              {item.is_blocked ? "Desbloquear" : "Bloquear"}
            </button>
            <button className="button secondary sm" type="button" onClick={() => openAdjustBalance(item)}>
              Ajustar saldo
            </button>
            <button className="button ghost sm" type="button" onClick={() => refreshTransactions(item)}>
              Transacoes
            </button>
          </>
        )}
      />

      {confirmCard ? (
        <section className="page" style={{ marginTop: "24px" }}>
          <div className="section">
            <div className="section-header">
              <div className="section-title">
                {confirmCard.is_blocked ? "Confirmar desbloqueio" : "Confirmar bloqueio"}
              </div>
            </div>
            <InlineAlert tone="warning">
              {confirmCard.is_blocked
                ? `Desbloquear cartao ${confirmCard.card_number}?`
                : `Bloquear cartao ${confirmCard.card_number}?`}
            </InlineAlert>
            <div className="form-actions">
              <button className="button secondary" type="button" onClick={() => setConfirmCard(null)}>
                Cancelar
              </button>
              <button className="button" type="button" onClick={handleBlockToggle} disabled={confirmLoading}>
                {confirmLoading ? "Salvando..." : "Confirmar"}
              </button>
            </div>
          </div>
        </section>
      ) : null}

      {adjustCard ? (
        <section className="page" style={{ marginTop: "24px" }}>
          <div className="section">
            <div className="section-header">
              <div className="section-title">Ajustar saldo do cartao {adjustCard.card_number}</div>
            </div>
            <form className="form-grid" onSubmit={handleAdjustBalance}>
              <FormField label="Tipo" required>
                <select
                  className="input"
                  value={adjustForm.transaction_type}
                  onChange={(e) =>
                    setAdjustForm((prev) => ({
                      ...prev,
                      transaction_type: e.target.value as AdjustmentFormState["transaction_type"],
                    }))
                  }
                  required
                >
                  <option value="CREDIT">Credito</option>
                  <option value="DEBIT">Debito</option>
                  <option value="ADJUSTMENT">Ajuste</option>
                  <option value="REFUND">Estorno</option>
                </select>
              </FormField>
              <FormField label="Valor" required>
                <input
                  className="input"
                  type="number"
                  min={0}
                  step={0.01}
                  value={adjustForm.amount}
                  onChange={(e) => setAdjustForm((prev) => ({ ...prev, amount: e.target.value }))}
                  required
                />
              </FormField>
              <FormField label="Descricao" hint="Opcional">
                <input
                  className="input"
                  value={adjustForm.description}
                  onChange={(e) => setAdjustForm((prev) => ({ ...prev, description: e.target.value }))}
                />
              </FormField>
              <div className="form-actions full-span">
                <button className="button secondary" type="button" onClick={() => setAdjustCard(null)}>
                  Cancelar
                </button>
                <button className="button" type="submit" disabled={adjustLoading}>
                  {adjustLoading ? "Salvando..." : "Salvar ajuste"}
                </button>
              </div>
            </form>
          </div>
        </section>
      ) : null}

      {selectedCard ? (
        <section className="page" style={{ marginTop: "24px" }}>
          <div className="section">
            <div className="section-header">
              <div className="section-title">Transacoes de {selectedCard.card_number}</div>
              <button className="button secondary sm" type="button" onClick={() => setSelectedCard(null)}>
                Fechar
              </button>
            </div>
            {transactionsError ? <InlineAlert tone="error">{transactionsError}</InlineAlert> : null}
            {transactionsLoading ? (
              <div className="text-muted">Carregando transacoes...</div>
            ) : transactions.length === 0 ? (
              <div className="text-muted">Nenhuma transacao registrada.</div>
            ) : (
              <div
                className="table"
                style={{ "--table-columns": "repeat(5, minmax(0, 1fr))" } as CSSProperties}
              >
                <div className="table-row table-head">
                  <div className="table-cell">Tipo</div>
                  <div className="table-cell">Valor</div>
                  <div className="table-cell">Saldo</div>
                  <div className="table-cell">Descricao</div>
                  <div className="table-cell">Data</div>
                </div>
                {transactions.map((tx) => (
                  <div className="table-row" key={tx.id}>
                    <div className="table-cell" data-label="Tipo">
                      {cardTransactionLabel[tx.transaction_type] ?? tx.transaction_type}
                    </div>
                    <div className="table-cell" data-label="Valor">
                      {formatCurrency(tx.amount)}
                    </div>
                    <div className="table-cell" data-label="Saldo">
                      {formatCurrency(tx.balance_after)}
                    </div>
                    <div className="table-cell" data-label="Descricao">
                      {tx.description || "-"}
                    </div>
                    <div className="table-cell" data-label="Data">
                      {formatDateTime(tx.created_at)}
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        </section>
      ) : null}
    </>
  );
}
