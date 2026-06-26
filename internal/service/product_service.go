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

	initialStock := p.Stock

	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(context.Background())

	if err := s.Repo.Create(ctx, tx, p); err != nil {
		return err
	}

	if initialStock != 0 {
		if err := s.AdjustStock(ctx, tx, p.ID, initialStock, "inventario inicial"); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
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

func (s *ProductService) AdjustStock(ctx context.Context, db repository.DBTX, id string, delta int, reason string) error {
	if delta == 0 {
		return nil
	}

	stock, err := s.Repo.GetStock(ctx, db, id)
	if err != nil {
		return err
	}

	if stock+delta < 0 && reason != "inventario inicial" {
		return fmt.Errorf("%w: for product %s", model.ErrStockInsufficient, id)
	}

	if err := s.Repo.DecrementStockAtomic(ctx, db, id, -delta); err != nil {
		return err
	}

	if err := s.InventoryRepo.RecordMovement(ctx, db, id, delta, reason, ""); err != nil {
		return err
	}

	return nil
}

