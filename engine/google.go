package engine

import (
	"net/url"

	"github.com/PuerkitoBio/goquery"
)

type Google struct {
}

func (g *Google) Flag() (flag, usage string) {
	flag, usage = "g", "Google"
	return
}

func (g *Google) URL(q string, domains []string, limit int) string {
	var dms string
	for i, dm := range domains {
		if i > 0 {
			dms += " OR"
		}
		dms += " site:" + dm
	}

	return "https://www.google.com/search?q=" + url.QueryEscape(q+dms)
}

func (g *Google) Parse(doc *goquery.Document, limit int) ([]*Result, error) {
	results := []*Result{}
	rNodes := doc.Find("#ires ol .g") // h3.r a")
	rNodes.Each(func(i int, n *goquery.Selection) {
		results = append(results, &Result{
			singleLine(n.Find("h3.r a").Text()),
			singleLine(n.Find(".st").Text()),
			singleLine(n.Find(".s .kv cite").Text()),
		})
	})

	return results, nil
}
