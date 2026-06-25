package service

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/solidbit/integritypos/internal/model"
)

type SalesSummary struct {
	TotalOrders   int
	TotalRevenue  int // cents
	AverageTicket int // cents
}

type TopProduct struct {
	Name     string
	Quantity int
	Revenue  int
}

type DailySale struct {
	Date    string
	Revenue int
	Orders  int
}

type ReportService struct {
	db *pgxpool.Pool
}

func NewReportService(db *pgxpool.Pool) *ReportService {
	return &ReportService{db: db}
}

func (s *ReportService) SalesSummary(ctx context.Context, dateFrom, dateTo time.Time) (SalesSummary, error) {
	query := `
		SELECT 
			COUNT(id) as total_orders, 
			COALESCE(SUM(total_cents), 0) as total_revenue 
		FROM orders 
		WHERE status = 'confirmed' 
		  AND created_at >= $1 
		  AND created_at <= $2
	`
	var summary SalesSummary
	err := s.db.QueryRow(ctx, query, dateFrom, dateTo).Scan(&summary.TotalOrders, &summary.TotalRevenue)
	if err != nil {
		return summary, err
	}
	if summary.TotalOrders > 0 {
		summary.AverageTicket = summary.TotalRevenue / summary.TotalOrders
	}
	return summary, nil
}

func (s *ReportService) TopProducts(ctx context.Context, dateFrom, dateTo time.Time, limit int) ([]TopProduct, error) {
	query := `
		SELECT 
			product_name, 
			SUM(quantity) as quantity, 
			SUM(price_cents * quantity) as revenue
		FROM order_items
		JOIN orders ON order_items.order_id = orders.id
		WHERE orders.status = 'confirmed'
		  AND orders.created_at >= $1 
		  AND orders.created_at <= $2
		GROUP BY product_name
		ORDER BY quantity DESC
		LIMIT $3
	`
	rows, err := s.db.Query(ctx, query, dateFrom, dateTo, limit)
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
	return products, nil
}

func (s *ReportService) LowStock(ctx context.Context, threshold int) ([]model.Product, error) {
	query := `
		SELECT id, name, price_cents, category_id, sku, barcode, track_inventory, is_available 
		FROM products 
		WHERE track_inventory = true 
		  AND id IN (
			  SELECT product_id FROM inventory_transactions 
			  GROUP BY product_id 
			  HAVING SUM(quantity) <= $1
		  )
	`
	rows, err := s.db.Query(ctx, query, threshold)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []model.Product
	for rows.Next() {
		var p model.Product
		if err := rows.Scan(&p.ID, &p.Name, &p.PriceCents, &p.CategoryID, &p.SKU, &p.Barcode, &p.TrackInventory, &p.IsAvailable); err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	return products, nil
}

func (s *ReportService) DailySales(ctx context.Context, days int) ([]DailySale, error) {
	query := `
		SELECT 
			DATE(created_at) as date, 
			COALESCE(SUM(total_cents), 0) as revenue, 
			COUNT(id) as orders
		FROM orders 
		WHERE status = 'confirmed' 
		  AND created_at >= CURRENT_DATE - INTERVAL '1 day' * $1
		GROUP BY DATE(created_at)
		ORDER BY DATE(created_at) ASC
	`
	rows, err := s.db.Query(ctx, query, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sales []DailySale
	for rows.Next() {
		var d DailySale
		var t time.Time
		if err := rows.Scan(&t, &d.Revenue, &d.Orders); err != nil {
			return nil, err
		}
		d.Date = t.Format("2006-01-02")
		sales = append(sales, d)
	}
	return sales, nil
}
