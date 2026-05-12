package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Benevanio/Jobs_Scraper_Global/scraper-go/internal/models"
)

// TheMuseAdapter espelha o theMuseAdapter do JS.
type TheMuseAdapter struct {
	client *http.Client
}

func NewTheMuse() *TheMuseAdapter {
	return &TheMuseAdapter{
		client: &http.Client{Timeout: 60 * time.Second},
	}
}

func (a *TheMuseAdapter) SourceName() string { return "The Muse" }

// buildTheMuseUrl espelha exatamente o buildTheMuseUrl do JS.
func buildTheMuseURL(keyword string, req models.ScrapeRequest, page int) string {
	u, _ := url.Parse("https://www.themuse.com/api/public/jobs")
	q := u.Query()

	q.Set("page", fmt.Sprintf("%d", page))
	q.Set("descending", "true")

	if keyword != "" {
		q.Set("category", keyword)
	}

	if req.SearchLocation != "" {
		q.Set("location", req.SearchLocation)
	}

	u.RawQuery = q.Encode()
	return u.String()
}

type theMuseResponse struct {
	Results []struct {
		Name    string `json:"name"`
		Company struct {
			Name string `json:"name"`
		} `json:"company"`
		Refs struct {
			LandingPage string `json:"landing_page"`
		} `json:"refs"`
		Locations []struct {
			Name string `json:"name"`
		} `json:"locations"`
		PublicationDate string `json:"publication_date"`
	} `json:"results"`
}

func (a *TheMuseAdapter) Search(ctx context.Context, keyword string, req models.ScrapeRequest) ([]models.Job, error) {
	maxPages := req.MaxPagesPerKeyword
	if maxPages <= 0 {
		maxPages = 3
	}

	pageTimeout := time.Duration(req.PageTimeoutMs) * time.Millisecond
	if pageTimeout <= 0 {
		pageTimeout = 15 * time.Second
	}

	waitBetween := time.Duration(req.WaitBetweenSearchesMs) * time.Millisecond
	if waitBetween <= 0 {
		waitBetween = 1000 * time.Millisecond
	}

	var allJobs []models.Job

	for page := 1; page <= maxPages; page++ {
		endpoint := buildTheMuseURL(keyword, req, page)

		pageCtx, cancel := context.WithTimeout(ctx, pageTimeout)
		jobs, err := a.fetchPage(pageCtx, endpoint, keyword)
		cancel()

		if err != nil {
			return nil, err
		}

		if len(jobs) == 0 {
			// Espelha o break do JS quando results está vazio.
			break
		}

		allJobs = append(allJobs, jobs...)

		if page < maxPages {
			select {
			case <-ctx.Done():
				return allJobs, nil
			case <-time.After(waitBetween):
			}
		}
	}

	return allJobs, nil
}

func (a *TheMuseAdapter) fetchPage(ctx context.Context, endpoint, keyword string) ([]models.Job, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	resp, err := a.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}

	var data theMuseResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	if len(data.Results) == 0 {
		return []models.Job{}, nil
	}

	jobs := make([]models.Job, 0, len(data.Results))
	for _, r := range data.Results {
		// Espelha: Array.isArray(job.locations) ? job.locations.map(i => i.name).join(", ") : ""
		locationParts := make([]string, 0, len(r.Locations))
		for _, loc := range r.Locations {
			locationParts = append(locationParts, loc.Name)
		}
		local := strings.Join(locationParts, ", ")

		jobs = append(jobs, models.Job{
			ID:       strings.TrimSpace(r.Refs.LandingPage),
			Title:    strings.TrimSpace(r.Name),
			Company:  strings.TrimSpace(r.Company.Name),
			Location: local,
			URL:      strings.TrimSpace(r.Refs.LandingPage),
			PostedAt: r.PublicationDate,
			Source:   "The Muse",
			Sources:  []string{"The Muse"},
			Keyword:  keyword,
			Keywords: []string{keyword},
		})
	}

	return jobs, nil
}
