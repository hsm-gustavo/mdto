package mdto

import (
	"github.com/hsm-gustavo/mdto/mdconv"
	"github.com/hsm-gustavo/mdto/renderer"
)

// Node e NodeType expõem a AST pública sem duplicar sua representação interna.
type Node = mdconv.Node
type NodeType = mdconv.NodeType

const (
	NodeParagraph      = mdconv.NodeParagraph
	NodeHeading        = mdconv.NodeHeading
	NodeCodeBlock      = mdconv.NodeCodeBlock
	NodeBlockquote     = mdconv.NodeBlockquote
	NodeUnorderedList  = mdconv.NodeUnorderedList
	NodeOrderedList    = mdconv.NodeOrderedList
	NodeListItem       = mdconv.NodeListItem
	NodeHorizontalRule = mdconv.NodeHorizontalRule
	NodeText           = mdconv.NodeText
	NodeItalic         = mdconv.NodeItalic
	NodeBold           = mdconv.NodeBold
	NodeStrikethrough  = mdconv.NodeStrikethrough
	NodeInlineCode     = mdconv.NodeInlineCode
	NodeLink           = mdconv.NodeLink
	NodeImage          = mdconv.NodeImage
	NodeAutolink       = mdconv.NodeAutolink
	NodeTable          = mdconv.NodeTable
	NodeTableHeader    = mdconv.NodeTableHeader
	NodeTableRow       = mdconv.NodeTableRow
	NodeTableCell      = mdconv.NodeTableCell
)

// Parse converte conteúdo Markdown em uma lista de nós da AST.
func Parse(input string) []*Node {
	lexer := mdconv.NewLexer(input)
	parser := mdconv.NewParser(lexer.Tokenize())
	return parser.Parse()
}

/*
Converte o conteúdo Markdown em HTML usando a pipeline completa: tokenização, parsing e renderização.
*/
func HTML(markdownContent string) string {
	r := renderer.NewHTMLRenderer()
	return r.Render(Parse(markdownContent))
}

// Text converte conteúdo Markdown em texto puro.
func Text(markdownContent string) string {
	r := renderer.NewTextRenderer()
	return r.Render(Parse(markdownContent))
}
