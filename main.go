package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"golang.org/x/net/html"

	"github.com/PuerkitoBio/goquery"
	"github.com/frizinak/tiif/engine"
	"github.com/frizinak/tiif/httpcache"
	"github.com/frizinak/tiif/provider"
	"github.com/frizinak/tiif/terminal"
	"github.com/skratchdot/open-golang/open"
)

var terminalWidth int
var cache *httpcache.Client

var noResults error = errors.New("No Results")

func init() {
	cache = httpcache.New(10)
}

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

func getResults(
	engine engine.Engine,
	providers []provider.Provider,
	query string,
) ([]*engine.Result, error) {
	_, h := terminal.Dimensions()
	h = intMax(10, h)
	limit := h/2 - 1

	domains := make([]string, len(providers))
	for i := range providers {
		domains[i] = providers[i].Domain()
	}
	url := engine.URL(query, domains, limit)

	doc, err := getDOM(url)
	if err != nil {
		return nil, err
	}

	results, err := engine.Parse(doc, limit)
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		log.Println(url)
		return nil, noResults
	}

	for _, r := range results {
		go func(url string) {
			req, err := createRequest(url)
			if err == nil {
				<-cache.Exec(req)
			}
		}(r.URL)
	}

	return results, nil
}

func printResults(results []*engine.Result) error {
	w, _ := terminal.Dimensions()
	w = intMax(60, w)
	terminalWidth = w

	err := execTpl(os.Stdout, "search-results", results, nil)
	if err != nil {
		return err
	}

	return nil
}

func getResult(results []*engine.Result) (*engine.Result, error) {
	var n int
	resetCursor := "\033[1A\033[1G\033[K"
	for {
		in, err := terminal.Prompt(resetCursor + "> ")
		if err != nil {
			return nil, err
		}

		if len(in) == 0 {
			continue
		}

		browser := false
		action := in[0]
		if action >= 'A' && action <= 'Z' {
			action += 'a' - 'A'
		}

		switch action {
		case 'o':
			browser = true
			in = in[1:]
		case 'q':
			return nil, nil
		case '?':
			fmt.Printf(
				"%s[1-%d] to read, prefix with o for browser. q to quit.",
				resetCursor,
				len(results),
			)
		}

		n, err = strconv.Atoi(in)
		if err != nil {
			continue
		}

		if n < 1 || n > len(results) {
			continue
		}

		if browser {
			go open.Run(results[n-1].URL)
			continue
		}

		break
	}

	return results[n-1], nil
}

func getPage(
	results []*engine.Result,
	r *engine.Result,
	providers []provider.Provider,
) error {
	doc, err := getDOM(r.URL)
	if err != nil {
		return err
	}

	var provider provider.Provider
	for _, p := range providers {
		if p.Match(r.URL) {
			provider = p
			break
		}
	}

	if provider == nil {
		return fmt.Errorf("No provider found for %s", r.URL)
	}

	result, err := provider.Parse(doc)
	if err != nil {
		return err
	}

	cmd := exec.Command("less", "-Rc")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	pipe, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	err = execTpl(pipe, "result-page", result, []string{provider.Name()})
	pipe.Close()
	cmd.Wait()
	if err != nil {
		return err
	}

	r, err = getResult(results)
	if err != nil {
		return err
	}

	if r == nil {
		return nil
	}

	return getPage(results, r, providers)
}

func main() {
	// Setup engines and providers
	availableEngines := []engine.Engine{
		&engine.DuckDuckGo{},
		&engine.Google{},
	}

	availableProviders := []provider.Provider{
		&provider.StackOverflow{},
		&provider.Wikipedia{},
	}

	enabledEngines := make([]*bool, len(availableEngines))
	enabledProviders := make([]*bool, len(availableProviders))

	engines := make([]engine.Engine, 0, len(availableEngines))
	providers := make([]provider.Provider, 0, len(availableProviders))

	for i := range availableEngines {
		f, u := availableEngines[i].Flag()
		enabledEngines[i] = flag.Bool(f, false, "Search engine: "+u)
	}

	for i := range availableProviders {
		f, u := availableProviders[i].Flag()
		enabledProviders[i] = flag.Bool(f, false, "Provider: "+u)
	}

	flag.Parse()

	for i := range availableEngines {
		if *enabledEngines[i] {
			engines = append(engines, availableEngines[i])
			break
		}
	}

	for i := range availableProviders {
		if *enabledProviders[i] {
			providers = append(providers, availableProviders[i])
		}
	}

	if len(engines) == 0 {
		engines = availableEngines
	}

	if len(providers) == 0 {
		providers = availableProviders
	}

	// Run query through engines until we get results
	query := strings.Join(flag.Args(), " ")
	var results []*engine.Result
	for _, engine := range engines {
		var err error
		results, err = getResults(engine, providers, query)
		if err == noResults {
			continue
		}

		if err != nil {
			log.Fatal(err)
		}

		break
	}

	// Nope
	if len(results) == 0 {
		log.Fatal(noResults)
	}

	if err := printResults(results); err != nil {
		log.Fatal(err)
	}

	// Run prompt
	result, err := getResult(results)
	if err != nil {
		log.Fatal(err)
	}

	if result == nil {
		return
	}

	// Got a result, pass it to the provider
	err = getPage(results, result, providers)
	if err != nil {
		log.Fatal(err)
	}
}
