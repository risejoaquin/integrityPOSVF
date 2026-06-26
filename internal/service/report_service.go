package service

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/solidbit/integritypos/internal/model"
)

type SalesSummary struct {
	TotalOrders  int     `json:"total_orders"`
	TotalRevenue float64 `json:"total_revenue"`
	AverageTicket float64 `json:"average_ticket"`
}

type TopProduct struct {
	ProductID string  `json:"product_id"`
	Name      string  `json:"name"`
	Quantity  int     `json:"quantity"`
	Revenue   float64 `json:"revenue"`
}

type DailySale struct {
	Date    string  `json:"date"`
	Revenue float64 `json:"revenue"`
	Orders  int     `json:"orders"`
}

type ReportService struct {
	DB *pgxpool.Pool
}

func NewReportService(db *pgxpool.Pool) *ReportService {
	return &ReportService{DB: db}
}

func (s *ReportService) SalesSummary(ctx context.Context, from, to time.Time) (*SalesSummary, error) {
	query := `SELECT COUNT(id), COALESCE(SUM(total_cents), 0) FROM orders WHERE status = 'completed' AND created_at >= $1 AND created_at <= $2`
	var totalOrders int
	var totalCents int64

	err := s.DB.QueryRow(ctx, query, from, to).Scan(&totalOrders, &totalCents)
	if err != nil {
		return nil, fmt.Errorf("ReportService.SalesSummary: %w", err)
	}

	revenue := float64(totalCents) / 100.0
	avgTicket := 0.0
	if totalOrders > 0 {
		avgTicket = revenue / float64(totalOrders)
	}

	return &SalesSummary{
		TotalOrders:   totalOrders,
		TotalRevenue:  revenue,
		AverageTicket: avgTicket,
	}, nil
}

func (s *ReportService) TopProducts(ctx context.Context, from, to time.Time, limit int) ([]TopProduct, error) {
	if limit <= 0 {
		limit = 10
	} else if limit > 50 {
		limit = 50
	}

	query := `
		SELECT product_id, product_name, SUM(quantity) as qty, SUM(total_cents) as rev
		FROM order_items oi
		JOIN orders o ON oi.order_id = o.id
		WHERE o.status = 'completed' AND o.created_at >= $1 AND o.created_at <= $2
		GROUP BY product_id, product_name
		ORDER BY qty DESC
		LIMIT $3
	`

	rows, err := s.DB.Query(ctx, query, from, to, limit)
	if err != nil {
		return nil, fmt.Errorf("ReportService.TopProducts: %w", err)
	}
	defer rows.Close()

	var result []TopProduct
	for rows.Next() {
		var tp TopProduct
		var pid *string
		var rev int64
		if err := rows.Scan(&pid, &tp.Name, &tp.Quantity, &rev); err != nil {
			return nil, fmt.Errorf("ReportService.TopProducts scan: %w", err)
		}
		if pid != nil {
			tp.ProductID = *pid
		}
		tp.Revenue = float64(rev) / 100.0
		result = append(result, tp)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ReportService.TopProducts rows error: %w", err)
	}

	if result == nil {
		result = make([]TopProduct, 0)
	}

	return result, nil
}

func (s *ReportService) LowStock(ctx context.Context, threshold int) ([]model.Product, error) {
	query := `SELECT id, name, price_cents, category, stock, is_available, attributes, created_at, updated_at FROM products WHERE stock <= $1 AND is_available = true ORDER BY stock ASC`
	rows, err := s.DB.Query(ctx, query, threshold)
	if err != nil {
		return nil, fmt.Errorf("ReportService.LowStock: %w", err)
	}
	defer rows.Close()

	var products = make([]model.Product, 0)
	for rows.Next() {
		var p model.Product
		if err := rows.Scan(&p.ID, &p.Name, &p.PriceCents, &p.Category, &p.Stock, &p.IsAvailable, &p.Attributes, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("ReportService.LowStock scan: %w", err)
		}
		products = append(products, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ReportService.LowStock rows error: %w", err)
	}

	return products, nil
}

func (s *ReportService) DailySales(ctx context.Context, days int) ([]DailySale, error) {
	if days <= 0 {
		days = 7
	} else if days > 90 {
		days = 90
	}

	query := `
		SELECT TO_CHAR(created_at, 'YYYY-MM-DD') as dt, COUNT(id), COALESCE(SUM(total_cents), 0)
		FROM orders
		WHERE status = 'completed' AND created_at >= CURRENT_DATE - ($1 || ' days')::interval
		GROUP BY dt
		ORDER BY dt ASC
	`
	rows, err := s.DB.Query(ctx, query, days)
	if err != nil {
		return nil, fmt.Errorf("ReportService.DailySales: %w", err)
	}
	defer rows.Close()

	var result = make([]DailySale, 0)
	for rows.Next() {
		var ds DailySale
		var rev int64
		if err := rows.Scan(&ds.Date, &ds.Orders, &rev); err != nil {
			return nil, fmt.Errorf("ReportService.DailySales scan: %w", err)
		}
		ds.Revenue = float64(rev) / 100.0
		result = append(result, ds)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ReportService.DailySales rows error: %w", err)
	}

	return result, nil
}
