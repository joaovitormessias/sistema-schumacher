package purchase_orders

import (
	"context"

	"github.com/google/uuid"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context, filter ListFilter) ([]PurchaseOrder, error) {
	return s.repo.List(ctx, filter)
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (PurchaseOrder, error) {
	return s.repo.Get(ctx, id)
}

func (s *Service) Create(ctx context.Context, input CreatePurchaseOrderInput) (PurchaseOrder, error) {
	return s.repo.Create(ctx, input)
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, input UpdatePurchaseOrderInput) (PurchaseOrder, error) {
	return s.repo.Update(ctx, id, input)
}

func (s *Service) AddItem(ctx context.Context, orderID uuid.UUID, input AddItemInput) (PurchaseOrder, error) {
	return s.repo.AddItem(ctx, orderID, input)
}

func (s *Service) RemoveItem(ctx context.Context, orderID, itemID uuid.UUID) error {
	return s.repo.RemoveItem(ctx, orderID, itemID)
}

func (s *Service) Send(ctx context.Context, id uuid.UUID) error {
	return s.repo.SetStatus(ctx, id, StatusSent)
}

func (s *Service) MarkReceived(ctx context.Context, id uuid.UUID) error {
	return s.repo.SetStatus(ctx, id, StatusReceived)
}

func (s *Service) Cancel(ctx context.Context, id uuid.UUID) error {
	return s.repo.SetStatus(ctx, id, StatusCancelled)
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}
