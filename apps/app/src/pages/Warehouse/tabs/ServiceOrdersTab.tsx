import { useMemo, useState } from "react";
import CRUDListPage, {
    type ColumnConfig,
    type FormFieldConfig,
} from "../../../components/layout/CRUDListPage";
import StatusBadge from "../../../components/StatusBadge";
import useToast from "../../../hooks/useToast";
import { apiGet, apiPatch, apiPost } from "../../../services/api";
import type { ServiceOrder, ServiceOrderType } from "../../../hooks/useServiceOrders";
import { formatDateTime, formatShortId } from "../../../utils/format";

type BusItem = { id: string; license_plate: string };
type DriverItem = { id: string; name: string };

type ServiceOrderForm = {
    bus_id: string;
    driver_id: string;
    order_type: ServiceOrderType | "";
    description: string;
    odometer_km: number | "";
    location: string;
    notes: string;
};

const statusLabels: Record<string, string> = {
    OPEN: "Aberta",
    IN_PROGRESS: "Em Andamento",
    CLOSED: "Fechada",
    CANCELLED: "Cancelada",
};

const typeLabels: Record<string, string> = {
    PREVENTIVE: "Preventiva",
    CORRECTIVE: "Corretiva",
};

function getStatusTone(status: string) {
    switch (status) {
        case "OPEN":
            return "warning";
        case "IN_PROGRESS":
            return "info";
        case "CLOSED":
            return "success";
        case "CANCELLED":
            return "danger";
        default:
            return "neutral";
    }
}

export default function ServiceOrdersTab() {
    const toast = useToast();
    const [buses, setBuses] = useState<BusItem[]>([]);
    const [drivers, setDrivers] = useState<DriverItem[]>([]);
    const [reloadKey, setReloadKey] = useState(0);

    const busMap = useMemo(() => new Map(buses.map((b) => [b.id, b.license_plate])), [buses]);
    const driverMap = useMemo(() => new Map(drivers.map((d) => [d.id, d.name])), [drivers]);

    const formFields: FormFieldConfig<ServiceOrderForm>[] = useMemo(
        () => [
            {
                key: "bus_id",
                label: "Veículo",
                type: "select",
                required: true,
                options: [
                    { label: "Selecione o veículo", value: "" },
                    ...buses.map((bus) => ({ label: bus.license_plate, value: bus.id })),
                ],
            },
            {
                key: "driver_id",
                label: "Motorista",
                type: "select",
                options: [
                    { label: "Selecione (opcional)", value: "" },
                    ...drivers.map((driver) => ({ label: driver.name, value: driver.id })),
                ],
            },
            {
                key: "order_type",
                label: "Tipo de OS",
                type: "select",
                required: true,
                options: [
                    { label: "Selecione o tipo", value: "" },
                    { label: "Preventiva", value: "PREVENTIVE" },
                    { label: "Corretiva", value: "CORRECTIVE" },
                ],
            },
            {
                key: "odometer_km",
                label: "Km Atual",
                type: "number",
                hint: "Quilometragem atual do veículo",
                inputProps: { min: 0 },
            },
            {
                key: "location",
                label: "Local",
                type: "select",
                options: [
                    { label: "Schumacher", value: "SCHUMACHER" },
                    { label: "Externo", value: "EXTERNO" },
                ],
            },
            {
                key: "description",
                label: "Descrição",
                type: "textarea",
                required: true,
                colSpan: "full",
                hint: "Descreva o serviço a ser realizado",
            },
            {
                key: "notes",
                label: "Observações",
                type: "textarea",
                colSpan: "full",
            },
        ],
        [buses, drivers]
    );

    const columns: ColumnConfig<ServiceOrder>[] = [
        { label: "Nº OS", accessor: (item) => `#${item.order_number}` },
        { label: "Veículo", accessor: (item) => item.bus_plate ?? busMap.get(item.bus_id) ?? formatShortId(item.bus_id) },
        {
            label: "Tipo",
            render: (item) => (
                <StatusBadge tone={item.order_type === "PREVENTIVE" ? "info" : "warning"}>
                    {typeLabels[item.order_type] ?? item.order_type}
                </StatusBadge>
            ),
        },
        { label: "Descrição", accessor: (item) => item.description.substring(0, 50) + (item.description.length > 50 ? "..." : "") },
        {
            label: "Status",
            render: (item) => (
                <StatusBadge tone={getStatusTone(item.status)}>
                    {statusLabels[item.status] ?? item.status}
                </StatusBadge>
            ),
        },
        { label: "Abertura", accessor: (item) => formatDateTime(item.opened_at) },
    ];

    const handleStart = async (item: ServiceOrder) => {
        if (!window.confirm("Iniciar execução da OS?")) return;
        try {
            await apiPost(`/service-orders/${item.id}/start`, {});
            toast.success("OS iniciada com sucesso.");
            setReloadKey((v) => v + 1);
        } catch (err: any) {
            toast.error(err.message || "Erro ao iniciar OS");
        }
    };

    const handleClose = async (item: ServiceOrder) => {
        if (!window.confirm("Fechar esta OS?")) return;
        try {
            await apiPost(`/service-orders/${item.id}/close`, {});
            toast.success("OS fechada com sucesso.");
            setReloadKey((v) => v + 1);
        } catch (err: any) {
            toast.error(err.message || "Erro ao fechar OS");
        }
    };

    return (
        <CRUDListPage<ServiceOrder, ServiceOrderForm>
            key={reloadKey}
            hidePageHeader
            title="Ordens de Serviço"
            subtitle="Gestão de manutenção preventiva e corretiva."
            formTitle="Nova OS"
            listTitle="Ordens de serviço"
            createLabel="Criar OS"
            updateLabel="Salvar OS"
            emptyState={{
                title: "Nenhuma OS encontrada",
                description: "Crie uma ordem de serviço para começar.",
            }}
            formFields={formFields}
            columns={columns}
            initialForm={{ bus_id: "", driver_id: "", order_type: "", description: "", odometer_km: "", location: "SCHUMACHER", notes: "" }}
            mapItemToForm={(item) => ({
                bus_id: item.bus_id,
                driver_id: item.driver_id ?? "",
                order_type: item.order_type,
                description: item.description,
                odometer_km: item.odometer_km ?? "",
                location: item.location,
                notes: item.notes ?? "",
            })}
            getId={(item) => item.id}
            fetchItems={async ({ page, pageSize }) => {
                const data = await apiGet<ServiceOrder[]>(
                    `/service-orders?limit=${pageSize}&offset=${page * pageSize}`
                );
                const [busesData, driversData] = await Promise.all([
                    apiGet<BusItem[]>("/buses?limit=500&offset=0"),
                    apiGet<DriverItem[]>("/drivers?limit=500&offset=0"),
                ]);
                setBuses(busesData);
                setDrivers(driversData);
                return data;
            }}
            createItem={(form) =>
                apiPost("/service-orders", {
                    bus_id: form.bus_id,
                    driver_id: form.driver_id || undefined,
                    order_type: form.order_type,
                    description: form.description,
                    odometer_km: form.odometer_km ? Number(form.odometer_km) : undefined,
                    location: form.location || undefined,
                    notes: form.notes || undefined,
                })
            }
            updateItem={(id, form) =>
                apiPatch(`/service-orders/${id}`, {
                    driver_id: form.driver_id || undefined,
                    description: form.description,
                    odometer_km: form.odometer_km ? Number(form.odometer_km) : undefined,
                    location: form.location || undefined,
                    notes: form.notes || undefined,
                })
            }
            searchFilter={(item, term) => {
                return (
                    item.order_number.toString().includes(term) ||
                    item.description.toLowerCase().includes(term) ||
                    (item.bus_plate?.toLowerCase().includes(term) ?? false)
                );
            }}
            rowActions={(item) =>
                item.status === "OPEN" ? (
                    <button className="button ghost sm" type="button" onClick={() => handleStart(item)}>
                        Iniciar
                    </button>
                ) : item.status === "IN_PROGRESS" ? (
                    <button className="button ghost sm" type="button" onClick={() => handleClose(item)}>
                        Fechar
                    </button>
                ) : null
            }
        />
    );
}
