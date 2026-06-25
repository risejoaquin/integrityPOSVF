package webhook

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/solidbit/integritypos/internal/config"
	"github.com/solidbit/integritypos/internal/service"
)

type WhatsAppWebhook struct {
	cfg         config.WhatsAppConfig
	whatsappSvc *service.WhatsAppService
}

func NewWhatsAppWebhook(mux *http.ServeMux, cfg config.WhatsAppConfig, whatsappSvc *service.WhatsAppService) {
	if !cfg.Enabled {
		return
	}
	h := &WhatsAppWebhook{
		cfg:         cfg,
		whatsappSvc: whatsappSvc,
	}
	mux.HandleFunc("GET /webhook/whatsapp", h.verifyWebhook)
	mux.HandleFunc("POST /webhook/whatsapp", h.handleWebhook)
}

func (h *WhatsAppWebhook) verifyWebhook(w http.ResponseWriter, r *http.Request) {
	mode := r.URL.Query().Get("hub.mode")
	token := r.URL.Query().Get("hub.verify_token")
	challenge := r.URL.Query().Get("hub.challenge")

	if mode == "subscribe" && token == h.cfg.VerifyToken {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(challenge))
	} else {
		w.WriteHeader(http.StatusForbidden)
	}
}

func (h *WhatsAppWebhook) handleWebhook(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Object string `json:"object"`
		Entry  []struct {
			Changes []struct {
				Value struct {
					MessagingProduct string `json:"messaging_product"`
					Messages         []struct {
						From        string `json:"from"`
						ID          string `json:"id"`
						Type        string `json:"type"`
						Text        struct {
							Body string `json:"body"`
						} `json:"text"`
						Interactive struct {
							ListReply struct {
								ID string `json:"id"`
							} `json:"list_reply"`
						} `json:"interactive"`
					} `json:"messages"`
				} `json:"value"`
			} `json:"changes"`
		} `json:"entry"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)

	go func() {
		ctx := context.Background()
		for _, entry := range payload.Entry {
			for _, change := range entry.Changes {
				for _, msg := range change.Value.Messages {
					if msg.Type == "text" {
						if err := h.whatsappSvc.ProcessIncomingMessage(ctx, msg.From, msg.Text.Body); err != nil {
							log.Printf("Error processing text message: %v", err)
						}
					} else if msg.Type == "interactive" {
						if err := h.whatsappSvc.HandleInteractiveReply(ctx, msg.From, msg.Interactive.ListReply.ID); err != nil {
							log.Printf("Error processing interactive reply: %v", err)
						}
					}
				}
			}
		}
	}()
}
