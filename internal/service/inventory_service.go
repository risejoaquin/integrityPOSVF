package service

import (
	"context"

	"github.com/solidbit/integritypos/internal/model"
	"github.com/solidbit/integritypos/internal/repository"
)

type InventoryService struct {
	Repo *repository.InventoryRepository
}

func (s *InventoryService) GetMovements(ctx context.Context, productID string, limit int) ([]model.InventoryMovement, error) {
	return s.Repo.GetMovements(ctx, productID, limit)
}
