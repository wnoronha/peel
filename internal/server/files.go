package server

import (
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

type FileEntry struct {
	Path      string `json:"path"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	IsDeck    bool   `json:"isDeck"`
	IsC4      bool   `json:"isC4"`
	Title     string `json:"title"`
	Size      int64  `json:"size"`
	ModTime   string `json:"modTime"`
	Words     int    `json:"words"`
	Links     int    `json:"links"`
	Backlinks int    `json:"backlinks"`
}

type frontMatter struct {
	Title string `yaml:"title"`
	Marp  bool   `yaml:"marp"`
}

var (
	wikiTargetRe = regexp.MustCompile("(?s)(```.*?```|`[^`\\n]+`)|" + `\[\[([^\]|#]+)(?:#[^\]|]+)?(?:\|[^\]]+)?\]\]`)
)

func parseFrontMatter(raw string) (frontMatter, string) {
	var fm frontMatter
	if !strings.HasPrefix(raw, "---\n") && !strings.HasPrefix(raw, "---\r\n") {
		return fm, raw
	}
	rest := raw[4:]
	idx := strings.Index(rest, "\n---")
	if idx < 0 {
		return fm, raw
	}
	fmText := rest[:idx]
	_ = yaml.Unmarshal([]byte(fmText), &fm)
	body := rest[idx+4:]
	body = strings.TrimLeft(body, "\r\n")
	return fm, body
}

// c4Title: first `title <name>` line in a .c4 source, else basename sans ext.
func c4Title(body, rel string) string {
	for _, ln := range strings.Split(body, "\n") {
		t := strings.TrimSpace(strings.TrimRight(ln, "\r"))
		if rest, ok := strings.CutPrefix(t, "title "); ok {
			if rest = strings.TrimSpace(rest); rest != "" {
				return rest
			}
		}
	}
	return strings.TrimSuffix(filepath.Base(rel), filepath.Ext(rel))
}

// c4Kind maps the mermaid C4 directive to the classic C4 label.
// C4Dynamic is treated as Code (the fourth C4 level).
func c4Kind(src string) string {
	for _, ln := range strings.Split(src, "\n") {
		t := strings.TrimSpace(strings.TrimRight(ln, "\r"))
		if t == "" || strings.HasPrefix(t, "%%") {
			continue
		}
		switch {
		case strings.HasPrefix(t, "C4Context"):
			return "Context"
		case strings.HasPrefix(t, "C4Container"):
			return "Container"
		case strings.HasPrefix(t, "C4Component"):
			return "Component"
		case strings.HasPrefix(t, "C4Code"), strings.HasPrefix(t, "C4Dynamic"):
			return "Code"
		case strings.HasPrefix(t, "C4Deployment"):
			return "Deployment"
		}
		break
	}
	return ""
}

// splitC4Levels: .c4 sources hold one C4 diagram per level, separated by
// a line containing exactly ---. Returns section sources.
func splitC4Levels(body string) []string {
	var out []string
	var cur []string
	flush := func() {
		s := strings.Trim(strings.Join(cur, "\n"), "\r\n")
		if strings.TrimSpace(s) != "" {
			out = append(out, s)
		}
		cur = nil
	}
	for _, ln := range strings.Split(body, "\n") {
		if strings.TrimSpace(strings.TrimRight(ln, "\r")) == "---" {
			flush()
			continue
		}
		cur = append(cur, ln)
	}
	flush()
	return out
}

func titleFor(fm frontMatter, body, rel string) string {
	if t := strings.TrimSpace(fm.Title); t != "" {
		return t
	}
	for _, ln := range strings.Split(body, "\n") {
		m := headingRe.FindStringSubmatch(strings.TrimRight(strings.TrimSpace(ln), "\r"))
		if m == nil {
			continue
		}
		text := strings.TrimSpace(strings.TrimRight(strings.TrimSpace(m[2]), "#"))
		if text != "" {
			return text
		}
	}
	return strings.TrimSuffix(filepath.Base(rel), filepath.Ext(rel))
}

type rawFile struct {
	rel string
	fm  frontMatter
	b   string
	bd  string
	sz  int64
	mt  string
}

func (s *Server) scanRawFiles() []rawFile {
	var raws []rawFile
	_ = filepath.WalkDir(s.root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			base := filepath.Base(p)
			if strings.HasPrefix(base, ".") || base == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		ext := strings.ToLower(filepath.Ext(p))
		if ext != ".md" && ext != ".c4" {
			return nil
		}
		if info, err := os.Lstat(p); err == nil && info.Mode()&os.ModeSymlink != 0 {
			return nil
		}
		rel, _ := filepath.Rel(s.root, p)
		rel = filepath.ToSlash(rel)
		info, _ := d.Info()
		content, err := os.ReadFile(p)
		if err != nil {
			return nil
		}
		fm, bd := parseFrontMatter(string(content))
		mt := ""
		var sz int64
		if info != nil {
			mt = info.ModTime().UTC().Format("2006-01-02T15:04:05Z")
			sz = info.Size()
		}
		raws = append(raws, rawFile{rel: rel, fm: fm, b: string(content), bd: bd, sz: sz, mt: mt})
		return nil
	})
	return raws
}

// targetMap returns lookup keys mapped to canonical relative paths.
func (s *Server) targetMap() map[string]string {
	lookup := make(map[string]string)
	_ = filepath.WalkDir(s.root, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			if d != nil && d.IsDir() {
				base := filepath.Base(p)
				if strings.HasPrefix(base, ".") || base == "node_modules" {
					return filepath.SkipDir
				}
			}
			return nil
		}
		ext := strings.ToLower(filepath.Ext(p))
		if ext != ".md" && ext != ".c4" {
			return nil
		}
		if info, err := os.Lstat(p); err == nil && info.Mode()&os.ModeSymlink != 0 {
			return nil
		}
		rel, _ := filepath.Rel(s.root, p)
		rel = filepath.ToSlash(rel)
		base := strings.ToLower(strings.TrimSuffix(filepath.Base(rel), filepath.Ext(rel)))
		noExt := strings.ToLower(strings.TrimSuffix(rel, filepath.Ext(rel)))
		full := strings.ToLower(rel)
		if _, ok := lookup[base]; !ok {
			lookup[base] = rel
		}
		if _, ok := lookup[noExt]; !ok {
			lookup[noExt] = rel
		}
		if _, ok := lookup[full]; !ok {
			lookup[full] = rel
		}
		return nil
	})
	return lookup
}

func (s *Server) listFiles() ([]FileEntry, error) {
	raws := s.scanRawFiles()
	targetLookup, _ := lookupFromRaws(raws)

	backlinks := make(map[string]map[string]bool)
	for _, r := range raws {
		backlinks[r.rel] = make(map[string]bool)
	}
	linkCount := make(map[string]int)

	for _, r := range raws {
		count := 0
		matches := wikiTargetRe.FindAllStringSubmatch(r.b, -1)
		for _, m := range matches {
			if m[1] != "" {
				continue
			}
			count++
			targetKey := strings.ToLower(strings.TrimSpace(m[2]))
			if dest, ok := targetLookup[targetKey]; ok && dest != r.rel {
				backlinks[dest][r.rel] = true
			}
		}
		for _, dest := range collectMarkdownDocLinks(r.rel, r.b) {
			count++
			if dest != r.rel && backlinks[dest] != nil {
				backlinks[dest][r.rel] = true
			}
		}
		linkCount[r.rel] = count
	}

	var out []FileEntry
	for _, r := range raws {
		isC4 := strings.ToLower(filepath.Ext(r.rel)) == ".c4"
		fileType := "doc"
		if isC4 {
			fileType = "c4"
		} else if r.fm.Marp {
			fileType = "deck"
		}

		title := titleFor(r.fm, r.bd, r.rel)
		if isC4 {
			title = c4Title(r.b, r.rel)
		}
		out = append(out, FileEntry{
			Path:      r.rel,
			Name:      r.rel,
			Type:      fileType,
			IsDeck:    fileType == "deck",
			IsC4:      fileType == "c4",
			Title:     title,
			Size:      r.sz,
			ModTime:   r.mt,
			Words:     len(strings.Fields(r.b)),
			Links:     linkCount[r.rel],
			Backlinks: len(backlinks[r.rel]),
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	return out, nil
}

func (s *Server) handleFiles(w http.ResponseWriter, r *http.Request) {
	files, err := s.listFiles()
	if err != nil {
		http.Error(w, "read", http.StatusInternalServerError)
		return
	}
	if files == nil {
		files = []FileEntry{}
	}
	writeJSON(w, map[string]any{"root": filepath.Base(s.root), "files": files})
}
