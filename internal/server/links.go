package server

import (
	"bytes"
	"net/url"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// mdLinkRe matches fenced/inline code (group 1) or [label](target) (groups 2–3).
var mdLinkRe = regexp.MustCompile("(?s)(```.*?```|`[^`\\n]+`)|" + `\[([^\]]*)\]\(([^)]+)\)`)

// splitHref separates path and fragment from an href (no query expected for docs).
func splitHref(href string) (pathPart, anchor string) {
	href = strings.TrimSpace(href)
	if i := strings.IndexByte(href, '#'); i >= 0 {
		return href[:i], href[i+1:]
	}
	return href, ""
}

// isExternalHref reports whether href should open outside the Peel SPA.
func isExternalHref(href string) bool {
	h := strings.TrimSpace(href)
	if h == "" {
		return false
	}
	lower := strings.ToLower(h)
	switch {
	case strings.HasPrefix(lower, "http://"), strings.HasPrefix(lower, "https://"):
		return true
	case strings.HasPrefix(lower, "mailto:"), strings.HasPrefix(lower, "//"):
		return true
	default:
		return false
	}
}

// isDocExt reports whether p ends in a served doc extension.
func isDocExt(p string) bool {
	ext := strings.ToLower(filepath.Ext(p))
	return ext == ".md" || ext == ".c4"
}

// resolveLocalDocRef resolves a markdown/local href against fromRel (slash path
// under the docs root). ok is true when the href looks like a local .md/.c4 path.
// rel is cleaned and root-relative with forward slashes; it may not exist on disk.
func resolveLocalDocRef(fromRel, href string) (rel, anchor string, ok bool) {
	href = strings.TrimSpace(href)
	if href == "" || isExternalHref(href) {
		return "", "", false
	}
	lower := strings.ToLower(href)
	if strings.HasPrefix(lower, "javascript:") || strings.HasPrefix(lower, "data:") {
		return "", "", false
	}
	if strings.HasPrefix(href, "?") {
		return "", "", false
	}

	pathPart, anchor := splitHref(href)
	if pathPart == "" {
		// Pure fragment (#section) — not a doc link.
		return "", "", false
	}
	if u, err := url.Parse(pathPart); err == nil && u.Scheme != "" {
		return "", "", false
	}
	if !isDocExt(pathPart) {
		return "", "", false
	}

	var joined string
	if strings.HasPrefix(pathPart, "/") {
		joined = strings.TrimPrefix(path.Clean("/"+pathPart), "/")
	} else {
		baseDir := path.Dir(filepath.ToSlash(fromRel))
		if baseDir == "." {
			baseDir = ""
		}
		joined = path.Clean(path.Join(baseDir, pathPart))
	}
	joined = filepath.ToSlash(joined)
	if joined == ".." || strings.HasPrefix(joined, "../") {
		return "", "", false
	}
	return joined, anchor, true
}

func attrVal(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}

func setAttr(n *html.Node, key, val string) {
	for i := range n.Attr {
		if n.Attr[i].Key == key {
			n.Attr[i].Val = val
			return
		}
	}
	n.Attr = append(n.Attr, html.Attribute{Key: key, Val: val})
}

func hasClass(n *html.Node, class string) bool {
	for _, c := range strings.Fields(attrVal(n, "class")) {
		if c == class {
			return true
		}
	}
	return false
}

func addClass(n *html.Node, class string) {
	cur := attrVal(n, "class")
	if cur == "" {
		setAttr(n, "class", class)
		return
	}
	if hasClass(n, class) {
		return
	}
	setAttr(n, "class", cur+" "+class)
}

func renderNodes(nodes []*html.Node) string {
	var buf bytes.Buffer
	for _, n := range nodes {
		_ = html.Render(&buf, n)
	}
	return buf.String()
}

// rewriteDocLinks rewrites rendered <a href> for local docs and external URLs.
func rewriteDocLinks(htmlStr, fromRel string, known map[string]bool) string {
	ctx := &html.Node{Type: html.ElementNode, Data: "div", DataAtom: atom.Div}
	nodes, err := html.ParseFragment(strings.NewReader(htmlStr), ctx)
	if err != nil {
		return htmlStr
	}
	// Attach under ctx so parent rewrites (missing → span) work.
	for _, n := range nodes {
		ctx.AppendChild(n)
	}

	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			// Already handled as a Peel wikilink / doclink.
			if attrVal(n, "data-wikilink") != "" || attrVal(n, "data-doclink") != "" {
				for c := n.FirstChild; c != nil; c = c.NextSibling {
					walk(c)
				}
				return
			}
			href := attrVal(n, "href")
			switch {
			case isExternalHref(href):
				setAttr(n, "target", "_blank")
				setAttr(n, "rel", "noopener noreferrer")
				addClass(n, "external-link")
			default:
				rel, anchor, ok := resolveLocalDocRef(fromRel, href)
				if ok {
					if known[rel] {
						newHref := "?file=" + rel
						if anchor != "" {
							newHref += "#" + anchor
						}
						setAttr(n, "href", newHref)
						setAttr(n, "data-doclink", rel)
						addClass(n, "doclink")
					} else {
						// Convert to missing span, preserving children.
						span := &html.Node{
							Type:     html.ElementNode,
							Data:     "span",
							DataAtom: atom.Span,
							Attr:     []html.Attribute{{Key: "class", Val: "wikilink-missing"}},
						}
						for c := n.FirstChild; c != nil; {
							next := c.NextSibling
							n.RemoveChild(c)
							span.AppendChild(c)
							c = next
						}
						if n.Parent != nil {
							n.Parent.InsertBefore(span, n)
							n.Parent.RemoveChild(n)
							n = span
						}
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; {
			next := c.NextSibling
			walk(c)
			c = next
		}
	}
	walk(ctx)

	var out []*html.Node
	for c := ctx.FirstChild; c != nil; c = c.NextSibling {
		out = append(out, c)
	}
	return renderNodes(out)
}

// collectMarkdownDocLinks returns local doc destinations referenced via [text](href).
func collectMarkdownDocLinks(fromRel, body string) []string {
	var out []string
	matches := mdLinkRe.FindAllStringSubmatch(body, -1)
	for _, m := range matches {
		if m[1] != "" {
			continue
		}
		target := strings.TrimSpace(m[3])
		// Angle-bracket destinations: <https://...> or <./doc.md>
		target = strings.TrimPrefix(target, "<")
		target = strings.TrimSuffix(target, ">")
		rel, _, ok := resolveLocalDocRef(fromRel, target)
		if !ok {
			continue
		}
		out = append(out, rel)
	}
	return out
}