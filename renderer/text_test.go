package renderer

import "testing"

func TestTextRendererRendersReadableMarkdown(t *testing.T) {
	renderer := NewTextRenderer()
	input := "# Hello\n\nThis is **bold** and [Google](https://google.com).\n\n- foo\n- bar"

	text := renderer.RenderMarkdown(input)
	expected := "Hello\n\nThis is bold and Google.\n\n- foo\n- bar"
	if text != expected {
		t.Fatalf("unexpected text\nexpected: %q\ngot: %q", expected, text)
	}
}

func TestTextRendererUsesImageAltTextAndPreservesCode(t *testing.T) {
	renderer := NewTextRenderer()
	input := "![Diagram](diagram.png)\n\n```go\nfmt.Println(\"hello\")\n```"

	text := renderer.RenderMarkdown(input)
	expected := "Diagram\n\nfmt.Println(\"hello\")"
	if text != expected {
		t.Fatalf("unexpected text\nexpected: %q\ngot: %q", expected, text)
	}
}

func TestTextRendererRendersNestedLists(t *testing.T) {
	renderer := NewTextRenderer()
	text := renderer.RenderMarkdown("- parent\n  - child\n\n1. first\n2. second")

	expected := "- parent\n  - child\n\n1. first\n2. second"
	if text != expected {
		t.Fatalf("unexpected nested list text\nexpected: %q\ngot: %q", expected, text)
	}
}
