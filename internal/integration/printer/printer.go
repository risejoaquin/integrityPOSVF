package printer

import (
	"fmt"
	"io"
	"time"

	"github.com/solidbit/integritypos/internal/config"
	"github.com/solidbit/integritypos/internal/model"
)

type Printer interface {
	PrintTicket(order model.Order) error
	OpenDrawer() error
}

func NewPrinter(cfg config.PrinterConfig, businessName string) (Printer, error) {
	switch cfg.Type {
	case "usb":
		return NewUSBPrinter(cfg.DevicePath, businessName), nil
	case "tcp":
		return NewTCPPrinter(cfg.TCPAddr, businessName), nil
	case "none", "":
		return NewNoopPrinter(), nil
	default:
		return nil, fmt.Errorf("unsupported printer type: %s", cfg.Type)
	}
}

func printTicketToWriter(w io.Writer, order model.Order, businessName string) error {
	w.Write([]byte{0x1b, 0x40})
	w.Write([]byte{0x1b, 0x61, 0x01})
	w.Write([]byte(fmt.Sprintf("%s\n\n", businessName)))
	
	w.Write([]byte(time.Now().Format("2006-01-02 15:04:05") + "\n"))
	w.Write([]byte(fmt.Sprintf("Orden: %s\n", order.ID)))
	w.Write([]byte("--------------------------------\n"))
	w.Write([]byte{0x1b, 0x61, 0x00})

	for _, item := range order.Items {
		w.Write([]byte(fmt.Sprintf("%d x %s\n", item.Quantity, item.ProductName)))
		w.Write([]byte(fmt.Sprintf("  $%.2f\n", float64(item.TotalCents)/100.0)))
	}
	w.Write([]byte("--------------------------------\n"))
	w.Write([]byte{0x1b, 0x61, 0x01})
	w.Write([]byte(fmt.Sprintf("TOTAL: $%.2f\n\n", float64(order.TotalCents)/100.0)))

	w.Write([]byte("¡Gracias por su compra!\n\n\n\n"))
	w.Write([]byte{0x1d, 0x56, 0x42, 0x01})
	w.Write([]byte{0x1b, 0x70, 0x00, 0x32, 0x32})

	return nil
}
