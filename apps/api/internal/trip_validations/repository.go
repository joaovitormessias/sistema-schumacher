package trip_validations

import (
  "context"
  "fmt"
  "strings"

  "github.com/google/uuid"
  "github.com/jackc/pgx/v5"
  "github.com/jackc/pgx/v5/pgconn"
  "github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
  pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
  return &Repository{pool: pool}
}

func (r *Repository) List(ctx context.Context, filter ListFilter) ([]TripValidation, error) {
  query := `select id, trip_id, odometer_initial, odometer_final, distance_km, passengers_expected, passengers_boarded, passengers_no_show, validation_notes, validated_by, validated_at, created_at, updated_at from trip_validations`
  args := []interface{}{}
  clauses := []string{}

  if filter.TripID != "" {
    args = append(args, filter.TripID)
    clauses = append(clauses, fmt.Sprintf("trip_id=$%d", len(args)))
  }

  if len(clauses) > 0 {
    query += " where " + strings.Join(clauses, " and ")
  }

  query += " order by created_at desc"

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

  items := []TripValidation{}
  for rows.Next() {
    var item TripValidation
    if err := rows.Scan(&item.ID, &item.TripID, &item.OdometerInitial, &item.OdometerFinal, &item.DistanceKM, &item.PassengersExpected, &item.PassengersBoarded, &item.PassengersNoShow, &item.ValidationNotes, &item.ValidatedBy, &item.ValidatedAt, &item.CreatedAt, &item.UpdatedAt); err != nil {
      return nil, err
    }
    items = append(items, item)
  }
  return items, rows.Err()
}

func (r *Repository) Get(ctx context.Context, id uuid.UUID) (TripValidation, error) {
  var item TripValidation
  row := r.pool.QueryRow(ctx, `select id, trip_id, odometer_initial, odometer_final, distance_km, passengers_expected, passengers_boarded, passengers_no_show, validation_notes, validated_by, validated_at, created_at, updated_at from trip_validations where id=$1`, id)
  if err := row.Scan(&item.ID, &item.TripID, &item.OdometerInitial, &item.OdometerFinal, &item.DistanceKM, &item.PassengersExpected, &item.PassengersBoarded, &item.PassengersNoShow, &item.ValidationNotes, &item.ValidatedBy, &item.ValidatedAt, &item.CreatedAt, &item.UpdatedAt); err != nil {
    return item, err
  }
  return item, nil
}

func (r *Repository) Create(ctx context.Context, input CreateTripValidationInput, validatedBy *string) (TripValidation, error) {
  var item TripValidation
  row := r.pool.QueryRow(ctx,
    `insert into trip_validations (trip_id, odometer_initial, odometer_final, passengers_expected, passengers_boarded, passengers_no_show, validation_notes, validated_by, validated_at)
     values ($1,$2,$3,$4,$5,$6,$7,$8,now())
     returning id, trip_id, odometer_initial, odometer_final, distance_km, passengers_expected, passengers_boarded, passengers_no_show, validation_notes, validated_by, validated_at, created_at, updated_at`,
    input.TripID, input.OdometerInitial, input.OdometerFinal, input.PassengersExpected, input.PassengersBoarded, input.PassengersNoShow, input.ValidationNotes, validatedBy,
  )
  if err := row.Scan(&item.ID, &item.TripID, &item.OdometerInitial, &item.OdometerFinal, &item.DistanceKM, &item.PassengersExpected, &item.PassengersBoarded, &item.PassengersNoShow, &item.ValidationNotes, &item.ValidatedBy, &item.ValidatedAt, &item.CreatedAt, &item.UpdatedAt); err != nil {
    return item, err
  }
  return item, nil
}

func (r *Repository) Update(ctx context.Context, id uuid.UUID, input UpdateTripValidationInput, validatedBy *string) (TripValidation, error) {
  sets := []string{}
  args := []interface{}{}
  idx := 1

  if input.OdometerInitial != nil {
    sets = append(sets, fmt.Sprintf("odometer_initial=$%d", idx))
    args = append(args, *input.OdometerInitial)
    idx++
  }
  if input.OdometerFinal != nil {
    sets = append(sets, fmt.Sprintf("odometer_final=$%d", idx))
    args = append(args, *input.OdometerFinal)
    idx++
  }
  if input.PassengersExpected != nil {
    sets = append(sets, fmt.Sprintf("passengers_expected=$%d", idx))
    args = append(args, *input.PassengersExpected)
    idx++
  }
  if input.PassengersBoarded != nil {
    sets = append(sets, fmt.Sprintf("passengers_boarded=$%d", idx))
    args = append(args, *input.PassengersBoarded)
    idx++
  }
  if input.PassengersNoShow != nil {
    sets = append(sets, fmt.Sprintf("passengers_no_show=$%d", idx))
    args = append(args, *input.PassengersNoShow)
    idx++
  }
  if input.ValidationNotes != nil {
    sets = append(sets, fmt.Sprintf("validation_notes=$%d", idx))
    args = append(args, *input.ValidationNotes)
    idx++
  }

  if len(sets) == 0 {
    return r.Get(ctx, id)
  }

  sets = append(sets, fmt.Sprintf("validated_by=$%d", idx))
  args = append(args, validatedBy)
  idx++
  sets = append(sets, "validated_at=now()")
  sets = append(sets, "updated_at=now()")

  args = append(args, id)
  query := fmt.Sprintf(`update trip_validations set %s where id=$%d returning id, trip_id, odometer_initial, odometer_final, distance_km, passengers_expected, passengers_boarded, passengers_no_show, validation_notes, validated_by, validated_at, created_at, updated_at`, strings.Join(sets, ", "), idx)

  var item TripValidation
  row := r.pool.QueryRow(ctx, query, args...)
  if err := row.Scan(&item.ID, &item.TripID, &item.OdometerInitial, &item.OdometerFinal, &item.DistanceKM, &item.PassengersExpected, &item.PassengersBoarded, &item.PassengersNoShow, &item.ValidationNotes, &item.ValidatedBy, &item.ValidatedAt, &item.CreatedAt, &item.UpdatedAt); err != nil {
    return item, err
  }
  return item, nil
}

func IsNotFound(err error) bool {
  return err == pgx.ErrNoRows
}

func IsUniqueViolation(err error) bool {
  pgErr, ok := err.(*pgconn.PgError)
  if !ok {
    return false
  }
  return pgErr.Code == "23505"
}
