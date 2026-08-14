package mdto

import "testing"

func TestParseReturnsASTNodes(t *testing.T) {
	nodes := Parse("# Title\n\nParagraph with **bold** text.")

	if len(nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %d: %#v", len(nodes), nodes)
	}
	if nodes[0].Type != NodeHeading || nodes[0].Value != "Title" {
		t.Fatalf("expected heading node, got %#v", nodes[0])
	}
	if nodes[1].Type != NodeParagraph {
		t.Fatalf("expected paragraph node, got %#v", nodes[1])
	}
	if len(nodes[1].Children) != 3 || nodes[1].Children[1].Type != NodeBold {
		t.Fatalf("expected parsed inline bold child, got %#v", nodes[1].Children)
	}
}
