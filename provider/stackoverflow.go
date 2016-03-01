package provider

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const (
	SECTION_TYPE_NORMAL = iota
	SECTION_TYPE_CODE
)

type StackoverflowSections []*StackoverflowSection

func (s StackoverflowSections) String() string {
	var body string
	for i, sec := range s {
		if i > 0 {
			body += "\n"
		}

		body += sec.Content
	}
	return body
}

type StackoverflowSection struct {
	Type    int
	Content string
}

func (s *StackoverflowSection) IsCode() bool {
	return s.Type == SECTION_TYPE_CODE
}

type StackOverflowResult struct {
	title            string
	author           string
	question         StackoverflowSections
	bestAnswer       StackoverflowSections
	bestAnswerAuthor string
}

func (s *StackOverflowResult) Title() string {
	return s.title
}

func (s *StackOverflowResult) Author() string {
	return s.author
}

func (s *StackOverflowResult) Body() string {
	return s.bestAnswer.String()
}

func (s *StackOverflowResult) Answer() StackoverflowSections {
	return s.bestAnswer
}

func (s *StackOverflowResult) AnswerAuthor() string {
	return s.bestAnswerAuthor
}

func (s *StackOverflowResult) Question() StackoverflowSections {
	return s.question
}

type StackOverflow struct {
}

func (s *StackOverflow) Flag() (flag, usage string) {
	flag, usage = "so", "StackOverflow"
	return
}

func (s *StackOverflow) Name() string {
	return "stackoverflow"
}

func (s *StackOverflow) Domain() string {
	return "https://stackoverflow.com/questions"
}

func (s *StackOverflow) Match(url string) bool {
	i := strings.Index(url, "stackoverflow.com")
	// https://www.
	return i > -1 && i < 13
}

func (s *StackOverflow) parseIntoSections(sel *goquery.Selection) StackoverflowSections {
	sections := make(StackoverflowSections, sel.Length())
	sel.Each(func(i int, n *goquery.Selection) {
		nType := SECTION_TYPE_NORMAL
		switch goquery.NodeName(n) {
		case "pre":
			fallthrough
		case "code":
			nType = SECTION_TYPE_CODE
		}

		sections[i] = &StackoverflowSection{
			nType,
			singleBreak(n.Text()),
		}
	})

	return sections
}

func (s *StackOverflow) Parse(doc *goquery.Document) (Result, error) {
	root := doc.Find("#content")
	title := doc.Find("#question-header h1").Text()

	questionN := root.Find("#question")
	author := questionN.Find(".post-signature.owner .user-details>a").Text()
	question := s.parseIntoSections(questionN.Find(".post-text").Children())

	bestAnswerN := root.Find("#answers .accepted-answer, #answers .answer").First()
	bestAnswer := s.parseIntoSections(bestAnswerN.Find(".post-text").Children())
	bestAnswerAuthor := bestAnswerN.Find(".user-details>a").Text()

	return &StackOverflowResult{
		singleBreak(title),
		singleBreak(author),
		question,
		bestAnswer,
		singleBreak(bestAnswerAuthor),
	}, nil
}
