package mdconv

// Parser transforma a sequência de tokens produzida pelo lexer em uma lista de nós.
type Parser struct {
	tokens []Token
	pos    int
}

// NewParser cria um parser pronto para consumir a sequência de tokens informada.
func NewParser(tokens []Token) *Parser {
	return &Parser{tokens: tokens}
}

// Parse percorre todos os tokens até EOF e converte cada token em um nó da árvore.
func (p *Parser) Parse() []*Node {
	nodes := make([]*Node, 0)

	for !p.atEnd() {
		token := p.current()
		if token.Type == EOF {
			break
		}

		if token.Type == ULIST || token.Type == OLIST {
			node := p.parseList()
			if node != nil {
				nodes = append(nodes, node)
			}
			continue
		}

		node := p.parseToken(token)
		if node != nil {
			nodes = append(nodes, node)
		}
		p.pos++
	}

	return nodes
}

// parseToken converte um token individual no nó equivalente da AST.
func (p *Parser) parseToken(token Token) *Node {
	switch token.Type {
	case HEADING:
		return &Node{
			Type:     NodeHeading,
			Level:    token.Level,
			Value:    token.Literal,
			Children: p.parseInlineChildren(token.Literal),
		}
	case PARAGRAPH:
		return &Node{
			Type:     NodeParagraph,
			Value:    token.Literal,
			Children: p.parseInlineChildren(token.Literal),
		}
	case CODE_BLOCK:
		return &Node{
			Type:     NodeCodeBlock,
			Language: token.Language,
			Value:    token.Literal,
		}
	case BLOCKQUOTE:
		quotedLexer := NewLexer(token.Literal)
		quotedParser := NewParser(quotedLexer.Tokenize())
		return &Node{
			Type:     NodeBlockquote,
			Value:    token.Literal,
			Children: quotedParser.Parse(),
		}
	case ULIST, OLIST:
		return &Node{
			Type:     nodeTypeFromTokenType(token.Type),
			Value:    token.Literal,
			Children: p.parseInlineChildren(token.Literal),
		}
	case HR:
		return &Node{Type: NodeHorizontalRule, Value: token.Literal}
	case TABLE:
		return p.parseTable(token)
	default:
		return nil
	}
}

// parseInlineChildren reaproveita a tokenização inline para montar os filhos de um bloco textual.
func (p *Parser) parseInlineChildren(input string) []*Node {
	inlineTokens := NewLexer("").TokenizeInline(input)
	children := make([]*Node, 0, len(inlineTokens))

	for _, token := range inlineTokens {
		if token.Type == EOF {
			break
		}
		child := p.inlineTokenToNode(token)
		if child != nil {
			children = append(children, child)
		}
	}

	return children
}

// inlineTokenToNode converte tokens de texto em nós folha da AST.
func (p *Parser) inlineTokenToNode(token Token) *Node {
	switch token.Type {
	case TEXT:
		return &Node{Type: NodeText, Value: token.Literal}
	case ITALIC:
		return &Node{Type: NodeItalic, Value: token.Literal, Children: p.parseInlineChildren(token.Literal)}
	case BOLD:
		return &Node{Type: NodeBold, Value: token.Literal, Children: p.parseInlineChildren(token.Literal)}
	case STRIKETHROUGH:
		return &Node{Type: NodeStrikethrough, Value: token.Literal, Children: p.parseInlineChildren(token.Literal)}
	case CODE_INLINE:
		return &Node{Type: NodeInlineCode, Value: token.Literal}
	case LINK:
		return &Node{Type: NodeLink, Value: token.Literal, URL: token.URL, Title: token.Title}
	case IMAGE:
		return &Node{Type: NodeImage, Value: token.Literal, URL: token.URL, Title: token.Title}
	case AUTOLINK:
		return &Node{Type: NodeAutolink, Value: token.Literal, URL: token.URL}
	default:
		return nil
	}
}

// parseList agrupa tokens consecutivos de lista na mesma estrutura, preservando listas aninhadas por indentação.
func (p *Parser) parseList() *Node {
	start := p.current()
	if start.Type != ULIST && start.Type != OLIST {
		return nil
	}

	listType := start.Type
	baseIndent := start.Position.Column
	items := make([]*Node, 0)

	for !p.atEnd() {
		token := p.current()
		if token.Type != listType || token.Position.Column < baseIndent {
			break
		}

		if token.Position.Column > baseIndent {
			break
		}

		item := p.parseListItem(token.Literal)
		p.pos++

		for !p.atEnd() {
			next := p.current()
			if !p.isListToken(next) || next.Position.Column <= token.Position.Column {
				break
			}

			nested := p.parseNestedList(next.Position.Column)
			if nested != nil {
				item.Children = append(item.Children, nested)
			}
		}

		items = append(items, item)
	}

	return &Node{Type: nodeTypeFromTokenType(listType), Children: items}
}

// parseNestedList cria uma lista filha quando a indentação aumenta em relação ao item atual.
func (p *Parser) parseNestedList(indent int) *Node {
	if p.atEnd() {
		return nil
	}

	start := p.current()
	if !p.isListToken(start) || start.Position.Column < indent {
		return nil
	}

	listType := start.Type
	items := make([]*Node, 0)

	for !p.atEnd() {
		token := p.current()
		if token.Type != listType || token.Position.Column < indent {
			break
		}
		if token.Position.Column > indent {
			break
		}

		item := p.parseListItem(token.Literal)
		p.pos++

		for !p.atEnd() {
			next := p.current()
			if !p.isListToken(next) || next.Position.Column <= token.Position.Column {
				break
			}

			nested := p.parseNestedList(next.Position.Column)
			if nested != nil {
				item.Children = append(item.Children, nested)
			}
		}

		items = append(items, item)
	}

	return &Node{Type: nodeTypeFromTokenType(listType), Children: items}
}

// current devolve o token na posição atual sem avançar o cursor do parser.
func (p *Parser) current() Token {
	if p.atEnd() {
		return Token{Type: EOF}
	}

	return p.tokens[p.pos]
}

// atEnd informa se o cursor já chegou ao final da lista de tokens.
func (p *Parser) atEnd() bool {
	return p.pos >= len(p.tokens)
}

// isListToken verifica se o token atual representa uma linha de lista.
func (p *Parser) isListToken(token Token) bool {
	return token.Type == ULIST || token.Type == OLIST
}

func (p *Parser) parseTable(token Token) *Node {
	headerCells := make([]*Node, 0, len(token.TableHeader))
	for _, value := range token.TableHeader {
		headerCells = append(headerCells, p.parseTableCell(value))
	}

	children := []*Node{{Type: NodeTableHeader, Children: headerCells}}
	for _, row := range token.TableRows {
		cells := make([]*Node, 0, len(row))
		for _, value := range row {
			cells = append(cells, p.parseTableCell(value))
		}
		children = append(children, &Node{Type: NodeTableRow, Children: cells})
	}

	return &Node{Type: NodeTable, Children: children}
}

func (p *Parser) parseTableCell(value string) *Node {
	return &Node{Type: NodeTableCell, Value: value, Children: p.parseInlineChildren(value)}
}

func (p *Parser) parseListItem(value string) *Node {
	content, checked := taskListState(value)
	return &Node{
		Type:     NodeListItem,
		Value:    content,
		Children: p.parseInlineChildren(content),
		Checked:  checked,
	}
}

func taskListState(value string) (string, *bool) {
	if len(value) < 3 || value[0] != '[' || value[2] != ']' {
		return value, nil
	}

	var checked bool
	switch value[1] {
	case ' ':
		checked = false
	case 'x', 'X':
		checked = true
	default:
		return value, nil
	}

	content := value[3:]
	if len(content) > 0 && (content[0] == ' ' || content[0] == '\t') {
		content = content[1:]
	}
	return content, &checked
}

func nodeTypeFromTokenType(tokenType TokenType) NodeType {
	switch tokenType {
	case ULIST:
		return NodeUnorderedList
	case OLIST:
		return NodeOrderedList
	default:
		return ""
	}
}
