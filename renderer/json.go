package renderer

import (
	"encoding/json"

	"github.com/hsm-gustavo/mdto/mdconv"
)

// JSONRenderer converte nós da AST em uma representação JSON pública.
type JSONRenderer struct{}

// NewJSONRenderer cria uma instância de renderer JSON reutilizável.
func NewJSONRenderer() *JSONRenderer {
	return &JSONRenderer{}
}

// Render converte uma lista de nós em JSON formatado.
func (r *JSONRenderer) Render(nodes []*mdconv.Node) (string, error) {
	encoded, err := json.MarshalIndent(jsonNodes(nodes), "", "  ")
	if err != nil {
		return "", err
	}
	return string(encoded), nil
}

type jsonNode struct {
	Type     string      `json:"type"`
	Level    int         `json:"level,omitempty"`
	Language string      `json:"language,omitempty"`
	Value    string      `json:"value,omitempty"`
	URL      string      `json:"url,omitempty"`
	Children []*jsonNode `json:"children,omitempty"`
}

func jsonNodes(nodes []*mdconv.Node) []*jsonNode {
	result := make([]*jsonNode, 0, len(nodes))
	for _, node := range nodes {
		if node != nil {
			result = append(result, jsonNodeFromNode(node))
		}
	}
	return result
}

func jsonNodeFromNode(node *mdconv.Node) *jsonNode {
	result := &jsonNode{
		Type:     jsonNodeType(node.Type),
		Level:    node.Level,
		Language: node.Language,
		Children: jsonNodes(node.Children),
	}

	switch node.Type {
	case mdconv.NodeText, mdconv.NodeInlineCode, mdconv.NodeCodeBlock, mdconv.NodeLink, mdconv.NodeImage:
		result.Value = node.Value
	}
	if node.Type == mdconv.NodeLink || node.Type == mdconv.NodeImage {
		result.URL = node.URL
	}

	return result
}

func jsonNodeType(nodeType mdconv.NodeType) string {
	switch nodeType {
	case mdconv.NodeParagraph:
		return "paragraph"
	case mdconv.NodeHeading:
		return "heading"
	case mdconv.NodeCodeBlock:
		return "code_block"
	case mdconv.NodeBlockquote:
		return "blockquote"
	case mdconv.NodeUnorderedList:
		return "unordered_list"
	case mdconv.NodeOrderedList:
		return "ordered_list"
	case mdconv.NodeListItem:
		return "list_item"
	case mdconv.NodeHorizontalRule:
		return "horizontal_rule"
	case mdconv.NodeText:
		return "text"
	case mdconv.NodeItalic:
		return "italic"
	case mdconv.NodeBold:
		return "bold"
	case mdconv.NodeStrikethrough:
		return "strikethrough"
	case mdconv.NodeInlineCode:
		return "inline_code"
	case mdconv.NodeLink:
		return "link"
	case mdconv.NodeImage:
		return "image"
	default:
		return "unknown"
	}
}
