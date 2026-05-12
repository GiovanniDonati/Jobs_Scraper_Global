package pipeline

import (
	"context"
	"log/slog"
	"sync"

	"github.com/Benevanio/Jobs_Scraper_Global/scraper-go/internal/adapters"
	"github.com/Benevanio/Jobs_Scraper_Global/scraper-go/internal/dedup"
	"github.com/Benevanio/Jobs_Scraper_Global/scraper-go/internal/models"
)

const defaultMaxConcurrency = 20

type result struct {
	jobs []models.Job
	err  error
}

// Run executa todas as combinações adapter × keyword em paralelo,
// respeitando o limite de concorrência, e devolve as vagas deduplicadas.
func Run(ctx context.Context, adapterList []adapters.Adapter, req models.ScrapeRequest) []models.Job {
	maxConcurrency := req.MaxConcurrency
	if maxConcurrency <= 0 {
		maxConcurrency = defaultMaxConcurrency
	}

	type task struct {
		adapter adapters.Adapter
		keyword string
	}

	tasks := make([]task, 0, len(adapterList)*len(req.Keywords))
	for _, a := range adapterList {
		for _, kw := range req.Keywords {
			tasks = append(tasks, task{adapter: a, keyword: kw})
		}
	}

	sem := make(chan struct{}, maxConcurrency)
	results := make(chan result, len(tasks))
	var wg sync.WaitGroup

	for _, t := range tasks {
		wg.Add(1)
		sem <- struct{}{}

		go func(t task) {
			defer wg.Done()
			defer func() { <-sem }()

			jobs, err := t.adapter.Search(ctx, t.keyword, req)
			results <- result{jobs: jobs, err: err}

			if err != nil {
				slog.Warn("adapter falhou",
					"source", t.adapter.SourceName(),
					"keyword", t.keyword,
					"error", err,
				)
				return
			}

			slog.Info("adapter concluído",
				"source", t.adapter.SourceName(),
				"keyword", t.keyword,
				"count", len(jobs),
			)
		}(t)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var allJobs []models.Job
	for r := range results {
		if r.err == nil {
			allJobs = append(allJobs, r.jobs...)
		}
	}

	return dedup.DedupeJobs(allJobs)
}
