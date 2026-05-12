package main

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/Benevanio/Jobs_Scraper_Global/scraper-go/internal/adapters"
	"github.com/Benevanio/Jobs_Scraper_Global/scraper-go/internal/keywords"
	"github.com/Benevanio/Jobs_Scraper_Global/scraper-go/internal/models"
	"github.com/Benevanio/Jobs_Scraper_Global/scraper-go/internal/pipeline"
)

func handleScrape(adapterList []adapters.Adapter, kwStore *keywords.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req models.ScrapeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid json body", http.StatusBadRequest)
			return
		}

		// --- LÓGICA DE CARREGAMENTO AUTOMÁTICO ---
		if len(req.Keywords) == 0 {
			// Se o usuário não enviou keywords, carregamos do JSON/Redis
			kws, err := kwStore.Load(r.Context())
			if err != nil || len(kws) == 0 {
				http.Error(w, "falha ao carregar keywords do sistema", http.StatusInternalServerError)
				return
			}
			req.Keywords = kws
		}
		// ------------------------------------------

		start := time.Now()

		// Com 120 keywords, 5 minutos pode ser apertado, considere 10 se necessário
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Minute)
		defer cancel()

		jobs := pipeline.Run(ctx, adapterList, req)

		duration := time.Since(start)
		printSummary(len(adapterList), req.Keywords, len(jobs), duration)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(models.ScrapeResponse{
			Jobs:     jobs,
			Total:    len(jobs),
			CachedAt: time.Now().UTC().Format(time.RFC3339),
		})
	}
}

func handleHealth() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	}
}

func handleGetKeywords(kwStore *keywords.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		kws, err := kwStore.Load(r.Context())
		if err != nil {
			http.Error(w, "erro ao carregar keywords", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"keywords": kws,
			"total":    len(kws),
		})
	}
}

func handleSaveKeywords(kwStore *keywords.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var body struct {
			Keywords []string `json:"keywords"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}

		if err := kwStore.Save(r.Context(), body.Keywords); err != nil {
			http.Error(w, "erro ao salvar keywords", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"ok":       true,
			"keywords": body.Keywords,
		})
	}
}
