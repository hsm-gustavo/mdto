package mdconv

import "testing"

func TestParserCreatesSingleLineParagraph(t *testing.T) {
	nodes := NewParser(NewLexer("A single paragraph.").Tokenize()).Parse()

	if len(nodes) != 1 || nodes[0].Type != NodeParagraph {
		t.Fatalf("expected one paragraph, got %#v", nodes)
	}
	if nodes[0].Value != "A single paragraph." {
		t.Fatalf("expected paragraph value %q, got %q", "A single paragraph.", nodes[0].Value)
	}
}

func TestParserCreatesMultilineParagraph(t *testing.T) {
	input := "This is a paragraph\nthat continues here\nand finishes here."
	nodes := NewParser(NewLexer(input).Tokenize()).Parse()

	if len(nodes) != 1 || nodes[0].Type != NodeParagraph {
		t.Fatalf("expected one paragraph, got %#v", nodes)
	}
	if nodes[0].Value != input {
		t.Fatalf("expected paragraph value %q, got %q", input, nodes[0].Value)
	}
	if len(nodes[0].Children) != 1 || nodes[0].Children[0].Value != input {
		t.Fatalf("expected inline parsing to preserve paragraph content, got %#v", nodes[0].Children)
	}
}

func TestParserSeparatesParagraphsAtBlankLine(t *testing.T) {
	nodes := NewParser(NewLexer("First paragraph.\n\nSecond paragraph.").Tokenize()).Parse()

	if len(nodes) != 2 || nodes[0].Type != NodeParagraph || nodes[1].Type != NodeParagraph {
		t.Fatalf("expected two paragraph nodes, got %#v", nodes)
	}
}

func TestParserSeparatesParagraphFromOtherBlocks(t *testing.T) {
	input := "Paragraph before.\n# Heading\n- List item\n> Quote\n---\n```go\nfmt.Println()\n```"
	nodes := NewParser(NewLexer(input).Tokenize()).Parse()

	expected := []NodeType{NodeParagraph, NodeHeading, NodeUnorderedList, NodeBlockquote, NodeHorizontalRule, NodeCodeBlock}
	if len(nodes) != len(expected) {
		t.Fatalf("expected %d nodes, got %d: %#v", len(expected), len(nodes), nodes)
	}
	for i, nodeType := range expected {
		if nodes[i].Type != nodeType {
			t.Fatalf("node %d: expected %s, got %s", i, nodeType, nodes[i].Type)
		}
	}
}
