import { useCallback, useEffect, useMemo, useState } from "react";
import { Link, useParams } from "react-router-dom";
import PageHeader from "../../components/PageHeader";
import Stepper, { type StepperStep } from "../../components/Stepper";
import InlineAlert from "../../components/InlineAlert";
import StatusBadge from "../../components/StatusBadge";
import { APIRequestError, apiGet, apiPost, apiPut } from "../../services/api";
import { getSupabaseClient } from "../../services/supabase";
import {
  type OperationalStatus,
  type TripAttachment,
  type TripAuthorization,
  type TripChecklist,
  type TripDriverReport,
  type TripManifestEntry,
  type TripOperationsTrip,
  type TripReconciliation,
  type TripRequest,
} from "../../types/tripOperations";
import { tripOperationalStatusLabel, tripStatusLabel } from "../../utils/labels";

type ExpenseItem = {
  id: string;
  description: string;
  amount: number;
  receipt_verified: boolean;
};

const flowOrder: OperationalStatus[] = [
  "REQUESTED",
  "PASSENGERS_READY",
  "ITINERARY_READY",
  "DISPATCH_VALIDATED",
  "AUTHORIZED",
  "IN_PROGRESS",
  "RETURNED",
  "RETURN_CHECKED",
  "SETTLED",
  "CLOSED",
];

const flowTitle: Record<OperationalStatus, string> = {
  REQUESTED: "Solicitacao",
  PASSENGERS_READY: "Manifesto",
  ITINERARY_READY: "Roteiro",
  DISPATCH_VALIDATED: "D-1",
  AUTHORIZED: "Autorizacao",
  IN_PROGRESS: "Partida",
  RETURNED: "Retorno",
  RETURN_CHECKED: "Conferencia",
  SETTLED: "Acerto",
  CLOSED: "Fechamento",
};

function fromInputDateTime(value: string) {
  if (!value) return undefined;
  return new Date(value).toISOString();
}

export default function TripOperations() {
  const { tripId = "" } = useParams();
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [workflowBlocked, setWorkflowBlocked] = useState<string[]>([]);

  const [trip, setTrip] = useState<TripOperationsTrip | null>(null);
  const [tripRequests, setTripRequests] = useState<TripRequest[]>([]);
  const [manifest, setManifest] = useState<TripManifestEntry[]>([]);
  const [authorizations, setAuthorizations] = useState<TripAuthorization[]>([]);
  const [preChecklist, setPreChecklist] = useState<TripChecklist | null>(null);
  const [returnChecklist, setReturnChecklist] = useState<TripChecklist | null>(null);
  const [driverReport, setDriverReport] = useState<TripDriverReport | null>(null);
  const [reconciliation, setReconciliation] = useState<TripReconciliation | null>(null);
  const [attachments, setAttachments] = useState<TripAttachment[]>([]);
  const [expenses, setExpenses] = useState<ExpenseItem[]>([]);
  const [fiscalDocs, setFiscalDocs] = useState<Array<{ id: string; status: string; document_type: string }>>([]);

  const [requestForm, setRequestForm] = useState({
    source: "SYSTEM",
    requester_name: "",
    requester_contact: "",
    requested_departure_at: "",
    notes: "",
  });
  const [manifestForm, setManifestForm] = useState({
    passenger_name: "",
    passenger_document: "",
    passenger_phone: "",
  });
  const [authorizationForm, setAuthorizationForm] = useState({
    authority: "ANTT",
    status: "PENDING",
    protocol_number: "",
    license_number: "",
    src_policy_number: "",
    src_valid_until: "",
    exceptional_deadline_ok: true,
    notes: "",
  });
  const [preChecklistForm, setPreChecklistForm] = useState({
    is_complete: false,
    documents_checked: false,
    tachograph_checked: false,
    receipts_checked: false,
    rest_compliance_ok: true,
    notes: "",
  });
  const [returnChecklistForm, setReturnChecklistForm] = useState({
    is_complete: false,
    documents_checked: false,
    tachograph_checked: false,
    receipts_checked: false,
    rest_compliance_ok: true,
    notes: "",
  });
  const [driverReportForm, setDriverReportForm] = useState({
    odometer_start: "",
    odometer_end: "",
    fuel_used_liters: "",
    incidents: "",
    delays: "",
    rest_hours: "",
    notes: "",
  });
  const [reconciliationForm, setReconciliationForm] = useState({
    total_receipts_amount: "",
    receipts_validated: false,
    verified_expense_ids: [] as string[],
    notes: "",
  });
  const [attachmentForm, setAttachmentForm] = useState({
    attachment_type: "OTHER",
    file: null as File | null,
  });

  const currentStepIndex = useMemo(() => {
    if (!trip) return 0;
    return Math.max(flowOrder.indexOf(trip.operational_status), 0);
  }, [trip]);

  const nextStatus = useMemo(() => {
    if (!trip) return null;
    const idx = flowOrder.indexOf(trip.operational_status);
    if (idx < 0 || idx === flowOrder.length - 1) return null;
    return flowOrder[idx + 1];
  }, [trip]);

  const stepperSteps: StepperStep[] = useMemo(
    () =>
      flowOrder.map((step, idx) => ({
        id: step,
        title: flowTitle[step],
        summary: tripOperationalStatusLabel[step],
        status: idx < currentStepIndex ? "complete" : idx === currentStepIndex ? "current" : "upcoming",
        disabled: true,
      })),
    [currentStepIndex]
  );

  const compliance = useMemo(() => {
    const issued = authorizations.some((item) => item.status === "ISSUED");
    const srcValid = authorizations.some(
      (item) => item.status === "ISSUED" && item.src_policy_number && item.src_valid_until
    );
    return {
      lvSisaut: issued,
      tafSrc: srcValid,
      tachograph: Boolean(returnChecklist?.tachograph_checked),
      receipts: Boolean(reconciliation?.receipts_validated),
      fiscal: fiscalDocs.length > 0,
    };
  }, [authorizations, returnChecklist, reconciliation, fiscalDocs.length]);

  const loadData = useCallback(async () => {
    if (!tripId) return;
    setLoading(true);
    setError(null);
    try {
      const safeGet = async <T,>(path: string): Promise<T | null> => {
        try {
          return await apiGet<T>(path);
        } catch (err) {
          if (err instanceof APIRequestError && err.code === "NOT_FOUND") {
            return null;
          }
          throw err;
        }
      };

      const [
        loadedTrip,
        loadedRequests,
        loadedManifest,
        loadedAuthorizations,
        loadedPreChecklist,
        loadedReturnChecklist,
        loadedDriverReport,
        loadedReconciliation,
        loadedAttachments,
        loadedExpenses,
        loadedFiscalDocs,
      ] = await Promise.all([
        apiGet<TripOperationsTrip>(`/trips/${tripId}`),
        apiGet<TripRequest[]>("/trip-requests?limit=200"),
        apiGet<TripManifestEntry[]>(`/trips/${tripId}/manifest`),
        apiGet<TripAuthorization[]>(`/trips/${tripId}/authorizations`),
        safeGet<TripChecklist>(`/trips/${tripId}/checklists/PRE_DEPARTURE`),
        safeGet<TripChecklist>(`/trips/${tripId}/checklists/RETURN`),
        safeGet<TripDriverReport>(`/trips/${tripId}/driver-report`),
        safeGet<TripReconciliation>(`/trips/${tripId}/reconciliation`),
        apiGet<TripAttachment[]>(`/trips/${tripId}/attachments`),
        apiGet<ExpenseItem[]>(`/trip-expenses?trip_id=${tripId}&approved=true&limit=500`),
        apiGet<Array<{ id: string; status: string; document_type: string }>>(`/fiscal-documents?trip_id=${tripId}&limit=200`),
      ]);

      setTrip(loadedTrip);
      setTripRequests(loadedRequests);
      setManifest(loadedManifest);
      setAuthorizations(loadedAuthorizations);
      setPreChecklist(loadedPreChecklist);
      setReturnChecklist(loadedReturnChecklist);
      setDriverReport(loadedDriverReport);
      setReconciliation(loadedReconciliation);
      setAttachments(loadedAttachments);
      setExpenses(loadedExpenses);
      setFiscalDocs(loadedFiscalDocs);

      if (loadedPreChecklist) {
        setPreChecklistForm({
          is_complete: loadedPreChecklist.is_complete,
          documents_checked: loadedPreChecklist.documents_checked,
          tachograph_checked: loadedPreChecklist.tachograph_checked,
          receipts_checked: loadedPreChecklist.receipts_checked,
          rest_compliance_ok: loadedPreChecklist.rest_compliance_ok,
          notes: loadedPreChecklist.notes ?? "",
        });
      }
      if (loadedReturnChecklist) {
        setReturnChecklistForm({
          is_complete: loadedReturnChecklist.is_complete,
          documents_checked: loadedReturnChecklist.documents_checked,
          tachograph_checked: loadedReturnChecklist.tachograph_checked,
          receipts_checked: loadedReturnChecklist.receipts_checked,
          rest_compliance_ok: loadedReturnChecklist.rest_compliance_ok,
          notes: loadedReturnChecklist.notes ?? "",
        });
      }
      if (loadedDriverReport) {
        setDriverReportForm({
          odometer_start: loadedDriverReport.odometer_start?.toString() ?? "",
          odometer_end: loadedDriverReport.odometer_end?.toString() ?? "",
          fuel_used_liters: loadedDriverReport.fuel_used_liters?.toString() ?? "",
          incidents: loadedDriverReport.incidents ?? "",
          delays: loadedDriverReport.delays ?? "",
          rest_hours: loadedDriverReport.rest_hours?.toString() ?? "",
          notes: loadedDriverReport.notes ?? "",
        });
      }
      if (loadedReconciliation) {
        setReconciliationForm({
          total_receipts_amount: loadedReconciliation.total_receipts_amount.toString(),
          receipts_validated: loadedReconciliation.receipts_validated,
          verified_expense_ids: loadedReconciliation.verified_expense_ids ?? [],
          notes: loadedReconciliation.notes ?? "",
        });
      }
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : "Falha ao carregar dados operacionais");
    } finally {
      setLoading(false);
    }
  }, [tripId]);

  useEffect(() => {
    loadData();
  }, [loadData]);

  const runAction = async (action: () => Promise<void>, successMessage: string) => {
    setSaving(true);
    setError(null);
    setMessage(null);
    setWorkflowBlocked([]);
    try {
      await action();
      setMessage(successMessage);
      await loadData();
    } catch (err: unknown) {
      if (err instanceof APIRequestError) {
        setError(err.message);
        setWorkflowBlocked(err.requirementsMissing ?? []);
      } else {
        setError(err instanceof Error ? err.message : "Falha na operacao");
      }
    } finally {
      setSaving(false);
    }
  };

  if (!tripId) {
    return <section className="page">Trip id invalido.</section>;
  }

  return (
    <section className="page">
      <PageHeader
        title={`Operacao da viagem ${tripId.slice(0, 8)}`}
        subtitle="Fluxo operacional rigido do fretamento, com bloqueio por etapa."
        secondaryActions={<Link className="button secondary sm" to="/trips">Voltar para viagens</Link>}
      />

      {loading ? <div className="section">Carregando...</div> : null}
      {error ? <InlineAlert tone="error">{error}</InlineAlert> : null}
      {workflowBlocked.length > 0 ? (
        <InlineAlert tone="warning">Requisitos pendentes: {workflowBlocked.join(" | ")}</InlineAlert>
      ) : null}
      {message ? <InlineAlert tone="success">{message}</InlineAlert> : null}

      {trip ? (
        <>
          <div className="section">
            <div className="section-header">
              <div className="section-title">Resumo operacional</div>
            </div>
            <div className="summary-panel">
              <div className="summary-item">
                <span className="summary-label">Status de viagem</span>
                <span className="summary-value">{tripStatusLabel[trip.status] ?? trip.status}</span>
              </div>
              <div className="summary-item">
                <span className="summary-label">Status operacional</span>
                <span className="summary-value">{tripOperationalStatusLabel[trip.operational_status] ?? trip.operational_status}</span>
              </div>
              <div className="summary-item">
                <span className="summary-label">KM planejada</span>
                <span className="summary-value">{trip.estimated_km ?? 0} km</span>
              </div>
              <div className="summary-item">
                <span className="summary-label">Saida</span>
                <span className="summary-value">{new Date(trip.departure_at).toLocaleString()}</span>
              </div>
            </div>
            <Stepper steps={stepperSteps} />
            {nextStatus ? (
              <button className="button" type="button" disabled={saving} onClick={() => runAction(() => apiPost(`/trips/${tripId}/workflow/advance`, { to_status: nextStatus }), `Fluxo avancou para ${tripOperationalStatusLabel[nextStatus] ?? nextStatus}.`)}>
                Avancar para {tripOperationalStatusLabel[nextStatus] ?? nextStatus}
              </button>
            ) : (
              <StatusBadge tone="success">Fluxo concluido</StatusBadge>
            )}
          </div>

          <div className="section">
            <div className="section-header"><div className="section-title">Painel de conformidade</div></div>
            <div className="summary-panel">
              <div className="summary-item"><span className="summary-label">LV/SISAUT</span><span className="summary-value">{compliance.lvSisaut ? "OK" : "Pendente"}</span></div>
              <div className="summary-item"><span className="summary-label">TAF/SRC</span><span className="summary-value">{compliance.tafSrc ? "OK" : "Pendente"}</span></div>
              <div className="summary-item"><span className="summary-label">Tacografo</span><span className="summary-value">{compliance.tachograph ? "OK" : "Pendente"}</span></div>
              <div className="summary-item"><span className="summary-label">Comprovantes</span><span className="summary-value">{compliance.receipts ? "OK" : "Pendente"}</span></div>
              <div className="summary-item"><span className="summary-label">Fiscal (manual)</span><span className="summary-value">{compliance.fiscal ? "OK" : "Pendente"}</span></div>
            </div>
          </div>

          <div className="section">
            <div className="section-header"><div className="section-title">1. Solicitacao de viagem</div></div>
            <form className="form-grid" onSubmit={(event) => { event.preventDefault(); runAction(() => apiPost("/trip-requests", { source: requestForm.source, requester_name: requestForm.requester_name || undefined, requester_contact: requestForm.requester_contact || undefined, requested_departure_at: fromInputDateTime(requestForm.requested_departure_at), notes: requestForm.notes || undefined }), "Solicitacao registrada."); }}>
              <label>Fonte<select className="input" value={requestForm.source} onChange={(e) => setRequestForm((prev) => ({ ...prev, source: e.target.value }))}><option value="SYSTEM">SYSTEM</option><option value="EMAIL">EMAIL</option></select></label>
              <label>Solicitante<input className="input" value={requestForm.requester_name} onChange={(e) => setRequestForm((prev) => ({ ...prev, requester_name: e.target.value }))} /></label>
              <label>Contato<input className="input" value={requestForm.requester_contact} onChange={(e) => setRequestForm((prev) => ({ ...prev, requester_contact: e.target.value }))} /></label>
              <label>Saida prevista<input className="input" type="datetime-local" value={requestForm.requested_departure_at} onChange={(e) => setRequestForm((prev) => ({ ...prev, requested_departure_at: e.target.value }))} /></label>
              <label className="full-span">Observacoes<textarea className="input" value={requestForm.notes} onChange={(e) => setRequestForm((prev) => ({ ...prev, notes: e.target.value }))} /></label>
              <button className="button" type="submit" disabled={saving}>Registrar solicitacao</button>
            </form>
            <p className="page-subtitle">Solicitacoes recentes: {tripRequests.length}</p>
          </div>

          <div className="section">
            <div className="section-header"><div className="section-title">2. Manifesto de passageiros</div></div>
            <div className="form-actions"><button className="button secondary" type="button" disabled={saving} onClick={() => runAction(() => apiPost(`/trips/${tripId}/manifest/sync`, {}), "Manifesto sincronizado com reservas.")}>Sincronizar reservas</button></div>
            <form className="form-grid" onSubmit={(event) => { event.preventDefault(); runAction(() => apiPost(`/trips/${tripId}/manifest`, { ...manifestForm }), "Passageiro adicionado ao manifesto."); }}>
              <label>Nome<input className="input" required value={manifestForm.passenger_name} onChange={(e) => setManifestForm((prev) => ({ ...prev, passenger_name: e.target.value }))} /></label>
              <label>Documento<input className="input" value={manifestForm.passenger_document} onChange={(e) => setManifestForm((prev) => ({ ...prev, passenger_document: e.target.value }))} /></label>
              <label>Telefone<input className="input" value={manifestForm.passenger_phone} onChange={(e) => setManifestForm((prev) => ({ ...prev, passenger_phone: e.target.value }))} /></label>
              <button className="button" type="submit" disabled={saving}>Adicionar manualmente</button>
            </form>
            <p className="page-subtitle">Ativos: {manifest.filter((item) => item.is_active && item.status !== "CANCELLED").length}</p>
          </div>

          <div className="section">
            <div className="section-header"><div className="section-title">3. Autorizacoes ANTT/DETER</div></div>
            <form className="form-grid" onSubmit={(event) => { event.preventDefault(); runAction(() => apiPost(`/trips/${tripId}/authorizations`, { authority: authorizationForm.authority, status: authorizationForm.status, protocol_number: authorizationForm.protocol_number || undefined, license_number: authorizationForm.license_number || undefined, src_policy_number: authorizationForm.src_policy_number || undefined, src_valid_until: fromInputDateTime(authorizationForm.src_valid_until), exceptional_deadline_ok: authorizationForm.exceptional_deadline_ok, notes: authorizationForm.notes || undefined }), "Autorizacao registrada."); }}>
              <label>Autoridade<select className="input" value={authorizationForm.authority} onChange={(e) => setAuthorizationForm((prev) => ({ ...prev, authority: e.target.value }))}><option value="ANTT">ANTT</option><option value="DETER">DETER</option><option value="EXCEPTIONAL">EXCEPTIONAL</option></select></label>
              <label>Status<select className="input" value={authorizationForm.status} onChange={(e) => setAuthorizationForm((prev) => ({ ...prev, status: e.target.value }))}><option value="PENDING">PENDING</option><option value="ISSUED">ISSUED</option><option value="REJECTED">REJECTED</option><option value="EXPIRED">EXPIRED</option></select></label>
              <label><input className="input" placeholder="Protocolo" value={authorizationForm.protocol_number} onChange={(e) => setAuthorizationForm((prev) => ({ ...prev, protocol_number: e.target.value }))} /></label>
              <label><input className="input" placeholder="Licenca (LV)" value={authorizationForm.license_number} onChange={(e) => setAuthorizationForm((prev) => ({ ...prev, license_number: e.target.value }))} /></label>
              <label><input className="input" placeholder="Apolice SRC" value={authorizationForm.src_policy_number} onChange={(e) => setAuthorizationForm((prev) => ({ ...prev, src_policy_number: e.target.value }))} /></label>
              <label>SRC valido ate<input className="input" type="datetime-local" value={authorizationForm.src_valid_until} onChange={(e) => setAuthorizationForm((prev) => ({ ...prev, src_valid_until: e.target.value }))} /></label>
              <label className="checkbox"><input type="checkbox" checked={authorizationForm.exceptional_deadline_ok} onChange={(e) => setAuthorizationForm((prev) => ({ ...prev, exceptional_deadline_ok: e.target.checked }))} />Prazo minimo cumprido (excepcional)</label>
              <button className="button" type="submit" disabled={saving}>Salvar autorizacao</button>
            </form>
            <p className="page-subtitle">Registros: {authorizations.length}</p>
          </div>

          <div className="section">
            <div className="section-header"><div className="section-title">4. Checklist pre-partida</div></div>
            <ChecklistForm value={preChecklistForm} onChange={setPreChecklistForm} onSubmit={() => runAction(() => apiPut(`/trips/${tripId}/checklists/PRE_DEPARTURE`, { ...preChecklistForm, checklist_data: { generated_at: new Date().toISOString() } }), "Checklist de pre-partida salvo.")} disabled={saving} />
          </div>

          <div className="section">
            <div className="section-header"><div className="section-title">5. Relatorio do motorista</div></div>
            <form className="form-grid" onSubmit={(event) => { event.preventDefault(); runAction(() => apiPut(`/trips/${tripId}/driver-report`, { odometer_start: numberOrUndefined(driverReportForm.odometer_start), odometer_end: numberOrUndefined(driverReportForm.odometer_end), fuel_used_liters: floatOrUndefined(driverReportForm.fuel_used_liters), incidents: driverReportForm.incidents || undefined, delays: driverReportForm.delays || undefined, rest_hours: floatOrUndefined(driverReportForm.rest_hours), notes: driverReportForm.notes || undefined }), "Relatorio do motorista salvo."); }}>
              <label><input className="input" placeholder="Odometro inicial" value={driverReportForm.odometer_start} onChange={(e) => setDriverReportForm((prev) => ({ ...prev, odometer_start: e.target.value }))} /></label>
              <label><input className="input" placeholder="Odometro final" value={driverReportForm.odometer_end} onChange={(e) => setDriverReportForm((prev) => ({ ...prev, odometer_end: e.target.value }))} /></label>
              <label><input className="input" placeholder="Litros consumidos" value={driverReportForm.fuel_used_liters} onChange={(e) => setDriverReportForm((prev) => ({ ...prev, fuel_used_liters: e.target.value }))} /></label>
              <label><input className="input" placeholder="Horas de descanso" value={driverReportForm.rest_hours} onChange={(e) => setDriverReportForm((prev) => ({ ...prev, rest_hours: e.target.value }))} /></label>
              <label className="full-span"><textarea className="input" placeholder="Incidentes" value={driverReportForm.incidents} onChange={(e) => setDriverReportForm((prev) => ({ ...prev, incidents: e.target.value }))} /></label>
              <label className="full-span"><textarea className="input" placeholder="Atrasos" value={driverReportForm.delays} onChange={(e) => setDriverReportForm((prev) => ({ ...prev, delays: e.target.value }))} /></label>
              <label className="full-span"><textarea className="input" placeholder="Notas" value={driverReportForm.notes} onChange={(e) => setDriverReportForm((prev) => ({ ...prev, notes: e.target.value }))} /></label>
              <button className="button" type="submit" disabled={saving}>Salvar relatorio</button>
            </form>
          </div>

          <div className="section">
            <div className="section-header"><div className="section-title">6. Checklist de retorno</div></div>
            <ChecklistForm value={returnChecklistForm} onChange={setReturnChecklistForm} onSubmit={() => runAction(() => apiPut(`/trips/${tripId}/checklists/RETURN`, { ...returnChecklistForm, checklist_data: { generated_at: new Date().toISOString() } }), "Checklist de retorno salvo.")} disabled={saving} />
          </div>

          <div className="section">
            <div className="section-header"><div className="section-title">7. Reconciliacao de comprovantes</div></div>
            <div className="summary-panel">
              {expenses.map((expense) => (
                <label className="checkbox" key={expense.id}>
                  <input
                    type="checkbox"
                    checked={reconciliationForm.verified_expense_ids.includes(expense.id)}
                    onChange={(e) =>
                      setReconciliationForm((prev) => ({
                        ...prev,
                        verified_expense_ids: e.target.checked
                          ? [...prev.verified_expense_ids, expense.id]
                          : prev.verified_expense_ids.filter((item) => item !== expense.id),
                      }))
                    }
                  />
                  {expense.description} - R$ {expense.amount.toFixed(2)} {expense.receipt_verified ? "(ja validado)" : ""}
                </label>
              ))}
            </div>
            <form className="form-grid" onSubmit={(event) => { event.preventDefault(); runAction(() => apiPut(`/trips/${tripId}/reconciliation`, { total_receipts_amount: floatOrZero(reconciliationForm.total_receipts_amount), receipts_validated: reconciliationForm.receipts_validated, verified_expense_ids: reconciliationForm.verified_expense_ids, notes: reconciliationForm.notes || undefined }), "Reconciliacao atualizada."); }}>
              <label>Total comprovantes<input className="input" value={reconciliationForm.total_receipts_amount} onChange={(e) => setReconciliationForm((prev) => ({ ...prev, total_receipts_amount: e.target.value }))} /></label>
              <label className="checkbox"><input type="checkbox" checked={reconciliationForm.receipts_validated} onChange={(e) => setReconciliationForm((prev) => ({ ...prev, receipts_validated: e.target.checked }))} />Comprovantes validados</label>
              <label className="full-span">Notas<textarea className="input" value={reconciliationForm.notes} onChange={(e) => setReconciliationForm((prev) => ({ ...prev, notes: e.target.value }))} /></label>
              <button className="button" type="submit" disabled={saving}>Salvar reconciliacao</button>
            </form>
          </div>

          <div className="section">
            <div className="section-header"><div className="section-title">8. Anexos</div></div>
            <form className="form-grid" onSubmit={(event) => { event.preventDefault(); runAction(async () => { if (!attachmentForm.file) throw new Error("Selecione um arquivo"); const client = getSupabaseClient(); if (!client) throw new Error("Supabase nao configurado para upload"); const file = attachmentForm.file; const storagePath = `${tripId}/${Date.now()}-${file.name.replace(/\s+/g, "-")}`; const upload = await client.storage.from("trip-documents").upload(storagePath, file, { upsert: false }); if (upload.error) throw upload.error; await apiPost(`/trips/${tripId}/attachments`, { attachment_type: attachmentForm.attachment_type, storage_bucket: "trip-documents", storage_path: storagePath, file_name: file.name, mime_type: file.type || undefined, file_size: file.size, metadata: { uploaded_via: "trip_operations_ui" } }); setAttachmentForm((prev) => ({ ...prev, file: null })); }, "Anexo enviado e registrado."); }}>
              <label>Tipo<select className="input" value={attachmentForm.attachment_type} onChange={(e) => setAttachmentForm((prev) => ({ ...prev, attachment_type: e.target.value }))}><option value="TRIP_REQUEST">TRIP_REQUEST</option><option value="AUTHORIZATION">AUTHORIZATION</option><option value="INSURANCE">INSURANCE</option><option value="CHECKLIST">CHECKLIST</option><option value="TACHOGRAPH">TACHOGRAPH</option><option value="RECEIPT">RECEIPT</option><option value="FISCAL">FISCAL</option><option value="DRIVER_REPORT">DRIVER_REPORT</option><option value="OTHER">OTHER</option></select></label>
              <label>Arquivo<input className="input" type="file" onChange={(e) => setAttachmentForm((prev) => ({ ...prev, file: e.target.files?.[0] ?? null }))} /></label>
              <button className="button" type="submit" disabled={saving}>Upload</button>
            </form>
            <p className="page-subtitle">Anexos registrados: {attachments.length}</p>
          </div>

          <div className="section">
            <div className="section-header"><div className="section-title">9. Registro fiscal manual (CT-e OS / NF)</div></div>
            <form className="form-grid" onSubmit={(event) => { event.preventDefault(); runAction(() => apiPost(`/fiscal-documents`, { trip_id: tripId, document_type: "CTE_OS", document_number: `MAN-${Date.now()}`, amount: 0, status: "ISSUED" }), "Registro fiscal manual criado."); }}>
              <button className="button" type="submit" disabled={saving}>Criar registro fiscal manual</button>
            </form>
            <p className="page-subtitle">Documentos fiscais vinculados: {fiscalDocs.length}</p>
          </div>
        </>
      ) : null}
    </section>
  );
}

function numberOrUndefined(value: string) {
  const trimmed = value.trim();
  if (!trimmed) return undefined;
  const parsed = Number(trimmed);
  return Number.isFinite(parsed) ? parsed : undefined;
}

function floatOrUndefined(value: string) {
  return numberOrUndefined(value);
}

function floatOrZero(value: string) {
  const parsed = numberOrUndefined(value);
  return parsed ?? 0;
}

type ChecklistFormValue = {
  is_complete: boolean;
  documents_checked: boolean;
  tachograph_checked: boolean;
  receipts_checked: boolean;
  rest_compliance_ok: boolean;
  notes: string;
};

function ChecklistForm({
  value,
  onChange,
  onSubmit,
  disabled,
}: {
  value: ChecklistFormValue;
  onChange: (next: ChecklistFormValue) => void;
  onSubmit: () => void;
  disabled: boolean;
}) {
  return (
    <form className="form-grid" onSubmit={(event) => { event.preventDefault(); onSubmit(); }}>
      <label className="checkbox"><input type="checkbox" checked={value.documents_checked} onChange={(e) => onChange({ ...value, documents_checked: e.target.checked })} />Documentacao conferida</label>
      <label className="checkbox"><input type="checkbox" checked={value.tachograph_checked} onChange={(e) => onChange({ ...value, tachograph_checked: e.target.checked })} />Tacografo conferido</label>
      <label className="checkbox"><input type="checkbox" checked={value.receipts_checked} onChange={(e) => onChange({ ...value, receipts_checked: e.target.checked })} />Comprovantes conferidos</label>
      <label className="checkbox"><input type="checkbox" checked={value.rest_compliance_ok} onChange={(e) => onChange({ ...value, rest_compliance_ok: e.target.checked })} />Jornada/descanso conforme</label>
      <label className="checkbox"><input type="checkbox" checked={value.is_complete} onChange={(e) => onChange({ ...value, is_complete: e.target.checked })} />Checklist completo</label>
      <label className="full-span">Notas<textarea className="input" value={value.notes} onChange={(e) => onChange({ ...value, notes: e.target.value })} /></label>
      <button className="button" type="submit" disabled={disabled}>Salvar checklist</button>
    </form>
  );
}
