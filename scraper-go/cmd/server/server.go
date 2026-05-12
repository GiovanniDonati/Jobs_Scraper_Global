// package main

// import (
// 	"context"
// 	"errors"
// 	"log/slog"
// 	"net/http"
// 	"os"
// 	"os/signal"
// 	"syscall"
// 	"time"

// 	"github.com/Benevanio/Jobs_Scraper_Global/scraper-go/internal/adapters"
// )

// func run(adapterList []adapters.Adapter) {
// 	addr := os.Getenv("GO_SCRAPER_ADDR")
// 	if addr == "" {
// 		addr = ":8081"
// 	}

// 	mux := http.NewServeMux()
// 	mux.Handle("POST /scrape", handleScrape(adapterList))
// 	mux.Handle("GET /health", handleHealth())

// 	srv := &http.Server{
// 		Addr:         addr,
// 		Handler:      mux,
// 		ReadTimeout:  10 * time.Second,
// 		WriteTimeout: 6 * time.Minute,
// 		IdleTimeout:  60 * time.Second,
// 	}

// 	stop := make(chan os.Signal, 1)
// 	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

// 	go func() {
// 		slog.Info("go-scraper em execução", "addr", addr)
// 		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
// 			slog.Error("falha crítica no servidor", "error", err)
// 			os.Exit(1)
// 		}
// 	}()

// 	<-stop
// 	slog.Info("sinal de interrupção recebido, desligando...")

// 	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
// 	defer cancel()

// 	if err := srv.Shutdown(ctx); err != nil {
// 		slog.Error("erro durante o shutdown", "error", err)
// 	}

// 	slog.Info("servidor encerrado com sucesso")
// }

package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Benevanio/Jobs_Scraper_Global/scraper-go/internal/adapters"
	"github.com/Benevanio/Jobs_Scraper_Global/scraper-go/internal/keywords"
	"github.com/redis/go-redis/v9"
)

func run(adapterList []adapters.Adapter) {
	addr := os.Getenv("GO_SCRAPER_ADDR")
	if addr == "" {
		addr = ":8081"
	}

	// 1. Inicializa o Redis (necessário para o kwStore)
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "redis:6379"
	}
	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	// 2. Instancia a Store de keywords (agora no novo caminho)
	kwStore := keywords.NewStore(rdb)

	mux := http.NewServeMux()

	// 3. CORREÇÃO: Passando o kwStore como segundo argumento
	mux.Handle("POST /scrape", handleScrape(adapterList, kwStore))
	mux.Handle("GET /health", handleHealth())
	mux.Handle("GET /api/keywords", handleGetKeywords(kwStore))
	mux.Handle("POST /api/keywords", handleSaveKeywords(kwStore))

	srv := &http.Server{
		Addr:        addr,
		Handler:     mux,
		ReadTimeout: 10 * time.Second,
		// Aumentei um pouco para aguentar as 120+ keywords
		WriteTimeout: 10 * time.Minute,
		IdleTimeout:  60 * time.Second,
	}

	// ... (restante do código de shutdown continua igual)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		slog.Info("go-scraper em execução", "addr", addr)
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			slog.Error("falha crítica no servidor", "error", err)
			os.Exit(1)
		}
	}()

	<-stop
	slog.Info("sinal de interrupção recebido, desligando...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("erro durante o shutdown", "error", err)
	}

	slog.Info("servidor encerrado com sucesso")
}
