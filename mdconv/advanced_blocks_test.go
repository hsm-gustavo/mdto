package mdconv

import "testing"

func TestParserBuildsTableWithInlineCells(t *testing.T) {
	input := "| Name | Age |\n|------|-----|\n| **Ana** | 20 |\n| Bob | 30 |"
	nodes := NewParser(NewLexer(input).Tokenize()).Parse()

	if len(nodes) != 1 || nodes[0].Type != NodeTable {
		t.Fatalf("expected one table node, got %#v", nodes)
	}
	table := nodes[0]
	if len(table.Children) != 3 || table.Children[0].Type != NodeTableHeader {
		t.Fatalf("unexpected table structure: %#v", table.Children)
	}
	if table.Children[0].Children[0].Value != "Name" {
		t.Fatalf("expected header cell Name, got %#v", table.Children[0].Children[0])
	}
	firstDataCell := table.Children[1].Children[0]
	if firstDataCell.Type != NodeTableCell || len(firstDataCell.Children) != 1 || firstDataCell.Children[0].Type != NodeBold {
		t.Fatalf("expected inline bold table cell, got %#v", firstDataCell)
	}
}

func TestParserRequiresValidTableSeparator(t *testing.T) {
	nodes := NewParser(NewLexer("| Name | Age |\n| invalid | separator |").Tokenize()).Parse()

	for _, node := range nodes {
		if node.Type == NodeTable {
			t.Fatalf("expected no table for an invalid separator, got %#v", nodes)
		}
	}
}

func TestParserBuildsBlockquoteWithNestedBlocks(t *testing.T) {
	input := "> ## Header\n>\n> - one\n> - two\n>\n> another paragraph"
	nodes := NewParser(NewLexer(input).Tokenize()).Parse()

	if len(nodes) != 1 || nodes[0].Type != NodeBlockquote {
		t.Fatalf("expected one blockquote, got %#v", nodes)
	}
	children := nodes[0].Children
	expected := []NodeType{NodeHeading, NodeUnorderedList, NodeParagraph}
	if len(children) != len(expected) {
		t.Fatalf("expected %d blockquote children, got %#v", len(expected), children)
	}
	for index, nodeType := range expected {
		if children[index].Type != nodeType {
			t.Fatalf("child %d: expected %s, got %s", index, nodeType, children[index].Type)
		}
	}
}

func TestParserSupportsNestedBlockquotes(t *testing.T) {
	nodes := NewParser(NewLexer("> outer\n> > inner").Tokenize()).Parse()

	if len(nodes) != 1 || len(nodes[0].Children) != 2 {
		t.Fatalf("unexpected blockquote structure: %#v", nodes)
	}
	if nodes[0].Children[1].Type != NodeBlockquote {
		t.Fatalf("expected nested blockquote, got %#v", nodes[0].Children)
	}
}

func TestParserSupportsSetextHeadings(t *testing.T) {
	nodes := NewParser(NewLexer("Heading 1\n=========\n\nHeading 2\n---------\n\n---").Tokenize()).Parse()

	if len(nodes) != 3 || nodes[0].Type != NodeHeading || nodes[0].Level != 1 || nodes[1].Level != 2 || nodes[2].Type != NodeHorizontalRule {
		t.Fatalf("unexpected setext headings: %#v", nodes)
	}
}

func TestParserRecognizesTaskListItems(t *testing.T) {
	nodes := NewParser(NewLexer("- [x] implemented\n- [X] reviewed\n- [ ] pending\n- normal").Tokenize()).Parse()

	items := nodes[0].Children
	if items[0].Checked == nil || !*items[0].Checked || items[0].Value != "implemented" {
		t.Fatalf("unexpected checked task: %#v", items[0])
	}
	if items[1].Checked == nil || !*items[1].Checked || items[1].Value != "reviewed" {
		t.Fatalf("unexpected uppercase checked task: %#v", items[1])
	}
	if items[2].Checked == nil || *items[2].Checked || items[2].Value != "pending" {
		t.Fatalf("unexpected unchecked task: %#v", items[2])
	}
	if items[3].Checked != nil || items[3].Value != "normal" {
		t.Fatalf("unexpected ordinary item: %#v", items[3])
	}
}

func TestTokenizeInlineSupportsEscapesAndAutolinks(t *testing.T) {
	tokens := NewLexer("").TokenizeInline("\\*not italic\\* <https://example.com>")

	if tokens[0].Type != TEXT || tokens[0].Literal != "*" || tokens[1].Literal != "not italic" || tokens[2].Literal != "*" {
		t.Fatalf("unexpected escaped tokens: %#v", tokens)
	}
	if tokens[4].Type != AUTOLINK || tokens[4].URL != "https://example.com" {
		t.Fatalf("unexpected autolink token: %#v", tokens)
	}
}
