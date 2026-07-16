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
			Type:     HEADING,
			Level:    token.Level,
			Value:    token.Literal,
			Children: p.parseInlineChildren(token.Literal),
		}
	case PARAGRAPH:
		return &Node{
			Type:     PARAGRAPH,
			Value:    token.Literal,
			Children: p.parseInlineChildren(token.Literal),
		}
	case CODE_BLOCK:
		return &Node{
			Type:     CODE_BLOCK,
			Language: token.Language,
			Value:    token.Literal,
		}
	case BLOCKQUOTE:
		return &Node{
			Type:     BLOCKQUOTE,
			Value:    token.Literal,
			Children: p.parseInlineChildren(token.Literal),
		}
	case ULIST, OLIST:
		return &Node{
			Type:     token.Type,
			Value:    token.Literal,
			Children: p.parseInlineChildren(token.Literal),
		}
	case HR:
		return &Node{Type: HR, Value: token.Literal}
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
		return &Node{Type: TEXT, Value: token.Literal}
	case ITALIC:
		return &Node{Type: ITALIC, Value: token.Literal, Children: p.parseInlineChildren(token.Literal)}
	case BOLD:
		return &Node{Type: BOLD, Value: token.Literal, Children: p.parseInlineChildren(token.Literal)}
	case STRIKETHROUGH:
		return &Node{Type: STRIKETHROUGH, Value: token.Literal, Children: p.parseInlineChildren(token.Literal)}
	case CODE_INLINE:
		return &Node{Type: CODE_INLINE, Value: token.Literal}
	case LINK:
		return &Node{Type: LINK, Value: token.Literal, URL: token.URL, Title: token.Title}
	case IMAGE:
		return &Node{Type: IMAGE, Value: token.Literal, URL: token.URL, Title: token.Title}
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

		item := &Node{
			Type:     LIST_ITEM,
			Value:    token.Literal,
			Children: p.parseInlineChildren(token.Literal),
		}
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

	return &Node{Type: listType, Children: items}
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

		item := &Node{
			Type:     LIST_ITEM,
			Value:    token.Literal,
			Children: p.parseInlineChildren(token.Literal),
		}
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

	return &Node{Type: listType, Children: items}
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
