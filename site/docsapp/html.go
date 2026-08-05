package main

import (
	"regexp"
	"strings"
)

var (
	reTag    = regexp.MustCompile(`(?s)<[^>]*>`)
	reEntity = strings.NewReplacer(
		"&nbsp;", " ",
		"&amp;", "&",
		"&lt;", "<",
		"&gt;", ">",
		"&#39;", "'",
		"&quot;", `"`,
		// Typographic entities: the renderer strips tags, so anything
		// left undecoded would print as literal "&ldquo;" in the prose.
		"&ldquo;", "\u201c",
		"&rdquo;", "\u201d",
		"&lsquo;", "\u2018",
		"&rsquo;", "\u2019",
		"&mdash;", "\u2014",
		"&ndash;", "\u2013",
		"&hellip;", "\u2026",
	)
)

// plainHTML strips tags for Gio label display.
func plainHTML(s string) string {
	s = reEntity.Replace(s)
	s = reTag.ReplaceAllString(s, "")
	s = strings.Join(strings.Fields(s), " ")
	return strings.TrimSpace(s)
}

// htmlBlock is one renderable chunk of docs prose.
type htmlBlock struct {
	Kind    string // p | h3 | note | takeaway | table | list | stats
	Text    string
	Headers []string
	Rows    [][]string
	Items   []string
	Stats   []htmlStat
}

type htmlStat struct {
	Num, Label, Note string
}

// htmlBlocks parses docs HTML into structured blocks so Performance
// tables / stats / takeaways survive in the Gio docsapp.
func htmlBlocks(html string) []htmlBlock {
	html = strings.TrimSpace(html)
	if html == "" {
		return nil
	}
	var out []htmlBlock
	rest := html
	for len(rest) > 0 {
		rest = strings.TrimSpace(rest)
		if rest == "" {
			break
		}
		low := strings.ToLower(rest)
		switch {
		case strings.HasPrefix(low, `<div class="statgrid"`):
			inner, next := cutElement(rest)
			out = append(out, htmlBlock{Kind: "stats", Stats: parseStats(inner)})
			rest = next
		case strings.HasPrefix(low, `<div class="perf-takeaway"`):
			inner, next := cutElement(rest)
			out = append(out, htmlBlock{Kind: "takeaway", Text: plainHTML(inner)})
			rest = next
		case strings.HasPrefix(low, `<div class="proptable-wrap"`):
			inner, next := cutElement(rest)
			h, rows := parseTable(inner)
			if len(h) > 0 || len(rows) > 0 {
				out = append(out, htmlBlock{Kind: "table", Headers: h, Rows: rows})
			}
			rest = next
		case strings.HasPrefix(low, `<div class="perfbars"`):
			inner, next := cutElement(rest)
			for _, row := range splitClass(inner, "perfrow") {
				if t := plainHTML(row); t != "" {
					out = append(out, htmlBlock{Kind: "note", Text: t})
				}
			}
			rest = next
		case strings.HasPrefix(low, "<ol"):
			inner, next := cutElement(rest)
			out = append(out, htmlBlock{Kind: "list", Items: parseListItems(inner)})
			rest = next
		case strings.HasPrefix(low, "<ul"):
			inner, next := cutElement(rest)
			out = append(out, htmlBlock{Kind: "list", Items: parseListItems(inner)})
			rest = next
		case strings.HasPrefix(low, "<h3"):
			inner, next := cutElement(rest)
			out = append(out, htmlBlock{Kind: "h3", Text: plainHTML(inner)})
			rest = next
		case strings.HasPrefix(low, "<p"):
			inner, next := cutElement(rest)
			kind := "p"
			openEnd := strings.Index(rest, ">")
			if openEnd > 0 && (strings.Contains(rest[:openEnd], `perfnote`) || strings.Contains(rest[:openEnd], `note`)) {
				kind = "note"
			}
			if t := plainHTML(inner); t != "" {
				out = append(out, htmlBlock{Kind: kind, Text: t})
			}
			rest = next
		case strings.HasPrefix(low, "<table"):
			inner, next := cutElement(rest)
			h, rows := parseTable(inner)
			if len(h) > 0 || len(rows) > 0 {
				out = append(out, htmlBlock{Kind: "table", Headers: h, Rows: rows})
			}
			rest = next
		case strings.HasPrefix(rest, "<"):
			_, next := cutElement(rest)
			if next == rest {
				// failed parse — consume one char
				rest = rest[1:]
			} else {
				rest = next
			}
		default:
			if i := strings.Index(rest, "<"); i > 0 {
				if t := plainHTML(rest[:i]); t != "" {
					out = append(out, htmlBlock{Kind: "p", Text: t})
				}
				rest = rest[i:]
				continue
			}
			if t := plainHTML(rest); t != "" {
				out = append(out, htmlBlock{Kind: "p", Text: t})
			}
			rest = ""
		}
	}
	if len(out) == 0 {
		if t := plainHTML(html); t != "" {
			out = append(out, htmlBlock{Kind: "p", Text: t})
		}
	}
	return out
}

func htmlParagraphs(html string) []string {
	var out []string
	for _, b := range htmlBlocks(html) {
		switch b.Kind {
		case "p", "note", "takeaway", "h3":
			if b.Text != "" {
				out = append(out, b.Text)
			}
		case "list":
			out = append(out, b.Items...)
		}
	}
	return out
}

func cutElement(s string) (inner, rest string) {
	if !strings.HasPrefix(s, "<") {
		return "", s
	}
	gt := strings.Index(s, ">")
	if gt < 0 {
		return "", ""
	}
	open := s[1:gt]
	if strings.HasPrefix(open, "/") {
		return "", s[gt+1:]
	}
	name := open
	if i := strings.IndexAny(open, " \t\n/"); i >= 0 {
		name = open[:i]
	}
	name = strings.ToLower(name)
	if strings.HasSuffix(strings.TrimSpace(open), "/") || voidTag(name) {
		return "", s[gt+1:]
	}
	start := gt + 1
	closePat := "</" + name + ">"
	openPat := "<" + name
	depth := 1
	low := strings.ToLower(s)
	i := start
	for i < len(s) && depth > 0 {
		nextClose := strings.Index(low[i:], closePat)
		if nextClose < 0 {
			return s[start:], ""
		}
		nextOpen := strings.Index(low[i:], openPat)
		if nextOpen >= 0 && nextOpen < nextClose {
			j := i + nextOpen + len(openPat)
			if j < len(low) && (low[j] == ' ' || low[j] == '>' || low[j] == '\t' || low[j] == '\n' || low[j] == '/') {
				depth++
			}
			i = i + nextOpen + 1
			continue
		}
		depth--
		if depth == 0 {
			end := i + nextClose
			return s[start:end], s[end+len(closePat):]
		}
		i = i + nextClose + len(closePat)
	}
	return s[start:], ""
}

func voidTag(name string) bool {
	switch name {
	case "br", "hr", "img", "input", "meta", "link":
		return true
	}
	return false
}

func parseTable(html string) (headers []string, rows [][]string) {
	low := strings.ToLower(html)
	if i := strings.Index(low, "<thead"); i >= 0 {
		inner, _ := cutElement(html[i:])
		if rs := tableRows(inner); len(rs) > 0 {
			headers = rs[0]
		}
	}
	body := html
	if i := strings.Index(low, "<tbody"); i >= 0 {
		body, _ = cutElement(html[i:])
	}
	rows = tableRows(body)
	return headers, rows
}

func tableRows(html string) [][]string {
	low := strings.ToLower(html)
	var rows [][]string
	for {
		i := strings.Index(low, "<tr")
		if i < 0 {
			break
		}
		inner, rest := cutElement(html[i:])
		cells := tableCells(inner)
		if len(cells) > 0 {
			rows = append(rows, cells)
		}
		html = rest
		low = strings.ToLower(html)
	}
	return rows
}

func tableCells(html string) []string {
	low := strings.ToLower(html)
	var cells []string
	for {
		iTD := strings.Index(low, "<td")
		iTH := strings.Index(low, "<th")
		i := -1
		if iTD < 0 {
			i = iTH
		} else if iTH < 0 {
			i = iTD
		} else if iTD < iTH {
			i = iTD
		} else {
			i = iTH
		}
		if i < 0 {
			break
		}
		inner, rest := cutElement(html[i:])
		cells = append(cells, plainHTML(inner))
		html = rest
		low = strings.ToLower(html)
	}
	return cells
}

func parseListItems(html string) []string {
	low := strings.ToLower(html)
	var items []string
	for {
		i := strings.Index(low, "<li")
		if i < 0 {
			break
		}
		inner, rest := cutElement(html[i:])
		if t := plainHTML(inner); t != "" {
			items = append(items, t)
		}
		html = rest
		low = strings.ToLower(html)
	}
	return items
}

func splitClass(html, class string) []string {
	// Match class="foo" or class="foo other" — not class="foobar".
	needles := []string{`class="` + class + `"`, `class="` + class + ` `}
	var out []string
	for {
		low := strings.ToLower(html)
		i := -1
		for _, n := range needles {
			j := strings.Index(low, n)
			if j >= 0 && (i < 0 || j < i) {
				i = j
			}
		}
		if i < 0 {
			break
		}
		start := strings.LastIndex(html[:i], "<")
		if start < 0 {
			break
		}
		inner, rest := cutElement(html[start:])
		out = append(out, inner)
		html = rest
	}
	return out
}

func parseStats(html string) []htmlStat {
	var out []htmlStat
	for _, inner := range splitClass(html, "stat") {
		st := htmlStat{
			Num:   plainHTML(firstClass(inner, "statnum")),
			Label: plainHTML(firstClass(inner, "statlabel")),
			Note:  plainHTML(firstClass(inner, "statnote")),
		}
		if st.Num != "" || st.Label != "" {
			out = append(out, st)
		}
	}
	return out
}

func firstClass(html, class string) string {
	parts := splitClass(html, class)
	if len(parts) == 0 {
		return ""
	}
	return parts[0]
}
