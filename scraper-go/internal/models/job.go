package models

type Job struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Company     string   `json:"company"`
	Location    string   `json:"location"`
	URL         string   `json:"url"`
	Salary      string   `json:"salary,omitempty"`
	Modality    string   `json:"modality,omitempty"`
	Description string   `json:"description,omitempty"`
	PostedAt    string   `json:"postedAt,omitempty"`
	Source      string   `json:"source"`
	Sources     []string `json:"sources"`
	Keyword     string   `json:"keyword"`
	Keywords    []string `json:"keywords"`
}

type ScrapeRequest struct {
	Keywords              []string `json:"keywords"`
	SearchLocation        string   `json:"searchLocation"`
	SearchGeoID           string   `json:"searchGeoId"`
	SearchLanguage        string   `json:"searchLanguage"`
	JobTypes              string   `json:"jobTypes"`
	TimeFilter            string   `json:"timeFilter"`
	RemoteOnly            bool     `json:"remoteOnly"`
	Sources               []string `json:"sources"`
	ResultsPerPage        int      `json:"resultsPerPage"`
	MaxPagesPerKeyword    int      `json:"maxPagesPerKeyword"`
	WaitBetweenSearchesMs int      `json:"waitBetweenSearchesMs"`
	PageTimeoutMs         int      `json:"pageTimeoutMs"`
	MaxConcurrency        int      `json:"maxConcurrency"`
}

type ScrapeResponse struct {
	Jobs      []Job  `json:"jobs"`
	Total     int    `json:"total"`
	CachedAt  string `json:"cachedAt"`
	FromCache bool   `json:"fromCache"`
}
