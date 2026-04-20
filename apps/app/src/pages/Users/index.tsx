import { useMemo, useState, type FormEvent } from "react";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import DataTable, { type DataTableColumn } from "../../components/table/DataTable";
import Drawer from "../../components/overlay/Drawer";
import EmptyState from "../../components/EmptyState";
import FormField from "../../components/FormField";
import InlineAlert from "../../components/InlineAlert";
import LoadingState from "../../components/LoadingState";
import PageHeader from "../../components/PageHeader";
import StatusBadge from "../../components/StatusBadge";
import PaginationControls from "../../components/input/PaginationControls";
import SearchToolbar from "../../components/input/SearchToolbar";
import useToast from "../../hooks/useToast";
import { apiGet, apiPatch, apiPost } from "../../services/api";

type UserControlItem = {
  user_id: string;
  email: string;
  full_name: string;
  roles: string[];
  can_access_saldo: boolean;
  recipient_id?: string | null;
  has_recipient: boolean;
  is_active: boolean;
};

type UserFormState = {
  email: string;
  full_name: string;
  password: string;
  can_access_saldo: boolean;
};

const PAGE_SIZE = 50;

function initialUserForm(): UserFormState {
  return {
    email: "",
    full_name: "",
    password: "",
    can_access_saldo: false,
  };
}

export default function Users() {
  const toast = useToast();
  const queryClient = useQueryClient();
  const [page, setPage] = useState(0);
  const [query, setQuery] = useState("");

  const [editingUser, setEditingUser] = useState<UserControlItem | null>(null);
  const [formOpen, setFormOpen] = useState(false);
  const [form, setForm] = useState<UserFormState>(initialUserForm());
  const [submitError, setSubmitError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  const [resetTarget, setResetTarget] = useState<UserControlItem | null>(null);
  const [resetPassword, setResetPassword] = useState("");
  const [resetOpen, setResetOpen] = useState(false);
  const [resetSubmitting, setResetSubmitting] = useState(false);
  const [resetError, setResetError] = useState<string | null>(null);

  const usersQuery = useQuery({
    queryKey: ["users-control", page, PAGE_SIZE, query.trim().toLowerCase()],
    queryFn: () =>
      apiGet<UserControlItem[]>(
        `/users?limit=${PAGE_SIZE}&offset=${page * PAGE_SIZE}&search=${encodeURIComponent(
          query.trim()
        )}`
      ),
  });

  const items = usersQuery.data ?? [];
  const creating = !editingUser;

  const columns = useMemo<DataTableColumn<UserControlItem>[]>(
    () => [
      {
        label: "Nome",
        accessor: (item) => item.full_name || "-",
      },
      {
        label: "Email",
        accessor: (item) => item.email || "-",
        hideOnMobile: true,
      },
      {
        label: "Permissao saldo",
        render: (item) => (
          <StatusBadge tone={item.can_access_saldo ? "success" : "warning"}>
            {item.can_access_saldo ? "Liberado" : "Bloqueado"}
          </StatusBadge>
        ),
      },
      {
        label: "Roles",
        accessor: (item) => (item.roles?.length ? item.roles.join(", ") : "-"),
        hideOnMobile: true,
      },
    ],
    []
  );

  const openCreate = () => {
    setEditingUser(null);
    setForm(initialUserForm());
    setSubmitError(null);
    setFormOpen(true);
  };

  const openEdit = (item: UserControlItem) => {
    setEditingUser(item);
    setForm({
      email: item.email || "",
      full_name: item.full_name || "",
      password: "",
      can_access_saldo: item.can_access_saldo,
    });
    setSubmitError(null);
    setFormOpen(true);
  };

  const handleSubmit = async (event: FormEvent) => {
    event.preventDefault();
    setSubmitError(null);

    if (!form.email.trim() || !form.full_name.trim()) {
      setSubmitError("Nome e email sao obrigatorios.");
      return;
    }
    if (creating && form.password.trim().length < 8) {
      setSubmitError("A senha inicial precisa ter pelo menos 8 caracteres.");
      return;
    }

    try {
      setSubmitting(true);
      if (creating) {
        await apiPost<UserControlItem>("/users", {
          email: form.email.trim(),
          full_name: form.full_name.trim(),
          password: form.password.trim(),
          can_access_saldo: form.can_access_saldo,
        });
        toast.success("Usuario criado com sucesso.");
      } else if (editingUser) {
        await apiPatch<UserControlItem>(`/users/${editingUser.user_id}`, {
          email: form.email.trim(),
          full_name: form.full_name.trim(),
          can_access_saldo: form.can_access_saldo,
        });
        toast.success("Usuario atualizado com sucesso.");
      }

      setFormOpen(false);
      await queryClient.invalidateQueries({ queryKey: ["users-control"] });
    } catch (err: any) {
      setSubmitError(err?.message || "Nao foi possivel salvar o usuario.");
    } finally {
      setSubmitting(false);
    }
  };

  const openResetPassword = (item: UserControlItem) => {
    setResetTarget(item);
    setResetPassword("");
    setResetError(null);
    setResetOpen(true);
  };

  const handleResetPassword = async (event: FormEvent) => {
    event.preventDefault();
    if (!resetTarget) return;
    setResetError(null);

    if (resetPassword.trim().length < 8) {
      setResetError("A nova senha precisa ter pelo menos 8 caracteres.");
      return;
    }

    try {
      setResetSubmitting(true);
      await apiPost(`/users/${resetTarget.user_id}/reset-password`, {
        password: resetPassword.trim(),
      });
      toast.success("Senha redefinida com sucesso.");
      setResetOpen(false);
    } catch (err: any) {
      setResetError(err?.message || "Nao foi possivel redefinir a senha.");
    } finally {
      setResetSubmitting(false);
    }
  };

  return (
    <section className="page">
      <PageHeader
        title="Usuarios"
        subtitle="Controle de acesso ao saldo e credenciais."
        primaryAction={
          <button className="button" type="button" onClick={openCreate}>
            Novo usuario
          </button>
        }
      />

      <div className="section">
        <div className="section-header">
          <div className="section-title">Usuarios cadastrados</div>
        </div>

        <SearchToolbar
          value={query}
          onChange={setQuery}
          placeholder="Buscar por nome, email ou UUID"
          inputLabel="Buscar usuarios"
          resultCount={items.length}
        />

        <PaginationControls
          page={page}
          pageSize={PAGE_SIZE}
          itemCount={items.length}
          onPageChange={setPage}
          disabled={usersQuery.isLoading}
        />

        {usersQuery.isLoading ? <LoadingState label="Carregando usuarios..." /> : null}
        {usersQuery.error ? (
          <InlineAlert tone="error">
            {(usersQuery.error as Error).message || "Erro ao carregar usuarios"}
          </InlineAlert>
        ) : null}

        {!usersQuery.isLoading && !usersQuery.error ? (
          <DataTable
            columns={columns}
            rows={items}
            rowKey={(item) => item.user_id}
            actions={(item) => (
              <div style={{ display: "flex", gap: "8px", flexWrap: "wrap" }}>
                <button className="button secondary sm" type="button" onClick={() => openEdit(item)}>
                  Editar
                </button>
                <button className="button ghost sm" type="button" onClick={() => openResetPassword(item)}>
                  Resetar senha
                </button>
              </div>
            )}
            emptyState={
              <EmptyState
                title="Nenhum usuario encontrado"
                description="Ajuste o filtro de busca para localizar usuarios."
              />
            }
          />
        ) : null}
      </div>

      <Drawer
        open={formOpen}
        title={creating ? "Criar usuario" : "Editar usuario"}
        description={
          creating
            ? "Crie um novo usuario com acesso opcional ao modulo de saldo."
            : editingUser
            ? `${editingUser.full_name || "-"} (${editingUser.email || "-"})`
            : ""
        }
        onClose={() => setFormOpen(false)}
        footer={
          <>
            <button className="button secondary" type="button" onClick={() => setFormOpen(false)}>
              Cancelar
            </button>
            <button className="button" type="submit" form="user-form" disabled={submitting}>
              {submitting ? "Salvando..." : creating ? "Criar usuario" : "Salvar alteracoes"}
            </button>
          </>
        }
      >
        <form id="user-form" className="form-grid" onSubmit={handleSubmit}>
          <FormField label="Nome completo" required>
            <input
              className="input"
              value={form.full_name}
              onChange={(event) => setForm((prev) => ({ ...prev, full_name: event.target.value }))}
              placeholder="Nome do usuario"
              required
            />
          </FormField>

          <FormField label="Email" required>
            <input
              className="input"
              type="email"
              value={form.email}
              onChange={(event) => setForm((prev) => ({ ...prev, email: event.target.value }))}
              placeholder="usuario@empresa.com"
              required
            />
          </FormField>

          {creating ? (
            <FormField label="Senha inicial" required hint="Minimo de 8 caracteres">
              <input
                className="input"
                type="password"
                value={form.password}
                onChange={(event) => setForm((prev) => ({ ...prev, password: event.target.value }))}
                placeholder="********"
                minLength={8}
                required
              />
            </FormField>
          ) : null}

          <FormField label="Acesso ao saldo">
            <label className="checkbox">
              <input
                type="checkbox"
                checked={form.can_access_saldo}
                onChange={(event) =>
                  setForm((prev) => ({ ...prev, can_access_saldo: event.target.checked }))
                }
              />
              Permitir tela e operacoes de saldo
            </label>
          </FormField>

          {submitError ? (
            <div className="full-span">
              <InlineAlert tone="error">{submitError}</InlineAlert>
            </div>
          ) : null}
        </form>
      </Drawer>

      <Drawer
        open={resetOpen}
        title="Resetar senha"
        description={
          resetTarget
            ? `Defina uma nova senha para ${resetTarget.full_name || resetTarget.email || resetTarget.user_id}.`
            : ""
        }
        onClose={() => setResetOpen(false)}
        footer={
          <>
            <button className="button secondary" type="button" onClick={() => setResetOpen(false)}>
              Cancelar
            </button>
            <button className="button" type="submit" form="reset-password-form" disabled={resetSubmitting}>
              {resetSubmitting ? "Redefinindo..." : "Resetar senha"}
            </button>
          </>
        }
      >
        <form id="reset-password-form" className="form-grid" onSubmit={handleResetPassword}>
          <FormField label="Nova senha" required hint="Minimo de 8 caracteres">
            <input
              className="input"
              type="password"
              value={resetPassword}
              onChange={(event) => setResetPassword(event.target.value)}
              placeholder="********"
              minLength={8}
              required
            />
          </FormField>

          {resetError ? (
            <div className="full-span">
              <InlineAlert tone="error">{resetError}</InlineAlert>
            </div>
          ) : null}
        </form>
      </Drawer>
    </section>
  );
}
