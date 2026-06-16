package chunk

import (
	"errors"
	"strings"
	"unicode"
)

// Chunk is a rough plain-text segment with byte offsets into the source text.
type Chunk struct {
	Index     int
	StartByte int
	EndByte   int
	Text      string
}

type Options struct {
	MaxWords int
}

// Split creates deterministic rough chunks before LLM boundary refinement.
func Split(text string, opts Options) ([]Chunk, error) {
	if opts.MaxWords < 1 {
		return nil, errors.New("max words must be at least 1")
	}

	spans := wordSpans(text)
	if len(spans) == 0 {
		return nil, nil
	}

	var chunks []Chunk
	for startWord := 0; startWord < len(spans); {
		endWord := startWord + opts.MaxWords
		if endWord > len(spans) {
			endWord = len(spans)
		}
		if endWord < len(spans) {
			if paragraphEnd, ok := paragraphBreakBefore(text, spans, startWord, endWord); ok {
				endWord = paragraphEnd
			}
		}

		startByte := spans[startWord].start
		endByte := spans[endWord-1].end
		chunks = append(chunks, Chunk{
			Index:     len(chunks),
			StartByte: startByte,
			EndByte:   endByte,
			Text:      strings.TrimSpace(text[startByte:endByte]),
		})
		startWord = endWord
	}
	return chunks, nil
}

type span struct {
	start int
	end   int
}

func wordSpans(text string) []span {
	var spans []span
	inWord := false
	start := 0
	for idx, r := range text {
		if unicode.IsSpace(r) {
			if inWord {
				spans = append(spans, span{start: start, end: idx})
				inWord = false
			}
			continue
		}
		if !inWord {
			start = idx
			inWord = true
		}
	}
	if inWord {
		spans = append(spans, span{start: start, end: len(text)})
	}
	return spans
}

func paragraphBreakBefore(text string, spans []span, startWord, maxEndWord int) (int, bool) {
	for word := maxEndWord; word > startWord+1; word-- {
		between := text[spans[word-1].end:spans[word].start]
		if strings.Contains(between, "\n\n") || strings.Contains(between, "\r\n\r\n") {
			return word, true
		}
	}
	return 0, false
}
