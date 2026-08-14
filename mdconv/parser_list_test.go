package mdconv

import "testing"

func TestParserGroupsListsAndKeepsNestedEmphasis(t *testing.T) {
	lexer := NewLexer("_You **can** combine them_\n\n- Item 1\n- Item 2\n- Item 2a\n- Item 2b\n  - Item 3a\n  - Item 3b\n\n1. Item 1\n2. Item 2\n3. Item 3\n   1. Item 3a\n   2. Item 3b\n")
	parser := NewParser(lexer.Tokenize())
	nodes := parser.Parse()

	if len(nodes) != 3 {
		t.Fatalf("esperava 3 nós, recebeu %d: %#v", len(nodes), nodes)
	}

	paragraph := nodes[0]
	if paragraph.Type != NodeParagraph {
		t.Fatalf("esperava parágrafo, recebeu %s", paragraph.Type)
	}
	if len(paragraph.Children) != 1 {
		t.Fatalf("esperava 1 filho no parágrafo, recebeu %d: %#v", len(paragraph.Children), paragraph.Children)
	}
	if paragraph.Children[0].Type != NodeItalic {
		t.Fatalf("esperava ITALIC externo, recebeu %s", paragraph.Children[0].Type)
	}
	if len(paragraph.Children[0].Children) != 3 {
		t.Fatalf("esperava 3 filhos dentro do ITALIC, recebeu %d: %#v", len(paragraph.Children[0].Children), paragraph.Children[0].Children)
	}
	if paragraph.Children[0].Children[1].Type != NodeBold {
		t.Fatalf("esperava BOLD aninhado dentro do ITALIC, recebeu %s", paragraph.Children[0].Children[1].Type)
	}

	unordered := nodes[1]
	if unordered.Type != NodeUnorderedList {
		t.Fatalf("esperava ULIST, recebeu %s", unordered.Type)
	}
	if len(unordered.Children) != 4 {
		t.Fatalf("esperava 4 itens no ULIST, recebeu %d: %#v", len(unordered.Children), unordered.Children)
	}
	if unordered.Children[3].Type != NodeListItem {
		t.Fatalf("esperava LIST_ITEM, recebeu %s", unordered.Children[3].Type)
	}
	if len(unordered.Children[3].Children) != 2 {
		t.Fatalf("esperava filhos inline + lista aninhada no último item, recebeu %#v", unordered.Children[3].Children)
	}

	ordered := nodes[2]
	if ordered.Type != NodeOrderedList {
		t.Fatalf("esperava OLIST, recebeu %s", ordered.Type)
	}
	if len(ordered.Children) != 3 {
		t.Fatalf("esperava 3 itens no OLIST, recebeu %d", len(ordered.Children))
	}
	if len(ordered.Children[2].Children) != 2 {
		t.Fatalf("esperava lista aninhada no terceiro item, recebeu %#v", ordered.Children[2].Children)
	}
}
