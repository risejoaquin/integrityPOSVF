package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/solidbit/integritypos/internal/model"
	"github.com/solidbit/integritypos/internal/repository"
)

type ProductSvcRepository interface {
	List(ctx context.Context, filter repository.ProductFilter) ([]model.Product, error)
	GetByID(ctx context.Context, id string) (model.Product, error)
	Create(ctx context.Context, p *model.Product) error
	Update(ctx context.Context, p *model.Product) error
	Delete(ctx context.Context, id string) error
	GetStock(ctx context.Context, id string) (int, error)
	UpdateStock(ctx context.Context, tx repository.Tx, productID string, delta int) error
	BeginTx(ctx context.Context) (repository.Tx, error)
}

type InventorySvcRepository interface {
	RecordMovement(ctx context.Context, tx repository.Tx, productID string, delta int, reason string, orderID string) error
	GetMovements(ctx context.Context, productID string, limit int) ([]model.InventoryMovement, error)
}

type ProductService struct {
	Repo          ProductSvcRepository
	InventoryRepo InventorySvcRepository
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
		return fmt.Errorf("product name cannot be empty")
	}
	if p.PriceCents < 0 {
		return fmt.Errorf("product price cannot be negative")
	}
	return s.Repo.Create(ctx, p)
}

func (s *ProductService) Update(ctx context.Context, p *model.Product) error {
	p.Name = strings.TrimSpace(p.Name)
	if p.Name == "" {
		return fmt.Errorf("product name cannot be empty")
	}
	if p.PriceCents < 0 {
		return fmt.Errorf("product price cannot be negative")
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

	tx, err := s.Repo.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	stock, err := s.Repo.GetStock(ctx, id)
	if err != nil {
		return err
	}

	if stock+delta < 0 && reason != "inventario inicial" {
		return fmt.Errorf("insufficient stock for product %s", id)
	}

	if err := s.Repo.UpdateStock(ctx, tx, id, delta); err != nil {
		return err
	}

	if err := s.InventoryRepo.RecordMovement(ctx, tx, id, delta, reason, ""); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
