package provider

import (
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var singleBreakRE *regexp.Regexp

func init() {
	singleBreakRE = regexp.MustCompile(`\n+`)
}

type Result interface {
	Title() string
	Author() string
	Body() string
}

type Provider interface {
	Flag() (flag, usage string)
	Name() string
	Domain() string
	Match(url string) bool
	Parse(doc *goquery.Document) (Result, error)
}

func singleBreak(s string) string {
	return strings.TrimSpace(singleBreakRE.ReplaceAllString(s, "\n"))
}
