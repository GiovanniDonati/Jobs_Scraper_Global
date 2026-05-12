package adapters

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Benevanio/Jobs_Scraper_Global/scraper-go/internal/models"
)

type JoobleAdapter struct {
	apiKey string
	client *http.Client
}

func NewJooble(apiKey string) *JoobleAdapter {
	return &JoobleAdapter{
		apiKey: apiKey,
		client: &http.Client{Timeout: 15 * time.Second},
	}
}

func (a *JoobleAdapter) SourceName() string {
	return "Jooble"
}

func (a *JoobleAdapter) Supports(source string) bool {
	return source == "Jooble"
}

func (a *JoobleAdapter) Search(ctx context.Context, keyword string, req models.ScrapeRequest) ([]models.Job, error) {
	url := "https://br.jooble.org/api/" + a.apiKey

	// Payload conforme a documentação que você printou
	payload := map[string]string{
		"keywords": keyword,
		"location": "Brasil",
	}

	body, _ := json.Marshal(payload)

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("jooble erro: status %d", resp.StatusCode)
	}

	// Estrutura de resposta da Jooble
	var joobleRes struct {
		Jobs []struct {
			Title    string `json:"title"`
			Location string `json:"location"`
			Snippet  string `json:"snippet"`
			Source   string `json:"source"`
			Type     string `json:"type"`
			Link     string `json:"link"`
			Company  string `json:"company"`
			Updated  string `json:"updated"`
			Salary   string `json:"salary"`
			ID       int64  `json:"id"`
		} `json:"jobs"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&joobleRes); err != nil {
		return nil, err
	}

	var jobs []models.Job
	for _, j := range joobleRes.Jobs {
		jobs = append(jobs, models.Job{
			ID:       j.Link,
			Title:    j.Title,
			Company:  j.Company,
			Location: j.Location,
			URL:      j.Link,
			Salary:   j.Salary,
			Source:   "Jooble",
			Sources:  []string{"Jooble"},
			Keyword:  keyword,
			Keywords: []string{keyword},
		})
	}

	return jobs, nil
}
