package imports_xlsx

import "time"

type UploadInput struct {
	SourceFileName string                   `json:"source_file_name"`
	Sheets         map[string][]interface{} `json:"sheets"`
}

type UploadResult struct {
	BatchID      string         `json:"batch_id"`
	Status       string         `json:"status"`
	SheetInserts map[string]int `json:"sheet_inserts"`
}

type ValidateResult struct {
	BatchID    string           `json:"batch_id"`
	Status     string           `json:"status"`
	ErrorCount int64            `json:"error_count"`
	BySheet    map[string]int64 `json:"by_sheet"`
	Errors     []ImportError    `json:"errors,omitempty"`
}

type PromoteResult struct {
	BatchID string `json:"batch_id"`
	Status  string `json:"status"`
}

type BatchReport struct {
	BatchID        string                 `json:"batch_id"`
	Status         string                 `json:"status"`
	SourceFile     *string                `json:"source_file_name,omitempty"`
	UploadedAt     time.Time              `json:"uploaded_at"`
	ValidatedAt    *time.Time             `json:"validated_at,omitempty"`
	PromotedAt     *time.Time             `json:"promoted_at,omitempty"`
	ErrorCount     int64                  `json:"error_count"`
	ErrorsBySheet  map[string]int64       `json:"errors_by_sheet"`
	Reconciliation map[string]SheetHash   `json:"reconciliation"`
	Report         map[string]interface{} `json:"report"`
}

type SheetHash struct {
	Count int64  `json:"count"`
	Hash  string `json:"hash_md5"`
}

type ImportError struct {
	SheetName    string                 `json:"sheet_name"`
	RowNumber    *int                   `json:"row_number,omitempty"`
	ErrorCode    string                 `json:"error_code"`
	ErrorMessage string                 `json:"error_message"`
	RowData      map[string]interface{} `json:"row_data,omitempty"`
}
