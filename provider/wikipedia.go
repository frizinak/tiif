package provider

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type WikipediaSection struct {
	Title string
	body  []string
}

func (s *WikipediaSection) Body() string {
	return strings.Join(s.body, "\n\n")
}

type WikipediaResult struct {
	title       string
	author      string
	description []string
	sections    []*WikipediaSection
}

func (s *WikipediaResult) Title() string {
	return s.title
}

func (s *WikipediaResult) Author() string {
	return s.author
}

func (s *WikipediaResult) Body() string {
	return strings.Join(s.description, "\n\n")
}

func (s *WikipediaResult) Sections() []*WikipediaSection {
	return s.sections
}

type Wikipedia struct {
}

func (s *Wikipedia) Flag() (flag, usage string) {
	flag, usage = "wiki", "Wikipedia"
	return
}

func (s *Wikipedia) Name() string {
	return "wikipedia"
}

func (s *Wikipedia) Domain() string {
	return "https://en.wikipedia.org"
}

func (s *Wikipedia) Match(url string) bool {
	i := strings.Index(url, "wikipedia.org")
	// https://www.
	return i > -1 && i < 13
}

func (s *Wikipedia) Parse(doc *goquery.Document) (Result, error) {
	root := doc.Find("#content")
	title := root.Find("h1#firstHeading").Text()
	contentN := root.Find("#mw-content-text")
	note := contentN.Find(".hatnote").Text()
	description := make([]string, 0, 1)

	sections := make([]*WikipediaSection, 0)

	contentN.Children().EachWithBreak(func(i int, s *goquery.Selection) bool {
		if title := s.Find(".mw-headline"); s.Is("h2, h3") && title.Length != nil {
			id, _ := title.Attr("id")
			if id == "References" || id == "External_links" {
				return false
			}

			section := &WikipediaSection{
				title.Text(),
				make([]string, 0),
			}
			sections = append(sections, section)
		} else if len(sections) > 0 {
			sections[len(sections)-1].body = append(
				sections[len(sections)-1].body,
				singleBreak(s.Text()),
			)
		} else {
			description = append(description, singleBreak(s.Text()))
		}

		return true
	})

	return &WikipediaResult{
		singleBreak(title),
		singleBreak(note),
		description,
		sections,
	}, nil
}
