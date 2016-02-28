package engine

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type Result struct {
	Title       string
	Description string
	URL         string
}

type Engine interface {
	Flag() (flag, usage string)
	URL(q string, domains []string, limit int) string
	Parse(doc *goquery.Document, limit int) ([]*Result, error)
}

func singleLine(s string) string {
	return strings.TrimSpace(strings.Replace(s, "\n", " ", -1))
}
