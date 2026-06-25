package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/solidbit/integritypos/internal/integration/gemini"
	"github.com/solidbit/integritypos/internal/repository"
)

type AIService struct {
	geminiClient *gemini.Client
	productRepo  *repository.ProductRepository
	businessName string
}

func NewAIService(geminiClient *gemini.Client, productRepo *repository.ProductRepository, businessName string) *AIService {
	return &AIService{
		geminiClient: geminiClient,
		productRepo:  productRepo,
		businessName: businessName,
	}
}

type OrderDraftItem struct {
	ProductName string `json:"product_name"`
	Quantity    int    `json:"quantity"`
}

type OrderDraft struct {
	Items               []OrderDraftItem `json:"items"`
	SpecialInstructions string           `json:"special_instructions,omitempty"`
}

func (s *AIService) InterpretMessage(ctx context.Context, message string) (OrderDraft, error) {
	products, err := s.productRepo.List(ctx, repository.ProductFilter{AvailableOnly: true})
	if err != nil {
		return OrderDraft{}, fmt.Errorf("failed to list products: %w", err)
	}

	var productLines []string
	for _, p := range products {
		productLines = append(productLines, fmt.Sprintf("- %s (ID: %s) $%.2f", p.Name, p.ID, float64(p.PriceCents)/100.0))
	}
	productList := strings.Join(productLines, "\n")

	systemPrompt := fmt.Sprintf(`Eres el asistente de pedidos de %s. Disponemos de los siguientes productos:
%s

El cliente enviará un mensaje. Extrae los productos y cantidades y devuelve SOLO un JSON válido con el formato:
{"items":[{"product_name":"nombre exacto","quantity":cantidad}],"special_instructions":"..."}
Si no entiendes algún producto, omítelo o usa special_instructions para indicarlo. Usa exactamente el nombre del producto de la lista proporcionada.`, s.businessName, productList)

	rawJSON, err := s.geminiClient.GenerateJSON(ctx, systemPrompt, message)
	if err != nil {
		return OrderDraft{}, fmt.Errorf("failed to generate JSON from gemini: %w", err)
	}

	var draft OrderDraft
	if err := json.Unmarshal(rawJSON, &draft); err != nil {
		return OrderDraft{}, fmt.Errorf("failed to parse order draft: %w", err)
	}

	return draft, nil
}
