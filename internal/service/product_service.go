package service

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/solidbit/integritypos/internal/model"
	"github.com/solidbit/integritypos/internal/repository"
)

type ProductService struct {
	DB            DBBeginner
	Repo          ProductRepo
	InventoryRepo InventoryRepo
}

func (s *ProductService) List(ctx context.Context, filter repository.ProductFilter) ([]model.Product, error) {
	return s.Repo.List(ctx, filter)
}

func (s *ProductService) GetByID(ctx context.Context, id string) (model.Product, error) {
	return s.Repo.GetByID(ctx, id)
}

func (s *ProductService) Create(ctx context.Context, p *model.Product) error {
	p.Name = strings.TrimSpace(p.Name)
	if p.Name == "" {
		return fmt.Errorf("%w: product name cannot be empty", model.ErrInvalidInput)
	}
	if p.PriceCents < 0 {
		return fmt.Errorf("%w: product price cannot be negative", model.ErrInvalidInput)
	}
	return s.Repo.Create(ctx, p)
}

func (s *ProductService) Update(ctx context.Context, p *model.Product) error {
	p.Name = strings.TrimSpace(p.Name)
	if p.Name == "" {
		return fmt.Errorf("%w: product name cannot be empty", model.ErrInvalidInput)
	}
	if p.PriceCents < 0 {
		return fmt.Errorf("%w: product price cannot be negative", model.ErrInvalidInput)
	}
	return s.Repo.Update(ctx, p)
}

func (s *ProductService) Delete(ctx context.Context, id string) error {
	return s.Repo.Delete(ctx, id)
}

func (s *ProductService) AdjustStock(ctx context.Context, id string, delta int, reason string) error {
	if delta == 0 {
		return nil
	}

	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	var commitErr error
	defer func() {
		if commitErr != nil {
			if rollbackErr := tx.Rollback(context.Background()); rollbackErr != nil {
				log.Printf("error rolling back transaction: %v", rollbackErr)
			}
		}
	}()

	stock, err := s.Repo.GetStock(ctx, id)
	if err != nil {
		commitErr = err
		return err
	}

	if stock+delta < 0 && reason != "inventario inicial" {
		commitErr = fmt.Errorf("%w: for product %s", model.ErrStockInsufficient, id)
		return commitErr
	}

	if err := s.Repo.DecrementStockAtomic(ctx, tx, id, -delta); err != nil {
		commitErr = err
		return err
	}

	if err := s.InventoryRepo.RecordMovement(ctx, tx, id, delta, reason, ""); err != nil {
		commitErr = err
		return err
	}

	commitErr = tx.Commit(ctx)
	if commitErr != nil {
		return fmt.Errorf("failed to commit transaction: %w", commitErr)
	}

	return nil
}

