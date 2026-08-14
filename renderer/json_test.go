package renderer

import (
	"testing"

	"github.com/hsm-gustavo/mdto/mdconv"
)

func TestJSONRendererRendersReadableAST(t *testing.T) {
	lexer := mdconv.NewLexer("# Hello\n\n[Google](https://google.com)")
	nodes := mdconv.NewParser(lexer.Tokenize()).Parse()

	output, err := NewJSONRenderer().Render(nodes)
	if err != nil {
		t.Fatal(err)
	}

	expected := `[
  {
    "type": "heading",
    "level": 1,
    "children": [
      {
        "type": "text",
        "value": "Hello"
      }
    ]
  },
  {
    "type": "paragraph",
    "children": [
      {
        "type": "link",
        "value": "Google",
        "url": "https://google.com"
      }
    ]
  }
]`
	if output != expected {
		t.Fatalf("unexpected JSON\nexpected: %s\ngot: %s", expected, output)
	}
}
