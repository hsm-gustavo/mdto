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