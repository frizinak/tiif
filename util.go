package main

import (
	"bytes"
	"net/http"
	"net/url"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
)

func intMax(x, y int) int {
	if x > y {
		return x
	}

	return y
}

func createRequest(u string) (*http.Request, error) {
	purl, err := url.Parse(u)
	if err != nil {
		return nil, err
	}

	if purl.Scheme == "" {
		purl.Scheme = "https"
	}

	u = purl.String()

	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set(
		"User-Agent",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 "+
			"(KHTML, like Gecko) Chrome/48.0.2564.116 Safari/537.36",
	)

	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Dnt", "1")
	req.Header.Set("X-Do-Not-Track", "1")
	req.Header.Set("Accept-Language", "en-US,en")

	return req, nil
}

func getDOM(u string) (*goquery.Document, error) {
	req, err := createRequest(u)
	if err != nil {
		return nil, err
	}

	b := bytes.NewBuffer(<-cache.Exec(req))
	node, err := html.Parse(b)
	if err != nil {
		return nil, err
	}

	dom := goquery.NewDocumentFromNode(node)

	return dom, nil
}
