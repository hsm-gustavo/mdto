package renderer

import (
	"strings"
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

func TestJSONRendererIncludesTableAndTaskListFields(t *testing.T) {
	lexer := mdconv.NewLexer("| Name | Age |\n|------|-----|\n| Ana | 20 |\n\n- [x] done")
	nodes := mdconv.NewParser(lexer.Tokenize()).Parse()

	output, err := NewJSONRenderer().Render(nodes)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output, `"type": "table"`) || !strings.Contains(output, `"type": "table_cell"`) || !strings.Contains(output, `"checked": true`) {
		t.Fatalf("expected table and task fields in JSON, got: %s", output)
	}
}
