package buses

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrCapacityInUse = errors.New("cannot reduce capacity; seats are in use")

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) List(ctx context.Context, filter ListFilter) ([]Bus, error) {
	limit := filter.Limit
	if limit <= 0 || limit > 500 {
		limit = 200
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	query := `select id, name, plate, capacity, seat_map_name, is_active, created_at
     from buses`
	args := []interface{}{}
	clauses := []string{}

	if filter.Search != "" {
		args = append(args, "%"+filter.Search+"%")
		clauses = append(clauses, fmt.Sprintf(`(
      name ilike $%d
      or plate ilike $%d
      or seat_map_name ilike $%d
      or id::text ilike $%d
    )`, len(args), len(args), len(args), len(args)))
	}

	if len(clauses) > 0 {
		query += " where " + strings.Join(clauses, " and ")
	}

	query += " order by created_at desc"
	args = append(args, limit)
	query += fmt.Sprintf(" limit $%d", len(args))
	args = append(args, offset)
	query += fmt.Sprintf(" offset $%d", len(args))

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []Bus{}
	for rows.Next() {
		var item Bus
		if err := rows.Scan(&item.ID, &item.Name, &item.Plate, &item.Capacity, &item.SeatMapName, &item.IsActive, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) Get(ctx context.Context, id uuid.UUID) (Bus, error) {
	var item Bus
	row := r.pool.QueryRow(ctx, `select id, name, plate, capacity, seat_map_name, is_active, created_at from buses where id=$1`, id)
	if err := row.Scan(&item.ID, &item.Name, &item.Plate, &item.Capacity, &item.SeatMapName, &item.IsActive, &item.CreatedAt); err != nil {
		return item, err
	}
	return item, nil
}

func (r *Repository) Create(ctx context.Context, input CreateBusInput) (Bus, error) {
	isActive := true
	if input.IsActive != nil {
		isActive = *input.IsActive
	}
	createSeats := true
	if input.CreateSeats != nil {
		createSeats = *input.CreateSeats
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return Bus{}, err
	}
	defer tx.Rollback(ctx)

	var item Bus
	row := tx.QueryRow(ctx,
		`insert into buses (name, plate, capacity, seat_map_name, is_active)
     values ($1,$2,$3,$4,$5)
     returning id, name, plate, capacity, seat_map_name, is_active, created_at`,
		input.Name, input.Plate, input.Capacity, input.SeatMapName, isActive,
	)
	if err := row.Scan(&item.ID, &item.Name, &item.Plate, &item.Capacity, &item.SeatMapName, &item.IsActive, &item.CreatedAt); err != nil {
		return Bus{}, err
	}

	if createSeats && item.Capacity > 0 {
		for i := 1; i <= item.Capacity; i++ {
			if _, err := tx.Exec(ctx, `insert into bus_seats (bus_id, seat_number) values ($1,$2) on conflict do nothing`, item.ID, i); err != nil {
				return Bus{}, err
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return Bus{}, err
	}
	return item, nil
}

func (r *Repository) Update(ctx context.Context, id uuid.UUID, input UpdateBusInput) (Bus, error) {
	if input.Capacity != nil {
		return r.updateWithCapacity(ctx, id, input)
	}

	sets := []string{}
	args := []interface{}{}
	idx := 1

	if input.Name != nil {
		sets = append(sets, fmt.Sprintf("name=$%d", idx))
		args = append(args, *input.Name)
		idx++
	}
	if input.Plate != nil {
		sets = append(sets, fmt.Sprintf("plate=$%d", idx))
		args = append(args, *input.Plate)
		idx++
	}
	if input.SeatMapName != nil {
		sets = append(sets, fmt.Sprintf("seat_map_name=$%d", idx))
		args = append(args, *input.SeatMapName)
		idx++
	}
	if input.IsActive != nil {
		sets = append(sets, fmt.Sprintf("is_active=$%d", idx))
		args = append(args, *input.IsActive)
		idx++
	}

	if len(sets) == 0 {
		return r.Get(ctx, id)
	}

	args = append(args, id)
	query := fmt.Sprintf(`update buses set %s where id=$%d returning id, name, plate, capacity, seat_map_name, is_active, created_at`, strings.Join(sets, ", "), idx)

	var item Bus
	row := r.pool.QueryRow(ctx, query, args...)
	if err := row.Scan(&item.ID, &item.Name, &item.Plate, &item.Capacity, &item.SeatMapName, &item.IsActive, &item.CreatedAt); err != nil {
		return item, err
	}
	return item, nil
}

func (r *Repository) updateWithCapacity(ctx context.Context, id uuid.UUID, input UpdateBusInput) (Bus, error) {
	newCap := *input.Capacity

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return Bus{}, err
	}
	defer tx.Rollback(ctx)

	var currentCap int
	if err := tx.QueryRow(ctx, `select capacity from buses where id=$1`, id).Scan(&currentCap); err != nil {
		return Bus{}, err
	}

	if newCap < currentCap {
		var inUse int
		if err := tx.QueryRow(ctx,
			`select count(1)
       from booking_passengers bp
       join bus_seats bs on bp.seat_id = bs.id
       join bookings b on b.id = bp.booking_id
       where bs.bus_id = $1
         and bs.seat_number > $2
         and bp.is_active = true
         and b.status not in ('CANCELLED','EXPIRED')`,
			id, newCap,
		).Scan(&inUse); err != nil {
			return Bus{}, err
		}
		if inUse > 0 {
			return Bus{}, ErrCapacityInUse
		}
	}

	sets := []string{"capacity=$1"}
	args := []interface{}{newCap}
	idx := 2

	if input.Name != nil {
		sets = append(sets, fmt.Sprintf("name=$%d", idx))
		args = append(args, *input.Name)
		idx++
	}
	if input.Plate != nil {
		sets = append(sets, fmt.Sprintf("plate=$%d", idx))
		args = append(args, *input.Plate)
		idx++
	}
	if input.SeatMapName != nil {
		sets = append(sets, fmt.Sprintf("seat_map_name=$%d", idx))
		args = append(args, *input.SeatMapName)
		idx++
	}
	if input.IsActive != nil {
		sets = append(sets, fmt.Sprintf("is_active=$%d", idx))
		args = append(args, *input.IsActive)
		idx++
	}

	args = append(args, id)
	query := fmt.Sprintf(`update buses set %s where id=$%d returning id, name, plate, capacity, seat_map_name, is_active, created_at`, strings.Join(sets, ", "), idx)

	var item Bus
	row := tx.QueryRow(ctx, query, args...)
	if err := row.Scan(&item.ID, &item.Name, &item.Plate, &item.Capacity, &item.SeatMapName, &item.IsActive, &item.CreatedAt); err != nil {
		return Bus{}, err
	}

	if newCap > currentCap {
		for i := currentCap + 1; i <= newCap; i++ {
			if _, err := tx.Exec(ctx, `insert into bus_seats (bus_id, seat_number) values ($1,$2) on conflict do nothing`, id, i); err != nil {
				return Bus{}, err
			}
		}
	} else if newCap < currentCap {
		if _, err := tx.Exec(ctx, `update bus_seats set is_active = (seat_number <= $2) where bus_id=$1`, id, newCap); err != nil {
			return Bus{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return Bus{}, err
	}

	return item, nil
}

func IsNotFound(err error) bool {
	return err == pgx.ErrNoRows
}
