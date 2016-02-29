package provider

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type StackOverflowResult struct {
	title            string
	author           string
	question         string
	bestAnswer       string
	bestAnswerAuthor string
}

func (s *StackOverflowResult) Title() string {
	return s.title
}

func (s *StackOverflowResult) Author() string {
	return s.author
}

func (s *StackOverflowResult) Body() string {
	return s.bestAnswer
}

func (s *StackOverflowResult) Answer() string {
	return s.Body()
}

func (s *StackOverflowResult) AnswerAuthor() string {
	return s.bestAnswerAuthor
}

func (s *StackOverflowResult) Question() string {
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

func (s *StackOverflow) Parse(doc *goquery.Document) (Result, error) {
	root := doc.Find("#content")
	title := doc.Find("#question-header h1").Text()

	questionN := root.Find("#question")
	author := questionN.Find(".post-signature.owner .user-details>a").Text()
	question := questionN.Find(".post-text").Text()

	bestAnswerN := root.Find("#answers .accepted-answer, #answers .answer").First()
	bestAnswer := bestAnswerN.Find(".post-text").Text()
	bestAnswerAuthor := bestAnswerN.Find(".user-details>a").Text()

	return &StackOverflowResult{
		singleBreak(title),
		singleBreak(author),
		singleBreak(question),
		singleBreak(bestAnswer),
		singleBreak(bestAnswerAuthor),
	}, nil
}
