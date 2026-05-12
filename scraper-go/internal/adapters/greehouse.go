package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/Benevanio/Jobs_Scraper_Global/scraper-go/internal/models"
)

var jsonLDPattern = regexp.MustCompile(`(?s)<script[^>]+type="application/ld\+json"[^>]*>(.*?)</script>`)

type greenhouseListResponse struct {
	Jobs []greenhouseListJob `json:"jobs"`
}

type greenhouseListJob struct {
	Title       string `json:"title"`
	Content     string `json:"content"`
	AbsoluteURL string `json:"absolute_url"`
	UpdatedAt   string `json:"updated_at"`
	Location    struct {
		Name string `json:"name"`
	} `json:"location"`
}

type jobPosting struct {
	Type        string `json:"@type"`
	Title       string `json:"title"`
	Description string `json:"description"`
	DatePosted  string `json:"datePosted"`

	HiringOrganization struct {
		Name string `json:"name"`
	} `json:"hiringOrganization"`

	JobLocation struct {
		Address struct {
			Locality string `json:"addressLocality"`
			Region   string `json:"addressRegion"`
			Country  string `json:"addressCountry"`
		} `json:"address"`
	} `json:"jobLocation"`

	BaseSalary *struct {
		Value struct {
			MinValue float64 `json:"minValue"`
			MaxValue float64 `json:"maxValue"`
			UnitText string  `json:"unitText"`
		} `json:"value"`
	} `json:"baseSalary"`
}

func extractJSONLD(html string) *jobPosting {
	matches := jsonLDPattern.FindAllStringSubmatch(html, -1)

	for _, match := range matches {
		if len(match) < 2 {
			continue
		}

		var posting jobPosting
		if err := json.Unmarshal([]byte(strings.TrimSpace(match[1])), &posting); err != nil {
			continue
		}

		if posting.Type == "JobPosting" && posting.Title != "" {
			return &posting
		}
	}

	return nil
}

type GreenhouseAdapter struct {
	client      *http.Client
	boardToken  string
	companyName string
}

func NewGreenhouse(boardToken, companyName string) *GreenhouseAdapter {
	return &GreenhouseAdapter{
		client:      &http.Client{Timeout: 30 * time.Second},
		boardToken:  boardToken,
		companyName: companyName,
	}
}

func (a *GreenhouseAdapter) SourceName() string {
	return fmt.Sprintf("Green House:%s", a.companyName)
}

func (a *GreenhouseAdapter) Search(ctx context.Context, keyword string, req models.ScrapeRequest) ([]models.Job, error) {
	pageTimeout := time.Duration(req.PageTimeoutMs) * time.Millisecond
	if pageTimeout <= 0 {
		pageTimeout = 15 * time.Second
	}

	listCtx, cancel := context.WithTimeout(ctx, pageTimeout)
	defer cancel()

	endpoint := fmt.Sprintf("https://boards-api.greenhouse.io/v1/boards/%s/jobs?content=true", a.boardToken)

	listReq, err := http.NewRequestWithContext(listCtx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	resp, err := a.client.Do(listReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}

	var data struct {
		Jobs []struct {
			Title       string `json:"title"`
			Content     string `json:"content"`
			AbsoluteURL string `json:"absolute_url"`
			UpdatedAt   string `json:"updated_at"`
			Location    struct {
				Name string `json:"name"`
			} `json:"location"`
		} `json:"jobs"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	kwLower := strings.ToLower(keyword)
	var jobs []models.Job

	for _, j := range data.Jobs {
		if strings.Contains(strings.ToLower(j.Title), kwLower) {
			jobs = append(jobs, models.Job{
				ID:          strings.TrimSpace(j.AbsoluteURL),
				Title:       strings.TrimSpace(j.Title),
				Description: strings.TrimSpace(j.Content),
				Company:     a.companyName,
				Location:    strings.TrimSpace(j.Location.Name),
				URL:         strings.TrimSpace(j.AbsoluteURL),
				PostedAt:    j.UpdatedAt,
				Source:      "Green House",
				Sources:     []string{"Green House"},
				Keyword:     keyword,
				Keywords:    []string{keyword},
			})
		}
	}

	return jobs, nil
}

func FetchGreenhouseSlugs(ctx context.Context) ([]string, error) {

	filename := "./internal/interfaces/greenhouseCompanies.json"

	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("não foi possível ler o arquivo %s: %w", filename, err)
	}

	var slugs []string
	if err := json.Unmarshal(data, &slugs); err != nil {
		return nil, fmt.Errorf("erro ao processar o JSON de empresas: %w", err)
	}

	slog.Info("Slugs da Greenhouse carregados com sucesso", "total", len(slugs))

	return slugs, nil
}
