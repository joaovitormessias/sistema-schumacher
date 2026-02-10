package products

import "time"

type Product struct {
	ID           string    `json:"id"`
	Code         string    `json:"code"`
	Name         string    `json:"name"`
	Category     *string   `json:"category"`
	Unit         string    `json:"unit"`
	MinStock     float64   `json:"min_stock"`
	CurrentStock float64   `json:"current_stock"`
	LastCost     *float64  `json:"last_cost"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type CreateProductInput struct {
	Code     string   `json:"code"`
	Name     string   `json:"name"`
	Category *string  `json:"category"`
	Unit     *string  `json:"unit"`
	MinStock *float64 `json:"min_stock"`
	IsActive *bool    `json:"is_active"`
}

type UpdateProductInput struct {
	Code     *string  `json:"code"`
	Name     *string  `json:"name"`
	Category *string  `json:"category"`
	Unit     *string  `json:"unit"`
	MinStock *float64 `json:"min_stock"`
	IsActive *bool    `json:"is_active"`
}

type ListFilter struct {
	Limit    int
	Offset   int
	Active   *bool
	Category *string
	Search   *string
}
