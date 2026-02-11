package trip_operations

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
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

func (r *Repository) ListTripRequests(ctx context.Context, limit int, offset int) ([]TripRequest, error) {
	if limit <= 0 || limit > 500 {
		limit = 200
	}
	if offset < 0 {
		offset = 0
	}
	rows, err := r.pool.Query(ctx, `
    select id, route_id, source, status, requester_name, requester_contact, requested_departure_at, notes, created_by, created_at, updated_at
    from trip_requests
    order by created_at desc
    limit $1 offset $2
  `, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []TripRequest{}
	for rows.Next() {
		var item TripRequest
		if err := rows.Scan(&item.ID, &item.RouteID, &item.Source, &item.Status, &item.RequesterName, &item.RequesterContact, &item.RequestedDepartureAt, &item.Notes, &item.CreatedBy, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) CreateTripRequest(ctx context.Context, input CreateTripRequestInput, createdBy *string) (TripRequest, error) {
	status := "OPEN"
	if input.Status != nil && *input.Status != "" {
		status = *input.Status
	}
	var item TripRequest
	row := r.pool.QueryRow(ctx, `
    insert into trip_requests (route_id, source, status, requester_name, requester_contact, requested_departure_at, notes, created_by, updated_at)
    values ($1,$2,$3,$4,$5,$6,$7,$8,now())
    returning id, route_id, source, status, requester_name, requester_contact, requested_departure_at, notes, created_by, created_at, updated_at
  `, nullableUUID(input.RouteID), input.Source, status, nullableText(input.RequesterName), nullableText(input.RequesterContact), input.RequestedDepartureAt, nullableText(input.Notes), nullableUUID(createdBy))
	if err := row.Scan(&item.ID, &item.RouteID, &item.Source, &item.Status, &item.RequesterName, &item.RequesterContact, &item.RequestedDepartureAt, &item.Notes, &item.CreatedBy, &item.CreatedAt, &item.UpdatedAt); err != nil {
		return TripRequest{}, err
	}
	return item, nil
}

func (r *Repository) ListManifestEntries(ctx context.Context, tripID uuid.UUID) ([]TripManifestEntry, error) {
	rows, err := r.pool.Query(ctx, `
    select id, trip_id, booking_passenger_id, passenger_name, passenger_document, passenger_phone, source, status, seat_number, is_active, created_at, updated_at
    from trip_manifest_entries
    where trip_id=$1
    order by created_at asc
  `, tripID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []TripManifestEntry{}
	for rows.Next() {
		var item TripManifestEntry
		if err := rowscanManifest(rows, &item); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) CreateManifestEntry(ctx context.Context, tripID uuid.UUID, input CreateManifestEntryInput) (TripManifestEntry, error) {
	status := "EXPECTED"
	if input.Status != nil && *input.Status != "" {
		status = *input.Status
	}

	var item TripManifestEntry
	row := r.pool.QueryRow(ctx, `
    insert into trip_manifest_entries (trip_id, booking_passenger_id, passenger_name, passenger_document, passenger_phone, source, status, seat_number, is_active, updated_at)
    values ($1,$2,$3,$4,$5,$6,$7,$8,true,now())
    returning id, trip_id, booking_passenger_id, passenger_name, passenger_document, passenger_phone, source, status, seat_number, is_active, created_at, updated_at
  `, tripID, nullableUUID(input.BookingPassengerID), input.PassengerName, nullableText(input.PassengerDocument), nullableText(input.PassengerPhone), manifestSource(input.BookingPassengerID), status, input.SeatNumber)
	if err := rowscanManifest(row, &item); err != nil {
		return TripManifestEntry{}, err
	}
	return item, nil
}

func (r *Repository) UpdateManifestEntry(ctx context.Context, tripID uuid.UUID, entryID uuid.UUID, input UpdateManifestEntryInput) (TripManifestEntry, error) {
	sets := []string{}
	args := []interface{}{}
	idx := 1

	if input.PassengerName != nil {
		sets = append(sets, fmt.Sprintf("passenger_name=$%d", idx))
		args = append(args, *input.PassengerName)
		idx++
	}
	if input.PassengerDocument != nil {
		sets = append(sets, fmt.Sprintf("passenger_document=$%d", idx))
		args = append(args, nullableText(input.PassengerDocument))
		idx++
	}
	if input.PassengerPhone != nil {
		sets = append(sets, fmt.Sprintf("passenger_phone=$%d", idx))
		args = append(args, nullableText(input.PassengerPhone))
		idx++
	}
	if input.Status != nil {
		sets = append(sets, fmt.Sprintf("status=$%d", idx))
		args = append(args, *input.Status)
		idx++
	}
	if input.SeatNumber != nil {
		sets = append(sets, fmt.Sprintf("seat_number=$%d", idx))
		args = append(args, input.SeatNumber)
		idx++
	}
	if input.IsActive != nil {
		sets = append(sets, fmt.Sprintf("is_active=$%d", idx))
		args = append(args, *input.IsActive)
		idx++
	}
	if len(sets) == 0 {
		return r.GetManifestEntry(ctx, tripID, entryID)
	}
	sets = append(sets, "updated_at=now()")
	args = append(args, tripID)
	tripArg := idx
	idx++
	args = append(args, entryID)
	entryArg := idx

	var item TripManifestEntry
	row := r.pool.QueryRow(ctx, fmt.Sprintf(`
    update trip_manifest_entries
    set %s
    where trip_id=$%d and id=$%d
    returning id, trip_id, booking_passenger_id, passenger_name, passenger_document, passenger_phone, source, status, seat_number, is_active, created_at, updated_at
  `, strings.Join(sets, ", "), tripArg, entryArg), args...)
	if err := rowscanManifest(row, &item); err != nil {
		return TripManifestEntry{}, err
	}
	return item, nil
}

func (r *Repository) GetManifestEntry(ctx context.Context, tripID uuid.UUID, entryID uuid.UUID) (TripManifestEntry, error) {
	var item TripManifestEntry
	row := r.pool.QueryRow(ctx, `
    select id, trip_id, booking_passenger_id, passenger_name, passenger_document, passenger_phone, source, status, seat_number, is_active, created_at, updated_at
    from trip_manifest_entries
    where trip_id=$1 and id=$2
  `, tripID, entryID)
	if err := rowscanManifest(row, &item); err != nil {
		return TripManifestEntry{}, err
	}
	return item, nil
}

func (r *Repository) SyncManifestFromBookings(ctx context.Context, tripID uuid.UUID) ([]TripManifestEntry, error) {
	_, err := r.pool.Exec(ctx, `
    insert into trip_manifest_entries (
      trip_id,
      booking_passenger_id,
      passenger_name,
      passenger_document,
      passenger_phone,
      source,
      status,
      seat_number,
      is_active,
      created_at,
      updated_at
    )
    select
      bp.trip_id,
      bp.id,
      bp.name,
      bp.document,
      bp.phone,
      'BOOKING'::manifest_entry_source,
      case bp.status
        when 'BOARDED' then 'BOARDED'::manifest_passenger_status
        when 'NO_SHOW' then 'NO_SHOW'::manifest_passenger_status
        when 'CANCELLED' then 'CANCELLED'::manifest_passenger_status
        else 'EXPECTED'::manifest_passenger_status
      end,
      bs.seat_number,
      bp.is_active,
      bp.created_at,
      now()
    from booking_passengers bp
    left join bus_seats bs on bs.id = bp.seat_id
    where bp.trip_id = $1 and bp.is_active = true
    on conflict (trip_id, booking_passenger_id)
    do update set
      passenger_name = excluded.passenger_name,
      passenger_document = excluded.passenger_document,
      passenger_phone = excluded.passenger_phone,
      status = excluded.status,
      seat_number = excluded.seat_number,
      is_active = excluded.is_active,
      updated_at = now()
  `, tripID)
	if err != nil {
		return nil, err
	}
	return r.ListManifestEntries(ctx, tripID)
}

func (r *Repository) ListTripAuthorizations(ctx context.Context, tripID uuid.UUID) ([]TripAuthorization, error) {
	rows, err := r.pool.Query(ctx, `
    select id, trip_id, authority, status, protocol_number, license_number, issued_at, valid_until,
      src_policy_number, src_valid_until::timestamptz, exceptional_deadline_ok, attachment_id, notes,
      created_by, created_at, updated_at
    from trip_authorizations
    where trip_id=$1
    order by created_at desc
  `, tripID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []TripAuthorization{}
	for rows.Next() {
		var item TripAuthorization
		if err := rowscanAuthorization(rows, &item); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) CreateTripAuthorization(ctx context.Context, tripID uuid.UUID, input CreateTripAuthorizationInput, createdBy *string) (TripAuthorization, error) {
	exceptionalDeadlineOK := true
	if input.ExceptionalDeadlineOK != nil {
		exceptionalDeadlineOK = *input.ExceptionalDeadlineOK
	}
	var item TripAuthorization
	row := r.pool.QueryRow(ctx, `
    insert into trip_authorizations (
      trip_id, authority, status, protocol_number, license_number, issued_at, valid_until,
      src_policy_number, src_valid_until, exceptional_deadline_ok, attachment_id, notes, created_by, updated_at
    )
    values ($1,$2,$3,$4,$5,$6,$7,$8,$9::date,$10,$11,$12,$13,now())
    returning id, trip_id, authority, status, protocol_number, license_number, issued_at, valid_until,
      src_policy_number, src_valid_until::timestamptz, exceptional_deadline_ok, attachment_id, notes,
      created_by, created_at, updated_at
  `, tripID, input.Authority, input.Status, nullableText(input.ProtocolNumber), nullableText(input.LicenseNumber), input.IssuedAt, input.ValidUntil, nullableText(input.SRCPolicyNumber), toDateTime(input.SRCValidUntil), exceptionalDeadlineOK, nullableUUID(input.AttachmentID), nullableText(input.Notes), nullableUUID(createdBy))
	if err := rowscanAuthorization(row, &item); err != nil {
		return TripAuthorization{}, err
	}
	return item, nil
}

func (r *Repository) UpdateTripAuthorization(ctx context.Context, tripID uuid.UUID, authorizationID uuid.UUID, input UpdateTripAuthorizationInput) (TripAuthorization, error) {
	sets := []string{}
	args := []interface{}{}
	idx := 1

	if input.Authority != nil {
		sets = append(sets, fmt.Sprintf("authority=$%d", idx))
		args = append(args, *input.Authority)
		idx++
	}
	if input.Status != nil {
		sets = append(sets, fmt.Sprintf("status=$%d", idx))
		args = append(args, *input.Status)
		idx++
	}
	if input.ProtocolNumber != nil {
		sets = append(sets, fmt.Sprintf("protocol_number=$%d", idx))
		args = append(args, nullableText(input.ProtocolNumber))
		idx++
	}
	if input.LicenseNumber != nil {
		sets = append(sets, fmt.Sprintf("license_number=$%d", idx))
		args = append(args, nullableText(input.LicenseNumber))
		idx++
	}
	if input.IssuedAt != nil {
		sets = append(sets, fmt.Sprintf("issued_at=$%d", idx))
		args = append(args, input.IssuedAt)
		idx++
	}
	if input.ValidUntil != nil {
		sets = append(sets, fmt.Sprintf("valid_until=$%d", idx))
		args = append(args, input.ValidUntil)
		idx++
	}
	if input.SRCPolicyNumber != nil {
		sets = append(sets, fmt.Sprintf("src_policy_number=$%d", idx))
		args = append(args, nullableText(input.SRCPolicyNumber))
		idx++
	}
	if input.SRCValidUntil != nil {
		sets = append(sets, fmt.Sprintf("src_valid_until=$%d::date", idx))
		args = append(args, toDateTime(input.SRCValidUntil))
		idx++
	}
	if input.ExceptionalDeadlineOK != nil {
		sets = append(sets, fmt.Sprintf("exceptional_deadline_ok=$%d", idx))
		args = append(args, *input.ExceptionalDeadlineOK)
		idx++
	}
	if input.AttachmentID != nil {
		sets = append(sets, fmt.Sprintf("attachment_id=$%d", idx))
		args = append(args, nullableUUID(input.AttachmentID))
		idx++
	}
	if input.Notes != nil {
		sets = append(sets, fmt.Sprintf("notes=$%d", idx))
		args = append(args, nullableText(input.Notes))
		idx++
	}
	if len(sets) == 0 {
		return r.GetTripAuthorization(ctx, tripID, authorizationID)
	}
	sets = append(sets, "updated_at=now()")
	args = append(args, tripID)
	tripArg := idx
	idx++
	args = append(args, authorizationID)
	authArg := idx

	var item TripAuthorization
	row := r.pool.QueryRow(ctx, fmt.Sprintf(`
    update trip_authorizations
    set %s
    where trip_id=$%d and id=$%d
    returning id, trip_id, authority, status, protocol_number, license_number, issued_at, valid_until,
      src_policy_number, src_valid_until::timestamptz, exceptional_deadline_ok, attachment_id, notes,
      created_by, created_at, updated_at
  `, strings.Join(sets, ", "), tripArg, authArg), args...)
	if err := rowscanAuthorization(row, &item); err != nil {
		return TripAuthorization{}, err
	}
	return item, nil
}

func (r *Repository) GetTripAuthorization(ctx context.Context, tripID uuid.UUID, authorizationID uuid.UUID) (TripAuthorization, error) {
	var item TripAuthorization
	row := r.pool.QueryRow(ctx, `
    select id, trip_id, authority, status, protocol_number, license_number, issued_at, valid_until,
      src_policy_number, src_valid_until::timestamptz, exceptional_deadline_ok, attachment_id, notes,
      created_by, created_at, updated_at
    from trip_authorizations
    where trip_id=$1 and id=$2
  `, tripID, authorizationID)
	if err := rowscanAuthorization(row, &item); err != nil {
		return TripAuthorization{}, err
	}
	return item, nil
}

func (r *Repository) GetTripChecklist(ctx context.Context, tripID uuid.UUID, stage string) (TripChecklist, error) {
	var item TripChecklist
	row := r.pool.QueryRow(ctx, `
    select id, trip_id, stage, checklist_data, is_complete, documents_checked, tachograph_checked,
      receipts_checked, rest_compliance_ok, completed_by, completed_at, notes, created_at, updated_at
    from trip_checklists
    where trip_id=$1 and stage=$2
  `, tripID, stage)
	if err := rowscanChecklist(row, &item); err != nil {
		return TripChecklist{}, err
	}
	return item, nil
}

func (r *Repository) UpsertTripChecklist(ctx context.Context, tripID uuid.UUID, stage string, input UpsertChecklistInput, completedBy *string) (TripChecklist, error) {
	restComplianceOK := true
	if input.RestComplianceOK != nil {
		restComplianceOK = *input.RestComplianceOK
	}
	checklistData := input.ChecklistData
	if len(checklistData) == 0 {
		checklistData = json.RawMessage("{}")
	}

	var item TripChecklist
	row := r.pool.QueryRow(ctx, `
    insert into trip_checklists (
      trip_id, stage, checklist_data, is_complete, documents_checked, tachograph_checked,
      receipts_checked, rest_compliance_ok, completed_by, completed_at, notes, updated_at
    )
    values ($1,$2,$3,$4,$5,$6,$7,$8,$9,case when $4 then now() else null end,$10,now())
    on conflict (trip_id, stage)
    do update set
      checklist_data=excluded.checklist_data,
      is_complete=excluded.is_complete,
      documents_checked=excluded.documents_checked,
      tachograph_checked=excluded.tachograph_checked,
      receipts_checked=excluded.receipts_checked,
      rest_compliance_ok=excluded.rest_compliance_ok,
      completed_by=excluded.completed_by,
      completed_at=case when excluded.is_complete then now() else null end,
      notes=excluded.notes,
      updated_at=now()
    returning id, trip_id, stage, checklist_data, is_complete, documents_checked, tachograph_checked,
      receipts_checked, rest_compliance_ok, completed_by, completed_at, notes, created_at, updated_at
  `, tripID, stage, checklistData, input.IsComplete, input.DocumentsChecked, input.TachographChecked, input.ReceiptsChecked, restComplianceOK, nullableUUID(completedBy), nullableText(input.Notes))
	if err := rowscanChecklist(row, &item); err != nil {
		return TripChecklist{}, err
	}
	return item, nil
}

func (r *Repository) GetTripDriverReport(ctx context.Context, tripID uuid.UUID) (TripDriverReport, error) {
	var item TripDriverReport
	row := r.pool.QueryRow(ctx, `
    select id, trip_id, driver_id, odometer_start, odometer_end, fuel_used_liters, incidents,
      delays, rest_hours, notes, submitted_by, submitted_at, created_at, updated_at
    from trip_driver_reports
    where trip_id=$1
  `, tripID)
	if err := rowscanDriverReport(row, &item); err != nil {
		return TripDriverReport{}, err
	}
	return item, nil
}

func (r *Repository) UpsertTripDriverReport(ctx context.Context, tripID uuid.UUID, input UpsertDriverReportInput, submittedBy *string) (TripDriverReport, error) {
	var item TripDriverReport
	row := r.pool.QueryRow(ctx, `
    insert into trip_driver_reports (
      trip_id, driver_id, odometer_start, odometer_end, fuel_used_liters,
      incidents, delays, rest_hours, notes, submitted_by, submitted_at, updated_at
    )
    values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,now(),now())
    on conflict (trip_id)
    do update set
      driver_id=excluded.driver_id,
      odometer_start=excluded.odometer_start,
      odometer_end=excluded.odometer_end,
      fuel_used_liters=excluded.fuel_used_liters,
      incidents=excluded.incidents,
      delays=excluded.delays,
      rest_hours=excluded.rest_hours,
      notes=excluded.notes,
      submitted_by=excluded.submitted_by,
      submitted_at=now(),
      updated_at=now()
    returning id, trip_id, driver_id, odometer_start, odometer_end, fuel_used_liters,
      incidents, delays, rest_hours, notes, submitted_by, submitted_at, created_at, updated_at
  `, tripID, nullableUUID(input.DriverID), input.OdometerStart, input.OdometerEnd, input.FuelUsedLiters, nullableText(input.Incidents), nullableText(input.Delays), input.RestHours, nullableText(input.Notes), nullableUUID(submittedBy))
	if err := rowscanDriverReport(row, &item); err != nil {
		return TripDriverReport{}, err
	}
	return item, nil
}

func (r *Repository) GetTripReceiptReconciliation(ctx context.Context, tripID uuid.UUID) (TripReceiptReconciliation, error) {
	var item TripReceiptReconciliation
	row := r.pool.QueryRow(ctx, `
    select id, trip_id, total_receipts_amount, total_approved_expenses, difference,
      receipts_validated,
      coalesce(array(select value::text from unnest(verified_expense_ids) as value), '{}'::text[]),
      notes, reconciled_by, reconciled_at, created_at, updated_at
    from trip_receipt_reconciliations
    where trip_id=$1
  `, tripID)
	if err := rowscanReconciliation(row, &item); err != nil {
		return TripReceiptReconciliation{}, err
	}
	return item, nil
}

func (r *Repository) UpsertTripReceiptReconciliation(ctx context.Context, tripID uuid.UUID, input UpsertReceiptReconciliationInput, reconciledBy *string) (TripReceiptReconciliation, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return TripReceiptReconciliation{}, err
	}
	defer tx.Rollback(ctx)

	verifiedUUIDs, verifiedIDs, err := parseUUIDList(input.VerifiedExpenseIDs)
	if err != nil {
		return TripReceiptReconciliation{}, err
	}

	if _, err := tx.Exec(ctx, `
    update trip_expenses
    set receipt_verified=false,
        receipt_verified_by=$2,
        receipt_verified_at=now(),
        receipt_verification_notes='updated by reconciliation'
    where trip_id=$1
  `, tripID, nullableUUID(reconciledBy)); err != nil {
		return TripReceiptReconciliation{}, err
	}

	if len(verifiedUUIDs) > 0 {
		rows, err := tx.Query(ctx, `
      select id::text
      from trip_expenses
      where trip_id=$1 and is_approved=true and id = any($2::uuid[])
    `, tripID, verifiedUUIDs)
		if err != nil {
			return TripReceiptReconciliation{}, err
		}
		defer rows.Close()

		matched := map[string]struct{}{}
		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err != nil {
				return TripReceiptReconciliation{}, err
			}
			matched[id] = struct{}{}
		}
		if err := rows.Err(); err != nil {
			return TripReceiptReconciliation{}, err
		}

		for _, id := range verifiedIDs {
			if _, ok := matched[id]; !ok {
				return TripReceiptReconciliation{}, fmt.Errorf("expense %s is not an approved expense for this trip", id)
			}
		}

		if _, err := tx.Exec(ctx, `
      update trip_expenses
      set receipt_verified=true,
          receipt_verified_by=$2,
          receipt_verified_at=now(),
          receipt_verification_notes='validated by reconciliation'
      where trip_id=$1 and is_approved=true and id = any($3::uuid[])
    `, tripID, nullableUUID(reconciledBy), verifiedUUIDs); err != nil {
			return TripReceiptReconciliation{}, err
		}
	}

	var totalApproved float64
	if err := tx.QueryRow(ctx, `
    select coalesce(sum(amount), 0)
    from trip_expenses
    where trip_id=$1 and is_approved=true and receipt_verified=true
  `, tripID).Scan(&totalApproved); err != nil {
		return TripReceiptReconciliation{}, err
	}

	var item TripReceiptReconciliation
	row := tx.QueryRow(ctx, `
    insert into trip_receipt_reconciliations (
      trip_id,
      total_receipts_amount,
      total_approved_expenses,
      receipts_validated,
      verified_expense_ids,
      notes,
      reconciled_by,
      reconciled_at,
      updated_at
    )
    values ($1,$2,$3,$4,$5::uuid[],$6,$7,now(),now())
    on conflict (trip_id)
    do update set
      total_receipts_amount=excluded.total_receipts_amount,
      total_approved_expenses=excluded.total_approved_expenses,
      receipts_validated=excluded.receipts_validated,
      verified_expense_ids=excluded.verified_expense_ids,
      notes=excluded.notes,
      reconciled_by=excluded.reconciled_by,
      reconciled_at=now(),
      updated_at=now()
    returning id, trip_id, total_receipts_amount, total_approved_expenses, difference,
      receipts_validated,
      coalesce(array(select value::text from unnest(verified_expense_ids) as value), '{}'::text[]),
      notes, reconciled_by, reconciled_at, created_at, updated_at
  `, tripID, input.TotalReceiptsAmount, totalApproved, input.ReceiptsValidated, verifiedUUIDs, nullableText(input.Notes), nullableUUID(reconciledBy))
	if err := rowscanReconciliation(row, &item); err != nil {
		return TripReceiptReconciliation{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return TripReceiptReconciliation{}, err
	}

	return item, nil
}

func (r *Repository) ListTripAttachments(ctx context.Context, tripID uuid.UUID) ([]TripAttachment, error) {
	rows, err := r.pool.Query(ctx, `
    select id, trip_id, attachment_type, storage_bucket, storage_path, file_name,
      mime_type, file_size, metadata, uploaded_by, uploaded_at, created_at
    from trip_attachments
    where trip_id=$1
    order by created_at desc
  `, tripID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []TripAttachment{}
	for rows.Next() {
		var item TripAttachment
		if err := rowscanAttachment(rows, &item); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) CreateTripAttachment(ctx context.Context, tripID uuid.UUID, input CreateAttachmentInput, uploadedBy *string) (TripAttachment, error) {
	bucket := "trip-documents"
	if input.StorageBucket != nil && *input.StorageBucket != "" {
		bucket = *input.StorageBucket
	}
	metadata := input.Metadata
	if len(metadata) == 0 {
		metadata = json.RawMessage("{}")
	}
	var item TripAttachment
	row := r.pool.QueryRow(ctx, `
    insert into trip_attachments (
      trip_id, attachment_type, storage_bucket, storage_path, file_name,
      mime_type, file_size, metadata, uploaded_by, uploaded_at
    )
    values ($1,$2,$3,$4,$5,$6,$7,$8,$9,now())
    returning id, trip_id, attachment_type, storage_bucket, storage_path, file_name,
      mime_type, file_size, metadata, uploaded_by, uploaded_at, created_at
  `, tripID, input.AttachmentType, bucket, input.StoragePath, input.FileName, nullableText(input.MimeType), input.FileSize, metadata, nullableUUID(uploadedBy))
	if err := rowscanAttachment(row, &item); err != nil {
		return TripAttachment{}, err
	}
	return item, nil
}

func (r *Repository) GetOperationalTripState(ctx context.Context, tripID uuid.UUID) (OperationalTripState, error) {
	var item OperationalTripState
	row := r.pool.QueryRow(ctx, `
    select id, status, operational_status, departure_at, estimated_km, dispatch_validated_at, dispatch_validated_by
    from trips
    where id=$1
  `, tripID)
	if err := row.Scan(&item.ID, &item.Status, &item.OperationalStatus, &item.DepartureAt, &item.EstimatedKM, &item.DispatchValidatedAt, &item.DispatchValidatedBy); err != nil {
		return OperationalTripState{}, err
	}
	return item, nil
}

func (r *Repository) MarkDispatchValidated(ctx context.Context, tripID uuid.UUID, validatedBy *string) error {
	_, err := r.pool.Exec(ctx, `
    update trips
    set dispatch_validated_at=now(),
        dispatch_validated_by=$2,
        updated_at=now()
    where id=$1
  `, tripID, nullableUUID(validatedBy))
	return err
}

func (r *Repository) SetOperationalStatus(ctx context.Context, tripID uuid.UUID, status string) (OperationalTripState, error) {
	var item OperationalTripState
	row := r.pool.QueryRow(ctx, `
    update trips
    set operational_status=$2,
        status=case
          when $2 = 'IN_PROGRESS' then 'IN_PROGRESS'::trip_status
          when $2 in ('SETTLED', 'CLOSED') then 'COMPLETED'::trip_status
          else status
        end,
        updated_at=now()
    where id=$1
    returning id, status, operational_status, departure_at, estimated_km, dispatch_validated_at, dispatch_validated_by
  `, tripID, status)
	if err := row.Scan(&item.ID, &item.Status, &item.OperationalStatus, &item.DepartureAt, &item.EstimatedKM, &item.DispatchValidatedAt, &item.DispatchValidatedBy); err != nil {
		return OperationalTripState{}, err
	}
	return item, nil
}

func (r *Repository) CountActiveManifestEntries(ctx context.Context, tripID uuid.UUID) (int, error) {
	var count int
	if err := r.pool.QueryRow(ctx, `
    select count(*)
    from trip_manifest_entries
    where trip_id=$1 and is_active=true and status != 'CANCELLED'
  `, tripID).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *Repository) CountTripStops(ctx context.Context, tripID uuid.UUID) (int, error) {
	var count int
	if err := r.pool.QueryRow(ctx, `select count(*) from trip_stops where trip_id=$1`, tripID).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *Repository) HasIssuedAuthorizationAndValidSRC(ctx context.Context, tripID uuid.UUID, departureAt time.Time) (bool, bool, bool, error) {
	var issued bool
	if err := r.pool.QueryRow(ctx, `
    select exists(
      select 1 from trip_authorizations
      where trip_id=$1 and status='ISSUED'
    )
  `, tripID).Scan(&issued); err != nil {
		return false, false, false, err
	}

	var srcValid bool
	if err := r.pool.QueryRow(ctx, `
    select exists(
      select 1
      from trip_authorizations
      where trip_id=$1
        and status='ISSUED'
        and src_policy_number is not null
        and src_valid_until is not null
        and src_valid_until >= $2::date
    )
  `, tripID, departureAt).Scan(&srcValid); err != nil {
		return false, false, false, err
	}

	var exceptionalOK bool
	if err := r.pool.QueryRow(ctx, `
    select coalesce(bool_and(exceptional_deadline_ok), true)
    from trip_authorizations
    where trip_id=$1 and authority='EXCEPTIONAL'
  `, tripID).Scan(&exceptionalOK); err != nil {
		return false, false, false, err
	}

	return issued, srcValid, exceptionalOK, nil
}

func (r *Repository) GetChecklistCompliance(ctx context.Context, tripID uuid.UUID, stage string) (TripChecklist, bool, error) {
	item, err := r.GetTripChecklist(ctx, tripID, stage)
	if err != nil {
		if err == pgx.ErrNoRows {
			return TripChecklist{}, false, nil
		}
		return TripChecklist{}, false, err
	}
	return item, true, nil
}

func (r *Repository) HasDriverReport(ctx context.Context, tripID uuid.UUID) (bool, error) {
	var exists bool
	if err := r.pool.QueryRow(ctx, `select exists(select 1 from trip_driver_reports where trip_id=$1)`, tripID).Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

func (r *Repository) GetReconciliationCompliance(ctx context.Context, tripID uuid.UUID) (TripReceiptReconciliation, bool, error) {
	item, err := r.GetTripReceiptReconciliation(ctx, tripID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return TripReceiptReconciliation{}, false, nil
		}
		return TripReceiptReconciliation{}, false, err
	}
	return item, true, nil
}

func (r *Repository) IsTripSettlementCompleted(ctx context.Context, tripID uuid.UUID) (bool, error) {
	var exists bool
	if err := r.pool.QueryRow(ctx, `
    select exists(
      select 1 from trip_settlements
      where trip_id=$1 and status='COMPLETED'
    )
  `, tripID).Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

func (r *Repository) HasFiscalComplianceRecord(ctx context.Context, tripID uuid.UUID) (bool, error) {
	var exists bool
	if err := r.pool.QueryRow(ctx, `
    select exists(
      select 1
      from fiscal_documents
      where trip_id=$1
        and document_type in ('CTE_OS','CTE','NF')
        and upper(status) not in ('PENDING','REJECTED','CANCELLED')
    )
  `, tripID).Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

func rowscanManifest(scanner interface {
	Scan(dest ...interface{}) error
}, item *TripManifestEntry) error {
	return scanner.Scan(&item.ID, &item.TripID, &item.BookingPassengerID, &item.PassengerName, &item.PassengerDocument, &item.PassengerPhone, &item.Source, &item.Status, &item.SeatNumber, &item.IsActive, &item.CreatedAt, &item.UpdatedAt)
}

func rowscanAuthorization(scanner interface {
	Scan(dest ...interface{}) error
}, item *TripAuthorization) error {
	return scanner.Scan(&item.ID, &item.TripID, &item.Authority, &item.Status, &item.ProtocolNumber, &item.LicenseNumber, &item.IssuedAt, &item.ValidUntil, &item.SRCPolicyNumber, &item.SRCValidUntil, &item.ExceptionalDeadlineOK, &item.AttachmentID, &item.Notes, &item.CreatedBy, &item.CreatedAt, &item.UpdatedAt)
}

func rowscanChecklist(scanner interface {
	Scan(dest ...interface{}) error
}, item *TripChecklist) error {
	return scanner.Scan(&item.ID, &item.TripID, &item.Stage, &item.ChecklistData, &item.IsComplete, &item.DocumentsChecked, &item.TachographChecked, &item.ReceiptsChecked, &item.RestComplianceOK, &item.CompletedBy, &item.CompletedAt, &item.Notes, &item.CreatedAt, &item.UpdatedAt)
}

func rowscanDriverReport(scanner interface {
	Scan(dest ...interface{}) error
}, item *TripDriverReport) error {
	return scanner.Scan(&item.ID, &item.TripID, &item.DriverID, &item.OdometerStart, &item.OdometerEnd, &item.FuelUsedLiters, &item.Incidents, &item.Delays, &item.RestHours, &item.Notes, &item.SubmittedBy, &item.SubmittedAt, &item.CreatedAt, &item.UpdatedAt)
}

func rowscanReconciliation(scanner interface {
	Scan(dest ...interface{}) error
}, item *TripReceiptReconciliation) error {
	return scanner.Scan(&item.ID, &item.TripID, &item.TotalReceiptsAmount, &item.TotalApprovedExpenses, &item.Difference, &item.ReceiptsValidated, &item.VerifiedExpenseIDs, &item.Notes, &item.ReconciledBy, &item.ReconciledAt, &item.CreatedAt, &item.UpdatedAt)
}

func rowscanAttachment(scanner interface {
	Scan(dest ...interface{}) error
}, item *TripAttachment) error {
	return scanner.Scan(&item.ID, &item.TripID, &item.AttachmentType, &item.StorageBucket, &item.StoragePath, &item.FileName, &item.MimeType, &item.FileSize, &item.Metadata, &item.UploadedBy, &item.UploadedAt, &item.CreatedAt)
}

func parseUUIDList(ids []string) ([]uuid.UUID, []string, error) {
	if len(ids) == 0 {
		return []uuid.UUID{}, []string{}, nil
	}
	unique := map[uuid.UUID]struct{}{}
	for _, id := range ids {
		trimmed := strings.TrimSpace(id)
		if trimmed == "" {
			continue
		}
		parsed, err := uuid.Parse(trimmed)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid expense id: %s", trimmed)
		}
		unique[parsed] = struct{}{}
	}
	uuids := make([]uuid.UUID, 0, len(unique))
	stringsList := make([]string, 0, len(unique))
	for id := range unique {
		uuids = append(uuids, id)
		stringsList = append(stringsList, id.String())
	}
	sort.Slice(uuids, func(i, j int) bool {
		return uuids[i].String() < uuids[j].String()
	})
	sort.Strings(stringsList)
	return uuids, stringsList, nil
}

func manifestSource(bookingPassengerID *string) string {
	if bookingPassengerID == nil || *bookingPassengerID == "" {
		return "MANUAL"
	}
	return "BOOKING"
}

func toDateTime(v *time.Time) *time.Time {
	if v == nil {
		return nil
	}
	value := time.Date(v.Year(), v.Month(), v.Day(), 0, 0, 0, 0, time.UTC)
	return &value
}

func nullableUUID(v *string) interface{} {
	if v == nil || strings.TrimSpace(*v) == "" {
		return nil
	}
	return strings.TrimSpace(*v)
}

func nullableText(v *string) interface{} {
	if v == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*v)
	if trimmed == "" {
		return nil
	}
	return trimmed
}

func IsNotFound(err error) bool {
	return err == pgx.ErrNoRows
}
