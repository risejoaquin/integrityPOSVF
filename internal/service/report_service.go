package service

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/solidbit/integritypos/internal/model"
)

type ReportService struct {
	DB *pgxpool.Pool
}

type SalesSummary struct {
	TotalOrders   int `json:"total_orders"`
	TotalRevenue  int `json:"total_revenue"`
	AverageTicket int `json:"average_ticket"`
}

type TopProduct struct {
	Name     string `json:"name"`
	Quantity int    `json:"quantity"`
	Revenue  int    `json:"revenue"`
}

type DailySale struct {
	Date    string `json:"date"`
	Revenue int    `json:"revenue"`
	Orders  int    `json:"orders"`
}

func NewReportService(db *pgxpool.Pool) *ReportService {
	return &ReportService{DB: db}
}

func (s *ReportService) SalesSummary(ctx context.Context, from, to time.Time) (*SalesSummary, error) {
	var summary SalesSummary
	err := s.DB.QueryRow(ctx,
		`SELECT 
			COUNT(*) as total_orders,
			COALESCE(SUM(total_cents), 0) as total_revenue,
			CASE WHEN COUNT(*) > 0 THEN COALESCE(SUM(total_cents), 0) / COUNT(*) ELSE 0 END as average_ticket
		FROM orders 
		WHERE status = 'completed' AND created_at >= $1 AND created_at <= $2`,
		from, to,
	).Scan(&summary.TotalOrders, &summary.TotalRevenue, &summary.AverageTicket)
	return &summary, err
}

func (s *ReportService) TopProducts(ctx context.Context, from, to time.Time, limit int) ([]TopProduct, error) {
	rows, err := s.DB.Query(ctx,
		`SELECT 
			oi.product_name,
			SUM(oi.quantity) as quantity,
			SUM(oi.total_cents) as revenue
		FROM order_items oi
		JOIN orders o ON oi.order_id = o.id
		WHERE o.status = 'completed' AND o.created_at >= $1 AND o.created_at <= $2
		GROUP BY oi.product_name
		ORDER BY revenue DESC
		LIMIT $3`,
		from, to, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []TopProduct
	for rows.Next() {
		var p TopProduct
		if err := rows.Scan(&p.Name, &p.Quantity, &p.Revenue); err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	return products, rows.Err()
}

func (s *ReportService) LowStock(ctx context.Context, threshold int) ([]model.Product, error) {
	rows, err := s.DB.Query(ctx,
		`SELECT id, name, price_cents, category, stock, is_available, attributes, created_at, updated_at
		FROM products 
		WHERE stock <= $1 AND is_available = true
		ORDER BY stock ASC`,
		threshold,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []model.Product
	for rows.Next() {
		var p model.Product
		if err := rows.Scan(&p.ID, &p.Name, &p.PriceCents, &p.Category, &p.Stock, &p.IsAvailable, &p.Attributes, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	return products, rows.Err()
}

func (s *ReportService) DailySales(ctx context.Context, days int) ([]DailySale, error) {
	rows, err := s.DB.Query(ctx,
		`SELECT 
			to_char(created_at, 'YYYY-MM-DD') as date,
			COALESCE(SUM(total_cents), 0) as revenue,
			COUNT(*) as orders
		FROM orders 
		WHERE status = 'completed' AND created_at >= now() - ($1 || ' days')::interval
		GROUP BY date
		ORDER BY date DESC`,
		days,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sales []DailySale
	for rows.Next() {
		var s DailySale
		if err := rows.Scan(&s.Date, &s.Revenue, &s.Orders); err != nil {
			return nil, err
		}
		sales = append(sales, s)
	}
	return sales, rows.Err()
}
