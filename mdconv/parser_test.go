package mdconv

import "testing"

func TestParserBuildsNodes(t *testing.T) {
	lexer := NewLexer("# Titulo\nTexto com **negrito** e [link](https://example.com)\n```go\nfmt.Println(\"hi\")\n```\n")
	parser := NewParser(lexer.Tokenize())
	nodes := parser.Parse()

	if len(nodes) != 3 {
		t.Fatalf("esperava 3 nós, recebeu %d: %#v", len(nodes), nodes)
	}

	heading := nodes[0]
	if heading.Type != NodeHeading {
		t.Fatalf("esperava heading, recebeu %s", heading.Type)
	}
	if heading.Level != 1 {
		t.Fatalf("esperava level 1, recebeu %d", heading.Level)
	}
	if heading.Value != "Titulo" {
		t.Fatalf("esperava valor %q, recebeu %q", "Titulo", heading.Value)
	}
	if len(heading.Children) != 1 || heading.Children[0].Type != NodeText {
		t.Fatalf("esperava um filho TEXT no heading, recebeu %#v", heading.Children)
	}

	paragraph := nodes[1]
	if paragraph.Type != NodeParagraph {
		t.Fatalf("esperava paragraph, recebeu %s", paragraph.Type)
	}
	if len(paragraph.Children) != 4 {
		t.Fatalf("esperava 4 filhos no parágrafo, recebeu %d: %#v", len(paragraph.Children), paragraph.Children)
	}
	if paragraph.Children[1].Type != NodeBold {
		t.Fatalf("esperava BOLD no segundo filho, recebeu %s", paragraph.Children[1].Type)
	}
	if paragraph.Children[3].Type != NodeLink || paragraph.Children[3].URL != "https://example.com" {
		t.Fatalf("esperava LINK com URL correta, recebeu %#v", paragraph.Children[3])
	}

	codeBlock := nodes[2]
	if codeBlock.Type != NodeCodeBlock {
		t.Fatalf("esperava code block, recebeu %s", codeBlock.Type)
	}
	if codeBlock.Language != "go" {
		t.Fatalf("esperava linguagem go, recebeu %q", codeBlock.Language)
	}
	if codeBlock.Value != "fmt.Println(\"hi\")" {
		t.Fatalf("esperava valor do bloco de código, recebeu %q", codeBlock.Value)
	}
}
