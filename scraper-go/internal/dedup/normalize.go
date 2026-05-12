package dedup

import (
	"net/url"
	"regexp"
	"strings"
)

var spaces = regexp.MustCompile(`\s+`)

func NormalizeText(v string) string {
	v = strings.ToLower(strings.TrimSpace(v))
	v = spaces.ReplaceAllString(v, " ")

	return v
}

func NormalizeURL(raw string) string {
	u, err := url.Parse(raw)

	if err != nil {
		return raw
	}

	u.RawQuery = ""
	u.Fragment = ""

	return strings.TrimRight(u.String(), "/")
}
