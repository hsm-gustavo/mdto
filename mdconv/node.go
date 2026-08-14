package mdconv

// NodeType representa os tipos de nós presentes na AST.
// Ele é separado de TokenType para que lexer e parser possam evoluir independentemente.
type NodeType string

const (
	NodeParagraph      NodeType = "PARAGRAPH"
	NodeHeading        NodeType = "HEADING"
	NodeCodeBlock      NodeType = "CODE_BLOCK"
	NodeBlockquote     NodeType = "BLOCKQUOTE"
	NodeUnorderedList  NodeType = "ULIST"
	NodeOrderedList    NodeType = "OLIST"
	NodeListItem       NodeType = "LIST_ITEM"
	NodeHorizontalRule NodeType = "HR"
	NodeText           NodeType = "TEXT"
	NodeItalic         NodeType = "ITALIC"
	NodeBold           NodeType = "BOLD"
	NodeStrikethrough  NodeType = "STRIKETHROUGH"
	NodeInlineCode     NodeType = "CODE_INLINE"
	NodeLink           NodeType = "LINK"
	NodeImage          NodeType = "IMAGE"
	NodeAutolink       NodeType = "AUTOLINK"
	NodeTable          NodeType = "TABLE"
	NodeTableHeader    NodeType = "TABLE_HEADER"
	NodeTableRow       NodeType = "TABLE_ROW"
	NodeTableCell      NodeType = "TABLE_CELL"
)

type Node struct {
	Type NodeType

	// metadados do bloco
	Level    int     // nível do heading
	Language string  // linguagem do bloco de código
	Children []*Node // filhos do nó

	// metadados de texto
	Value   string // conteúdo textual bruto
	URL     string // endereço de destino para link e image
	Title   string // texto descritivo ou alternativo para link e image
	Checked *bool  // estado de um item de task list; nil quando não é uma task
}
