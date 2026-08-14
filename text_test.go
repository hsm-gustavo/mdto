package mdto

import "testing"

func TestTextConvertsMarkdownToPlainText(t *testing.T) {
	text := Text("# Hello\n\nThis is **bold**.")

	expected := "Hello\n\nThis is bold."
	if text != expected {
		t.Fatalf("expected %q, got %q", expected, text)
	}
}
