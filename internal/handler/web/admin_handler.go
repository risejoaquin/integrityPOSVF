package web

import (
	"html/template"
	"net/http"

	"github.com/solidbit/integritypos/internal/config"
	"github.com/solidbit/integritypos/internal/service"
)

type AdminHandler struct {
	productSvc *service.ProductService
	cfg        *config.Config
	tmpl       *template.Template
}

func NewAdminHandler(mux *http.ServeMux, productSvc *service.ProductService, cfg *config.Config, tmpl *template.Template) {
	h := &AdminHandler{
		productSvc: productSvc,
		cfg:        cfg,
		tmpl:       tmpl,
	}
	mux.HandleFunc("GET /admin", h.serveAdmin)
}

func (h *AdminHandler) serveAdmin(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"Config": h.cfg,
	}

	if err := h.tmpl.ExecuteTemplate(w, "admin.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
