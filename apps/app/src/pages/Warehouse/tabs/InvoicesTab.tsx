import { useMemo, useState } from "react";
import CRUDListPage, {
    type ColumnConfig,
    type FormFieldConfig,
} from "../../../components/layout/CRUDListPage";
import StatusBadge from "../../../components/StatusBadge";
import useToast from "../../../hooks/useToast";
import { apiGet, apiPatch, apiPost } from "../../../services/api";
import type { Invoice } from "../../../hooks/useInvoices";
import { formatCurrency } from "../../../utils/financialLabels";
import { formatDateTime, formatShortId } from "../../../utils/format";
import type { Supplier } from "../../../hooks/useSuppliers";

type InvoiceForm = {
    invoice_number: string;
    barcode: string;
    supplier_id: string;
    issue_date: string;
    cfop: string;
    payment_type: string;
    due_date: string;
    discount: number | "";
    freight: number | "";
    notes: string;
};

const statusLabels: Record<string, string> = {
    PENDING: "Pendente",
    PROCESSED: "Processada",
    CANCELLED: "Cancelada",
};

function getStatusTone(status: string) {
    switch (status) {
        case "PENDING":
            return "warning";
        case "PROCESSED":
            return "success";
        case "CANCELLED":
            return "danger";
        default:
            return "neutral";
    }
}

export default function InvoicesTab() {
    const toast = useToast();
    const [suppliers, setSuppliers] = useState<Supplier[]>([]);
    const [reloadKey, setReloadKey] = useState(0);

    const supplierMap = useMemo(() => new Map(suppliers.map((s) => [s.id, s.name])), [suppliers]);

    const formFields: FormFieldConfig<InvoiceForm>[] = useMemo(
        () => [
            {
                key: "invoice_number",
                label: "Número da NF",
                type: "text",
                required: true,
            },
            {
                key: "barcode",
                label: "Código de Barras",
                type: "text",
            },
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
                key: "issue_date",
                label: "Data de Emissão",
                type: "text",
                required: true,
                hint: "Formato: AAAA-MM-DD",
            },
            {
                key: "cfop",
                label: "CFOP",
                type: "text",
                hint: "Ex: 1000, 2403",
            },
            {
                key: "payment_type",
                label: "Tipo de Pagamento",
                type: "select",
                options: [
                    { label: "Selecione", value: "" },
                    { label: "Boleto", value: "BOLETO" },
                    { label: "Pix", value: "PIX" },
                    { label: "Transferência", value: "TRANSFERENCIA" },
                    { label: "Cartão", value: "CARTAO" },
                    { label: "Dinheiro", value: "DINHEIRO" },
                ],
            },
            {
                key: "due_date",
                label: "Vencimento",
                type: "text",
                hint: "Formato: AAAA-MM-DD",
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

    const columns: ColumnConfig<Invoice>[] = [
        { label: "Nº NF", accessor: (item) => item.invoice_number },
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
        { label: "Emissão", accessor: (item) => formatDateTime(item.issue_date) },
        { label: "Entrada", accessor: (item) => formatDateTime(item.entry_date) },
    ];

    const handleProcess = async (item: Invoice) => {
        if (!window.confirm("Processar esta nota fiscal?")) return;
        try {
            await apiPost(`/invoices/${item.id}/process`, {});
            toast.success("NF processada com sucesso.");
            setReloadKey((v) => v + 1);
        } catch (err: any) {
            toast.error(err.message || "Erro ao processar NF");
        }
    };

    return (
        <CRUDListPage<Invoice, InvoiceForm>
            key={reloadKey}
            hidePageHeader
            title="Notas Fiscais"
            subtitle="Entrada de notas fiscais de fornecedores."
            formTitle="Nova NF"
            listTitle="Notas fiscais"
            createLabel="Lançar NF"
            updateLabel="Salvar NF"
            emptyState={{
                title: "Nenhuma NF encontrada",
                description: "Lance uma nota fiscal para começar.",
            }}
            formFields={formFields}
            columns={columns}
            initialForm={{ invoice_number: "", barcode: "", supplier_id: "", issue_date: "", cfop: "1000", payment_type: "", due_date: "", discount: "", freight: "", notes: "" }}
            mapItemToForm={(item) => ({
                invoice_number: item.invoice_number,
                barcode: item.barcode ?? "",
                supplier_id: item.supplier_id,
                issue_date: item.issue_date?.split("T")[0] ?? "",
                cfop: item.cfop,
                payment_type: item.payment_type ?? "",
                due_date: item.due_date?.split("T")[0] ?? "",
                discount: item.discount ?? "",
                freight: item.freight ?? "",
                notes: item.notes ?? "",
            })}
            getId={(item) => item.id}
            fetchItems={async ({ page, pageSize }) => {
                const data = await apiGet<Invoice[]>(
                    `/invoices?limit=${pageSize}&offset=${page * pageSize}`
                );
                const suppliersData = await apiGet<Supplier[]>("/suppliers?limit=500&offset=0&active=true");
                setSuppliers(suppliersData);
                return data;
            }}
            createItem={(form) =>
                apiPost("/invoices", {
                    invoice_number: form.invoice_number,
                    barcode: form.barcode || undefined,
                    supplier_id: form.supplier_id,
                    issue_date: form.issue_date,
                    cfop: form.cfop || undefined,
                    payment_type: form.payment_type || undefined,
                    due_date: form.due_date || undefined,
                    discount: form.discount ? Number(form.discount) : undefined,
                    freight: form.freight ? Number(form.freight) : undefined,
                    notes: form.notes || undefined,
                    items: [], // Items will be added in detail view
                })
            }
            updateItem={(id, form) =>
                apiPatch(`/invoices/${id}`, {
                    barcode: form.barcode || undefined,
                    payment_type: form.payment_type || undefined,
                    due_date: form.due_date || undefined,
                    notes: form.notes || undefined,
                })
            }
            searchFilter={(item, term) => {
                return (
                    item.invoice_number.toLowerCase().includes(term) ||
                    (item.supplier_name?.toLowerCase().includes(term) ?? false)
                );
            }}
            rowActions={(item) =>
                item.status === "PENDING" ? (
                    <button className="button ghost sm" type="button" onClick={() => handleProcess(item)}>
                        Processar
                    </button>
                ) : null
            }
        />
    );
}
