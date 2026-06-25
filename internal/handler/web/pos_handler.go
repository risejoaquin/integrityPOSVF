package web

import (
	"encoding/json"
	"html/template"
	"net/http"

	"github.com/solidbit/integritypos/internal/config"
	"github.com/solidbit/integritypos/internal/module"
	"github.com/solidbit/integritypos/internal/repository"
	"github.com/solidbit/integritypos/internal/service"
)

type POSHandler struct {
	productSvc *service.ProductService
	cfg        *config.Config
	tmpl       *template.Template
}

func NewPOSHandler(mux *http.ServeMux, productSvc *service.ProductService, cfg *config.Config, tmpl *template.Template) {
	h := &POSHandler{
		productSvc: productSvc,
		cfg:        cfg,
		tmpl:       tmpl,
	}
	mux.HandleFunc("GET /pos", h.servePOS)
}

func (h *POSHandler) servePOS(w http.ResponseWriter, r *http.Request) {
	filter := repository.ProductFilter{AvailableOnly: true}
	products, err := h.productSvc.List(r.Context(), filter)
	if err != nil {
		http.Error(w, "Error fetching products", http.StatusInternalServerError)
		return
	}

	productsJSON, _ := json.Marshal(products)

	modulesMap := map[string]bool{
		"tables":  module.IsActive("tables"),
		"loyalty": module.IsActive("loyalty"),
	}

	data := map[string]interface{}{
		"Products":     products,
		"ProductsJSON": template.JS(productsJSON),
		"Config":       h.cfg,
		"Modules":      modulesMap,
	}

	if err := h.tmpl.ExecuteTemplate(w, "pos.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
