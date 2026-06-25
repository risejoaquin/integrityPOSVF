package printer

import (
	"log"

	"github.com/solidbit/integritypos/internal/model"
)

type NoopPrinter struct{}

func NewNoopPrinter() *NoopPrinter {
	return &NoopPrinter{}
}

func (p *NoopPrinter) PrintTicket(order model.Order) error {
	log.Printf("NoopPrinter: PrintTicket called for order %s", order.ID)
	return nil
}

func (p *NoopPrinter) OpenDrawer() error {
	log.Printf("NoopPrinter: OpenDrawer called")
	return nil
}
