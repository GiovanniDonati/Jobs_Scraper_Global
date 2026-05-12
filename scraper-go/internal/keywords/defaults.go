package keywords

import (
	"encoding/json"
	"log/slog"
	"os"
)

type keywordsFile struct {
	KEYWORDS []string `json:"KEYWORDS"`
}

func LoadDefaultKeywords() []string {
	// Agora que voltou para a pasta keywords:
	paths := []string{
		"/app/internal/keywords/keywords.json", // Caminho no Docker
		"./internal/keywords/keywords.json",    // Caminho local
	}

	var data []byte
	var err error

	for _, p := range paths {
		data, err = os.ReadFile(p)
		if err == nil {
			break
		}
	}

	if err != nil {
		slog.Warn("Arquivo keywords.json não encontrado, usando .env ou lista vazia")
		if env := os.Getenv("KEYWORDS"); env != "" {
			return []string{env}
		}
		return []string{}
	}

	var parsed keywordsFile
	if err := json.Unmarshal(data, &parsed); err != nil {
		slog.Error("Erro no unmarshal do JSON", "error", err)
		return []string{}
	}

	return NormalizeKeywords(parsed.KEYWORDS)
}
