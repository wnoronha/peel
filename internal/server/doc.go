package server

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

var md = goldmark.New(
	goldmark.WithExtensions(extension.GFM, extension.Footnote, extension.Typographer),
	goldmark.WithParserOptions(parser.WithAutoHeadingID()),
	goldmark.WithRendererOptions(html.WithUnsafe()),
)

var (
	headingRe  = regexp.MustCompile(`^(#{1,6})\s+(.+)$`)
	wikiSafeRe = regexp.MustCompile("(?s)(```.*?```|`[^`\\n]+`)|" + `\[\[([^\]|#]+)(?:#([^\]|]+))?(?:\|([^\]]+))?\]\]`)
)

type OutlineItem struct {
	Level int    `json:"level"`
	Text  string `json:"text"`
	Line  int    `json:"line"`
	ID    string `json:"id"`
}

type C4Level struct {
	Name   string `json:"name"`
	Kind   string `json:"kind"` // Context | Container | Component | Code | Deployment
	Source string `json:"source"`
}

type DocResponse struct {
	Path      string        `json:"path"`
	Title     string        `json:"title"`
	Type      string        `json:"type"`
	IsDeck    bool          `json:"isDeck"`
	IsC4      bool          `json:"isC4"`
	Raw       string        `json:"raw"`
	Body      string        `json:"body"`
	HTML      string        `json:"html"`
	Slides    []string      `json:"slides"`
	Levels    []C4Level     `json:"levels"`
	Outline   []OutlineItem `json:"outline"`
	Backlinks []string      `json:"backlinks"`
}

func slugify(t string) string {
	t = strings.ToLower(t)
	var b strings.Builder
	lastHyphen := true
	for _, r := range t {
		switch {
		case (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'):
			b.WriteRune(r)
			lastHyphen = false
		case r == ' ' || r == '-' || r == '_':
			if !lastHyphen {
				b.WriteRune('-')
				lastHyphen = true
			}
		}
	}
	return strings.Trim(b.String(), "-")
}

func (s *Server) expandWikilinks(body string, targetMap map[string]string) string {
	return wikiSafeRe.ReplaceAllStringFunc(body, func(m string) string {
		sub := wikiSafeRe.FindStringSubmatch(m)
		if sub[1] != "" {
			return sub[1]
		}
		target := strings.TrimSpace(sub[2])
		anchor := strings.TrimSpace(sub[3])
		label := strings.TrimSpace(sub[4])
		if label == "" {
			if anchor != "" {
				label = target + " #" + anchor
			} else {
				label = target
			}
		}
		dest := targetMap[strings.ToLower(target)]
		if dest == "" {
			return fmt.Sprintf(`<span class="wikilink-missing">%s</span>`, label)
		}
		href := fmt.Sprintf("?file=%s", dest)
		if anchor != "" {
			href += "#" + slugify(anchor)
		}
		return fmt.Sprintf(`<a class="wikilink" data-wikilink="%s" href="%s">%s</a>`, dest, href, label)
	})
}

func renderMarkdown(src string) string {
	var buf bytes.Buffer
	if err := md.Convert([]byte(src), &buf); err != nil {
		return "<pre>" + err.Error() + "</pre>"
	}
	return buf.String()
}

func tagDiagrams(htmlStr string) string {
	htmlStr = strings.ReplaceAll(htmlStr, `<pre><code class="language-mermaid"`, `<pre class="diagram" data-diagram="mermaid"><code class="language-pele"`)
	htmlStr = strings.ReplaceAll(htmlStr, `<pre><code class="language-pele"`, `<pre class="diagram" data-diagram="pele"><code class="language-pele"`)
	return htmlStr
}

func splitSlides(body string) []string {
	lines := strings.Split(body, "\n")
	var slides []string
	var cur []string
	for _, ln := range lines {
		if strings.TrimSpace(ln) == "---" {
			slides = append(slides, strings.Join(cur, "\n"))
			cur = nil
			continue
		}
		cur = append(cur, ln)
	}
	slides = append(slides, strings.Join(cur, "\n"))
	var nonEmpty []string
	for _, sl := range slides {
		if strings.TrimSpace(sl) != "" {
			nonEmpty = append(nonEmpty, sl)
		}
	}
	if len(nonEmpty) == 0 {
		return []string{body}
	}
	return nonEmpty
}

func (s *Server) handleDoc(w http.ResponseWriter, r *http.Request) {
	rel := r.URL.Query().Get("path")
	if rel == "" {
		rel = r.URL.Query().Get("file")
	}
	if rel == "" || strings.Contains(rel, "..") || filepath.IsAbs(rel) {
		http.Error(w, "bad path", http.StatusBadRequest)
		return
	}
	ext := strings.ToLower(filepath.Ext(rel))
	if ext != ".md" && ext != ".c4" {
		http.Error(w, "bad file type", http.StatusBadRequest)
		return
	}

	abs := filepath.Join(s.root, filepath.FromSlash(rel))
	rawBytes, err := os.ReadFile(abs)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	raw := string(rawBytes)

	resp := DocResponse{
		Path:      rel,
		Raw:       raw,
		Body:      raw,
		Slides:    []string{},
		Levels:    []C4Level{},
		Outline:   []OutlineItem{},
		Backlinks: s.backlinksFor(rel),
	}

	if ext == ".c4" {
		resp.Type = "c4"
		resp.IsC4 = true
		resp.Title = c4Title(raw, rel)
		secs := splitC4Levels(raw)
		for i, src := range secs {
			name := ""
			for _, ln := range strings.Split(src, "\n") {
				if rest, ok := strings.CutPrefix(strings.TrimSpace(strings.TrimRight(ln, "\r")), "title "); ok {
					if rest = strings.TrimSpace(rest); rest != "" {
						name = rest
						break
					}
				}
			}
			if name == "" {
				name = fmt.Sprintf("Level %d", i+1)
			}
			resp.Levels = append(resp.Levels, C4Level{Name: name, Kind: c4Kind(src), Source: src})
			resp.Outline = append(resp.Outline, OutlineItem{Level: 1, Text: name, Line: i + 1, ID: slugify(name)})
		}
		writeJSON(w, resp)
		return
	}

	fm, body := parseFrontMatter(raw)
	resp.Body = body
	resp.Title = titleFor(fm, body, rel)

	lines := strings.Split(body, "\n")
	for i, ln := range lines {
		m := headingRe.FindStringSubmatch(strings.TrimRight(ln, "\r"))
		if m == nil {
			continue
		}
		text := strings.TrimSpace(m[2])
		resp.Outline = append(resp.Outline, OutlineItem{Level: len(m[1]), Text: text, Line: i + 1, ID: slugify(text)})
	}

	targetMap := s.targetMap()
	known := s.knownDocSet()
	render := func(src string) string {
		return rewriteDocLinks(tagDiagrams(renderMarkdown(s.expandWikilinks(src, targetMap))), rel, known)
	}
	if fm.Marp {
		resp.Type = "deck"
		resp.IsDeck = true
		for _, sl := range splitSlides(body) {
			resp.Slides = append(resp.Slides, render(sl))
		}
		resp.HTML = strings.Join(resp.Slides, "\n")
	} else {
		resp.Type = "doc"
		resp.HTML = render(body)
	}

	writeJSON(w, resp)
}

func (s *Server) backlinksFor(rel string) []string {
	targetLookup := s.targetMap()
	raws := s.scanRawFiles()
	seen := make(map[string]bool)
	var out []string

	add := func(src string) {
		if !seen[src] {
			seen[src] = true
			out = append(out, src)
		}
	}

	for _, o := range raws {
		if o.rel == rel {
			continue
		}
		linked := false
		matches := wikiTargetRe.FindAllStringSubmatch(o.b, -1)
		for _, m := range matches {
			if m[1] != "" {
				continue
			}
			targetKey := strings.ToLower(strings.TrimSpace(m[2]))
			if dest, ok := targetLookup[targetKey]; ok && dest == rel {
				linked = true
				break
			}
		}
		if !linked {
			for _, dest := range collectMarkdownDocLinks(o.rel, o.b) {
				if dest == rel {
					linked = true
					break
				}
			}
		}
		if linked {
			add(o.rel)
		}
	}
	if out == nil {
		out = []string{}
	}
	sort.Strings(out)
	return out
}
