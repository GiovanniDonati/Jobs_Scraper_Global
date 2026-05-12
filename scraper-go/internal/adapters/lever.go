package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Benevanio/Jobs_Scraper_Global/scraper-go/internal/models"
)

type LeverAdapter struct {
	client      *http.Client
	companySlug string
	companyName string
}

type leverCompany struct {
	Slug string `json:"slug"`
	Name string `json:"name"`
}

type leverPosting struct {
	ID        string `json:"id"`
	Text      string `json:"text"`
	HostedURL string `json:"hostedUrl"`
	CreatedAt int64  `json:"createdAt"`

	Categories struct {
		Team       string `json:"team"`
		Department string `json:"department"`
		Location   string `json:"location"`
		Commitment string `json:"commitment"`
		Level      string `json:"level"`
	} `json:"categories"`

	Description string `json:"description"`
	State       string `json:"state"`
}

func NewLever(companySlug, companyName string) *LeverAdapter {
	return &LeverAdapter{
		client:      &http.Client{Timeout: 60 * time.Second},
		companySlug: companySlug,
		companyName: companyName,
	}
}

func FetchLeverSlugs(_ context.Context) ([]leverCompany, error) {
	path := os.Getenv("LEVER_COMPANIES_FILE")
	if path == "" {
		path = "./internal/interfaces/leverCompanies.json"
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("lever: leitura do arquivo '%s': %w", path, err)
	}

	var companies []leverCompany
	if err := json.Unmarshal(data, &companies); err != nil {
		return nil, fmt.Errorf("lever: parse do arquivo '%s': %w", path, err)
	}

	if len(companies) == 0 {
		return nil, fmt.Errorf("lever: nenhuma empresa encontrada em '%s'", path)
	}

	return companies, nil
}

func (a *LeverAdapter) SourceName() string {
	return fmt.Sprintf("Lever:%s", a.companyName)
}

func (a *LeverAdapter) Search(ctx context.Context, keyword string, req models.ScrapeRequest) ([]models.Job, error) {
	pageTimeout := time.Duration(req.PageTimeoutMs) * time.Millisecond
	if pageTimeout <= 0 {
		pageTimeout = 15 * time.Second
	}

	endpoint := fmt.Sprintf(
		"https://api.lever.co/v0/postings/%s?mode=json",
		a.companySlug,
	)

	reqCtx, cancel := context.WithTimeout(ctx, pageTimeout)
	defer cancel()

	httpReq, err := http.NewRequestWithContext(reqCtx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("lever: build request: %w", err)
	}

	httpReq.Header.Set("User-Agent", "JobsScraper/1.0")
	httpReq.Header.Set("Accept", "application/json")

	resp, err := a.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("lever: http do: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		// ok
	case http.StatusTooManyRequests:
		return nil, fmt.Errorf("lever: rate limit atingido (429)")
	case http.StatusNotFound:
		return nil, fmt.Errorf("lever: empresa '%s' não encontrada (404)", a.companySlug)
	default:
		return nil, fmt.Errorf("lever: status inesperado %d", resp.StatusCode)
	}

	var raw []leverPosting
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("lever: decode json: %w", err)
	}

	kwLower := strings.ToLower(keyword)

	var jobs []models.Job
	for _, j := range raw {
		if keyword != "" && !strings.Contains(strings.ToLower(j.Text), kwLower) {
			continue
		}

		dataPublicacao := ""
		if j.CreatedAt != 0 {
			dataPublicacao = time.UnixMilli(j.CreatedAt).UTC().Format(time.RFC3339)
		}

		local := strings.TrimSpace(j.Categories.Location)
		if local == "" {
			local = strings.TrimSpace(j.Categories.Department)
		}

		jobs = append(jobs, models.Job{
			ID:       strings.TrimSpace(j.HostedURL),
			Title:    strings.TrimSpace(j.Text),
			Company:  a.companyName,
			Location: local,
			URL:      strings.TrimSpace(j.HostedURL),
			Modality: strings.TrimSpace(j.Categories.Commitment),
			PostedAt: dataPublicacao,
			Source:   "Lever",
			Sources:  []string{"Lever"},
			Keyword:  keyword,
			Keywords: []string{keyword},
		})
	}

	return jobs, nil
}
