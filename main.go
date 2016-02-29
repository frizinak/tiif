package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

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

func getResult(results []*engine.Result) (*engine.Result, string, error) {
	var n int
	resetCursor := "\033[1A\033[1G\033[K"
	for {
		in, err := terminal.Prompt(resetCursor + "> ")
		if err != nil {
			return nil, "", err
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
			return nil, "", nil
		case '?':
			fmt.Printf(
				"[1-%d] to read, prefix with o for browser. q to quit.",
				len(results),
			)

			continue
		}

		n, err = strconv.Atoi(in)
		if err != nil {
			return nil, in, nil
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

	return results[n-1], "", nil
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

	return nil
}

func run(query string, engines []engine.Engine, providers []provider.Provider) error {
	// Run query through engines until we get results
	var results []*engine.Result
	for _, engine := range engines {
		var err error
		results, err = getResults(engine, providers, query)
		if err == noResults {
			continue
		}

		if err != nil {
			return err
		}

		break
	}

	// Nope
	if len(results) == 0 {
		return noResults
	}

	if err := printResults(results); err != nil {
		return err
	}

	// Run prompt
	for {
		result, q, err := getResult(results)
		if err != nil {
			return err
		}

		if q != "" {
			return run(q, engines, providers)
		}

		if result == nil {
			return nil
		}

		// Got a result, pass it to the provider
		if err := getPage(results, result, providers); err != nil {
			return err
		}
	}
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

	query := strings.Join(flag.Args(), " ")
	if err := run(query, engines, providers); err != nil {
		log.Fatal(err)
	}
}
