package pipeline

import (
	"context"
	"log/slog"

	"github.com/Benevanio/Jobs_Scraper_Global/scraper-go/internal/adapters"
	"github.com/Benevanio/Jobs_Scraper_Global/scraper-go/internal/dedup"
	"github.com/Benevanio/Jobs_Scraper_Global/scraper-go/internal/models"
)

type SearchConfig struct {
	Keywords       []string `json:"keywords"`
	SearchLocation string   `json:"searchLocation"`
	JobTypes       string   `json:"jobTypes"`
	TimeFilter     string   `json:"timeFilter"`
	RemoteOnly     bool     `json:"remoteOnly"`
}

func ScrapeAllSources(
	ctx context.Context,
	config SearchConfig,
) ([]models.Job, error) {
	slog.Info("starting scrape", "keywords", config.Keywords)

	adapterList := adapters.GetAdapters()

	req := models.ScrapeRequest{
		Keywords:       config.Keywords,
		SearchLocation: config.SearchLocation,
		JobTypes:       config.JobTypes,
		TimeFilter:     config.TimeFilter,
		RemoteOnly:     config.RemoteOnly,
	}

	jobs := Run(ctx, adapterList, req)

	deduped := dedup.DedupeJobs(jobs)

	slog.Info("scrape finished",
		"raw_jobs", len(jobs),
		"deduped_jobs", len(deduped),
	)

	return deduped, nil
}
