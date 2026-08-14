package renderer

import (
	"html"
	"strings"

	"github.com/hsm-gustavo/mdto/mdconv"
)

// HTMLRenderer converte nós da AST em HTML.
type HTMLRenderer struct{}

// NewHTMLRenderer cria uma instância de renderer HTML reutilizável.
func NewHTMLRenderer() *HTMLRenderer {
	return &HTMLRenderer{}
}

// Render converte uma lista de nós em um fragmento HTML.
func (r *HTMLRenderer) Render(nodes []*mdconv.Node) string {
	var builder strings.Builder
	for _, node := range nodes {
		builder.WriteString(r.renderNode(node))
	}
	return builder.String()
}

// RenderMarkdown executa a pipeline completa: tokeniza, faz parse e renderiza em HTML.
func (r *HTMLRenderer) RenderMarkdown(input string) string {
	lexer := mdconv.NewLexer(input)
	parser := mdconv.NewParser(lexer.Tokenize())
	return r.Render(parser.Parse())
}

// renderNode converte um nó individual no HTML correspondente.
func (r *HTMLRenderer) renderNode(node *mdconv.Node) string {
	if node == nil {
		return ""
	}

	switch node.Type {
	case mdconv.NodeHeading:
		return r.renderHeading(node)
	case mdconv.NodeParagraph:
		return r.renderParagraph(node)
	case mdconv.NodeCodeBlock:
		return r.renderCodeBlock(node)
	case mdconv.NodeBlockquote:
		return r.renderBlockquote(node)
	case mdconv.NodeUnorderedList:
		return r.renderList("ul", node)
	case mdconv.NodeOrderedList:
		return r.renderList("ol", node)
	case mdconv.NodeHorizontalRule:
		return "<hr />"
	case mdconv.NodeTable:
		return r.renderTable(node)
	case mdconv.NodeText:
		return html.EscapeString(node.Value)
	case mdconv.NodeItalic:
		return "<em>" + r.renderChildren(node.Children, node.Value) + "</em>"
	case mdconv.NodeBold:
		return "<strong>" + r.renderChildren(node.Children, node.Value) + "</strong>"
	case mdconv.NodeStrikethrough:
		return "<del>" + r.renderChildren(node.Children, node.Value) + "</del>"
	case mdconv.NodeInlineCode:
		return "<code>" + html.EscapeString(node.Value) + "</code>"
	case mdconv.NodeLink:
		return r.renderLink(node)
	case mdconv.NodeImage:
		return r.renderImage(node)
	case mdconv.NodeAutolink:
		return r.renderAutolink(node)
	case mdconv.NodeListItem:
		return r.renderChildren(node.Children, node.Value)
	default:
		return html.EscapeString(node.Value)
	}
}

// renderHeading monta a tag h1-h6 com o conteúdo inline já renderizado.
func (r *HTMLRenderer) renderHeading(node *mdconv.Node) string {
	level := node.Level
	if level < 1 {
		level = 1
	}
	if level > 6 {
		level = 6
	}

	return "<h" + string(rune('0'+level)) + ">" + r.renderChildren(node.Children, node.Value) + "</h" + string(rune('0'+level)) + ">"
}

// renderParagraph gera uma tag p com o conteúdo inline renderizado.
func (r *HTMLRenderer) renderParagraph(node *mdconv.Node) string {
	return "<p>" + r.renderChildren(node.Children, node.Value) + "</p>"
}

// renderCodeBlock preserva o conteúdo literal dentro de pre/code.
func (r *HTMLRenderer) renderCodeBlock(node *mdconv.Node) string {
	classAttr := ""
	if node.Language != "" {
		classAttr = " class=\"language-" + html.EscapeString(node.Language) + "\""
	}

	return "<pre><code" + classAttr + ">" + html.EscapeString(node.Value) + "</code></pre>"
}

// renderBlockquote cria um bloco de citação com o conteúdo inline interno.
func (r *HTMLRenderer) renderBlockquote(node *mdconv.Node) string {
	return "<blockquote>" + r.renderChildren(node.Children, node.Value) + "</blockquote>"
}

// renderList renderiza listas ordenadas e não ordenadas a partir dos filhos do nó.
func (r *HTMLRenderer) renderList(tag string, node *mdconv.Node) string {
	var builder strings.Builder
	builder.WriteString("<")
	builder.WriteString(tag)
	builder.WriteString(">")

	for _, item := range node.Children {
		builder.WriteString("<li>")
		if item.Checked != nil {
			builder.WriteString("<input type=\"checkbox\"")
			if *item.Checked {
				builder.WriteString(" checked")
			}
			builder.WriteString(" disabled />")
		}
		builder.WriteString(r.renderChildren(item.Children, item.Value))
		builder.WriteString("</li>")
	}

	builder.WriteString("</")
	builder.WriteString(tag)
	builder.WriteString(">")
	return builder.String()
}

// renderLink converte um nó de link em âncora HTML escapada.
func (r *HTMLRenderer) renderLink(node *mdconv.Node) string {
	return "<a href=\"" + html.EscapeString(node.URL) + "\">" + html.EscapeString(node.Title) + "</a>"
}

// renderImage converte um nó de imagem em uma tag img autocontida.
func (r *HTMLRenderer) renderImage(node *mdconv.Node) string {
	return "<img src=\"" + html.EscapeString(node.URL) + "\" alt=\"" + html.EscapeString(node.Title) + "\" />"
}

func (r *HTMLRenderer) renderAutolink(node *mdconv.Node) string {
	return "<a href=\"" + html.EscapeString(node.URL) + "\">" + html.EscapeString(node.Value) + "</a>"
}

func (r *HTMLRenderer) renderTable(node *mdconv.Node) string {
	var builder strings.Builder
	bodyStarted := false
	builder.WriteString("<table>")
	for _, child := range node.Children {
		switch child.Type {
		case mdconv.NodeTableHeader:
			builder.WriteString("<thead>")
			builder.WriteString(r.renderTableRow(child, "th"))
			builder.WriteString("</thead>")
		case mdconv.NodeTableRow:
			if !bodyStarted {
				builder.WriteString("<tbody>")
				bodyStarted = true
			}
			builder.WriteString(r.renderTableRow(child, "td"))
		}
	}
	if bodyStarted {
		builder.WriteString("</tbody>")
	}
	builder.WriteString("</table>")
	return builder.String()
}

func (r *HTMLRenderer) renderTableRow(row *mdconv.Node, tag string) string {
	var builder strings.Builder
	builder.WriteString("<tr>")
	for _, cell := range row.Children {
		builder.WriteString("<")
		builder.WriteString(tag)
		builder.WriteString(">")
		builder.WriteString(r.renderChildren(cell.Children, cell.Value))
		builder.WriteString("</")
		builder.WriteString(tag)
		builder.WriteString(">")
	}
	builder.WriteString("</tr>")
	return builder.String()
}

// renderChildren junta os filhos inline e, se não houver filhos, renderiza o valor bruto do nó.
func (r *HTMLRenderer) renderChildren(children []*mdconv.Node, fallback string) string {
	if len(children) == 0 {
		return html.EscapeString(fallback)
	}

	var builder strings.Builder
	for _, child := range children {
		builder.WriteString(r.renderNode(child))
	}
	return builder.String()
}
