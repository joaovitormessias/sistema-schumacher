import { useMemo, useState } from "react";
import CRUDListPage, { type ColumnConfig, type FormFieldConfig } from "../../components/layout/CRUDListPage";
import StatusBadge from "../../components/StatusBadge";
import InlineAlert from "../../components/InlineAlert";
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

export default function DriverCards() {
  const toast = useToast();
  const [drivers, setDrivers] = useState<DriverItem[]>([]);
  const [reloadKey, setReloadKey] = useState(0);
  const [selectedCard, setSelectedCard] = useState<DriverCard | null>(null);
  const [transactions, setTransactions] = useState<DriverCardTransaction[]>([]);
  const [transactionsLoading, setTransactionsLoading] = useState(false);
  const [transactionsError, setTransactionsError] = useState<string | null>(null);

  const driverMap = useMemo(
    () => new Map(drivers.map((driver) => [driver.id, driver.name])),
    [drivers]
  );

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
      label: "Número do cartão",
      required: true,
    },
    {
      key: "card_type",
      label: "Tipo",
      type: "select",
      required: true,
      options: [
        { label: "Selecione o tipo", value: "" },
        { label: "Combustível", value: "FUEL" },
        { label: "Múltiplo propósito", value: "MULTIPURPOSE" },
        { label: "Alimentação", value: "FOOD" },
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
      hint: "Cartão ativo",
    },
    {
      key: "notes",
      label: "Observações",
      type: "textarea",
      colSpan: "full",
    },
  ];

  const columns: ColumnConfig<DriverCard>[] = [
    {
      label: "Motorista",
      accessor: (item) => driverMap.get(item.driver_id) ?? formatShortId(item.driver_id),
    },
    { label: "Cartão", accessor: (item) => item.card_number },
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
      const data = await apiGet<DriverCardTransaction[]>(
        `/driver-cards/${card.id}/transactions?limit=50&offset=0`
      );
      setTransactions(data);
      setSelectedCard(card);
    } catch (err: any) {
      setTransactionsError(err.message || "Erro ao carregar transações");
    } finally {
      setTransactionsLoading(false);
    }
  };

  const handleBlockToggle = async (item: DriverCard) => {
    const action = item.is_blocked ? "unblock" : "block";
    const confirmText = item.is_blocked
      ? "Deseja desbloquear este cartão?"
      : "Deseja bloquear este cartão?";
    if (!window.confirm(confirmText)) return;
    try {
      await apiPost(`/driver-cards/${item.id}/${action}`, {});
      toast.success(item.is_blocked ? "Cartão desbloqueado." : "Cartão bloqueado.");
      setReloadKey((value) => value + 1);
    } catch (err: any) {
      toast.error(err.message || "Erro ao atualizar cartão");
    }
  };

  const handleAdjustBalance = async (item: DriverCard) => {
    const type = window
      .prompt("Tipo (CREDIT, DEBIT, ADJUSTMENT, REFUND)", "CREDIT")
      ?.trim()
      .toUpperCase();
    if (!type) return;
    const amountInput = window.prompt("Valor", "0");
    if (!amountInput) return;
    const amount = Number(amountInput);
    if (!Number.isFinite(amount) || amount <= 0) {
      toast.error("Informe um valor válido.");
      return;
    }
    const description = window.prompt("Descrição (opcional)", "Ajuste manual") ?? "";

    try {
      await apiPost(`/driver-cards/${item.id}/transactions`, {
        transaction_type: type,
        amount,
        description: description || undefined,
      });
      toast.success("Saldo ajustado com sucesso.");
      setReloadKey((value) => value + 1);
      if (selectedCard?.id === item.id) {
        refreshTransactions(item);
      }
    } catch (err: any) {
      toast.error(err.message || "Erro ao ajustar saldo");
    }
  };

  return (
    <>
      <CRUDListPage<DriverCard, DriverCardForm>
        key={reloadKey}
        title="Cartões de Motorista"
        subtitle="Controle de cartões e saldos para despesas."
        formTitle="Novo cartão"
        listTitle="Cartões cadastrados"
        createLabel="Criar cartão"
        updateLabel="Salvar cartão"
        emptyState={{
          title: "Nenhum cartão encontrado",
          description: "Cadastre um cartão para começar.",
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
          const data = await apiGet<DriverCard[]>(
            `/driver-cards?limit=${pageSize}&offset=${page * pageSize}`
          );
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
              onClick={() => handleBlockToggle(item)}
            >
              {item.is_blocked ? "Desbloquear" : "Bloquear"}
            </button>
            <button
              className="button secondary sm"
              type="button"
              onClick={() => handleAdjustBalance(item)}
            >
              Ajustar saldo
            </button>
            <button
              className="button ghost sm"
              type="button"
              onClick={() => refreshTransactions(item)}
            >
              Transações
            </button>
          </>
        )}
      />

      {selectedCard ? (
        <section className="page" style={{ marginTop: "24px" }}>
          <div className="section">
            <div className="section-header">
              <div className="section-title">
                Transações de {selectedCard.card_number}
              </div>
              <button
                className="button secondary sm"
                type="button"
                onClick={() => setSelectedCard(null)}
              >
                Fechar
              </button>
            </div>
            {transactionsError ? <InlineAlert tone="error">{transactionsError}</InlineAlert> : null}
            {transactionsLoading ? (
              <div className="text-muted">Carregando transações...</div>
            ) : transactions.length === 0 ? (
              <div className="text-muted">Nenhuma transação registrada.</div>
            ) : (
              <div
                className="table"
                style={{ "--table-columns": "repeat(5, minmax(0, 1fr))" } as any}
              >
                <div className="table-row table-head">
                  <div className="table-cell">Tipo</div>
                  <div className="table-cell">Valor</div>
                  <div className="table-cell">Saldo</div>
                  <div className="table-cell">Descrição</div>
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
                    <div className="table-cell" data-label="Descrição">
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
