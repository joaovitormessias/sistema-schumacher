import CRUDListPage, {
  type ColumnConfig,
  type FormFieldConfig,
  type VisibilityOption,
} from "../../components/layout/CRUDListPage";
import StatusBadge from "../../components/StatusBadge";
import { apiGet, apiPatch, apiPost } from "../../services/api";

type DriverItem = {
  id: string;
  name: string;
  document: string;
  phone: string;
  is_active: boolean;
};

type DriverForm = {
  name: string;
  document: string;
  phone: string;
};

const formFields: FormFieldConfig<DriverForm>[] = [
  {
    key: "name",
    label: "Nome do motorista",
    required: true,
    placeholder: "Ex: Carlos Silva",
  },
  {
    key: "document",
    label: "Documento",
    hint: "Opcional",
    placeholder: "CPF ou CNH",
  },
  {
    key: "phone",
    label: "Telefone",
    hint: "Opcional",
    placeholder: "(00) 00000-0000",
  },
];

const columns: ColumnConfig<DriverItem>[] = [
  {
    label: "Nome",
    accessor: (item) => item.name,
  },
  {
    label: "Documento",
    accessor: (item) => item.document || "-",
    hideOnMobile: true,
  },
  {
    label: "Telefone",
    accessor: (item) => item.phone || "-",
    hideOnMobile: true,
  },
  {
    label: "Status",
    render: (item) => (
      <StatusBadge tone={item.is_active ? "success" : "warning"}>
        {item.is_active ? "Ativo" : "Inativo"}
      </StatusBadge>
    ),
  },
];

const visibilityOptions: VisibilityOption<DriverItem>[] = [
  { label: "Ativos", value: "active", predicate: (item) => item.is_active },
  { label: "Inativos", value: "inactive", predicate: (item) => !item.is_active },
  { label: "Todos", value: "all", predicate: () => true },
];

export default function Drivers() {
  return (
    <CRUDListPage<DriverItem, DriverForm>
      title="Motoristas"
      subtitle="Cadastro e vínculo com viagens."
      meta={<span className="badge">MVP</span>}
      formTitle="Cadastro rápido"
      listTitle="Equipe cadastrada"
      createLabel="Criar motorista"
      updateLabel="Salvar motorista"
      emptyState={{
        title: "Nenhum motorista encontrado",
        description: "Tente ajustar a busca ou cadastre um novo motorista.",
      }}
      formFields={formFields}
      columns={columns}
      initialForm={{ name: "", document: "", phone: "" }}
      mapItemToForm={(item) => ({
        name: item.name,
        document: item.document || "",
        phone: item.phone || "",
      })}
      getId={(item) => item.id}
      fetchItems={async ({ page, pageSize, search }) => {
        const searchParam = search ? `&search=${encodeURIComponent(search)}` : "";
        return apiGet<DriverItem[]>(`/drivers?limit=${pageSize}&offset=${page * pageSize}${searchParam}`);
      }}
      createItem={(form) => apiPost("/drivers", form)}
      updateItem={(id, form) => apiPatch(`/drivers/${id}`, form)}
      softDeleteItem={(item) => apiPatch(`/drivers/${item.id}`, { is_active: false })}
      restoreItem={(item) => apiPatch(`/drivers/${item.id}`, { is_active: true })}
      getIsActive={(item) => item.is_active}
      searchFilter={(item, term) =>
        [item.name, item.document, item.phone].some((value) =>
          value?.toLowerCase().includes(term)
        )
      }
      visibilityOptions={visibilityOptions}
      visibilityDefault="active"
      layout="split"
      queryKey={["drivers"]}
      serverSideSearch
    />
  );
}
