import { useEffect, useMemo, useState, type FormEvent, type ReactNode } from "react";
import EmptyState from "../EmptyState";
import FormField from "../FormField";
import InlineAlert from "../InlineAlert";
import LoadingState from "../LoadingState";
import PageHeader from "../PageHeader";
import StatusBadge from "../StatusBadge";
import PaginationControls from "../input/PaginationControls";
import SearchToolbar from "../input/SearchToolbar";
import useToast from "../../hooks/useToast";
import TableActionButtons from "./TableActionButtons";
import DataTable, { type DataTableColumn } from "../table/DataTable";

type SelectOption = { label: string; value: string };

export type FormFieldConfig<F> = {
  key: keyof F;
  label: string;
  type?: "text" | "number" | "select" | "checkbox" | "datetime" | "textarea";
  required?: boolean;
  hint?: string;
  placeholder?: string;
  options?: SelectOption[];
  colSpan?: "full";
  disabled?: boolean;
  inputProps?: Record<string, string | number | boolean>;
  showWhen?: "create" | "edit" | "always";
};

export type ColumnConfig<T> = DataTableColumn<T>;

export type VisibilityOption<T> = {
  label: string;
  value: string;
  predicate: (item: T) => boolean;
};

type EmptyStateConfig = {
  title: string;
  description: string;
};

type CRUDListPageProps<T, F> = {
  title: string;
  subtitle?: string;
  eyebrow?: string;
  meta?: ReactNode;
  primaryAction?: ReactNode;
  secondaryActions?: ReactNode;
  formTitle: string;
  listTitle: string;
  createLabel?: string;
  updateLabel?: string;
  emptyState: EmptyStateConfig;
  formFields: FormFieldConfig<F>[];
  columns: ColumnConfig<T>[];
  initialForm: F;
  mapItemToForm: (item: T) => F;
  getId: (item: T) => string;
  fetchItems: (params: { page: number; pageSize: number }) => Promise<T[]>;
  createItem: (form: F) => Promise<void>;
  updateItem?: (id: string, form: F) => Promise<void>;
  softDeleteItem?: (item: T) => Promise<void>;
  restoreItem?: (item: T) => Promise<void>;
  getIsActive?: (item: T) => boolean;
  searchFilter: (item: T, term: string) => boolean;
  visibilityOptions?: VisibilityOption<T>[];
  visibilityDefault?: string;
  extraFilters?: ReactNode;
  rowActions?: (item: T) => ReactNode;
  pageSize?: number;
  layout?: "stacked" | "split";
};

export default function CRUDListPage<T, F>({
  title,
  subtitle,
  eyebrow,
  meta,
  primaryAction,
  secondaryActions,
  formTitle,
  listTitle,
  createLabel = "Criar",
  updateLabel = "Salvar",
  emptyState,
  formFields,
  columns,
  initialForm,
  mapItemToForm,
  getId,
  fetchItems,
  createItem,
  updateItem,
  softDeleteItem,
  restoreItem,
  getIsActive,
  searchFilter,
  visibilityOptions,
  visibilityDefault,
  extraFilters,
  rowActions,
  pageSize = 50,
  layout = "split",
}: CRUDListPageProps<T, F>) {
  const toast = useToast();
  const [items, setItems] = useState<T[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [page, setPage] = useState(0);
  const [query, setQuery] = useState("");
  const [form, setForm] = useState<F>(initialForm);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const [visibility, setVisibility] = useState(
    visibilityDefault ?? visibilityOptions?.[0]?.value ?? ""
  );

  const canEdit = Boolean(updateItem);
  const isEditing = canEdit && Boolean(editingId);

  const load = async () => {
    try {
      setLoading(true);
      const data = await fetchItems({ page, pageSize });
      setItems(data);
    } catch (err: any) {
      setError(err.message || "Erro ao carregar dados");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    load();
  }, [page]);

  const filteredItems = useMemo(() => {
    const term = query.trim().toLowerCase();
    let next = items;
    if (visibilityOptions && visibility) {
      const option = visibilityOptions.find((opt) => opt.value === visibility);
      if (option) {
        next = next.filter(option.predicate);
      }
    }
    if (!term) return next;
    return next.filter((item) => searchFilter(item, term));
  }, [items, query, visibility, visibilityOptions, searchFilter]);

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError(null);
    setSubmitting(true);
    try {
      if (isEditing && updateItem) {
        await updateItem(editingId as string, form);
        toast.success("Alterações salvas com sucesso.");
      } else {
        await createItem(form);
        toast.success("Cadastro criado com sucesso.");
      }
      setEditingId(null);
      setForm(initialForm);
      setPage(0);
      await load();
    } catch (err: any) {
      setError(err.message || "Erro ao salvar");
    } finally {
      setSubmitting(false);
    }
  };

  const startEdit = (item: T) => {
    if (!canEdit) return;
    setEditingId(getId(item));
    setForm(mapItemToForm(item));
  };

  const cancelEdit = () => {
    setEditingId(null);
    setForm(initialForm);
  };

  const handleDeactivate = async (item: T) => {
    if (!softDeleteItem) return;
    setError(null);
    try {
      await softDeleteItem(item);
      toast.success("Item desativado.");
      await load();
    } catch (err: any) {
      setError(err.message || "Erro ao desativar");
    }
  };

  const handleRestore = async (item: T) => {
    if (!restoreItem) return;
    setError(null);
    try {
      await restoreItem(item);
      toast.success("Item reativado.");
      await load();
    } catch (err: any) {
      setError(err.message || "Erro ao reativar");
    }
  };

  const renderField = (field: FormFieldConfig<F>) => {
    const value = form[field.key];
    const inputProps = field.inputProps ?? {};
    const type = field.type ?? "text";

    if (type === "checkbox") {
      return (
        <div className="form-field" key={String(field.key)}>
          <span className="form-label">{field.label}</span>
          <label className="checkbox">
            <input
              type="checkbox"
              checked={Boolean(value)}
              onChange={(e) =>
                setForm((prev) => ({ ...prev, [field.key]: e.target.checked } as F))
              }
              disabled={field.disabled}
              {...inputProps}
            />
            {field.hint ?? "Ativo"}
          </label>
        </div>
      );
    }

    return (
      <FormField
        key={String(field.key)}
        label={field.label}
        hint={field.hint}
        required={field.required}
      >
        {type === "select" ? (
          <select
            className="input"
            value={String(value ?? "")}
            onChange={(e) =>
              setForm((prev) => ({ ...prev, [field.key]: e.target.value } as F))
            }
            required={field.required}
            disabled={field.disabled}
            {...inputProps}
          >
            {field.options?.map((opt) => (
              <option key={opt.value} value={opt.value}>
                {opt.label}
              </option>
            ))}
          </select>
        ) : type === "textarea" ? (
          <textarea
            className="input"
            value={String(value ?? "")}
            placeholder={field.placeholder}
            onChange={(e) =>
              setForm((prev) => ({ ...prev, [field.key]: e.target.value } as F))
            }
            required={field.required}
            disabled={field.disabled}
            {...inputProps}
          />
        ) : (
          <input
            className="input"
            type={type === "datetime" ? "datetime-local" : type}
            value={value === undefined || value === null ? "" : String(value)}
            placeholder={field.placeholder}
            onChange={(e) => {
              const nextValue =
                type === "number"
                  ? e.target.value === ""
                    ? ""
                    : Number(e.target.value)
                  : e.target.value;
              setForm((prev) => ({ ...prev, [field.key]: nextValue } as F));
            }}
            required={field.required}
            disabled={field.disabled}
            {...inputProps}
          />
        )}
      </FormField>
    );
  };

  const filterControls = visibilityOptions ? (
    <select
      className="input"
      value={visibility}
      onChange={(e) => setVisibility(e.target.value)}
      aria-label="Filtrar status"
    >
      {visibilityOptions.map((option) => (
        <option key={option.value} value={option.value}>
          {option.label}
        </option>
      ))}
    </select>
  ) : null;

  return (
    <section className="page">
      <PageHeader
        title={title}
        subtitle={subtitle}
        eyebrow={eyebrow}
        meta={meta}
        primaryAction={primaryAction}
        secondaryActions={secondaryActions}
      />

      <div className={`crud-layout ${layout === "split" ? "split-layout" : "stacked-layout"}`}>
        <div className="crud-form section" id="crud-form">
          <div className="section-header">
            <div className="section-title">{formTitle}</div>
            {isEditing ? <StatusBadge tone="warning">Editando</StatusBadge> : null}
          </div>
          <form className="form-grid" onSubmit={handleSubmit}>
            {formFields.map((field) => {
              const showWhen = field.showWhen ?? "always";
              if ((showWhen === "create" && isEditing) || (showWhen === "edit" && !isEditing)) {
                return null;
              }
              return field.colSpan === "full" ? (
                <div key={String(field.key)} className="full-span">
                  {renderField(field)}
                </div>
              ) : (
                renderField(field)
              );
            })}
            <div className="form-actions full-width-mobile full-span">
              <button className="button" type="submit" disabled={submitting}>
                {isEditing ? updateLabel : createLabel}
              </button>
              {isEditing ? (
                <button className="button secondary" type="button" onClick={cancelEdit}>
                  Cancelar
                </button>
              ) : null}
            </div>
          </form>
        </div>

        <div className="crud-list section">
          <div className="section-header">
            <div className="section-title">{listTitle}</div>
          </div>
          <SearchToolbar
            value={query}
            onChange={setQuery}
            placeholder="Buscar"
            inputLabel="Buscar"
            filters={
              <>
                {filterControls}
                {extraFilters}
              </>
            }
            resultCount={filteredItems.length}
          />
          <PaginationControls
            page={page}
            pageSize={pageSize}
            itemCount={items.length}
            onPageChange={setPage}
            disabled={loading}
          />

          {error ? <InlineAlert tone="error">{error}</InlineAlert> : null}

          {loading ? (
            <LoadingState label="Carregando..." />
          ) : (
            <DataTable
              columns={columns}
              rows={filteredItems}
              rowKey={getId}
              actions={(item) => {
                const isActive = getIsActive ? getIsActive(item) : true;
                return (
                  <>
                    <TableActionButtons
                      isActive={isActive}
                      onEdit={canEdit ? () => startEdit(item) : undefined}
                      onDeactivate={softDeleteItem ? () => handleDeactivate(item) : undefined}
                      onRestore={restoreItem ? () => handleRestore(item) : undefined}
                      disableEdit={submitting}
                    />
                    {rowActions ? rowActions(item) : null}
                  </>
                );
              }}
              emptyState={
                <EmptyState title={emptyState.title} description={emptyState.description} />
              }
            />
          )}
        </div>
      </div>
    </section>
  );
}
