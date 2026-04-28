package parser

import (
	"bytes"
	"encoding/json"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
)

// SelectionToMarkdown renders a goquery selection into markdown text.
func SelectionToMarkdown(sel *goquery.Selection) string {
	var buf bytes.Buffer
	for _, n := range sel.Nodes {
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			renderMarkdownNode(child, &buf)
		}
	}
	out := strings.TrimSpace(buf.String())
	return normalizeMarkdown(out)
}

func renderMarkdownNode(n *html.Node, buf *bytes.Buffer) {
	if n == nil {
		return
	}
	switch n.Type {
	case html.TextNode:
		text := strings.ReplaceAll(n.Data, "\n", " ")
		text = strings.ReplaceAll(text, "\t", " ")
		if text != "" {
			buf.WriteString(text)
		}
	case html.ElementNode:
		switch n.Data {
		case "h1":
			buf.WriteString("\n# ")
			renderChildren(n, buf)
			buf.WriteString("\n\n")
		case "h2":
			buf.WriteString("\n## ")
			renderChildren(n, buf)
			buf.WriteString("\n\n")
		case "h3":
			buf.WriteString("\n### ")
			renderChildren(n, buf)
			buf.WriteString("\n\n")
		case "h4":
			buf.WriteString("\n#### ")
			renderChildren(n, buf)
			buf.WriteString("\n\n")
		case "h5":
			buf.WriteString("\n##### ")
			renderChildren(n, buf)
			buf.WriteString("\n\n")
		case "h6":
			buf.WriteString("\n###### ")
			renderChildren(n, buf)
			buf.WriteString("\n\n")
		case "p", "section", "article", "div":
			renderChildren(n, buf)
			buf.WriteString("\n\n")
		case "strong", "b":
			buf.WriteString("**")
			renderChildren(n, buf)
			buf.WriteString("**")
		case "em", "i":
			buf.WriteString("*")
			renderChildren(n, buf)
			buf.WriteString("*")
		case "code":
			if n.Parent != nil && n.Parent.Data == "pre" {
				return
			}
			buf.WriteString("`")
			buf.WriteString(strings.TrimSpace(nodeText(n)))
			buf.WriteString("`")
		case "pre":
			lang := codeLanguage(n)
			buf.WriteString("\n```")
			buf.WriteString(lang)
			buf.WriteString("\n")
			buf.WriteString(strings.TrimSpace(codeTextFromPre(n)))
			buf.WriteString("\n```\n\n")
		case "blockquote":
			quoted := expandInlineBullets(renderNodeToString(n))
			lines := strings.Split(strings.TrimSpace(quoted), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line != "" {
					buf.WriteString("> ")
					buf.WriteString(line)
					buf.WriteString("\n")
				}
			}
			buf.WriteString("\n")
		case "a":
			href := attr(n, "href")
			text := strings.TrimSpace(nodeText(n))
			if text == "" {
				text = href
			}
			if href == "" {
				buf.WriteString(text)
			} else {
				buf.WriteString("[")
				buf.WriteString(text)
				buf.WriteString("](")
				buf.WriteString(href)
				buf.WriteString(")")
			}
		case "img":
			src := attr(n, "data-src")
			if src == "" {
				src = attr(n, "src")
			}
			alt := attr(n, "alt")
			buf.WriteString("![")
			buf.WriteString(alt)
			buf.WriteString("](")
			buf.WriteString(src)
			buf.WriteString(")\n\n")
		case "ul":
			for child := n.FirstChild; child != nil; child = child.NextSibling {
				if child.Type == html.ElementNode && child.Data == "li" {
					item := normalizeMarkdown(strings.TrimSpace(renderNodeToString(child)))
					item = strings.TrimSpace(item)
					if item == "" {
						continue
					}
					itemLines := strings.Split(item, "\n")
					buf.WriteString("- ")
					buf.WriteString(itemLines[0])
					buf.WriteString("\n")
					for i := 1; i < len(itemLines); i++ {
						if strings.TrimSpace(itemLines[i]) == "" {
							buf.WriteString("\n")
							continue
						}
						buf.WriteString("  ")
						buf.WriteString(itemLines[i])
						buf.WriteString("\n")
					}
					continue
				}
				if child.Type == html.ElementNode && (child.Data == "ul" || child.Data == "ol") {
					renderMarkdownNode(child, buf)
				}
			}
			buf.WriteString("\n")
		case "ol":
			index := 1
			for child := n.FirstChild; child != nil; child = child.NextSibling {
				if child.Type == html.ElementNode && child.Data == "li" {
					item := normalizeMarkdown(strings.TrimSpace(renderNodeToString(child)))
					item = strings.TrimSpace(item)
					if item == "" {
						continue
					}
					itemLines := strings.Split(item, "\n")
					buf.WriteString(strconv.Itoa(index))
					buf.WriteString(". ")
					buf.WriteString(itemLines[0])
					buf.WriteString("\n")
					for i := 1; i < len(itemLines); i++ {
						if strings.TrimSpace(itemLines[i]) == "" {
							buf.WriteString("\n")
							continue
						}
						buf.WriteString("   ")
						buf.WriteString(itemLines[i])
						buf.WriteString("\n")
					}
					index++
					continue
				}
				if child.Type == html.ElementNode && (child.Data == "ul" || child.Data == "ol") {
					renderMarkdownNode(child, buf)
				}
			}
			buf.WriteString("\n")
		case "br":
			buf.WriteString("\n")
		default:
			renderChildren(n, buf)
		}
	}
}

func renderChildren(n *html.Node, buf *bytes.Buffer) {
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		renderMarkdownNode(child, buf)
	}
}

func nodeText(n *html.Node) string {
	if n == nil {
		return ""
	}
	var buf bytes.Buffer
	var walk func(*html.Node)
	walk = func(cur *html.Node) {
		if cur.Type == html.TextNode {
			buf.WriteString(cur.Data)
		}
		for c := cur.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return strings.TrimSpace(buf.String())
}

func attr(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if a.Key == key {
			return strings.TrimSpace(a.Val)
		}
	}
	return ""
}

func normalizeMarkdown(s string) string {
	lines := strings.Split(s, "\n")
	out := make([]string, 0, len(lines))
	blank := false
	for _, line := range lines {
		t := strings.TrimRight(line, " \t")
		if strings.TrimSpace(t) == "" {
			if !blank {
				out = append(out, "")
			}
			blank = true
			continue
		}
		blank = false
		out = append(out, t)
	}
	return strings.TrimSpace(strings.Join(out, "\n")) + "\n"
}

func codeLanguage(pre *html.Node) string {
	className := attr(pre, "class")
	if strings.Contains(className, "language-") {
		parts := strings.Split(className, "language-")
		if len(parts) > 1 {
			lang := strings.Fields(parts[1])
			if len(lang) > 0 {
				return lang[0]
			}
		}
	}
	return ""
}

func renderNodeToString(n *html.Node) string {
	var b bytes.Buffer
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		renderMarkdownNode(child, &b)
	}
	return strings.TrimSpace(b.String())
}

func codeTextFromPre(pre *html.Node) string {
	var lines []string
	var walk func(*html.Node)
	walk = func(cur *html.Node) {
		if cur == nil {
			return
		}
		if cur.Type == html.ElementNode {
			if shouldIgnoreCodeNode(cur) {
				return
			}
			if cur.Data == "br" {
				lines = append(lines, "\n")
				return
			}
		}
		if cur.Type == html.TextNode {
			text := strings.ReplaceAll(cur.Data, "\u00a0", " ")
			text = strings.TrimRight(text, "\r")
			if text != "" {
				lines = append(lines, text)
			}
		}
		for c := cur.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
		if cur.Type == html.ElementNode && isCodeLineNode(cur) {
			lines = append(lines, "\n")
		}
	}
	walk(pre)
	out := strings.Join(lines, "")
	out = strings.ReplaceAll(out, "\r\n", "\n")
	out = strings.ReplaceAll(out, "\r", "\n")
	out = strings.ReplaceAll(out, "\n\n\n", "\n\n")
	out = cleanupWechatCodeArtifacts(strings.TrimSpace(out))
	out = tryPrettyJSON(out)
	return out
}

func shouldIgnoreCodeNode(n *html.Node) bool {
	className := attr(n, "class")
	if className == "" {
		return false
	}
	return containsAnyClassToken(className,
		"line-number",
		"line_numbers",
		"lineNum",
		"code-snippet__js__line-num",
		"line-counter",
	)
}

func isCodeLineNode(n *html.Node) bool {
	className := attr(n, "class")
	if className == "" {
		return false
	}
	return containsAnyClassToken(className,
		"code-snippet__line",
		"hljs-ln-line",
	)
}

func containsAnyClassToken(className string, targets ...string) bool {
	lower := strings.ToLower(className)
	for _, t := range targets {
		if strings.Contains(lower, strings.ToLower(t)) {
			return true
		}
	}
	return false
}

var (
	artifactPrefixRE = regexp.MustCompile(`^(?:[a-zA-Z]*ounter\()+`)
)

func cleanupWechatCodeArtifacts(s string) string {
	s = strings.TrimSpace(s)
	s = artifactPrefixRE.ReplaceAllString(s, "")
	// Keep content from first useful token for common snippets.
	if idx := strings.IndexAny(s, "{["); idx > 0 {
		prefix := strings.TrimSpace(s[:idx])
		if !strings.Contains(prefix, "\n") {
			s = s[idx:]
		}
	}
	// Handle compressed numbered text like "1.xxx2.yyy3.zzz".
	s = regexp.MustCompile(`([^\n])(\d+\.\s*)`).ReplaceAllString(s, "$1\n$2")
	// Some snippets leak a standalone "line" token before numbered content.
	s = regexp.MustCompile(`(?i)^line\s*\n([0-9]+\.)`).ReplaceAllString(s, "$1")
	// Some WeChat snippets leak trailing ')' artifacts after JSON bodies.
	if strings.HasPrefix(strings.TrimSpace(s), "{") || strings.HasPrefix(strings.TrimSpace(s), "[") {
		s = strings.TrimRight(s, ") \t")
	}
	return strings.TrimSpace(s)
}

func tryPrettyJSON(s string) string {
	s = strings.TrimSpace(s)
	if !(strings.HasPrefix(s, "{") && strings.HasSuffix(s, "}")) &&
		!(strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]")) {
		return s
	}
	var v any
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		return s
	}
	out, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return s
	}
	return string(out)
}

func expandInlineBullets(s string) string {
	s = strings.TrimSpace(s)
	if strings.Count(s, "•") <= 1 {
		return s
	}
	parts := strings.Split(s, "•")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		out = append(out, "• "+part)
	}
	return strings.Join(out, "\n")
}
