package printer

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/solidbit/integritypos/internal/model"
)

type TCPPrinter struct {
	mu           sync.Mutex
	addr         string
	businessName string
}

func NewTCPPrinter(addr, businessName string) *TCPPrinter {
	return &TCPPrinter{addr: addr, businessName: businessName}
}

func (p *TCPPrinter) PrintTicket(order model.Order) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	conn, err := net.DialTimeout("tcp", p.addr, 5*time.Second)
	if err != nil {
		return fmt.Errorf("failed to connect to printer: %w", err)
	}
	defer conn.Close()

	return printTicketToWriter(conn, order, p.businessName)
}

func (p *TCPPrinter) OpenDrawer() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	conn, err := net.DialTimeout("tcp", p.addr, 5*time.Second)
	if err != nil {
		return fmt.Errorf("failed to connect to printer: %w", err)
	}
	defer conn.Close()

	_, err = conn.Write([]byte{0x1b, 0x70, 0x00, 0x32, 0x32})
	return err
}

func (p *TCPPrinter) Ping() error {
	conn, err := net.DialTimeout("tcp", p.addr, 2*time.Second)
	if err != nil {
		return err
	}
	conn.Close()
	return nil
}
