package drivers

import "time"

type Driver struct {
  ID        string    `json:"id"`
  Name      string    `json:"name"`
  Document  string    `json:"document"`
  Phone     string    `json:"phone"`
  IsActive  bool      `json:"is_active"`
  CreatedAt time.Time `json:"created_at"`
}

type CreateDriverInput struct {
  Name     string `json:"name"`
  Document string `json:"document"`
  Phone    string `json:"phone"`
  IsActive *bool  `json:"is_active"`
}

type UpdateDriverInput struct {
  Name     *string `json:"name"`
  Document *string `json:"document"`
  Phone    *string `json:"phone"`
  IsActive *bool   `json:"is_active"`
}

type ListFilter struct {
  Limit  int
  Offset int
}
