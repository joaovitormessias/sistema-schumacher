import CRUDListPage, {
  type ColumnConfig,
  type FormFieldConfig,
  type VisibilityOption,
} from "../../components/layout/CRUDListPage";
import StatusBadge from "../../components/StatusBadge";
import { apiGet, apiPatch, apiPost } from "../../services/api";

type BusItem = {
  id: string;
  name: string;
  plate: string;
  capacity: number;
  seat_map_name: string;
  is_active: boolean;
};

type BusForm = {
  name: string;
  plate: string;
  capacity: number;
  seat_map_name: string;
  create_seats: boolean;
};

const formFields: FormFieldConfig<BusForm>[] = [
  {
    key: "name",
    label: "Nome do ônibus",
    required: true,
    placeholder: "Ex: Executivo 1",
  },
  {
    key: "plate",
    label: "Placa",
    hint: "Opcional",
    placeholder: "ABC-1D23",
  },
  {
    key: "capacity",
    label: "Capacidade",
    type: "number",
    required: true,
    placeholder: "40",
    inputProps: { min: 1 },
  },
  {
    key: "seat_map_name",
    label: "Mapa de poltronas",
    hint: "Nome do layout",
    placeholder: "Semi-leito",
  },
  {
    key: "create_seats",
    label: "Automação",
    type: "checkbox",
    hint: "Gerar poltronas automaticamente",
    showWhen: "create",
  },
];

const columns: ColumnConfig<BusItem>[] = [
  { label: "Nome", accessor: (item) => item.name },
  { label: "Placa", accessor: (item) => item.plate || "-", hideOnMobile: true },
  { label: "Capacidade", accessor: (item) => item.capacity },
  { label: "Mapa", accessor: (item) => item.seat_map_name || "-", hideOnMobile: true },
  {
    label: "Status",
    render: (item) => (
      <StatusBadge tone={item.is_active ? "success" : "warning"}>
        {item.is_active ? "Ativo" : "Inativo"}
      </StatusBadge>
    ),
  },
];

const visibilityOptions: VisibilityOption<BusItem>[] = [
  { label: "Ativos", value: "active", predicate: (item) => item.is_active },
  { label: "Inativos", value: "inactive", predicate: (item) => !item.is_active },
  { label: "Todos", value: "all", predicate: () => true },
];

export default function Buses() {
  return (
    <CRUDListPage<BusItem, BusForm>
      title="Ônibus"
      subtitle="Frota, capacidade e mapa de poltronas."
      meta={<span className="badge">MVP</span>}
      formTitle="Cadastro rápido"
      listTitle="Frota cadastrada"
      createLabel="Criar ônibus"
      updateLabel="Salvar ônibus"
      emptyState={{
        title: "Nenhum ônibus encontrado",
        description: "Tente ajustar a busca ou cadastre um novo ônibus.",
      }}
      formFields={formFields}
      columns={columns}
      initialForm={{ name: "", plate: "", capacity: 40, seat_map_name: "", create_seats: true }}
      mapItemToForm={(item) => ({
        name: item.name,
        plate: item.plate || "",
        capacity: item.capacity,
        seat_map_name: item.seat_map_name || "",
        create_seats: true,
      })}
      getId={(item) => item.id}
      fetchItems={async ({ page, pageSize, search }) => {
        const searchParam = search ? `&search=${encodeURIComponent(search)}` : "";
        return apiGet<BusItem[]>(`/buses?limit=${pageSize}&offset=${page * pageSize}${searchParam}`);
      }}
      createItem={(form) =>
        apiPost("/buses", {
          name: form.name,
          plate: form.plate,
          capacity: Number(form.capacity),
          seat_map_name: form.seat_map_name,
          create_seats: form.create_seats,
        })
      }
      updateItem={(id, form) =>
        apiPatch(`/buses/${id}`, {
          name: form.name,
          plate: form.plate,
          capacity: Number(form.capacity),
          seat_map_name: form.seat_map_name,
        })
      }
      softDeleteItem={(item) => apiPatch(`/buses/${item.id}`, { is_active: false })}
      restoreItem={(item) => apiPatch(`/buses/${item.id}`, { is_active: true })}
      getIsActive={(item) => item.is_active}
      searchFilter={(item, term) =>
        [item.name, item.plate, item.seat_map_name].some((value) =>
          value?.toLowerCase().includes(term)
        )
      }
      visibilityOptions={visibilityOptions}
      visibilityDefault="active"
      layout="split"
      queryKey={["buses"]}
      serverSideSearch
    />
  );
}
