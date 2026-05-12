package main

import (
	"log/slog"
	"os"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	adapterList := buildAdapters()
	slog.Info("servidor inicializado", "adapters_total", len(adapterList))

	run(adapterList)
}
