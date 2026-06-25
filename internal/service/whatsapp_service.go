package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/solidbit/integritypos/internal/integration/whatsapp"
	"github.com/solidbit/integritypos/internal/repository"
)

type WhatsAppService struct {
	whatsappClient *whatsapp.Client
	aiService      *AIService
	orderService   *OrderService
	customerRepo   *repository.CustomerRepository
	msgRepo        *repository.WhatsAppMessageRepo
	productRepo    *repository.ProductRepository
}

func NewWhatsAppService(
	whatsappClient *whatsapp.Client,
	aiService *AIService,
	orderService *OrderService,
	customerRepo *repository.CustomerRepository,
	msgRepo *repository.WhatsAppMessageRepo,
	productRepo *repository.ProductRepository,
) *WhatsAppService {
	return &WhatsAppService{
		whatsappClient: whatsappClient,
		aiService:      aiService,
		orderService:   orderService,
		customerRepo:   customerRepo,
		msgRepo:        msgRepo,
		productRepo:    productRepo,
	}
}

func (s *WhatsAppService) ProcessIncomingMessage(ctx context.Context, from, text string) error {
	if err := s.msgRepo.SaveMessage(ctx, "inbound", from, s.whatsappClient.GetPhoneNumberID(), text); err != nil {
		return fmt.Errorf("failed to save inbound message: %w", err)
	}

	draft, err := s.aiService.InterpretMessage(ctx, text)
	if err != nil {
		return fmt.Errorf("failed to interpret message: %w", err)
	}

	if len(draft.Items) == 0 {
		reply := "Lo siento, no pude entender qué productos deseas. ¿Podrías ser más específico?"
		s.whatsappClient.SendTextMessage(ctx, from, reply)
		s.msgRepo.SaveMessage(ctx, "outbound", s.whatsappClient.GetPhoneNumberID(), from, reply)
		return nil
	}

	cart := Cart{Items: make([]CartItem, 0)}
	var itemDescriptions []string

	for _, dItem := range draft.Items {
		// Fuzzy search in product repo
		filter := repository.ProductFilter{Search: dItem.ProductName, AvailableOnly: true}
		products, err := s.productRepo.List(ctx, filter)
		if err != nil || len(products) == 0 {
			// Try without strict matching, maybe first word
			words := strings.Fields(dItem.ProductName)
			if len(words) > 0 {
				filter.Search = words[0]
				products, _ = s.productRepo.List(ctx, filter)
			}
		}

		if len(products) > 0 {
			p := products[0]
			cart.Items = append(cart.Items, CartItem{
				ProductID: p.ID,
				Quantity:  dItem.Quantity,
			})
			itemDescriptions = append(itemDescriptions, fmt.Sprintf("%d x %s ($%.2f)", dItem.Quantity, p.Name, float64(p.PriceCents)/100.0))
		}
	}

	if len(cart.Items) == 0 {
		reply := "No pude encontrar los productos que solicitaste en nuestro menú. ¿Podrías intentar de nuevo?"
		s.whatsappClient.SendTextMessage(ctx, from, reply)
		s.msgRepo.SaveMessage(ctx, "outbound", s.whatsappClient.GetPhoneNumberID(), from, reply)
		return nil
	}

	// Obtener o crear cliente
	custInfo := CustomerInfo{Phone: from}
	customer, err := s.customerRepo.GetByPhone(ctx, from)
	if err == nil {
		custInfo.Name = customer.Name
	}

	order, err := s.orderService.CreateOrder(ctx, cart, "whatsapp", custInfo)
	if err != nil {
		return fmt.Errorf("failed to create order: %w", err)
	}

	body := fmt.Sprintf("He registrado tu pedido:\n%s\nTotal: $%.2f\n\n¿Deseas confirmarlo?", strings.Join(itemDescriptions, "\n"), float64(order.TotalCents)/100.0)
	options := []whatsapp.InteractiveOption{
		{ID: "confirm_" + order.ID, Title: "Confirmar pedido", Description: "Enviar a preparación"},
		{ID: "cancel_" + order.ID, Title: "Cancelar pedido", Description: "Anular el pedido"},
	}

	if err := s.whatsappClient.SendInteractiveList(ctx, from, "Confirmación de Pedido", body, "Responde para continuar", options); err != nil {
		return fmt.Errorf("failed to send interactive list: %w", err)
	}
	s.msgRepo.SaveMessage(ctx, "outbound", s.whatsappClient.GetPhoneNumberID(), from, body)

	return nil
}

func (s *WhatsAppService) HandleInteractiveReply(ctx context.Context, from, buttonID string) error {
	if err := s.msgRepo.SaveMessage(ctx, "inbound", from, s.whatsappClient.GetPhoneNumberID(), "Interactive Reply: "+buttonID); err != nil {
		return fmt.Errorf("failed to save inbound message: %w", err)
	}

	var reply string
	if strings.HasPrefix(buttonID, "confirm_") {
		orderID := strings.TrimPrefix(buttonID, "confirm_")
		if err := s.orderService.ConfirmOrder(ctx, orderID); err != nil {
			reply = "Hubo un error al confirmar tu pedido. Por favor contáctanos."
		} else {
			reply = "¡Tu pedido ha sido confirmado y está en preparación! Gracias por tu compra."
		}
	} else if strings.HasPrefix(buttonID, "cancel_") {
		orderID := strings.TrimPrefix(buttonID, "cancel_")
		if err := s.orderService.CancelOrder(ctx, orderID); err != nil {
			reply = "Hubo un error al cancelar tu pedido."
		} else {
			reply = "Tu pedido ha sido cancelado exitosamente."
		}
	} else {
		reply = "No entendí esa opción."
	}

	s.whatsappClient.SendTextMessage(ctx, from, reply)
	s.msgRepo.SaveMessage(ctx, "outbound", s.whatsappClient.GetPhoneNumberID(), from, reply)
	return nil
}

func (s *WhatsAppService) SendOrderConfirmation(ctx context.Context, to, orderID string) error {
	order, err := s.orderService.GetOrder(ctx, orderID)
	if err != nil {
		return err
	}
	msg := fmt.Sprintf("Tu orden %s está confirmada. Total: $%.2f", order.ID, float64(order.TotalCents)/100.0)
	return s.whatsappClient.SendTextMessage(ctx, to, msg)
}
