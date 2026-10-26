package server

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseFrontMatter(t *testing.T) {
	tests := []struct {
		name      string
		raw       string
		wantTitle string
		wantMarp  bool
		wantBody  string
	}{
		{
			name:      "with marp and title",
			raw:       "---\ntitle: My Deck\nmarp: true\n---\n# Slide 1",
			wantTitle: "My Deck",
			wantMarp:  true,
			wantBody:  "# Slide 1",
		},
		{
			name:      "windows newlines",
			raw:       "---\r\ntitle: Windows Doc\r\nmarp: false\r\n---\r\nBody text",
			wantTitle: "Windows Doc",
			wantMarp:  false,
			wantBody:  "Body text",
		},
		{
			name:      "no frontmatter",
			raw:       "# Plain Title\nJust text",
			wantTitle: "",
			wantMarp:  false,
			wantBody:  "# Plain Title\nJust text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fm, body := parseFrontMatter(tt.raw)
			if fm.Title != tt.wantTitle {
				t.Errorf("got title %q, want %q", fm.Title, tt.wantTitle)
			}
			if fm.Marp != tt.wantMarp {
				t.Errorf("got marp %v, want %v", fm.Marp, tt.wantMarp)
			}
			if body != tt.wantBody {
				t.Errorf("got body %q, want %q", body, tt.wantBody)
			}
		})
	}
}

func TestC4TitleAndSplitLevels(t *testing.T) {
	c4Source := `C4Context
title System Context Diagram
Person(u, "User")
---
C4Container
title Container View
Container(c, "Web App")`

	title := c4Title(c4Source, "c4-test.c4")
	if title != "System Context Diagram" {
		t.Fatalf("expected 'System Context Diagram', got %q", title)
	}

	levels := splitC4Levels(c4Source)
	if got := c4Kind(levels[0]); got != "Context" {
		t.Fatalf("expected kind Context, got %q", got)
	}
	if got := c4Kind(levels[1]); got != "Container" {
		t.Fatalf("expected kind Container, got %q", got)
	}
	if got := c4Kind("C4Dynamic\ntitle Flow"); got != "Code" {
		t.Fatalf("expected Dynamic→Code, got %q", got)
	}
	if len(levels) != 2 {
		t.Fatalf("expected 2 levels, got %d", len(levels))
	}

	titleFallback := c4Title("C4Context\nPerson(a, b)", "nested/my-arch.c4")
	if titleFallback != "my-arch" {
		t.Fatalf("expected fallback 'my-arch', got %q", titleFallback)
	}
}

func TestTitleFor(t *testing.T) {
	fm := frontMatter{Title: "From FM"}
	if got := titleFor(fm, "# Heading\nbody", "doc.md"); got != "From FM" {
		t.Errorf("expected 'From FM', got %q", got)
	}

	fmEmpty := frontMatter{}
	if got := titleFor(fmEmpty, "# Markdown Heading\nbody", "doc.md"); got != "Markdown Heading" {
		t.Errorf("expected 'Markdown Heading', got %q", got)
	}

	if got := titleFor(fmEmpty, "no headings at all", "sub/my-notes.md"); got != "my-notes" {
		t.Errorf("expected 'my-notes', got %q", got)
	}
}

func TestListFilesAndBacklinks(t *testing.T) {
	tmp := t.TempDir()

	docA := "# Document A\n\nLinks to [[doc-b]] and [[doc-b|Doc B Piped]] and [[c4-arch#container]]."
	docB := "# Document B\n\nLinks to [[doc-a]]."
	c4Arch := "C4Context\ntitle Architecture\nPerson(u, \"User\")"

	if err := os.WriteFile(filepath.Join(tmp, "doc-a.md"), []byte(docA), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "doc-b.md"), []byte(docB), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "c4-arch.c4"), []byte(c4Arch), 0o644); err != nil {
		t.Fatal(err)
	}
	// should ignore .txt and hidden files
	_ = os.WriteFile(filepath.Join(tmp, "ignore.txt"), []byte("ignored"), 0o644)
	_ = os.Mkdir(filepath.Join(tmp, ".git"), 0o755)
	_ = os.WriteFile(filepath.Join(tmp, ".git", "hidden.md"), []byte("hidden"), 0o644)

	s := New(tmp)
	files, err := s.listFiles()
	if err != nil {
		t.Fatal(err)
	}

	if len(files) != 3 {
		t.Fatalf("expected 3 files, got %d", len(files))
	}

	fileMap := make(map[string]FileEntry)
	for _, f := range files {
		fileMap[f.Path] = f
	}

	entryA, okA := fileMap["doc-a.md"]
	if !okA || entryA.Type != "doc" || entryA.Backlinks != 1 || entryA.Links != 3 {
		t.Errorf("unexpected doc-a stats: %+v", entryA)
	}

	entryB, okB := fileMap["doc-b.md"]
	if !okB || entryB.Type != "doc" || entryB.Backlinks != 1 || entryB.Links != 1 {
		t.Errorf("unexpected doc-b stats: %+v", entryB)
	}

	entryC4, okC4 := fileMap["c4-arch.c4"]
	if !okC4 || entryC4.Type != "c4" || entryC4.Backlinks != 1 {
		t.Errorf("unexpected c4-arch stats: %+v", entryC4)
	}
}
