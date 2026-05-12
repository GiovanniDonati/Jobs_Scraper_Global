package adapters

import (
	"context"

	"github.com/Benevanio/Jobs_Scraper_Global/scraper-go/internal/models"
)

// Adapter é a interface que todo scraper de fonte deve implementar.
type Adapter interface {
	// SourceName retorna o identificador da fonte (ex: "linkedin", "Adzuna:br").
	SourceName() string

	// Search busca vagas para uma keyword e configuração específica.
	// Deve respeitar o contexto (timeout / cancelamento).
	Search(ctx context.Context, keyword string, req models.ScrapeRequest) ([]models.Job, error)
}
