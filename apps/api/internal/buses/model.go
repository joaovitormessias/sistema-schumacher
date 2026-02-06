package buses

import "time"

type Bus struct {
  ID          string    `json:"id"`
  Name        string    `json:"name"`
  Plate       string    `json:"plate"`
  Capacity    int       `json:"capacity"`
  SeatMapName string    `json:"seat_map_name"`
  IsActive    bool      `json:"is_active"`
  CreatedAt   time.Time `json:"created_at"`
}

type CreateBusInput struct {
  Name        string `json:"name"`
  Plate       string `json:"plate"`
  Capacity    int    `json:"capacity"`
  SeatMapName string `json:"seat_map_name"`
  IsActive    *bool  `json:"is_active"`
  CreateSeats *bool  `json:"create_seats"`
}

type UpdateBusInput struct {
  Name        *string `json:"name"`
  Plate       *string `json:"plate"`
  Capacity    *int    `json:"capacity"`
  SeatMapName *string `json:"seat_map_name"`
  IsActive    *bool   `json:"is_active"`
}

type ListFilter struct {
  Limit  int
  Offset int
}
