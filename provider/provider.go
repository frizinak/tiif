package provider

import "github.com/PuerkitoBio/goquery"

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
