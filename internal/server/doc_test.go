package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSlugify(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Hello World", "hello-world"},
		{"API v2.0 & Design", "api-v20-design"},
		{"--leading-and-trailing--", "leading-and-trailing"},
		{"Section #3: Details!", "section-3-details"},
	}
	for _, tt := range tests {
		if got := slugify(tt.input); got != tt.want {
			t.Errorf("slugify(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestExpandWikilinks(t *testing.T) {
	s := &Server{}
	targets := map[string]string{
		"architecture": "architecture.md",
		"c4-search":    "c4-search.c4",
	}

	body := "Read [[architecture]] or [[architecture|Arch Overview]] or [[architecture#server]] or [[unknown]]."
	expanded := s.expandWikilinks(body, targets)

	if !strings.Contains(expanded, `href="?file=architecture.md"`) {
		t.Errorf("expected href=?file=architecture.md in %s", expanded)
	}
	if !strings.Contains(expanded, `>Arch Overview</a>`) {
		t.Errorf("expected label 'Arch Overview' in %s", expanded)
	}
	if !strings.Contains(expanded, `href="?file=architecture.md#server"`) {
		t.Errorf("expected anchor in %s", expanded)
	}
	if !strings.Contains(expanded, `<span class="wikilink-missing">unknown</span>`) {
		t.Errorf("expected missing link span in %s", expanded)
	}
}

func TestSplitSlides(t *testing.T) {
	deck := "Slide 1\n---\nSlide 2\n---\nSlide 3"
	slides := splitSlides(deck)
	if len(slides) != 3 {
		t.Fatalf("expected 3 slides, got %d", len(slides))
	}
	if !strings.Contains(slides[0], "Slide 1") || !strings.Contains(slides[1], "Slide 2") {
		t.Errorf("unexpected slide contents: %+v", slides)
	}
}

func TestSplitSlidesIgnoresFenceHR(t *testing.T) {
	deck := "## Levels\n\n```\nC4Context\n---\nC4Container\n```\n\n---\n\n## Next"
	slides := splitSlides(deck)
	if len(slides) != 2 {
		t.Fatalf("expected 2 slides (fence --- ignored), got %d: %+v", len(slides), slides)
	}
	if !strings.Contains(slides[0], "C4Context") || !strings.Contains(slides[0], "C4Container") {
		t.Fatalf("code example should stay on slide 1: %q", slides[0])
	}
	if !strings.Contains(slides[1], "## Next") {
		t.Fatalf("slide 2 missing: %q", slides[1])
	}
}

func TestHandleDocModes(t *testing.T) {
	tmp := t.TempDir()

	docMd := "# Overview\n\nSome markdown content."
	deckMd := "---\nmarp: true\n---\n# Slide 1\n---\n# Slide 2"
	c4File := "C4Context\ntitle Architecture\nPerson(a, b)"

	_ = os.WriteFile(filepath.Join(tmp, "overview.md"), []byte(docMd), 0o644)
	_ = os.WriteFile(filepath.Join(tmp, "deck.md"), []byte(deckMd), 0o644)
	_ = os.WriteFile(filepath.Join(tmp, "arch.c4"), []byte(c4File), 0o644)

	s := New(tmp)

	// 1. Standard doc
	req := httptest.NewRequest("GET", "/api/doc?path=overview.md", nil)
	rec := httptest.NewRecorder()
	s.handleDoc(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var docResp DocResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &docResp); err != nil {
		t.Fatal(err)
	}
	if docResp.Type != "doc" || docResp.IsDeck || docResp.IsC4 || len(docResp.Levels) != 0 {
		t.Errorf("unexpected doc response: %+v", docResp)
	}

	// 2. Deck doc
	req = httptest.NewRequest("GET", "/api/doc?path=deck.md", nil)
	rec = httptest.NewRecorder()
	s.handleDoc(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var deckResp DocResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &deckResp)
	if deckResp.Type != "deck" || !deckResp.IsDeck || len(deckResp.Slides) != 2 {
		t.Errorf("unexpected deck response: %+v", deckResp)
	}

	// 3. C4 doc
	req = httptest.NewRequest("GET", "/api/doc?path=arch.c4", nil)
	rec = httptest.NewRecorder()
	s.handleDoc(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var c4Resp DocResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &c4Resp)
	if c4Resp.Type != "c4" || !c4Resp.IsC4 || len(c4Resp.Levels) != 1 || c4Resp.Title != "Architecture" {
		t.Errorf("unexpected c4 response: %+v", c4Resp)
	}

	// 4. Invalid file extension rejected
	req = httptest.NewRequest("GET", "/api/doc?path=main.go", nil)
	rec = httptest.NewRecorder()
	s.handleDoc(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for main.go, got %d", rec.Code)
	}

	// 5. Path traversal rejected
	req = httptest.NewRequest("GET", "/api/doc?path=../etc/passwd", nil)
	rec = httptest.NewRecorder()
	s.handleDoc(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for traversal, got %d", rec.Code)
	}
}

func TestWikilinksIgnoreCodeBlocks(t *testing.T) {
	s := &Server{}
	targets := map[string]string{
		"real-doc": "real-doc.md",
	}

	body := "Here is `[[ignored_code]]` and a fenced block:\n```\n[[ignored_fence]]\n```\nAnd a real link [[real-doc]]."
	expanded := s.expandWikilinks(body, targets)

	if !strings.Contains(expanded, "`[[ignored_code]]`") {
		t.Errorf("expected inline code to remain unchanged, got %s", expanded)
	}
	if !strings.Contains(expanded, "```\n[[ignored_fence]]\n```") {
		t.Errorf("expected code fence to remain unchanged, got %s", expanded)
	}
	if !strings.Contains(expanded, `href="?file=real-doc.md"`) {
		t.Errorf("expected real-doc to be expanded, got %s", expanded)
	}
}
