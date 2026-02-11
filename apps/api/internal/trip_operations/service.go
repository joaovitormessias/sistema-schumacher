package trip_operations

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
)

var (
	ErrInvalidSource            = errors.New("invalid source")
	ErrInvalidStatus            = errors.New("invalid status")
	ErrInvalidStage             = errors.New("invalid stage")
	ErrInvalidAuthority         = errors.New("invalid authority")
	ErrInvalidAuthorization     = errors.New("invalid authorization status")
	ErrInvalidAttachmentType    = errors.New("invalid attachment type")
	ErrInvalidOperationalStatus = errors.New("invalid operational status")
)

type WorkflowBlockedError struct {
	Response WorkflowBlockedResponse
}

func (e WorkflowBlockedError) Error() string {
	return e.Response.Message
}

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListTripRequests(ctx context.Context, limit int, offset int) ([]TripRequest, error) {
	return s.repo.ListTripRequests(ctx, limit, offset)
}

func (s *Service) CreateTripRequest(ctx context.Context, input CreateTripRequestInput, createdBy *string) (TripRequest, error) {
	if !inSet(input.Source, "EMAIL", "SYSTEM") {
		return TripRequest{}, ErrInvalidSource
	}
	if input.Status != nil && !inSet(*input.Status, "OPEN", "IN_REVIEW", "APPROVED", "REJECTED") {
		return TripRequest{}, ErrInvalidStatus
	}
	return s.repo.CreateTripRequest(ctx, input, createdBy)
}

func (s *Service) ListManifestEntries(ctx context.Context, tripID uuid.UUID) ([]TripManifestEntry, error) {
	return s.repo.ListManifestEntries(ctx, tripID)
}

func (s *Service) CreateManifestEntry(ctx context.Context, tripID uuid.UUID, input CreateManifestEntryInput) (TripManifestEntry, error) {
	if strings.TrimSpace(input.PassengerName) == "" {
		return TripManifestEntry{}, errors.New("passenger_name is required")
	}
	if input.Status != nil && !inSet(*input.Status, "EXPECTED", "BOARDED", "NO_SHOW", "CANCELLED") {
		return TripManifestEntry{}, ErrInvalidStatus
	}
	return s.repo.CreateManifestEntry(ctx, tripID, input)
}

func (s *Service) UpdateManifestEntry(ctx context.Context, tripID uuid.UUID, entryID uuid.UUID, input UpdateManifestEntryInput) (TripManifestEntry, error) {
	if input.Status != nil && !inSet(*input.Status, "EXPECTED", "BOARDED", "NO_SHOW", "CANCELLED") {
		return TripManifestEntry{}, ErrInvalidStatus
	}
	return s.repo.UpdateManifestEntry(ctx, tripID, entryID, input)
}

func (s *Service) SyncManifestFromBookings(ctx context.Context, tripID uuid.UUID) ([]TripManifestEntry, error) {
	return s.repo.SyncManifestFromBookings(ctx, tripID)
}

func (s *Service) ListTripAuthorizations(ctx context.Context, tripID uuid.UUID) ([]TripAuthorization, error) {
	return s.repo.ListTripAuthorizations(ctx, tripID)
}

func (s *Service) CreateTripAuthorization(ctx context.Context, tripID uuid.UUID, input CreateTripAuthorizationInput, createdBy *string) (TripAuthorization, error) {
	if !inSet(input.Authority, "ANTT", "DETER", "EXCEPTIONAL") {
		return TripAuthorization{}, ErrInvalidAuthority
	}
	if !inSet(input.Status, "PENDING", "ISSUED", "REJECTED", "EXPIRED") {
		return TripAuthorization{}, ErrInvalidAuthorization
	}
	return s.repo.CreateTripAuthorization(ctx, tripID, input, createdBy)
}

func (s *Service) UpdateTripAuthorization(ctx context.Context, tripID uuid.UUID, authorizationID uuid.UUID, input UpdateTripAuthorizationInput) (TripAuthorization, error) {
	if input.Authority != nil && !inSet(*input.Authority, "ANTT", "DETER", "EXCEPTIONAL") {
		return TripAuthorization{}, ErrInvalidAuthority
	}
	if input.Status != nil && !inSet(*input.Status, "PENDING", "ISSUED", "REJECTED", "EXPIRED") {
		return TripAuthorization{}, ErrInvalidAuthorization
	}
	return s.repo.UpdateTripAuthorization(ctx, tripID, authorizationID, input)
}

func (s *Service) GetTripChecklist(ctx context.Context, tripID uuid.UUID, stage string) (TripChecklist, error) {
	if !inSet(stage, "PRE_DEPARTURE", "RETURN") {
		return TripChecklist{}, ErrInvalidStage
	}
	return s.repo.GetTripChecklist(ctx, tripID, stage)
}

func (s *Service) UpsertTripChecklist(ctx context.Context, tripID uuid.UUID, stage string, input UpsertChecklistInput, completedBy *string) (TripChecklist, error) {
	if !inSet(stage, "PRE_DEPARTURE", "RETURN") {
		return TripChecklist{}, ErrInvalidStage
	}
	return s.repo.UpsertTripChecklist(ctx, tripID, stage, input, completedBy)
}

func (s *Service) GetTripDriverReport(ctx context.Context, tripID uuid.UUID) (TripDriverReport, error) {
	return s.repo.GetTripDriverReport(ctx, tripID)
}

func (s *Service) UpsertTripDriverReport(ctx context.Context, tripID uuid.UUID, input UpsertDriverReportInput, submittedBy *string) (TripDriverReport, error) {
	return s.repo.UpsertTripDriverReport(ctx, tripID, input, submittedBy)
}

func (s *Service) GetTripReceiptReconciliation(ctx context.Context, tripID uuid.UUID) (TripReceiptReconciliation, error) {
	return s.repo.GetTripReceiptReconciliation(ctx, tripID)
}

func (s *Service) UpsertTripReceiptReconciliation(ctx context.Context, tripID uuid.UUID, input UpsertReceiptReconciliationInput, reconciledBy *string) (TripReceiptReconciliation, error) {
	if input.TotalReceiptsAmount < 0 {
		return TripReceiptReconciliation{}, errors.New("total_receipts_amount must be >= 0")
	}
	return s.repo.UpsertTripReceiptReconciliation(ctx, tripID, input, reconciledBy)
}

func (s *Service) ListTripAttachments(ctx context.Context, tripID uuid.UUID) ([]TripAttachment, error) {
	return s.repo.ListTripAttachments(ctx, tripID)
}

func (s *Service) CreateTripAttachment(ctx context.Context, tripID uuid.UUID, input CreateAttachmentInput, uploadedBy *string) (TripAttachment, error) {
	if strings.TrimSpace(input.AttachmentType) == "" {
		return TripAttachment{}, ErrInvalidAttachmentType
	}
	if strings.TrimSpace(input.StoragePath) == "" || strings.TrimSpace(input.FileName) == "" {
		return TripAttachment{}, errors.New("storage_path and file_name are required")
	}
	return s.repo.CreateTripAttachment(ctx, tripID, input, uploadedBy)
}

func (s *Service) AdvanceWorkflow(ctx context.Context, tripID uuid.UUID, toStatus string, actorID *string) (OperationalTripState, error) {
	if _, ok := ValidOperationalStatuses[toStatus]; !ok {
		return OperationalTripState{}, ErrInvalidOperationalStatus
	}

	state, err := s.repo.GetOperationalTripState(ctx, tripID)
	if err != nil {
		return OperationalTripState{}, err
	}

	if state.OperationalStatus == toStatus {
		return state, nil
	}

	if !isAllowedTransition(state.OperationalStatus, toStatus) {
		return OperationalTripState{}, WorkflowBlockedError{
			Response: WorkflowBlockedResponse{
				Code:                "WORKFLOW_TRANSITION_INVALID",
				Message:             "transicao operacional fora da ordem",
				RequirementsMissing: []string{"seguir ordem sequencial do workflow"},
			},
		}
	}

	if toStatus == TripOperationalStatusDispatchValidated {
		if err := s.repo.MarkDispatchValidated(ctx, tripID, actorID); err != nil {
			return OperationalTripState{}, err
		}
		state, err = s.repo.GetOperationalTripState(ctx, tripID)
		if err != nil {
			return OperationalTripState{}, err
		}
	}

	validation, err := s.validateTransitionRequirements(ctx, state, toStatus)
	if err != nil {
		return OperationalTripState{}, err
	}
	if !validation.Allowed {
		return OperationalTripState{}, WorkflowBlockedError{Response: WorkflowBlockedResponse{
			Code:                validation.Code,
			Message:             validation.Message,
			RequirementsMissing: validation.RequirementsMissing,
		}}
	}

	return s.repo.SetOperationalStatus(ctx, tripID, toStatus)
}

func (s *Service) validateTransitionRequirements(ctx context.Context, state OperationalTripState, toStatus string) (WorkflowValidation, error) {
	switch toStatus {
	case TripOperationalStatusPassengersReady:
		count, err := s.repo.CountActiveManifestEntries(ctx, uuid.MustParse(state.ID))
		if err != nil {
			return WorkflowValidation{}, err
		}
		if count == 0 {
			return blocked("WORKFLOW_REQUIREMENTS_MISSING", "manifesto de passageiros obrigatorio", "manifesto com pelo menos 1 passageiro ativo"), nil
		}
	case TripOperationalStatusItineraryReady:
		count, err := s.repo.CountTripStops(ctx, uuid.MustParse(state.ID))
		if err != nil {
			return WorkflowValidation{}, err
		}
		missing := []string{}
		if count == 0 {
			missing = append(missing, "roteiro com paradas")
		}
		if state.EstimatedKM <= 0 {
			missing = append(missing, "estimated_km > 0")
		}
		if len(missing) > 0 {
			return blocked("WORKFLOW_REQUIREMENTS_MISSING", "roteiro incompleto", missing...), nil
		}
	case TripOperationalStatusDispatchValidated:
		if state.DispatchValidatedAt == nil {
			return blocked("WORKFLOW_REQUIREMENTS_MISSING", "validacao D-1 obrigatoria", "dispatch_validated_at"), nil
		}
	case TripOperationalStatusAuthorized:
		issued, srcValid, exceptionalOK, err := s.repo.HasIssuedAuthorizationAndValidSRC(ctx, uuid.MustParse(state.ID), state.DepartureAt)
		if err != nil {
			return WorkflowValidation{}, err
		}
		missing := []string{}
		if !issued {
			missing = append(missing, "autorizacao emitida (LV/SISAUT ou DETER)")
		}
		if !srcValid {
			missing = append(missing, "seguro SRC valido na data da viagem")
		}
		if !exceptionalOK {
			missing = append(missing, "autorizacao excepcional dentro do prazo minimo")
		}
		if len(missing) > 0 {
			return blocked("WORKFLOW_REQUIREMENTS_MISSING", "autorizacao regulatoria incompleta", missing...), nil
		}
	case TripOperationalStatusInProgress:
		checklist, exists, err := s.repo.GetChecklistCompliance(ctx, uuid.MustParse(state.ID), "PRE_DEPARTURE")
		if err != nil {
			return WorkflowValidation{}, err
		}
		if !exists || !checklist.IsComplete {
			return blocked("WORKFLOW_REQUIREMENTS_MISSING", "checklist de pre-partida obrigatorio", "trip_checklists[PRE_DEPARTURE].is_complete=true"), nil
		}
	case TripOperationalStatusReturned:
		hasReport, err := s.repo.HasDriverReport(ctx, uuid.MustParse(state.ID))
		if err != nil {
			return WorkflowValidation{}, err
		}
		if !hasReport {
			return blocked("WORKFLOW_REQUIREMENTS_MISSING", "relatorio do motorista obrigatorio", "trip_driver_reports"), nil
		}
	case TripOperationalStatusReturnChecked:
		checklist, exists, err := s.repo.GetChecklistCompliance(ctx, uuid.MustParse(state.ID), "RETURN")
		if err != nil {
			return WorkflowValidation{}, err
		}
		reconciliation, recExists, err := s.repo.GetReconciliationCompliance(ctx, uuid.MustParse(state.ID))
		if err != nil {
			return WorkflowValidation{}, err
		}
		missing := []string{}
		if !exists || !checklist.IsComplete {
			missing = append(missing, "checklist retorno completo")
		}
		if !exists || !checklist.DocumentsChecked {
			missing = append(missing, "documentacao conferida")
		}
		if !exists || !checklist.TachographChecked {
			missing = append(missing, "tacografo conferido")
		}
		if !exists || !checklist.ReceiptsChecked {
			missing = append(missing, "comprovantes conferidos")
		}
		if !recExists || !reconciliation.ReceiptsValidated {
			missing = append(missing, "reconciliacao de comprovantes validada")
		}
		if len(missing) > 0 {
			return blocked("WORKFLOW_REQUIREMENTS_MISSING", "retorno sem conferencia completa", missing...), nil
		}
	case TripOperationalStatusSettled:
		settled, err := s.repo.IsTripSettlementCompleted(ctx, uuid.MustParse(state.ID))
		if err != nil {
			return WorkflowValidation{}, err
		}
		if !settled {
			return blocked("WORKFLOW_REQUIREMENTS_MISSING", "acerto financeiro pendente", "trip_settlements.status=COMPLETED"), nil
		}
	case TripOperationalStatusClosed:
		fiscalOK, err := s.repo.HasFiscalComplianceRecord(ctx, uuid.MustParse(state.ID))
		if err != nil {
			return WorkflowValidation{}, err
		}
		if !fiscalOK {
			return blocked("WORKFLOW_REQUIREMENTS_MISSING", "compliance fiscal manual obrigatorio", "registro fiscal manual (CT-e OS/NF)"), nil
		}
	}

	return WorkflowValidation{Allowed: true}, nil
}

func blocked(code string, message string, requirements ...string) WorkflowValidation {
	return WorkflowValidation{
		Allowed:             false,
		Code:                code,
		Message:             message,
		RequirementsMissing: requirements,
	}
}

func isAllowedTransition(from string, to string) bool {
	allowed := map[string]string{
		TripOperationalStatusRequested:         TripOperationalStatusPassengersReady,
		TripOperationalStatusPassengersReady:   TripOperationalStatusItineraryReady,
		TripOperationalStatusItineraryReady:    TripOperationalStatusDispatchValidated,
		TripOperationalStatusDispatchValidated: TripOperationalStatusAuthorized,
		TripOperationalStatusAuthorized:        TripOperationalStatusInProgress,
		TripOperationalStatusInProgress:        TripOperationalStatusReturned,
		TripOperationalStatusReturned:          TripOperationalStatusReturnChecked,
		TripOperationalStatusReturnChecked:     TripOperationalStatusSettled,
		TripOperationalStatusSettled:           TripOperationalStatusClosed,
	}
	return allowed[from] == to
}

func inSet(value string, options ...string) bool {
	for _, item := range options {
		if value == item {
			return true
		}
	}
	return false
}
