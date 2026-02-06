package users

import (
  "github.com/go-chi/chi/v5"
  httpx "schumacher-tur/api/internal/shared/http"
)

type Handler struct {}

func NewHandler() *Handler { return &Handler{} }

func (h *Handler) RegisterRoutes(r chi.Router) {
  r.Route("/users", func(r chi.Router) {
    r.Get("/me", httpx.NotImplemented)
  })
}
