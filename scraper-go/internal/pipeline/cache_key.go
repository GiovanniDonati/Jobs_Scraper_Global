package pipeline

import (
	"sort"
	"strings"
)

func BuildCacheKey(config SearchConfig) string {
	keywords := normalizeKeywords(config.Keywords)

	return strings.Join([]string{
		"jobs",
		strings.Join(keywords, ","),
		strings.ToLower(strings.TrimSpace(config.SearchLocation)),
		config.JobTypes,
		config.TimeFilter,
	}, ":")
}

func normalizeKeywords(keywords []string) []string {
	var normalized []string

	for _, k := range keywords {
		k = strings.ToLower(strings.TrimSpace(k))

		if k != "" {
			normalized = append(normalized, k)
		}
	}

	sort.Strings(normalized)

	return normalized
}
