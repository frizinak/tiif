package engine

import (
	"fmt"
	"net/url"

	"github.com/PuerkitoBio/goquery"
)

type DuckDuckGo struct {
}

func (g *DuckDuckGo) Flag() (flag, usage string) {
	flag, usage = "ddg", "DuckDuckGo"
	return
}

func (g *DuckDuckGo) URL(q string, domains []string, limit int) string {
	var dms string
	for i, dm := range domains {
		if i > 0 {
			dms += " OR"
		}
		dms += " site:" + dm
	}

	return "https://duckduckgo.com/html?q=" +
		url.QueryEscape(fmt.Sprintf("%s (%s)", q, dms))
}

func (g *DuckDuckGo) Parse(doc *goquery.Document, limit int) ([]*Result, error) {
	results := []*Result{}
	rNodes := doc.Find("#links > .results_links > .links_main") // h3.r a")
	if rNodes.Find(".no-results").Length() != 0 {
		return nil, nil
	}

	rNodes.EachWithBreak(func(i int, n *goquery.Selection) bool {
		limit--
		a := n.Find("a.result__a")
		href, _ := a.Attr("href")
		results = append(results, &Result{
			singleLine(a.Text()),
			singleLine(n.Find(".result__snippet").Text()),
			href,
		})

		return limit > 0
	})

	return results, nil
}
