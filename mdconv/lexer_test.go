package mdconv

import "testing"

func TestTokenizeBlocks(t *testing.T) {
	lexer := NewLexer("# Titulo\n> citacao\n- item\n1. primeiro\n---\n```go\nfmt.Println(\"hi\")\n```\n\nparagrafo\n")

	tokens := lexer.Tokenize()

	expected := []struct {
		tokenType TokenType
		literal   string
		level     int
		language  string
	}{
		{tokenType: HEADING, literal: "Titulo", level: 1},
		{tokenType: BLOCKQUOTE, literal: "citacao"},
		{tokenType: ULIST, literal: "item"},
		{tokenType: OLIST, literal: "primeiro"},
		{tokenType: HR, literal: "---"},
		{tokenType: CODE_BLOCK, literal: "fmt.Println(\"hi\")", language: "go"},
		{tokenType: PARAGRAPH, literal: "paragrafo"},
		{tokenType: EOF},
	}

	if len(tokens) != len(expected) {
		t.Fatalf("esperava %d tokens, recebeu %d: %#v", len(expected), len(tokens), tokens)
	}

	for i, want := range expected {
		got := tokens[i]
		if got.Type != want.tokenType {
			t.Fatalf("token %d: esperava tipo %s, recebeu %s", i, want.tokenType, got.Type)
		}
		if got.Literal != want.literal {
			t.Fatalf("token %d: esperava literal %q, recebeu %q", i, want.literal, got.Literal)
		}
		if got.Level != want.level {
			t.Fatalf("token %d: esperava level %d, recebeu %d", i, want.level, got.Level)
		}
		if got.Language != want.language {
			t.Fatalf("token %d: esperava language %q, recebeu %q", i, want.language, got.Language)
		}
	}
}

func TestTokenizeInline(t *testing.T) {
	lexer := NewLexer("")
	tokens := lexer.TokenizeInline("Olá **mundo** e [link](https://example.com)")

	expected := []struct {
		tokenType TokenType
		literal   string
		url       string
		title     string
	}{
		{tokenType: TEXT, literal: "Olá "},
		{tokenType: BOLD, literal: "mundo"},
		{tokenType: TEXT, literal: " e "},
		{tokenType: LINK, literal: "link", url: "https://example.com", title: "link"},
		{tokenType: EOF},
	}

	if len(tokens) != len(expected) {
		t.Fatalf("esperava %d tokens, recebeu %d: %#v", len(expected), len(tokens), tokens)
	}

	for i, want := range expected {
		got := tokens[i]
		if got.Type != want.tokenType {
			t.Fatalf("token %d: esperava tipo %s, recebeu %s", i, want.tokenType, got.Type)
		}
		if got.Literal != want.literal {
			t.Fatalf("token %d: esperava literal %q, recebeu %q", i, want.literal, got.Literal)
		}
		if got.URL != want.url {
			t.Fatalf("token %d: esperava url %q, recebeu %q", i, want.url, got.URL)
		}
		if got.Title != want.title {
			t.Fatalf("token %d: esperava title %q, recebeu %q", i, want.title, got.Title)
		}
	}
}