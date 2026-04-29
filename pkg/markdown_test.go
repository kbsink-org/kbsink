package kbsink

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	prs "github.com/kbsink-org/kbsink/pkg/parser"
)

func TestSelectionToMarkdown(t *testing.T) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(`
<div id="js_content">
  <h2>Title</h2>
  <p>text <strong>bold</strong></p>
  <pre class="language-go"><code>fmt.Println("hi")</code></pre>
  <ul><li>a</li><li>b</li></ul>
</div>`))
	if err != nil {
		t.Fatalf("build doc: %v", err)
	}
	md := prs.SelectionToMarkdown(doc.Find("#js_content"))
	wantTokens := []string{
		"## Title",
		"text **bold**",
		"```go",
		`fmt.Println("hi")`,
		"- a",
		"- b",
	}
	for _, token := range wantTokens {
		if !strings.Contains(md, token) {
			t.Fatalf("markdown missing token %q:\n%s", token, md)
		}
	}
}

func TestSelectionToMarkdown_SkipEmptyListItemsAndFormatCodeLines(t *testing.T) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(`
<div id="js_content">
  <ul>
    <li></li>
    <li><p>first line</p><p>second line</p></li>
  </ul>
  <pre class="language-json">
    <span class="code-snippet__js__line-num">1</span>
    <span class="code-snippet__line">{</span>
    <span class="code-snippet__line">  "id": "123"</span>
    <span class="code-snippet__line">}</span>
  </pre>
</div>`))
	if err != nil {
		t.Fatalf("build doc: %v", err)
	}
	md := prs.SelectionToMarkdown(doc.Find("#js_content"))
	if strings.Contains(md, "- \n") {
		t.Fatalf("unexpected empty list item in markdown:\n%s", md)
	}
	wantTokens := []string{
		"- first line",
		"second line",
		"```json",
		"{",
		`"id": "123"`,
		"}",
	}
	for _, token := range wantTokens {
		if !strings.Contains(md, token) {
			t.Fatalf("markdown missing token %q:\n%s", token, md)
		}
	}
	if strings.Contains(strings.ToLower(md), "line-num") {
		t.Fatalf("line number nodes should be filtered:\n%s", md)
	}
}

func TestSelectionToMarkdown_CleanWechatCodeArtifacts(t *testing.T) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(`
<div id="js_content">
  <pre class="language-json">eounter(lineounter(lineounter(line{"id":"abc","enabled":true})</pre>
</div>`))
	if err != nil {
		t.Fatalf("build doc: %v", err)
	}
	md := prs.SelectionToMarkdown(doc.Find("#js_content"))
	if strings.Contains(md, "lineounter(") || strings.Contains(md, "eounter(") {
		t.Fatalf("artifact prefix not cleaned:\n%s", md)
	}
	if !strings.Contains(md, `"id": "abc"`) {
		t.Fatalf("expected pretty json content:\n%s", md)
	}
}

func TestSelectionToMarkdown_BlockquoteInlineBullets(t *testing.T) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(`
<div id="js_content">
  <blockquote>• A item text • B item text</blockquote>
</div>`))
	if err != nil {
		t.Fatalf("build doc: %v", err)
	}
	md := prs.SelectionToMarkdown(doc.Find("#js_content"))
	if !strings.Contains(md, "> • A item text") {
		t.Fatalf("missing first bullet quote:\n%s", md)
	}
	if !strings.Contains(md, "> • B item text") {
		t.Fatalf("missing second bullet quote:\n%s", md)
	}
}
