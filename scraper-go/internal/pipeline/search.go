package pipeline

import (
	"context"
	"time"

	"github.com/Benevanio/Jobs_Scraper_Global/scraper-go/internal/cache"
	"github.com/Benevanio/Jobs_Scraper_Global/scraper-go/internal/inflight"
	"github.com/Benevanio/Jobs_Scraper_Global/scraper-go/internal/models"
)

type SearchResult struct {
	Jobs      []models.Job `json:"jobs"`
	Total     int          `json:"total"`
	CachedAt  time.Time    `json:"cachedAt"`
	FromCache bool         `json:"fromCache"`
}

var searchCache = cache.NewMemoryCache()

func SearchJobs(
	ctx context.Context,
	config SearchConfig,
	ttl time.Duration,
) (SearchResult, error) {
	cacheKey := BuildCacheKey(config)

	var cached SearchResult
	if found, _ := searchCache.Get(cacheKey, &cached); found {
		cached.FromCache = true
		return cached, nil
	}

	return inflight.Do(cacheKey, func() (SearchResult, error) {
		var cached SearchResult
		if found, _ := searchCache.Get(cacheKey, &cached); found {
			cached.FromCache = true
			return cached, nil
		}

		jobs, err := ScrapeAllSources(ctx, config)
		if err != nil {
			return SearchResult{}, err
		}

		result := SearchResult{
			Jobs:      jobs,
			Total:     len(jobs),
			CachedAt:  time.Now(),
			FromCache: false,
		}

		_ = searchCache.Set(cacheKey, result, ttl)
		return result, nil
	})
}
