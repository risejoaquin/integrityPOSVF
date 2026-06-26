package service

import (
	"context"

	"github.com/solidbit/integritypos/internal/model"
	"github.com/solidbit/integritypos/internal/repository"
)

type CustomerService struct {
	Repo *repository.CustomerRepository
}

func (s *CustomerService) List(ctx context.Context, limit, offset int) ([]model.Customer, error) {
	return s.Repo.List(ctx, limit, offset)
}

func (s *CustomerService) GetByID(ctx context.Context, id string) (model.Customer, error) {
	return s.Repo.GetByID(ctx, id)
}

func (s *CustomerService) Create(ctx context.Context, c *model.Customer) error {
	return s.Repo.Create(ctx, c)
}

func (s *CustomerService) Update(ctx context.Context, c *model.Customer) error {
	return s.Repo.Update(ctx, c)
}
