package fiscal_documents

import (
  "context"
  "fmt"
  "strings"
  "time"

  "github.com/google/uuid"
  "github.com/jackc/pgx/v5"
  "github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
  pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
  return &Repository{pool: pool}
}

func (r *Repository) List(ctx context.Context, filter ListFilter) ([]FiscalDocument, error) {
  query := `select id, trip_id, document_type, document_number, issue_date, amount, recipient_name, recipient_document, status, external_id, metadata, created_by, created_at from fiscal_documents`
  args := []interface{}{}
  clauses := []string{}

  if filter.TripID != "" {
    args = append(args, filter.TripID)
    clauses = append(clauses, fmt.Sprintf("trip_id=$%d", len(args)))
  }
  if filter.DocumentType != "" {
    args = append(args, filter.DocumentType)
    clauses = append(clauses, fmt.Sprintf("document_type=$%d", len(args)))
  }
  if filter.Status != "" {
    args = append(args, filter.Status)
    clauses = append(clauses, fmt.Sprintf("status=$%d", len(args)))
  }

  if len(clauses) > 0 {
    query += " where " + strings.Join(clauses, " and ")
  }

  query += " order by issue_date desc"

  limit := filter.Limit
  if limit <= 0 || limit > 500 {
    limit = 200
  }
  args = append(args, limit)
  query += fmt.Sprintf(" limit $%d", len(args))

  if filter.Offset > 0 {
    args = append(args, filter.Offset)
    query += fmt.Sprintf(" offset $%d", len(args))
  }

  rows, err := r.pool.Query(ctx, query, args...)
  if err != nil {
    return nil, err
  }
  defer rows.Close()

  items := []FiscalDocument{}
  for rows.Next() {
    var item FiscalDocument
    if err := rows.Scan(&item.ID, &item.TripID, &item.DocumentType, &item.DocumentNumber, &item.IssueDate, &item.Amount, &item.RecipientName, &item.RecipientDocument, &item.Status, &item.ExternalID, &item.Metadata, &item.CreatedBy, &item.CreatedAt); err != nil {
      return nil, err
    }
    items = append(items, item)
  }
  return items, rows.Err()
}

func (r *Repository) Get(ctx context.Context, id uuid.UUID) (FiscalDocument, error) {
  var item FiscalDocument
  row := r.pool.QueryRow(ctx, `select id, trip_id, document_type, document_number, issue_date, amount, recipient_name, recipient_document, status, external_id, metadata, created_by, created_at from fiscal_documents where id=$1`, id)
  if err := row.Scan(&item.ID, &item.TripID, &item.DocumentType, &item.DocumentNumber, &item.IssueDate, &item.Amount, &item.RecipientName, &item.RecipientDocument, &item.Status, &item.ExternalID, &item.Metadata, &item.CreatedBy, &item.CreatedAt); err != nil {
    return item, err
  }
  return item, nil
}

func (r *Repository) Create(ctx context.Context, input CreateFiscalDocumentInput, createdBy *string) (FiscalDocument, error) {
  issueDate := time.Now()
  if input.IssueDate != nil {
    issueDate = *input.IssueDate
  }
  status := "PENDING"
  if input.Status != nil && *input.Status != "" {
    status = *input.Status
  }

  var item FiscalDocument
  row := r.pool.QueryRow(ctx,
    `insert into fiscal_documents (trip_id, document_type, document_number, issue_date, amount, recipient_name, recipient_document, status, external_id, metadata, created_by)
     values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
     returning id, trip_id, document_type, document_number, issue_date, amount, recipient_name, recipient_document, status, external_id, metadata, created_by, created_at`,
    input.TripID, input.DocumentType, input.DocumentNumber, issueDate, input.Amount, input.RecipientName, input.RecipientDocument, status, input.ExternalID, input.Metadata, createdBy,
  )
  if err := row.Scan(&item.ID, &item.TripID, &item.DocumentType, &item.DocumentNumber, &item.IssueDate, &item.Amount, &item.RecipientName, &item.RecipientDocument, &item.Status, &item.ExternalID, &item.Metadata, &item.CreatedBy, &item.CreatedAt); err != nil {
    return item, err
  }
  return item, nil
}

func (r *Repository) Update(ctx context.Context, id uuid.UUID, input UpdateFiscalDocumentInput) (FiscalDocument, error) {
  sets := []string{}
  args := []interface{}{}
  idx := 1

  if input.DocumentNumber != nil {
    sets = append(sets, fmt.Sprintf("document_number=$%d", idx))
    args = append(args, *input.DocumentNumber)
    idx++
  }
  if input.IssueDate != nil {
    sets = append(sets, fmt.Sprintf("issue_date=$%d", idx))
    args = append(args, *input.IssueDate)
    idx++
  }
  if input.Amount != nil {
    sets = append(sets, fmt.Sprintf("amount=$%d", idx))
    args = append(args, *input.Amount)
    idx++
  }
  if input.RecipientName != nil {
    sets = append(sets, fmt.Sprintf("recipient_name=$%d", idx))
    args = append(args, *input.RecipientName)
    idx++
  }
  if input.RecipientDocument != nil {
    sets = append(sets, fmt.Sprintf("recipient_document=$%d", idx))
    args = append(args, *input.RecipientDocument)
    idx++
  }
  if input.Status != nil {
    sets = append(sets, fmt.Sprintf("status=$%d", idx))
    args = append(args, *input.Status)
    idx++
  }
  if input.ExternalID != nil {
    sets = append(sets, fmt.Sprintf("external_id=$%d", idx))
    args = append(args, *input.ExternalID)
    idx++
  }
  if input.Metadata != nil {
    sets = append(sets, fmt.Sprintf("metadata=$%d", idx))
    args = append(args, input.Metadata)
    idx++
  }

  if len(sets) == 0 {
    return r.Get(ctx, id)
  }

  args = append(args, id)
  query := fmt.Sprintf(`update fiscal_documents set %s where id=$%d returning id, trip_id, document_type, document_number, issue_date, amount, recipient_name, recipient_document, status, external_id, metadata, created_by, created_at`, strings.Join(sets, ", "), idx)

  var item FiscalDocument
  row := r.pool.QueryRow(ctx, query, args...)
  if err := row.Scan(&item.ID, &item.TripID, &item.DocumentType, &item.DocumentNumber, &item.IssueDate, &item.Amount, &item.RecipientName, &item.RecipientDocument, &item.Status, &item.ExternalID, &item.Metadata, &item.CreatedBy, &item.CreatedAt); err != nil {
    return item, err
  }
  return item, nil
}

func IsNotFound(err error) bool {
  return err == pgx.ErrNoRows
}
