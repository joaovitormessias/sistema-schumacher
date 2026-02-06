package fiscal_documents

import (
  "encoding/json"
  "time"
)

type FiscalDocument struct {
  ID                string          `json:"id"`
  TripID            string          `json:"trip_id"`
  DocumentType      string          `json:"document_type"`
  DocumentNumber    *string         `json:"document_number"`
  IssueDate         time.Time       `json:"issue_date"`
  Amount            float64         `json:"amount"`
  RecipientName     *string         `json:"recipient_name"`
  RecipientDocument *string         `json:"recipient_document"`
  Status            string          `json:"status"`
  ExternalID        *string         `json:"external_id"`
  Metadata          json.RawMessage `json:"metadata"`
  CreatedBy         *string         `json:"created_by"`
  CreatedAt         time.Time       `json:"created_at"`
}

type CreateFiscalDocumentInput struct {
  TripID            string          `json:"trip_id"`
  DocumentType      string          `json:"document_type"`
  DocumentNumber    *string         `json:"document_number"`
  IssueDate         *time.Time      `json:"issue_date"`
  Amount            float64         `json:"amount"`
  RecipientName     *string         `json:"recipient_name"`
  RecipientDocument *string         `json:"recipient_document"`
  Status            *string         `json:"status"`
  ExternalID        *string         `json:"external_id"`
  Metadata          json.RawMessage `json:"metadata"`
}

type UpdateFiscalDocumentInput struct {
  DocumentNumber    *string         `json:"document_number"`
  IssueDate         *time.Time      `json:"issue_date"`
  Amount            *float64        `json:"amount"`
  RecipientName     *string         `json:"recipient_name"`
  RecipientDocument *string         `json:"recipient_document"`
  Status            *string         `json:"status"`
  ExternalID        *string         `json:"external_id"`
  Metadata          json.RawMessage `json:"metadata"`
}

type ListFilter struct {
  TripID       string
  DocumentType string
  Status       string
  Limit        int
  Offset       int
}
