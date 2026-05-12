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

type AdzunaAdapter struct {
	client  *http.Client
	appID   string
	appKey  string
	country string
	// mu      sync.Mutex
	semaphore chan struct{}
}

func NewAdzuna(appID, appKey, country string) *AdzunaAdapter {
	return &AdzunaAdapter{
		client:    &http.Client{Timeout: 60 * time.Second},
		appID:     appID,
		appKey:    appKey,
		country:   strings.ToLower(strings.TrimSpace(country)),
		semaphore: make(chan struct{}, 3),
	}
}

func (a *AdzunaAdapter) SourceName() string {
	return fmt.Sprintf("Adzuna:%s", a.country)
}

// buildURL espelha exatamente o buildAdzunaUrl do JS.
func (a *AdzunaAdapter) buildURL(keyword string, req models.ScrapeRequest, page int) string {
	resultsPerPage := req.ResultsPerPage
	if resultsPerPage <= 0 {
		resultsPerPage = 20
	}

	endpoint := fmt.Sprintf(
		"https://api.adzuna.com/v1/api/jobs/%s/search/%d",
		a.country, page,
	)

	u, _ := url.Parse(endpoint)
	q := u.Query()
	q.Set("app_id", a.appID)
	q.Set("app_key", a.appKey)
	q.Set("results_per_page", fmt.Sprintf("%d", resultsPerPage))
	q.Set("what", keyword)

	if req.SearchLocation != "" {
		q.Set("where", req.SearchLocation)
	}

	// Comentado para espelhar o JS:
	// if req.RemoteOnly {
	// 	q.Set("work_from_home", "1")
	// }

	u.RawQuery = q.Encode()
	return u.String()
}

func (a *AdzunaAdapter) Search(ctx context.Context, keyword string, req models.ScrapeRequest) ([]models.Job, error) {
	// Adzuna é sensível a concorrência, então usamos um canal como semáforo para limitar.
	a.semaphore <- struct{}{}
	defer func() { <-a.semaphore }()

	maxPages := req.MaxPagesPerKeyword
	if maxPages <= 0 {
		maxPages = 3
	}

	// Intervalo entre requisições (aumentamos a segurança)
	waitDuration := time.Duration(req.WaitBetweenSearchesMs) * time.Millisecond
	if waitDuration <= 0 {
		waitDuration = 2000 * time.Millisecond // Adzuna é sensível, 2s é mais seguro
	}

	pageTimeout := time.Duration(req.PageTimeoutMs) * time.Millisecond
	if pageTimeout <= 0 {
		pageTimeout = 15 * time.Second
	}

	var allJobs []models.Job

	for page := 1; page <= maxPages; page++ {
		endpoint := a.buildURL(keyword, req, page)

		pageCtx, cancel := context.WithTimeout(ctx, pageTimeout)
		jobs, err := a.fetchPage(pageCtx, endpoint, keyword)
		cancel()

		if err != nil {
			// Se der erro 400 ou 429, o log detalhado agora sairá no fetchPage
			return nil, fmt.Errorf("adzuna erro na página %d: %w", page, err)
		}

		if len(jobs) == 0 {
			break
		}

		allJobs = append(allJobs, jobs...)

		// Pausa obrigatória entre páginas.
		// O semáforo limita QUANTAS keywords rodam juntas,
		// e este sleep garante o respiro entre as PÁGINAS de cada keyword.
		select {
		case <-ctx.Done():
			return allJobs, ctx.Err()
		case <-time.After(waitDuration):
			// Continua para a próxima página ou libera para a próxima keyword
		}
	}

	return allJobs, nil
}

type adzunaResponse struct {
	Results []struct {
		Title   string `json:"title"`
		Company struct {
			DisplayName string `json:"display_name"`
		} `json:"company"`
		Location struct {
			DisplayName string `json:"display_name"`
		} `json:"location"`
		RedirectURL string  `json:"redirect_url"`
		URL         string  `json:"url"`
		SalaryMin   float64 `json:"salary_min"`
		SalaryMax   float64 `json:"salary_max"`
		Created     string  `json:"created"`
	} `json:"results"`
}

func (a *AdzunaAdapter) fetchPage(ctx context.Context, endpoint, keyword string) ([]models.Job, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}

	var data adzunaResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	if len(data.Results) == 0 {
		return []models.Job{}, nil
	}

	jobs := make([]models.Job, 0, len(data.Results))
	for _, r := range data.Results {
		// Espelha o salario do JS: "min-max" ou vazio.
		salario := ""
		if r.SalaryMin != 0 || r.SalaryMax != 0 {
			salario = fmt.Sprintf("%g-%g", r.SalaryMin, r.SalaryMax)
		}

		link := strings.TrimSpace(r.RedirectURL)
		if link == "" {
			link = strings.TrimSpace(r.URL)
		}

		jobs = append(jobs, models.Job{
			ID:       link,
			Title:    strings.TrimSpace(r.Title),
			Company:  strings.TrimSpace(r.Company.DisplayName),
			Location: strings.TrimSpace(r.Location.DisplayName),
			URL:      link,
			Salary:   salario,
			PostedAt: r.Created,
			Source:   "Adzuna",
			Sources:  []string{"Adzuna"},
			Keyword:  keyword,
			Keywords: []string{keyword},
		})
	}

	return jobs, nil
}
