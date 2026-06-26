package printer

import (
	"fmt"
	"os"
	"runtime"
	"sync"

	"github.com/solidbit/integritypos/internal/model"
)

type USBPrinter struct {
	mu           sync.Mutex
	devicePath   string
	businessName string
}

func NewUSBPrinter(devicePath, businessName string) *USBPrinter {
	return &USBPrinter{devicePath: devicePath, businessName: businessName}
}

func (p *USBPrinter) PrintTicket(order model.Order) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	f, err := os.OpenFile(p.devicePath, os.O_RDWR, 0)
	if err != nil {
		if runtime.GOOS == "windows" {
			return fmt.Errorf("failed to open printer on Windows (ensure path is like LPT1 or \\\\Computer\\Printer): %w", err)
		}
		return fmt.Errorf("failed to open printer device: %w", err)
	}
	defer f.Close()

	return printTicketToWriter(f, order, p.businessName)
}

func (p *USBPrinter) OpenDrawer() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	f, err := os.OpenFile(p.devicePath, os.O_RDWR, 0)
	if err != nil {
		if runtime.GOOS == "windows" {
			return fmt.Errorf("failed to open printer on Windows: %w", err)
		}
		return fmt.Errorf("failed to open printer device: %w", err)
	}
	defer f.Close()

	_, err = f.Write([]byte{0x1b, 0x70, 0x00, 0x32, 0x32})
	return err
}

func (p *USBPrinter) Ping() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	f, err := os.OpenFile(p.devicePath, os.O_RDWR, 0)
	if err != nil {
		return err
	}
	f.Close()
	return nil
}
