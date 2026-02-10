import { useMemo, useState } from "react";
import CRUDListPage, {
    type ColumnConfig,
    type FormFieldConfig,
} from "../../../components/layout/CRUDListPage";
import StatusBadge from "../../../components/StatusBadge";
import useToast from "../../../hooks/useToast";
import { apiGet, apiPatch, apiPost } from "../../../services/api";
import type { PurchaseOrder } from "../../../hooks/usePurchaseOrders";
import { formatCurrency } from "../../../utils/financialLabels";
import { formatDateTime, formatShortId } from "../../../utils/format";
import type { Supplier } from "../../../hooks/useSuppliers";

type PurchaseOrderForm = {
    supplier_id: string;
    expected_delivery: string;
    own_delivery: boolean;
    discount: number | "";
    freight: number | "";
    notes: string;
};

const statusLabels: Record<string, string> = {
    DRAFT: "Rascunho",
    SENT: "Enviado",
    PARTIAL: "Parcial",
    RECEIVED: "Recebido",
    CANCELLED: "Cancelado",
};

function getStatusTone(status: string) {
    switch (status) {
        case "DRAFT":
            return "neutral";
        case "SENT":
            return "info";
        case "PARTIAL":
            return "warning";
        case "RECEIVED":
            return "success";
        case "CANCELLED":
            return "danger";
        default:
            return "neutral";
    }
}

export default function PurchaseOrdersTab() {
    const toast = useToast();
    const [suppliers, setSuppliers] = useState<Supplier[]>([]);
    const [reloadKey, setReloadKey] = useState(0);

    const supplierMap = useMemo(() => new Map(suppliers.map((s) => [s.id, s.name])), [suppliers]);

    const formFields: FormFieldConfig<PurchaseOrderForm>[] = useMemo(
        () => [
            {
                key: "supplier_id",
                label: "Fornecedor",
                type: "select",
                required: true,
                options: [
                    { label: "Selecione o fornecedor", value: "" },
                    ...suppliers.map((sup) => ({ label: sup.name, value: sup.id })),
                ],
            },
            {
                key: "expected_delivery",
                label: "Previsão de Entrega",
                type: "text",
                hint: "Formato: AAAA-MM-DD",
            },
            {
                key: "own_delivery",
                label: "Retirada Própria",
                type: "checkbox",
            },
            {
                key: "discount",
                label: "Desconto (R$)",
                type: "number",
                inputProps: { min: 0, step: 0.01 },
            },
            {
                key: "freight",
                label: "Frete (R$)",
                type: "number",
                inputProps: { min: 0, step: 0.01 },
            },
            {
                key: "notes",
                label: "Observações",
                type: "textarea",
                colSpan: "full",
            },
        ],
        [suppliers]
    );

    const columns: ColumnConfig<PurchaseOrder>[] = [
        { label: "Nº Pedido", accessor: (item) => `#${item.order_number}` },
        { label: "Fornecedor", accessor: (item) => item.supplier_name ?? supplierMap.get(item.supplier_id) ?? formatShortId(item.supplier_id) },
        { label: "Total", accessor: (item) => formatCurrency(item.total) },
        {
            label: "Status",
            render: (item) => (
                <StatusBadge tone={getStatusTone(item.status)}>
                    {statusLabels[item.status] ?? item.status}
                </StatusBadge>
            ),
        },
        { label: "Data", accessor: (item) => formatDateTime(item.order_date) },
    ];

    const handleSend = async (item: PurchaseOrder) => {
        if (!window.confirm("Enviar pedido ao fornecedor?")) return;
        try {
            await apiPost(`/purchase-orders/${item.id}/send`, {});
            toast.success("Pedido enviado com sucesso.");
            setReloadKey((v) => v + 1);
        } catch (err: any) {
            toast.error(err.message || "Erro ao enviar pedido");
        }
    };

    const handleReceive = async (item: PurchaseOrder) => {
        if (!window.confirm("Marcar pedido como recebido?")) return;
        try {
            await apiPost(`/purchase-orders/${item.id}/receive`, {});
            toast.success("Pedido marcado como recebido.");
            setReloadKey((v) => v + 1);
        } catch (err: any) {
            toast.error(err.message || "Erro ao marcar recebimento");
        }
    };

    return (
        <CRUDListPage<PurchaseOrder, PurchaseOrderForm>
            key={reloadKey}
            hidePageHeader
            title="Pedidos de Compra"
            subtitle="Gestão de pedidos de compra."
            formTitle="Novo pedido"
            listTitle="Pedidos de compra"
            createLabel="Criar pedido"
            updateLabel="Salvar pedido"
            emptyState={{
                title: "Nenhum pedido encontrado",
                description: "Crie um pedido de compra para começar.",
            }}
            formFields={formFields}
            columns={columns}
            initialForm={{ supplier_id: "", expected_delivery: "", own_delivery: true, discount: "", freight: "", notes: "" }}
            mapItemToForm={(item) => ({
                supplier_id: item.supplier_id,
                expected_delivery: item.expected_delivery?.split("T")[0] ?? "",
                own_delivery: item.own_delivery,
                discount: item.discount ?? "",
                freight: item.freight ?? "",
                notes: item.notes ?? "",
            })}
            getId={(item) => item.id}
            fetchItems={async ({ page, pageSize }) => {
                const data = await apiGet<PurchaseOrder[]>(
                    `/purchase-orders?limit=${pageSize}&offset=${page * pageSize}`
                );
                const suppliersData = await apiGet<Supplier[]>("/suppliers?limit=500&offset=0&active=true");
                setSuppliers(suppliersData);
                return data;
            }}
            createItem={(form) =>
                apiPost("/purchase-orders", {
                    supplier_id: form.supplier_id,
                    expected_delivery: form.expected_delivery || undefined,
                    own_delivery: form.own_delivery,
                    discount: form.discount ? Number(form.discount) : undefined,
                    freight: form.freight ? Number(form.freight) : undefined,
                    notes: form.notes || undefined,
                    items: [], // Items will be added separately
                })
            }
            updateItem={(id, form) =>
                apiPatch(`/purchase-orders/${id}`, {
                    expected_delivery: form.expected_delivery || undefined,
                    own_delivery: form.own_delivery,
                    discount: form.discount ? Number(form.discount) : undefined,
                    freight: form.freight ? Number(form.freight) : undefined,
                    notes: form.notes || undefined,
                })
            }
            searchFilter={(item, term) => {
                return (
                    item.order_number.toString().includes(term) ||
                    (item.supplier_name?.toLowerCase().includes(term) ?? false)
                );
            }}
            rowActions={(item) =>
                item.status === "DRAFT" ? (
                    <button className="button ghost sm" type="button" onClick={() => handleSend(item)}>
                        Enviar
                    </button>
                ) : item.status === "SENT" ? (
                    <button className="button ghost sm" type="button" onClick={() => handleReceive(item)}>
                        Receber
                    </button>
                ) : null
            }
        />
    );
}
