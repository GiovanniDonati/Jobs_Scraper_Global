package adapters

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"

	"github.com/Benevanio/Jobs_Scraper_Global/scraper-go/internal/models"
)

const linkedinSearchURL = "https://www.linkedin.com/jobs-guest/jobs/api/seeMoreJobPostings/search"
const linkedinPageStep = 25

// LinkedInAdapter busca vagas no LinkedIn via scraping HTML,
// espelhando exatamente o linkedinAdapter.js com cheerio.
type LinkedInAdapter struct {
	client *http.Client
}

func NewLinkedIn() *LinkedInAdapter {
	return &LinkedInAdapter{
		// Timeout do client é maior que o pageTimeout individual,
		// porque o Search pode fazer várias páginas.
		client: &http.Client{Timeout: 2 * time.Minute},
	}
}

func (a *LinkedInAdapter) SourceName() string { return "linkedin" }

// buildSearchUrl espelha exatamente o buildSearchUrl do JS.
func buildLinkedInURL(keyword string, req models.ScrapeRequest, start int) string {
	u, _ := url.Parse(linkedinSearchURL)
	q := u.Query()

	q.Set("keywords", keyword)

	if req.SearchLocation != "" {
		q.Set("location", req.SearchLocation)
	}
	if req.SearchGeoID != "" {
		q.Set("geoId", req.SearchGeoID)
	}
	if req.SearchLanguage != "" {
		q.Set("lang", req.SearchLanguage)
	}
	// Espelha: if (config.remoteOnly !== false)
	// O zero-value de bool em Go é false, então usamos um ponteiro
	// no ScrapeRequest para distinguir "não informado" de false explícito.
	// Como o Node envia remoteOnly apenas quando true, checamos direto.
	if req.RemoteOnly {
		q.Set("f_WT", "2")
	}
	if req.JobTypes != "" {
		q.Set("f_JT", req.JobTypes)
	}
	if req.TimeFilter != "" {
		q.Set("f_TPR", req.TimeFilter)
	}

	q.Set("start", fmt.Sprintf("%d", start))
	u.RawQuery = q.Encode()
	return u.String()
}

// fetchJobsChunk espelha o fetchJobsChunk do JS.
// Faz o GET e parseia o HTML com goquery (equivalente ao cheerio).
func (a *LinkedInAdapter) fetchJobsChunk(ctx context.Context, keyword string, req models.ScrapeRequest, start int) ([]models.Job, error) {
	pageTimeout := time.Duration(req.PageTimeoutMs) * time.Millisecond
	if pageTimeout <= 0 {
		pageTimeout = 15 * time.Second
	}

	endpoint := buildLinkedInURL(keyword, req, start)

	pageCtx, cancel := context.WithTimeout(ctx, pageTimeout)
	defer cancel()

	httpReq, err := http.NewRequestWithContext(pageCtx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	// Mesmos headers do JS.
	httpReq.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Safari/537.36")
	httpReq.Header.Set("accept-language", "pt-BR,pt;q=0.9,en-US;q=0.8,en;q=0.7")

	resp, err := a.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}

	// goquery.NewDocumentFromReader espelha o cheerio.load(response.data).
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	var jobs []models.Job

	// Espelha: $(".base-card, .job-search-card").each(...)
	doc.Find(".base-card, .job-search-card").Each(func(_ int, card *goquery.Selection) {
		// Espelha cada node.find(...).text().trim() || fallback
		titulo := strings.TrimSpace(card.Find(".base-search-card__title").Text())
		if titulo == "" {
			titulo = strings.TrimSpace(card.Find("h3").First().Text())
		}

		empresa := strings.TrimSpace(card.Find(".base-search-card__subtitle").Text())
		if empresa == "" {
			empresa = strings.TrimSpace(card.Find("h4").First().Text())
		}

		local := strings.TrimSpace(card.Find(".job-search-card__location").Text())

		link, _ := card.Find("a.base-card__full-link").Attr("href")
		if link == "" {
			link, _ = card.Find("a[href*='/jobs/view/']").Attr("href")
		}

		jobs = append(jobs, models.Job{
			Title:    titulo,
			Company:  empresa,
			Location: local,
			URL:      link,
		})
	})

	return jobs, nil
}

// normalizeJob espelha o normalizeJob do JS.
func normalizeLinkedInJob(keyword string, job models.Job) models.Job {
	u := strings.TrimSpace(job.URL)
	return models.Job{
		ID:       u,
		Title:    strings.TrimSpace(job.Title),
		Company:  strings.TrimSpace(job.Company),
		Location: strings.TrimSpace(job.Location),
		URL:      u,
		Source:   "LinkedIn",
		Sources:  []string{"LinkedIn"},
		Keyword:  keyword,
		Keywords: []string{keyword},
	}
}

// dedupeLinkedIn espelha o dedupeJobs interno do linkedinAdapter.js.
func dedupeLinkedIn(jobs []models.Job) []models.Job {
	unique := make(map[string]models.Job, len(jobs))
	order := make([]string, 0, len(jobs))

	for _, job := range jobs {
		key := job.URL
		if key == "" {
			key = job.Title + "-" + job.Company + "-" + job.Location
		}
		if key == "" {
			continue
		}
		if _, exists := unique[key]; !exists {
			unique[key] = job
			order = append(order, key)
		}
	}

	result := make([]models.Job, 0, len(order))
	for _, key := range order {
		result = append(result, unique[key])
	}
	return result
}

// Search espelha o linkedinAdapter.search do JS.
func (a *LinkedInAdapter) Search(ctx context.Context, keyword string, req models.ScrapeRequest) ([]models.Job, error) {
	maxPages := req.MaxPagesPerKeyword
	if maxPages <= 0 {
		maxPages = 5
	}

	waitBetween := time.Duration(req.WaitBetweenSearchesMs) * time.Millisecond
	if waitBetween <= 0 {
		waitBetween = 1000 * time.Millisecond
	}

	var allJobs []models.Job

	for pageIndex := 0; pageIndex < maxPages; pageIndex++ {
		start := pageIndex * linkedinPageStep

		jobs, err := a.fetchJobsChunk(ctx, keyword, req, start)
		if err != nil {
			// Espelha o catch com logWarn — não interrompe, apenas avisa.
			// O pipeline vai logar o warn; aqui retornamos o que já temos.
			if pageIndex == 0 {
				return allJobs, fmt.Errorf("linkedin: falha HTTP na busca para %q (start=%d): %w", keyword, start, err)
			}
			break
		}

		if len(jobs) == 0 {
			// Espelha o break quando não há cards.
			break
		}

		for _, job := range jobs {
			allJobs = append(allJobs, normalizeLinkedInJob(keyword, job))
		}

		if pageIndex < maxPages-1 {
			select {
			case <-ctx.Done():
				return dedupeLinkedIn(allJobs), nil
			case <-time.After(waitBetween):
			}
		}
	}

	return dedupeLinkedIn(allJobs), nil
}
