import { useEffect, useMemo, useState, type FormEvent, type ReactNode } from "react";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import EmptyState from "../EmptyState";
import FormField from "../FormField";
import InlineAlert from "../InlineAlert";
import PageHeader from "../PageHeader";
import StatusBadge from "../StatusBadge";
import { Skeleton } from "../feedback/SkeletonLoader";
import PaginationControls from "../input/PaginationControls";
import SearchToolbar from "../input/SearchToolbar";
import useToast from "../../hooks/useToast";
import useDebouncedValue from "../../hooks/useDebouncedValue";
import useMediaQuery from "../../hooks/useMediaQuery";
import ConfirmDialog from "../overlay/ConfirmDialog";
import TableActionButtons from "./TableActionButtons";
import DataTable, { type DataTableColumn } from "../table/DataTable";
import VirtualDataTable from "../data-display/VirtualDataTable";

type SelectOption = { label: string; value: string };

export type FormFieldConfig<F> = {
  key: keyof F;
  label: string;
  type?: "text" | "number" | "select" | "checkbox" | "datetime" | "textarea" | "file";
  required?: boolean;
  hint?: string;
  placeholder?: string;
  options?: SelectOption[];
  colSpan?: "full";
  disabled?: boolean;
  inputProps?: Record<string, string | number | boolean>;
  showWhen?: "create" | "edit" | "always";
  customRender?: (args: {
    value: unknown;
    onChange: (nextValue: unknown) => void;
    form: F;
    isEditing: boolean;
  }) => ReactNode;
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
  fetchItems: (params: {
    page: number;
    pageSize: number;
    search: string;
    visibility: string;
  }) => Promise<T[]>;
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
  hidePageHeader?: boolean;
  queryKey?: readonly unknown[];
  serverSideSearch?: boolean;
  searchDebounceMs?: number;
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
  hidePageHeader = false,
  queryKey,
  serverSideSearch = false,
  searchDebounceMs = 500,
}: CRUDListPageProps<T, F>) {
  const toast = useToast();
  const queryClient = useQueryClient();
  const [actionError, setActionError] = useState<string | null>(null);
  const [page, setPage] = useState(0);
  const [query, setQuery] = useState("");
  const [form, setForm] = useState<F>(initialForm);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const [visibility, setVisibility] = useState(
    visibilityDefault ?? visibilityOptions?.[0]?.value ?? ""
  );
  const [confirmState, setConfirmState] = useState<{
    mode: "deactivate" | "restore";
    item: T;
  } | null>(null);
  const isMobile = useMediaQuery("(max-width: 900px)");

  const canEdit = Boolean(updateItem);
  const isEditing = canEdit && Boolean(editingId);
  const hasActionColumn =
    canEdit || Boolean(softDeleteItem) || Boolean(restoreItem) || Boolean(rowActions);

  const baseQueryKey = useMemo<readonly unknown[]>(
    () => (queryKey && queryKey.length > 0 ? queryKey : ["crud-list", title]),
    [queryKey, title]
  );

  const debouncedQuery = useDebouncedValue(query, searchDebounceMs);
  const effectiveSearch = (serverSideSearch ? debouncedQuery : query).trim().toLowerCase();

  useEffect(() => {
    setPage(0);
  }, [effectiveSearch, visibility]);

  const itemsQuery = useQuery({
    queryKey: [...baseQueryKey, page, pageSize, visibility, effectiveSearch],
    queryFn: () =>
      fetchItems({
        page,
        pageSize,
        search: effectiveSearch,
        visibility,
      }),
  });

  const items = itemsQuery.data ?? [];

  const filteredItems = useMemo(() => {
    let next = items;

    if (visibilityOptions && visibility) {
      const option = visibilityOptions.find((opt) => opt.value === visibility);
      if (option) {
        next = next.filter(option.predicate);
      }
    }

    if (serverSideSearch || !effectiveSearch) {
      return next;
    }

    return next.filter((item) => searchFilter(item, effectiveSearch));
  }, [
    items,
    visibilityOptions,
    visibility,
    serverSideSearch,
    effectiveSearch,
    searchFilter,
  ]);

  const errorMessage = actionError || (itemsQuery.error as Error | undefined)?.message || null;
  const loading = itemsQuery.isPending;

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setActionError(null);
    setSubmitting(true);
    try {
      if (isEditing && updateItem) {
        await updateItem(editingId as string, form);
        toast.success("Alteracoes salvas com sucesso.");
      } else {
        await createItem(form);
        toast.success("Cadastro criado com sucesso.");
      }
      setEditingId(null);
      setForm(initialForm);
      setPage(0);
      await queryClient.invalidateQueries({ queryKey: baseQueryKey });
    } catch (err: any) {
      setActionError(err.message || "Erro ao salvar");
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
    setActionError(null);
    try {
      await softDeleteItem(item);
      toast.success("Item desativado.");
      await queryClient.invalidateQueries({ queryKey: baseQueryKey });
    } catch (err: any) {
      setActionError(err.message || "Erro ao desativar");
    }
  };

  const handleRestore = async (item: T) => {
    if (!restoreItem) return;
    setActionError(null);
    try {
      await restoreItem(item);
      toast.success("Item reativado.");
      await queryClient.invalidateQueries({ queryKey: baseQueryKey });
    } catch (err: any) {
      setActionError(err.message || "Erro ao reativar");
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
        {field.customRender ? (
          field.customRender({
            value,
            onChange: (nextValue) =>
              setForm((prev) => ({ ...prev, [field.key]: nextValue } as F)),
            form,
            isEditing,
          })
        ) : null}
        {!field.customRender && type === "select" ? (
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
        ) : null}
        {!field.customRender && type === "textarea" ? (
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
        ) : null}
        {!field.customRender && type === "file" ? (
          <input
            className="input"
            type="file"
            onChange={(e) => {
              const file = e.target.files?.[0] ?? null;
              setForm((prev) => ({ ...prev, [field.key]: file } as F));
            }}
            required={field.required}
            disabled={field.disabled}
            {...inputProps}
          />
        ) : null}
        {!field.customRender && !["select", "textarea", "file"].includes(type) ? (
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
        ) : null}
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
  const shouldVirtualize = !isMobile && filteredItems.length > 100;

  return (
    <section className="page">
      {!hidePageHeader ? (
        <PageHeader
          title={title}
          subtitle={subtitle}
          eyebrow={eyebrow}
          meta={meta}
          primaryAction={primaryAction}
          secondaryActions={secondaryActions}
        />
      ) : null}

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

          {errorMessage ? <InlineAlert tone="error">{errorMessage}</InlineAlert> : null}

          {loading ? (
            <Skeleton.Table
              rows={6}
              columns={Math.max(columns.length + (hasActionColumn ? 1 : 0), 2)}
            />
          ) : shouldVirtualize ? (
            <VirtualDataTable
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
                      onDeactivate={
                        softDeleteItem
                          ? () => setConfirmState({ mode: "deactivate", item })
                          : undefined
                      }
                      onRestore={
                        restoreItem ? () => setConfirmState({ mode: "restore", item }) : undefined
                      }
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
                      onDeactivate={
                        softDeleteItem
                          ? () => setConfirmState({ mode: "deactivate", item })
                          : undefined
                      }
                      onRestore={
                        restoreItem ? () => setConfirmState({ mode: "restore", item }) : undefined
                      }
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
      <ConfirmDialog
        open={Boolean(confirmState)}
        title={confirmState?.mode === "restore" ? "Reativar item" : "Desativar item"}
        description={
          confirmState?.mode === "restore"
            ? "Deseja reativar este registro?"
            : "Deseja desativar este registro?"
        }
        confirmLabel={confirmState?.mode === "restore" ? "Reativar" : "Desativar"}
        tone={confirmState?.mode === "restore" ? "default" : "danger"}
        onCancel={() => setConfirmState(null)}
        onConfirm={async () => {
          if (!confirmState) return;
          if (confirmState.mode === "restore") {
            await handleRestore(confirmState.item);
          } else {
            await handleDeactivate(confirmState.item);
          }
          setConfirmState(null);
        }}
      />
    </section>
  );
}
