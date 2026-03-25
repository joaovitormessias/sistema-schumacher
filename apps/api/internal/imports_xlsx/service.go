package imports_xlsx

import "context"

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Upload(ctx context.Context, input UploadInput) (UploadResult, error) {
	batchID, status, err := s.repo.CreateBatch(ctx, input.SourceFileName)
	if err != nil {
		return UploadResult{}, err
	}

	counts, err := s.repo.InsertSheets(ctx, batchID, input.Sheets)
	if err != nil {
		return UploadResult{}, err
	}

	return UploadResult{
		BatchID:      batchID,
		Status:       status,
		SheetInserts: counts,
	}, nil
}

func (s *Service) Validate(ctx context.Context, batchID string) (ValidateResult, error) {
	bySheet, err := s.repo.ValidateBatch(ctx, batchID)
	if err != nil {
		return ValidateResult{}, err
	}

	errors, total, _, err := s.repo.GetErrors(ctx, batchID)
	if err != nil {
		return ValidateResult{}, err
	}

	status, err := s.repo.GetBatchStatus(ctx, batchID)
	if err != nil {
		return ValidateResult{}, err
	}

	return ValidateResult{
		BatchID:    batchID,
		Status:     status,
		ErrorCount: total,
		BySheet:    bySheet,
		Errors:     errors,
	}, nil
}

func (s *Service) Promote(ctx context.Context, batchID string) (PromoteResult, error) {
	if err := s.repo.PromoteBatch(ctx, batchID); err != nil {
		return PromoteResult{}, err
	}

	status, err := s.repo.GetBatchStatus(ctx, batchID)
	if err != nil {
		return PromoteResult{}, err
	}

	return PromoteResult{BatchID: batchID, Status: status}, nil
}

func (s *Service) Report(ctx context.Context, batchID string) (BatchReport, error) {
	return s.repo.GetBatchReport(ctx, batchID)
}
