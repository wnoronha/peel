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

func TestResolveLocalDocRef(t *testing.T) {
	tests := []struct {
		from, href, wantRel, wantAnchor string
		ok                              bool
	}{
		{"index.md", "./local-development.md", "local-development.md", "", true},
		{"guides/intro.md", "./local-development.md", "guides/local-development.md", "", true},
		{"guides/intro.md", "../architecture.md", "architecture.md", "", true},
		{"guides/intro.md", "../architecture.md#server", "architecture.md", "server", true},
		{"index.md", "/read.md", "read.md", "", true},
		{"index.md", "https://www.drupal.org/docs", "", "", false},
		{"index.md", "#section", "", "", false},
		{"index.md", "./assets/logo.png", "", "", false},
		{"index.md", "javascript:alert(1)", "", "", false},
		{"guides/a.md", "../../etc/passwd.md", "", "", false},
	}
	for _, tt := range tests {
		rel, anchor, ok := resolveLocalDocRef(tt.from, tt.href)
		if ok != tt.ok || rel != tt.wantRel || anchor != tt.wantAnchor {
			t.Errorf("resolveLocalDocRef(%q, %q) = (%q, %q, %v), want (%q, %q, %v)",
				tt.from, tt.href, rel, anchor, ok, tt.wantRel, tt.wantAnchor, tt.ok)
		}
	}
}

func TestRewriteDocLinks(t *testing.T) {
	known := map[string]bool{
		"local-development.md": true,
		"drupal.md":            true,
		"read.md":              true,
	}
	in := `<p><a href="./local-development.md">DDEV Setup</a></p>` +
		`<p><a href="https://www.drupal.org/docs/develop/standards">Drupal Development</a></p>` +
		`<p><a href="./missing.md">Missing</a></p>` +
		`<p><a href="./drupal.md#overview">Drupal 10</a></p>` +
		`<p><a class="wikilink" data-wikilink="read.md" href="?file=read.md">read</a></p>`

	out := rewriteDocLinks(in, "index.md", known)

	if !strings.Contains(out, `data-doclink="local-development.md"`) {
		t.Errorf("expected data-doclink for local-development, got %s", out)
	}
	if !strings.Contains(out, `href="?file=local-development.md"`) {
		t.Errorf("expected ?file= href, got %s", out)
	}
	if !strings.Contains(out, `target="_blank"`) || !strings.Contains(out, `rel="noopener noreferrer"`) {
		t.Errorf("expected external link attrs, got %s", out)
	}
	if !strings.Contains(out, `class="wikilink-missing"`) || !strings.Contains(out, `Missing`) {
		t.Errorf("expected missing span, got %s", out)
	}
	if !strings.Contains(out, `href="?file=drupal.md#overview"`) {
		t.Errorf("expected anchor preserved, got %s", out)
	}
	if !strings.Contains(out, `data-wikilink="read.md"`) {
		t.Errorf("expected wikilink left alone, got %s", out)
	}
}

func TestCollectMarkdownDocLinks(t *testing.T) {
	body := "See [DDEV](./local-development.md) and [ext](https://example.com) and `./ignored.md`.\n" +
		"```\n[fence](./fenced.md)\n```\n" +
		"Also [Drupal](./drupal.md#overview)."
	got := collectMarkdownDocLinks("guides/intro.md", body)
	want := []string{"guides/local-development.md", "guides/drupal.md"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("got[%d]=%q want %q", i, got[i], want[i])
		}
	}
}

func TestHandleDocMarkdownLinks(t *testing.T) {
	tmp := t.TempDir()
	_ = os.WriteFile(filepath.Join(tmp, "index.md"), []byte(`# Index

- [Local](./local-development.md)
- [External](https://www.drupal.org/docs/develop/standards)
- [Missing](./nope.md)
`), 0o644)
	_ = os.WriteFile(filepath.Join(tmp, "local-development.md"), []byte("# Local\n"), 0o644)

	s := New(tmp)
	req := httptest.NewRequest("GET", "/api/doc?path=index.md", nil)
	rec := httptest.NewRecorder()
	s.handleDoc(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	var resp DocResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(resp.HTML, `data-doclink="local-development.md"`) {
		t.Errorf("expected doclink rewrite in HTML: %s", resp.HTML)
	}
	if !strings.Contains(resp.HTML, `target="_blank"`) {
		t.Errorf("expected external target in HTML: %s", resp.HTML)
	}
	if !strings.Contains(resp.HTML, `wikilink-missing`) {
		t.Errorf("expected missing link styling: %s", resp.HTML)
	}
}

func TestMarkdownLinksInGraph(t *testing.T) {
	tmp := t.TempDir()
	_ = os.Mkdir(filepath.Join(tmp, "guides"), 0o755)
	_ = os.WriteFile(filepath.Join(tmp, "guides", "intro.md"), []byte("# Intro\n\nSee [Local](./local-development.md) and [[overview]].\n"), 0o644)
	_ = os.WriteFile(filepath.Join(tmp, "guides", "local-development.md"), []byte("# Local\n"), 0o644)
	_ = os.WriteFile(filepath.Join(tmp, "overview.md"), []byte("# Overview\n"), 0o644)

	s := New(tmp)
	files, err := s.listFiles()
	if err != nil {
		t.Fatal(err)
	}
	fileMap := map[string]FileEntry{}
	for _, f := range files {
		fileMap[f.Path] = f
	}
	intro := fileMap["guides/intro.md"]
	if intro.Links != 2 {
		t.Errorf("intro links=%d want 2: %+v", intro.Links, intro)
	}
	local := fileMap["guides/local-development.md"]
	if local.Backlinks != 1 {
		t.Errorf("local-development backlinks=%d want 1: %+v", local.Backlinks, local)
	}
	overview := fileMap["overview.md"]
	if overview.Backlinks != 1 {
		t.Errorf("overview backlinks=%d want 1: %+v", overview.Backlinks, overview)
	}

	bl := s.backlinksFor("guides/local-development.md")
	if len(bl) != 1 || bl[0] != "guides/intro.md" {
		t.Errorf("backlinksFor local = %v", bl)
	}
}
