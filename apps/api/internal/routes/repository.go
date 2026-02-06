package routes

import (
  "context"
  "fmt"
  "strings"

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

func (r *Repository) List(ctx context.Context, filter ListFilter) ([]Route, error) {
  limit := filter.Limit
  if limit <= 0 || limit > 500 {
    limit = 200
  }
  offset := filter.Offset
  if offset < 0 {
    offset = 0
  }

  rows, err := r.pool.Query(ctx,
    `select id, name, origin_city, destination_city, is_active, created_at
     from routes
     order by created_at desc
     limit $1 offset $2`, limit, offset,
  )
  if err != nil {
    return nil, err
  }
  defer rows.Close()

  items := []Route{}
  for rows.Next() {
    var item Route
    if err := rows.Scan(&item.ID, &item.Name, &item.OriginCity, &item.DestinationCity, &item.IsActive, &item.CreatedAt); err != nil {
      return nil, err
    }
    items = append(items, item)
  }
  return items, rows.Err()
}

func (r *Repository) Get(ctx context.Context, id uuid.UUID) (Route, error) {
  var item Route
  row := r.pool.QueryRow(ctx, `select id, name, origin_city, destination_city, is_active, created_at from routes where id=$1`, id)
  if err := row.Scan(&item.ID, &item.Name, &item.OriginCity, &item.DestinationCity, &item.IsActive, &item.CreatedAt); err != nil {
    return item, err
  }
  return item, nil
}

func (r *Repository) Create(ctx context.Context, input CreateRouteInput) (Route, error) {
  isActive := true
  if input.IsActive != nil {
    isActive = *input.IsActive
  }

  var item Route
  row := r.pool.QueryRow(ctx,
    `insert into routes (name, origin_city, destination_city, is_active)
     values ($1,$2,$3,$4)
     returning id, name, origin_city, destination_city, is_active, created_at`,
    input.Name, input.OriginCity, input.DestinationCity, isActive,
  )
  if err := row.Scan(&item.ID, &item.Name, &item.OriginCity, &item.DestinationCity, &item.IsActive, &item.CreatedAt); err != nil {
    return item, err
  }
  return item, nil
}

func (r *Repository) Update(ctx context.Context, id uuid.UUID, input UpdateRouteInput) (Route, error) {
  sets := []string{}
  args := []interface{}{}
  idx := 1

  if input.Name != nil {
    sets = append(sets, fmt.Sprintf("name=$%d", idx))
    args = append(args, *input.Name)
    idx++
  }
  if input.OriginCity != nil {
    sets = append(sets, fmt.Sprintf("origin_city=$%d", idx))
    args = append(args, *input.OriginCity)
    idx++
  }
  if input.DestinationCity != nil {
    sets = append(sets, fmt.Sprintf("destination_city=$%d", idx))
    args = append(args, *input.DestinationCity)
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
  query := fmt.Sprintf(`update routes set %s where id=$%d returning id, name, origin_city, destination_city, is_active, created_at`, strings.Join(sets, ", "), idx)

  var item Route
  row := r.pool.QueryRow(ctx, query, args...)
  if err := row.Scan(&item.ID, &item.Name, &item.OriginCity, &item.DestinationCity, &item.IsActive, &item.CreatedAt); err != nil {
    return item, err
  }
  return item, nil
}

func IsNotFound(err error) bool {
  return err == pgx.ErrNoRows
}

func (r *Repository) ListStops(ctx context.Context, routeID uuid.UUID) ([]RouteStop, error) {
  rows, err := r.pool.Query(ctx,
    `select id, route_id, city, stop_order, eta_offset_minutes, notes, created_at
     from route_stops
     where route_id = $1
     order by stop_order asc`, routeID,
  )
  if err != nil {
    return nil, err
  }
  defer rows.Close()

  items := []RouteStop{}
  for rows.Next() {
    var item RouteStop
    if err := rows.Scan(&item.ID, &item.RouteID, &item.City, &item.StopOrder, &item.ETAOffsetMinutes, &item.Notes, &item.CreatedAt); err != nil {
      return nil, err
    }
    items = append(items, item)
  }
  return items, rows.Err()
}

func (r *Repository) CreateStop(ctx context.Context, routeID uuid.UUID, input CreateRouteStopInput) (RouteStop, error) {
  var item RouteStop
  row := r.pool.QueryRow(ctx,
    `insert into route_stops (route_id, city, stop_order, eta_offset_minutes, notes)
     values ($1,$2,$3,$4,$5)
     returning id, route_id, city, stop_order, eta_offset_minutes, notes, created_at`,
    routeID, input.City, input.StopOrder, input.ETAOffsetMinutes, input.Notes,
  )
  if err := row.Scan(&item.ID, &item.RouteID, &item.City, &item.StopOrder, &item.ETAOffsetMinutes, &item.Notes, &item.CreatedAt); err != nil {
    return item, err
  }
  return item, nil
}

func (r *Repository) UpdateStop(ctx context.Context, routeID uuid.UUID, stopID uuid.UUID, input UpdateRouteStopInput) (RouteStop, error) {
  sets := []string{}
  args := []interface{}{}
  idx := 1

  if input.City != nil {
    sets = append(sets, fmt.Sprintf("city=$%d", idx))
    args = append(args, *input.City)
    idx++
  }
  if input.StopOrder != nil {
    sets = append(sets, fmt.Sprintf("stop_order=$%d", idx))
    args = append(args, *input.StopOrder)
    idx++
  }
  if input.ETAOffsetMinutes != nil {
    sets = append(sets, fmt.Sprintf("eta_offset_minutes=$%d", idx))
    args = append(args, *input.ETAOffsetMinutes)
    idx++
  }
  if input.Notes != nil {
    sets = append(sets, fmt.Sprintf("notes=$%d", idx))
    args = append(args, *input.Notes)
    idx++
  }

  if len(sets) == 0 {
    return r.GetStop(ctx, routeID, stopID)
  }

  args = append(args, routeID, stopID)
  query := fmt.Sprintf(`update route_stops set %s where route_id=$%d and id=$%d returning id, route_id, city, stop_order, eta_offset_minutes, notes, created_at`, strings.Join(sets, ", "), idx, idx+1)

  var item RouteStop
  row := r.pool.QueryRow(ctx, query, args...)
  if err := row.Scan(&item.ID, &item.RouteID, &item.City, &item.StopOrder, &item.ETAOffsetMinutes, &item.Notes, &item.CreatedAt); err != nil {
    return item, err
  }
  return item, nil
}

func (r *Repository) GetStop(ctx context.Context, routeID uuid.UUID, stopID uuid.UUID) (RouteStop, error) {
  var item RouteStop
  row := r.pool.QueryRow(ctx,
    `select id, route_id, city, stop_order, eta_offset_minutes, notes, created_at
     from route_stops
     where route_id=$1 and id=$2`, routeID, stopID,
  )
  if err := row.Scan(&item.ID, &item.RouteID, &item.City, &item.StopOrder, &item.ETAOffsetMinutes, &item.Notes, &item.CreatedAt); err != nil {
    return item, err
  }
  return item, nil
}
