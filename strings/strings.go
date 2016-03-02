package strings

import (
	"strings"
	"unicode/utf8"
)

func Fold(s string, width int) string {
	runes := []rune(s)
	lines := []string{}
	linew := 0
	lastSlice := 0
	lastSpace := 0
	widthSinceSpace := 0
	for i := 0; i < len(runes); i++ {
		b := utf8.RuneLen(runes[i])
		w := 1
		// Make the rough and severely wrong assumption
		// that a multibyte utf8 rune contains a multi
		// column grapheme
		if b > 1 {
			w = 2
		}

		switch runes[i] {
		case '\n':
			lines = append(
				lines,
				string(runes[lastSlice:i]),
			)

			linew = 0
			lastSpace = 0
			widthSinceSpace = 0
			lastSlice = i + 1
			//i++
			continue
		case ' ':
			lastSpace = i
			widthSinceSpace = 0
		case '\t':
			// TODO determine tabsize
			w = 8
		}

		if linew+w > width {
			offset := 1

			// No space found but limit reached
			if lastSpace <= lastSlice {
				offset = 0 // No space to skip
				lastSpace = i
				widthSinceSpace = 1
			}

			lines = append(
				lines,
				string(runes[lastSlice:lastSpace]),
			)

			lastSlice = lastSpace + offset
			linew = widthSinceSpace - 1
		}

		widthSinceSpace += w
		linew += w
	}

	if lastSlice <= len(runes) {
		lines = append(
			lines,
			string(runes[lastSlice:]),
		)
	}

	return strings.Join(lines, "\n")
}
