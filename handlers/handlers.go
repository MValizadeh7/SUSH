package handlers

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"net/url"
	"sush/store"
	"time"
)

type Handler struct {
	Store   *store.Store
	BaseURL string
}

type ShortenRequest struct {
	URL string `json:"url"`
}

type ShortenResponse struct {
	ShortURL string `json:"short_url"`
}

type StatsResponse struct {
	Clicks    int    `json:"clicks"`
	CreatedAt string `json:"created_at"`
}

const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// generateSlug creates a random 6-character slug.
func generateSlug() (string, error) {
	slug := make([]byte, 6)
	for i := range slug {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		if err != nil {
			return "", err
		}
		slug[i] = chars[n.Int64()]
	}
	return string(slug), nil
}

func New(s *store.Store, baseURL string) *Handler {
	return &Handler{Store: s, BaseURL: baseURL}
}

func (h *Handler) HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h *Handler) ShortenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ShortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request body", http.StatusBadRequest)
		return
	}

	if req.URL == "" {
		http.Error(w, "url is required", http.StatusBadRequest)
		return
	}

	u, err := url.Parse(req.URL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		http.Error(w, "bad url", http.StatusBadRequest)
		return
	}

	slug, err := generateSlug()
	if err != nil {
		http.Error(w, "failed to generate slug", http.StatusInternalServerError)
		return
	}

	if err := h.Store.Save(slug, req.URL); err != nil {
		http.Error(w, "failed to save url", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(ShortenResponse{
		ShortURL: h.BaseURL + "/" + slug,
	})
}

func (h *Handler) RedirectHandler(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")

	original, err := h.Store.Get(slug)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	h.Store.IncrementClicks(slug)
	http.Redirect(w, r, original, http.StatusFound)
}

func (h *Handler) StatsHandler(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")

	clicks, createdAt, err := h.Store.GetStats(slug)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(StatsResponse{
		Clicks:    clicks,
		CreatedAt: createdAt.Format(time.RFC3339),
	})
}

func (h *Handler) MetricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("# HELP sush_urls_total Total shortened URLs\n# TYPE sush_urls_total gauge\nsush_urls_total " + fmt.Sprintf("%d", h.Store.Count()) + "\n"))
}
