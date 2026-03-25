package imports_xlsx

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) CreateBatch(ctx context.Context, sourceFileName string) (string, string, error) {
	var id, status string
	err := r.pool.QueryRow(ctx, `
    insert into xlsx_import_batches (source_file_name, status)
    values ($1, 'UPLOADED')
    returning id::text, status
  `, nullString(sourceFileName)).Scan(&id, &status)
	return id, status, err
}

func (r *Repository) InsertSheets(ctx context.Context, batchID string, sheets map[string][]interface{}) (map[string]int, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	counts := map[string]int{}
	for name, rows := range sheets {
		count, err := r.insertSheetRows(ctx, tx, batchID, name, rows)
		if err != nil {
			return nil, err
		}
		counts[name] = count
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return counts, nil
}

func (r *Repository) insertSheetRows(ctx context.Context, tx pgx.Tx, batchID string, sheetName string, rows []interface{}) (int, error) {
	if len(rows) == 0 {
		return 0, nil
	}

	payload, err := json.Marshal(rows)
	if err != nil {
		return 0, err
	}

	query := insertQueryBySheet(sheetName)
	if query == "" {
		return 0, fmt.Errorf("unsupported sheet: %s", sheetName)
	}

	tag, err := tx.Exec(ctx, query, batchID, payload)
	if err != nil {
		return 0, err
	}
	return int(tag.RowsAffected()), nil
}

func insertQueryBySheet(sheetName string) string {
	switch sheetName {
	case "routes":
		return `
      insert into stg_xlsx_routes(batch_id, row_number, route_id, route_name, active, raw_row)
      select $1,
        e.ordinality::int,
        nullif(trim(e.elem->>'route_id'), ''),
        nullif(trim(e.elem->>'route_name'), ''),
        case when lower(coalesce(e.elem->>'active', '')) in ('true','t','1','yes','sim') then true else false end,
        e.elem
      from jsonb_array_elements($2::jsonb) with ordinality as e(elem, ordinality)
    `
	case "stops":
		return `
      insert into stg_xlsx_stops(batch_id, row_number, stop_id, stop_name, state, display_name, active, route_id, stop_order, eta_offset_minutes, raw_row)
      select $1,
        e.ordinality::int,
        nullif(trim(e.elem->>'stop_id'), ''),
        nullif(trim(e.elem->>'stop_name'), ''),
        nullif(trim(e.elem->>'state'), ''),
        nullif(trim(e.elem->>'display_name'), ''),
        case when lower(coalesce(e.elem->>'active', '')) in ('true','t','1','yes','sim') then true else false end,
        nullif(trim(e.elem->>'route_id'), ''),
        case when nullif(trim(e.elem->>'seq'), '') is not null then (e.elem->>'seq')::int else null end,
        case when nullif(trim(e.elem->>'eta_offset_minutes'), '') is not null then (e.elem->>'eta_offset_minutes')::int else null end,
        e.elem
      from jsonb_array_elements($2::jsonb) with ordinality as e(elem, ordinality)
    `
	case "trips":
		return `
      insert into stg_xlsx_trips(batch_id, row_number, trip_id, route_id, trip_date, bus_id, price_default, seats_total, seats_available, duration_hours, status, package, raw_row)
      select $1,
        e.ordinality::int,
        nullif(trim(e.elem->>'trip_id'), ''),
        nullif(trim(e.elem->>'route_id'), ''),
        case when nullif(trim(e.elem->>'trip_date'), '') is not null then (e.elem->>'trip_date')::date else null end,
        nullif(trim(e.elem->>'bus_id'), ''),
        case when nullif(trim(e.elem->>'price_default'), '') is not null then (e.elem->>'price_default')::numeric else null end,
        case when nullif(trim(e.elem->>'seats_total'), '') is not null then (e.elem->>'seats_total')::int else null end,
        case when nullif(trim(e.elem->>'seats_available'), '') is not null then (e.elem->>'seats_available')::int else null end,
        case when nullif(trim(e.elem->>'duration_hours'), '') is not null then (e.elem->>'duration_hours')::numeric else null end,
        nullif(trim(e.elem->>'status'), ''),
        nullif(trim(e.elem->>'package'), ''),
        e.elem
      from jsonb_array_elements($2::jsonb) with ordinality as e(elem, ordinality)
    `
	case "trip_stops":
		return `
      insert into stg_xlsx_trip_stops(batch_id, row_number, trip_stop_id, trip_id, stop_id, seq, depart_time, active, raw_row)
      select $1,
        e.ordinality::int,
        nullif(trim(e.elem->>'trip_stop_id'), ''),
        nullif(trim(e.elem->>'trip_id'), ''),
        nullif(trim(e.elem->>'stop_id'), ''),
        case when nullif(trim(e.elem->>'seq'), '') is not null then (e.elem->>'seq')::int else null end,
        nullif(trim(e.elem->>'depart_time'), ''),
        case when lower(coalesce(e.elem->>'active', '')) in ('true','t','1','yes','sim') then true else false end,
        e.elem
      from jsonb_array_elements($2::jsonb) with ordinality as e(elem, ordinality)
    `
	case "available_segments":
		return `
      insert into stg_xlsx_available_segments(batch_id, row_number, segment_id, trip_id, origin_stop_id, origin_display, origin_depart_time, dest_stop_id, dest_display, trip_date, price, seats_available, status, route_id, package, raw_row)
      select $1,
        e.ordinality::int,
        nullif(trim(e.elem->>'segment_id'), ''),
        nullif(trim(e.elem->>'trip_id'), ''),
        nullif(trim(e.elem->>'origin_stop_id'), ''),
        nullif(trim(e.elem->>'origin_display'), ''),
        nullif(trim(e.elem->>'origin_depart_time'), ''),
        nullif(trim(e.elem->>'dest_stop_id'), ''),
        nullif(trim(e.elem->>'dest_display'), ''),
        case when nullif(trim(e.elem->>'trip_date'), '') is not null then (e.elem->>'trip_date')::date else null end,
        case when nullif(trim(e.elem->>'price'), '') is not null then (e.elem->>'price')::numeric else null end,
        case when nullif(trim(e.elem->>'seats_available'), '') is not null then (e.elem->>'seats_available')::int else null end,
        nullif(trim(e.elem->>'status'), ''),
        nullif(trim(e.elem->>'route_id'), ''),
        nullif(trim(e.elem->>'package'), ''),
        e.elem
      from jsonb_array_elements($2::jsonb) with ordinality as e(elem, ordinality)
    `
	case "bookings":
		return `
      insert into stg_xlsx_bookings(batch_id, row_number, booking_id, trip_id, origin_stop_id, dest_stop_id, qty, customer_name, customer_phone, payment_type, payment_status, amount_total, amount_paid, created_at, status, updated_at, reservation_code, reserved_until, row_type, idempotency_key, payment_method, amount_due, abacate_brcode, abacate_brcode_base64, abacate_expires_at, remaining_at_boarding, abacate_pix_id, stripe_payment_intent_id, stripe_pix_copy_paste, stripe_pix_expires_at, stripe_hosted_instructions_url, raw_row)
      select $1,
        e.ordinality::int,
        nullif(trim(e.elem->>'booking_id'), ''),
        nullif(trim(e.elem->>'trip_id'), ''),
        nullif(trim(e.elem->>'origin_stop_id'), ''),
        nullif(trim(e.elem->>'dest_stop_id'), ''),
        case when nullif(trim(e.elem->>'qty'), '') is not null then (e.elem->>'qty')::int else null end,
        nullif(trim(e.elem->>'customer_name'), ''),
        nullif(trim(e.elem->>'customer_phone'), ''),
        nullif(trim(e.elem->>'payment_type'), ''),
        nullif(trim(e.elem->>'payment_status'), ''),
        case when nullif(trim(e.elem->>'amount_total'), '') is not null then (e.elem->>'amount_total')::numeric else null end,
        case when nullif(trim(e.elem->>'amount_paid'), '') is not null then (e.elem->>'amount_paid')::numeric else null end,
        case when nullif(trim(e.elem->>'created_at'), '') is not null then (e.elem->>'created_at')::timestamptz else null end,
        nullif(trim(e.elem->>'status'), ''),
        case when nullif(trim(e.elem->>'updated_at'), '') is not null then (e.elem->>'updated_at')::timestamptz else null end,
        nullif(trim(e.elem->>'reservation_code'), ''),
        case when nullif(trim(e.elem->>'reserved_until'), '') is not null then (e.elem->>'reserved_until')::timestamptz else null end,
        nullif(trim(e.elem->>'row_type'), ''),
        nullif(trim(e.elem->>'idempotency_key'), ''),
        nullif(trim(e.elem->>'payment_method'), ''),
        case when nullif(trim(e.elem->>'amount_due'), '') is not null then (e.elem->>'amount_due')::numeric else null end,
        nullif(trim(e.elem->>'abacate_brcode'), ''),
        nullif(trim(e.elem->>'abacate_brcode_base64'), ''),
        case when nullif(trim(e.elem->>'abacate_expires_at'), '') is not null then (e.elem->>'abacate_expires_at')::timestamptz else null end,
        case when nullif(trim(e.elem->>'remaining_at_boarding'), '') is not null then (e.elem->>'remaining_at_boarding')::numeric else null end,
        nullif(trim(e.elem->>'abacate_pix_id'), ''),
        nullif(trim(e.elem->>'stripe_payment_intent_id'), ''),
        nullif(trim(e.elem->>'stripe_pix_copy_paste'), ''),
        case when nullif(trim(e.elem->>'stripe_pix_expires_at'), '') is not null then (e.elem->>'stripe_pix_expires_at')::timestamptz else null end,
        nullif(trim(e.elem->>'stripe_hosted_instructions_url'), ''),
        e.elem
      from jsonb_array_elements($2::jsonb) with ordinality as e(elem, ordinality)
    `
	case "passengers":
		return `
      insert into stg_xlsx_passengers(batch_id, row_number, passenger_id, booking_id, trip_id, full_name, document, seat_number, origin_stop_id, dest_stop_id, phone, notes, status, created_at, updated_at, row_type, raw_row)
      select $1,
        e.ordinality::int,
        nullif(trim(e.elem->>'passenger_id'), ''),
        nullif(trim(e.elem->>'booking_id'), ''),
        nullif(trim(e.elem->>'trip_id'), ''),
        nullif(trim(e.elem->>'full_name'), ''),
        nullif(trim(e.elem->>'document'), ''),
        case when nullif(trim(e.elem->>'seat_number'), '') is not null then (e.elem->>'seat_number')::int else null end,
        nullif(trim(e.elem->>'origin_stop_id'), ''),
        nullif(trim(e.elem->>'dest_stop_id'), ''),
        nullif(trim(e.elem->>'phone'), ''),
        nullif(trim(e.elem->>'notes'), ''),
        nullif(trim(e.elem->>'status'), ''),
        case when nullif(trim(e.elem->>'created_at'), '') is not null then (e.elem->>'created_at')::timestamptz else null end,
        case when nullif(trim(e.elem->>'updated_at'), '') is not null then (e.elem->>'updated_at')::timestamptz else null end,
        nullif(trim(e.elem->>'row_type'), ''),
        e.elem
      from jsonb_array_elements($2::jsonb) with ordinality as e(elem, ordinality)
    `
	case "manifest_data":
		return `
      insert into stg_xlsx_manifest_data(batch_id, row_number, passenger_id, booking_id, trip_id, trip_date, full_name, document, phone, origin, destination, seat_number, payment_summary, payment_status, amount_total, amount_paid, amount_remaining, update_at, raw_row)
      select $1,
        e.ordinality::int,
        nullif(trim(e.elem->>'passenger_id'), ''),
        nullif(trim(e.elem->>'booking_id'), ''),
        nullif(trim(e.elem->>'trip_id'), ''),
        case when nullif(trim(e.elem->>'trip_date'), '') is not null then (e.elem->>'trip_date')::date else null end,
        nullif(trim(e.elem->>'full_name'), ''),
        nullif(trim(e.elem->>'document'), ''),
        nullif(trim(e.elem->>'phone'), ''),
        nullif(trim(e.elem->>'origin'), ''),
        nullif(trim(e.elem->>'destination'), ''),
        case when nullif(trim(e.elem->>'seat_number'), '') is not null then (e.elem->>'seat_number')::int else null end,
        nullif(trim(e.elem->>'payment_summary'), ''),
        nullif(trim(e.elem->>'payment_status'), ''),
        case when nullif(trim(e.elem->>'amount_total'), '') is not null then (e.elem->>'amount_total')::numeric else null end,
        case when nullif(trim(e.elem->>'amount_paid'), '') is not null then (e.elem->>'amount_paid')::numeric else null end,
        case when nullif(trim(e.elem->>'amount_remaining'), '') is not null then (e.elem->>'amount_remaining')::numeric else null end,
        case when nullif(trim(e.elem->>'update_at'), '') is not null then (e.elem->>'update_at')::timestamptz else null end,
        e.elem
      from jsonb_array_elements($2::jsonb) with ordinality as e(elem, ordinality)
    `
	default:
		return ""
	}
}

func (r *Repository) ValidateBatch(ctx context.Context, batchID string) (map[string]int64, error) {
	rows, err := r.pool.Query(ctx, `select sheet_name, error_count from fn_validate_xlsx_batch($1)`, batchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := map[string]int64{}
	for rows.Next() {
		var sheet string
		var count int64
		if err := rows.Scan(&sheet, &count); err != nil {
			return nil, err
		}
		result[sheet] = count
	}
	return result, rows.Err()
}

func (r *Repository) PromoteBatch(ctx context.Context, batchID string) error {
	_, err := r.pool.Exec(ctx, `select fn_promote_xlsx_batch($1)`, batchID)
	return err
}

func (r *Repository) GetErrors(ctx context.Context, batchID string) ([]ImportError, int64, map[string]int64, error) {
	items := []ImportError{}

	rows, err := r.pool.Query(ctx, `
    select sheet_name, row_number, error_code, error_message, row_data
    from xlsx_import_errors
    where batch_id = $1
    order by sheet_name, row_number nulls last
  `, batchID)
	if err != nil {
		return nil, 0, nil, err
	}
	defer rows.Close()

	var total int64
	bySheet := map[string]int64{}
	for rows.Next() {
		var item ImportError
		var rowData []byte
		if err := rows.Scan(&item.SheetName, &item.RowNumber, &item.ErrorCode, &item.ErrorMessage, &rowData); err != nil {
			return nil, 0, nil, err
		}
		if len(rowData) > 0 {
			_ = json.Unmarshal(rowData, &item.RowData)
		}
		items = append(items, item)
		total++
		bySheet[item.SheetName]++
	}
	if err := rows.Err(); err != nil {
		return nil, 0, nil, err
	}
	return items, total, bySheet, nil
}

func (r *Repository) GetBatchReport(ctx context.Context, batchID string) (BatchReport, error) {
	var report BatchReport
	report.BatchID = batchID
	var rawReport []byte
	if err := r.pool.QueryRow(ctx, `
    select status, source_file_name, uploaded_at, validated_at, promoted_at, report
    from xlsx_import_batches
    where id = $1
  `, batchID).Scan(&report.Status, &report.SourceFile, &report.UploadedAt, &report.ValidatedAt, &report.PromotedAt, &rawReport); err != nil {
		return report, err
	}
	if len(rawReport) > 0 {
		_ = json.Unmarshal(rawReport, &report.Report)
	} else {
		report.Report = map[string]interface{}{}
	}

	errors, total, bySheet, err := r.GetErrors(ctx, batchID)
	if err != nil {
		return report, err
	}
	_ = errors
	report.ErrorCount = total
	report.ErrorsBySheet = bySheet

	report.Reconciliation = map[string]SheetHash{}
	sheets := []string{"routes", "stops", "trips", "trip_stops", "available_segments", "bookings", "passengers", "manifest_data"}
	for _, sheet := range sheets {
		rows, err := r.pool.Query(ctx, `select metric, value from fn_compare_xlsx_vs_db($1)`, sheet)
		if err != nil {
			continue
		}
		hash := SheetHash{}
		for rows.Next() {
			var metric string
			var value string
			if scanErr := rows.Scan(&metric, &value); scanErr != nil {
				continue
			}
			if metric == "count" {
				var count int64
				fmt.Sscan(value, &count)
				hash.Count = count
			}
			if metric == "hash_md5" {
				hash.Hash = value
			}
		}
		rows.Close()
		report.Reconciliation[sheet] = hash
	}

	return report, nil
}

func (r *Repository) GetBatchStatus(ctx context.Context, batchID string) (string, error) {
	var status string
	err := r.pool.QueryRow(ctx, `select status from xlsx_import_batches where id = $1`, batchID).Scan(&status)
	return status, err
}

func nullString(v string) interface{} {
	if v == "" {
		return nil
	}
	return v
}
