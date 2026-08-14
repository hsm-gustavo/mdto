package renderer

import (
	"strconv"
	"strings"

	"github.com/hsm-gustavo/mdto/mdconv"
)

// TextRenderer converte nós da AST em texto puro legível.
type TextRenderer struct{}

// NewTextRenderer cria uma instância de renderer de texto reutilizável.
func NewTextRenderer() *TextRenderer {
	return &TextRenderer{}
}

// Render converte uma lista de nós em texto puro.
func (r *TextRenderer) Render(nodes []*mdconv.Node) string {
	blocks := make([]string, 0, len(nodes))
	for _, node := range nodes {
		if block := r.renderBlock(node); block != "" {
			blocks = append(blocks, block)
		}
	}
	return strings.Join(blocks, "\n\n")
}

// RenderMarkdown executa a pipeline completa e devolve texto puro.
func (r *TextRenderer) RenderMarkdown(input string) string {
	lexer := mdconv.NewLexer(input)
	parser := mdconv.NewParser(lexer.Tokenize())
	return r.Render(parser.Parse())
}

func (r *TextRenderer) renderBlock(node *mdconv.Node) string {
	if node == nil {
		return ""
	}

	switch node.Type {
	case mdconv.NodeHeading, mdconv.NodeParagraph:
		return r.renderChildren(node.Children, node.Value)
	case mdconv.NodeCodeBlock:
		return node.Value
	case mdconv.NodeBlockquote:
		return prefixLines(r.renderChildren(node.Children, node.Value), "> ")
	case mdconv.NodeUnorderedList:
		return r.renderList(node, false)
	case mdconv.NodeOrderedList:
		return r.renderList(node, true)
	case mdconv.NodeHorizontalRule:
		return "---"
	default:
		return r.renderInline(node)
	}
}

func (r *TextRenderer) renderInline(node *mdconv.Node) string {
	if node == nil {
		return ""
	}

	switch node.Type {
	case mdconv.NodeText, mdconv.NodeInlineCode, mdconv.NodeLink, mdconv.NodeImage:
		return node.Value
	case mdconv.NodeItalic, mdconv.NodeBold, mdconv.NodeStrikethrough, mdconv.NodeListItem:
		return r.renderChildren(node.Children, node.Value)
	default:
		return r.renderChildren(node.Children, node.Value)
	}
}

func (r *TextRenderer) renderChildren(children []*mdconv.Node, fallback string) string {
	if len(children) == 0 {
		return fallback
	}

	var builder strings.Builder
	for _, child := range children {
		builder.WriteString(r.renderInline(child))
	}
	return builder.String()
}

func (r *TextRenderer) renderList(node *mdconv.Node, ordered bool) string {
	items := make([]string, 0, len(node.Children))
	for index, item := range node.Children {
		prefix := "- "
		if ordered {
			prefix = strconv.Itoa(index+1) + ". "
		}
		items = append(items, prefix+r.renderListItem(item))
	}
	return strings.Join(items, "\n")
}

func (r *TextRenderer) renderListItem(item *mdconv.Node) string {
	if item == nil {
		return ""
	}

	inlineChildren := make([]*mdconv.Node, 0, len(item.Children))
	nestedLists := make([]*mdconv.Node, 0)
	for _, child := range item.Children {
		if child.Type == mdconv.NodeUnorderedList || child.Type == mdconv.NodeOrderedList {
			nestedLists = append(nestedLists, child)
			continue
		}
		inlineChildren = append(inlineChildren, child)
	}

	content := r.renderChildren(inlineChildren, item.Value)
	for _, nested := range nestedLists {
		content += "\n" + prefixLines(r.renderBlock(nested), "  ")
	}
	return content
}

func prefixLines(input, prefix string) string {
	if input == "" {
		return ""
	}

	lines := strings.Split(input, "\n")
	for index, line := range lines {
		lines[index] = prefix + line
	}
	return strings.Join(lines, "\n")
}
