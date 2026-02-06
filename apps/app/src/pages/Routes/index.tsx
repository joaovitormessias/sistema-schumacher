import CRUDListPage, {
  type ColumnConfig,
  type FormFieldConfig,
  type VisibilityOption,
} from "../../components/layout/CRUDListPage";
import StatusBadge from "../../components/StatusBadge";
import { apiGet, apiPatch, apiPost } from "../../services/api";

type RouteItem = {
  id: string;
  name: string;
  origin_city: string;
  destination_city: string;
  is_active: boolean;
};

type RouteForm = {
  name: string;
  origin_city: string;
  destination_city: string;
};

const formFields: FormFieldConfig<RouteForm>[] = [
  {
    key: "name",
    label: "Nome da rota",
    required: true,
    placeholder: "Ex: Executivo Sul",
  },
  {
    key: "origin_city",
    label: "Cidade de origem",
    required: true,
    placeholder: "Ex: Curitiba",
  },
  {
    key: "destination_city",
    label: "Cidade de destino",
    required: true,
    placeholder: "Ex: Florianópolis",
  },
];

const columns: ColumnConfig<RouteItem>[] = [
  { label: "Nome", accessor: (item) => item.name },
  { label: "Origem", accessor: (item) => item.origin_city },
  { label: "Destino", accessor: (item) => item.destination_city },
  {
    label: "Status",
    render: (item) => (
      <StatusBadge tone={item.is_active ? "success" : "warning"}>
        {item.is_active ? "Ativa" : "Inativa"}
      </StatusBadge>
    ),
  },
];

const visibilityOptions: VisibilityOption<RouteItem>[] = [
  { label: "Ativas", value: "active", predicate: (item) => item.is_active },
  { label: "Inativas", value: "inactive", predicate: (item) => !item.is_active },
  { label: "Todas", value: "all", predicate: () => true },
];

export default function RoutesPage() {
  return (
    <CRUDListPage<RouteItem, RouteForm>
      title="Rotas"
      subtitle="Rotas, paradas e cidades atendidas."
      meta={<span className="badge">MVP</span>}
      formTitle="Cadastro rápido"
      listTitle="Rotas cadastradas"
      createLabel="Criar rota"
      updateLabel="Salvar rota"
      emptyState={{
        title: "Nenhuma rota encontrada",
        description: "Tente ajustar a busca ou crie uma nova rota.",
      }}
      formFields={formFields}
      columns={columns}
      initialForm={{ name: "", origin_city: "", destination_city: "" }}
      mapItemToForm={(item) => ({
        name: item.name,
        origin_city: item.origin_city,
        destination_city: item.destination_city,
      })}
      getId={(item) => item.id}
      fetchItems={async ({ page, pageSize }) =>
        apiGet<RouteItem[]>(`/routes?limit=${pageSize}&offset=${page * pageSize}`)
      }
      createItem={(form) => apiPost("/routes", form)}
      updateItem={(id, form) => apiPatch(`/routes/${id}`, form)}
      softDeleteItem={(item) => apiPatch(`/routes/${item.id}`, { is_active: false })}
      restoreItem={(item) => apiPatch(`/routes/${item.id}`, { is_active: true })}
      getIsActive={(item) => item.is_active}
      searchFilter={(item, term) =>
        [item.name, item.origin_city, item.destination_city].some((value) =>
          value?.toLowerCase().includes(term)
        )
      }
      visibilityOptions={visibilityOptions}
      visibilityDefault="active"
      layout="split"
    />
  );
}
