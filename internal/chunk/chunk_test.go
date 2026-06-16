package chunk

import "testing"

func TestSplitReturnsNoChunksForWhitespace(t *testing.T) {
	chunks, err := Split(" \n\t ", Options{MaxWords: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(chunks) != 0 {
		t.Fatalf("chunks length = %d, want 0", len(chunks))
	}
}

func TestSplitByMaxWords(t *testing.T) {
	chunks, err := Split("one two three four five", Options{MaxWords: 2})
	if err != nil {
		t.Fatal(err)
	}
	if len(chunks) != 3 {
		t.Fatalf("chunks length = %d, want 3", len(chunks))
	}
	if chunks[0].Text != "one two" || chunks[1].Text != "three four" || chunks[2].Text != "five" {
		t.Fatalf("unexpected chunks: %+v", chunks)
	}
}

func TestSplitPrefersParagraphBreak(t *testing.T) {
	text := "one two three\n\nfour five six seven"
	chunks, err := Split(text, Options{MaxWords: 5})
	if err != nil {
		t.Fatal(err)
	}
	if len(chunks) != 2 {
		t.Fatalf("chunks length = %d, want 2", len(chunks))
	}
	if chunks[0].Text != "one two three" {
		t.Fatalf("first chunk = %q, want first paragraph", chunks[0].Text)
	}
	if chunks[1].Text != "four five six seven" {
		t.Fatalf("second chunk = %q, want second paragraph", chunks[1].Text)
	}
}

func TestSplitTracksByteOffsets(t *testing.T) {
	text := "  alpha beta gamma  "
	chunks, err := Split(text, Options{MaxWords: 2})
	if err != nil {
		t.Fatal(err)
	}
	if got := text[chunks[0].StartByte:chunks[0].EndByte]; got != "alpha beta" {
		t.Fatalf("first byte range = %q, want alpha beta", got)
	}
	if got := text[chunks[1].StartByte:chunks[1].EndByte]; got != "gamma" {
		t.Fatalf("second byte range = %q, want gamma", got)
	}
}

func TestSplitHandlesUTF8(t *testing.T) {
	text := "cafe\nnaive\n東京"
	chunks, err := Split(text, Options{MaxWords: 2})
	if err != nil {
		t.Fatal(err)
	}
	if len(chunks) != 2 {
		t.Fatalf("chunks length = %d, want 2", len(chunks))
	}
	if got := text[chunks[1].StartByte:chunks[1].EndByte]; got != "東京" {
		t.Fatalf("second byte range = %q, want 東京", got)
	}
}

func TestSplitRejectsInvalidOptions(t *testing.T) {
	if _, err := Split("text", Options{}); err == nil {
		t.Fatal("expected invalid options error")
	}
}
