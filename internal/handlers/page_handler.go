package handlers

import (
	"html/template"
	"net/http"
)

// PageHandler - handler dla stron WWW
type PageHandler struct {
	templates *template.Template
}

// NewPageHandler - tworzy nowy handler stron
func NewPageHandler() *PageHandler {
	// Załaduj szablony HTML
	tmpl := template.Must(template.ParseGlob("web/templates/*.html"))
	
	return &PageHandler{
		templates: tmpl,
	}
}

// Index - renderuje stronę główną
func (h *PageHandler) Index(w http.ResponseWriter, r *http.Request) {
	err := h.templates.ExecuteTemplate(w, "index.html", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}