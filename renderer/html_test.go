package renderer

import "testing"

func TestRenderMarkdownToHTML(t *testing.T) {
	r := NewHTMLRenderer()
	html := r.RenderMarkdown("# Titulo\n\nTexto com **negrito** e [link](https://example.com)\n\n```go\nfmt.Println(\"hi\")\n```\n")

	expected := "<h1>Titulo</h1><p>Texto com <strong>negrito</strong> e <a href=\"https://example.com\">link</a></p><pre><code class=\"language-go\">fmt.Println(&#34;hi&#34;)</code></pre>"
	if html != expected {
		t.Fatalf("html inesperado\nquerido: %s\nrecebido: %s", expected, html)
	}
}

func TestRenderMarkdownRendersSingleListBlock(t *testing.T) {
	r := NewHTMLRenderer()
	html := r.RenderMarkdown("- Item 1\n- Item 2\n- Item 2a\n- Item 2b\n  - Item 3a\n  - Item 3b\n")

	expected := "<ul><li>Item 1</li><li>Item 2</li><li>Item 2a</li><li>Item 2b<ul><li>Item 3a</li><li>Item 3b</li></ul></li></ul>"
	if html != expected {
		t.Fatalf("html de lista inesperado\nquerido: %s\nrecebido: %s", expected, html)
	}
}

func TestRenderMarkdownHandlesNestedEmphasis(t *testing.T) {
	r := NewHTMLRenderer()
	html := r.RenderMarkdown("_You **can** combine them_")

	expected := "<p><em>You <strong>can</strong> combine them</em></p>"
	if html != expected {
		t.Fatalf("html de ênfase aninhada inesperado\nquerido: %s\nrecebido: %s", expected, html)
	}
}

func TestRenderMarkdownRendersMultilineParagraph(t *testing.T) {
	r := NewHTMLRenderer()
	html := r.RenderMarkdown("First line\nsecond line")

	expected := "<p>First line\nsecond line</p>"
	if html != expected {
		t.Fatalf("unexpected multiline paragraph HTML\nexpected: %s\ngot: %s", expected, html)
	}
}

func TestRenderMarkdownRendersTable(t *testing.T) {
	r := NewHTMLRenderer()
	html := r.RenderMarkdown("| Name | Age |\n|------|-----|\n| **Ana** | 20 |\n| Bob | 30 |")

	expected := "<table><thead><tr><th>Name</th><th>Age</th></tr></thead><tbody><tr><td><strong>Ana</strong></td><td>20</td></tr><tr><td>Bob</td><td>30</td></tr></tbody></table>"
	if html != expected {
		t.Fatalf("unexpected table HTML\nexpected: %s\ngot: %s", expected, html)
	}
}

func TestRenderMarkdownRendersNestedBlockquoteAndTasks(t *testing.T) {
	r := NewHTMLRenderer()
	html := r.RenderMarkdown("> ## Header\n>\n> - [x] done\n> - [ ] pending\n\n<https://example.com>")

	expected := "<blockquote><h2>Header</h2><ul><li><input type=\"checkbox\" checked disabled />done</li><li><input type=\"checkbox\" disabled />pending</li></ul></blockquote><p><a href=\"https://example.com\">https://example.com</a></p>"
	if html != expected {
		t.Fatalf("unexpected HTML\nexpected: %s\ngot: %s", expected, html)
	}
}
