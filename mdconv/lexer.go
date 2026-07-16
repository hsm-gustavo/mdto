package mdconv

import (
	"strings"
)

type Lexer struct {
	reader *Reader
}

func NewLexer(input string) *Lexer {
	return &Lexer{
		reader: NewReader(input),
	}
}

/* 
Lê uma linha do texto de entrada
Enquanto não chegar no fim do arquivo, lê o caractere atual e adiciona ao builder.
Se encontrar uma quebra de linha, lê o próximo caractere e retorna a linha lida.
Se encontrar um retorno de carret, lê o próximo caractere e verifica se é uma quebra de linha. Se for, lê o próximo caractere e retorna a linha lida.
*/
func (l *Lexer) readLine() (string, Position) {
	start := l.reader.Position()
	var builder strings.Builder

	for !l.reader.EOF() {
		current := l.reader.Current()
		if current == '\n' {
			l.reader.Next()
			break
		}
		if current == '\r' {
			l.reader.Next()
			if !l.reader.EOF() && l.reader.Current() == '\n' {
				l.reader.Next()
			}
			break
		}

		builder.WriteByte(current)
		l.reader.Next()
	}

	return builder.String(), start
}

/* 
Lê o próximo token do texto de entrada.
Enquanto não chegar no fim do arquivo, lê uma linha do texto de entrada.
Se a linha lida for vazia, continua para a próxima iteração.
Tenta identificar o tipo de token da linha lida, verificando se é um bloco de código, título, linha horizontal, citação, lista não ordenada ou lista ordenada.
Se encontrar um token válido, retorna o token.
Se não encontrar nenhum token válido, retorna um token do tipo PARAGRAPH com o conteúdo da linha lida.
Se chegar no fim do arquivo, retorna um token do tipo EOF.
*/
func (l *Lexer) NextToken() Token {
	for !l.reader.EOF() {
		line, pos := l.readLine()
		if strings.TrimSpace(line) == "" {
			continue
		}

		if token, ok := l.lexCodeBlock(line, pos); ok {
			return token
		}
		if token, ok := l.lexHeading(line, pos); ok {
			return token
		}
		if token, ok := l.lexHorizontalRule(line, pos); ok {
			return token
		}
		if token, ok := l.lexBlockQuote(line, pos); ok {
			return token
		}
		if token, ok := l.lexUnorderedList(line, pos); ok {
			return token
		}
		if token, ok := l.lexOrderedList(line, pos); ok {
			return token
		}

		return Token{
			Type:     PARAGRAPH,
			Literal:  strings.TrimSpace(line),
			Position: pos,
		}
	}

	return Token{
		Type:     EOF,
		Position: l.reader.Position(),
	}
}

/* 
Tokeniza o texto de entrada em uma sequência de tokens.
Enquanto não chegar no fim do arquivo, lê o próximo token e adiciona à lista de tokens.
Se o token lido for do tipo EOF, retorna a lista de tokens.
*/
func (l *Lexer) Tokenize() []Token {
	tokens := make([]Token, 0)

	for {
		token := l.NextToken()
		tokens = append(tokens, token)
		if token.Type == EOF {
			return tokens
		}
	}
}

/* 
Tokeniza o texto inline em uma sequência de tokens.
*/
func (l *Lexer) TokenizeInline(input string) []Token {
	tokens := make([]Token, 0)
	index := 0
	position := Position{Line: 1, Column: 1}

	for index < len(input) {
		if token, consumed, ok := lexInlineToken(input, index, position); ok {
			tokens = append(tokens, token)
			index += consumed
			position.Column += consumed
			position.Offset += consumed
			continue
		}

		start := index
		for index < len(input) && !isInlineMarker(input, index) {
			index++
		}
		if index > start {
			literal := input[start:index]
			tokens = append(tokens, Token{
				Type:     TEXT,
				Literal:  literal,
				Position: shiftPosition(position, start),
			})
			position.Column += index - start
			position.Offset += index - start
			continue
		}

		tokens = append(tokens, Token{
			Type:     TEXT,
			Literal:  string(input[index]),
			Position: position,
		})
		index++
		position.Column++
		position.Offset++
	}

	tokens = append(tokens, Token{Type: EOF, Position: position})
	return tokens
}

func (l *Lexer) lexCodeBlock(line string, pos Position) (Token, bool) {
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, "```") {
		return Token{}, false
	}

	language := strings.TrimSpace(trimmed[3:])
	body := l.readFencedBlockBody()

	return Token{
		Type:     CODE_BLOCK,
		Literal:  body,
		Position: pos,
		Language: language,
	}, true
}

func (l *Lexer) readFencedBlockBody() string {
	var builder strings.Builder
	firstLine := true

	for !l.reader.EOF() {
		line, _ := l.readLine()
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") {
			break
		}
		if !firstLine {
			builder.WriteByte('\n')
		}
		builder.WriteString(line)
		firstLine = false
	}

	return builder.String()
}

func (l *Lexer) lexHeading(line string, pos Position) (Token, bool) {
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, "#") {
		return Token{}, false
	}

	level := 0
	for level < len(trimmed) && trimmed[level] == '#' {
		level++
	}
	if level == 0 || level > 6 {
		return Token{}, false
	}
	if level >= len(trimmed) || trimmed[level] != ' ' {
		return Token{}, false
	}

	return Token{
		Type:     HEADING,
		Literal:  strings.TrimSpace(trimmed[level:]),
		Position: pos,
		Level:    level,
	}, true
}

func (l *Lexer) lexHorizontalRule(line string, pos Position) (Token, bool) {
	trimmed := strings.TrimSpace(line)
	if !isHorizontalRule(trimmed) {
		return Token{}, false
	}

	return Token{
		Type:     HR,
		Literal:  trimmed,
		Position: pos,
	}, true
}

func (l *Lexer) lexBlockQuote(line string, pos Position) (Token, bool) {
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, ">") {
		return Token{}, false
	}

	content := strings.TrimSpace(trimmed[1:])
	return Token{
		Type:     BLOCKQUOTE,
		Literal:  content,
		Position: pos,
	}, true
}

func (l *Lexer) lexUnorderedList(line string, pos Position) (Token, bool) {
	leading := leadingWhitespaceCount(line)
	trimmed := strings.TrimLeft(line, " \t")
	if len(trimmed) < 2 {
		return Token{}, false
	}

	marker := trimmed[0]
	if marker != '-' && marker != '*' && marker != '+' {
		return Token{}, false
	}
	if trimmed[1] != ' ' && trimmed[1] != '\t' {
		return Token{}, false
	}

	return Token{
		Type:     ULIST,
		Literal:  strings.TrimSpace(trimmed[1:]),
		Position: shiftedPosition(pos, leading),
	}, true
}

func (l *Lexer) lexOrderedList(line string, pos Position) (Token, bool) {
	leading := leadingWhitespaceCount(line)
	trimmed := strings.TrimLeft(line, " \t")
	index := 0
	for index < len(trimmed) && trimmed[index] >= '0' && trimmed[index] <= '9' {
		index++
	}
	if index == 0 || index+1 >= len(trimmed) {
		return Token{}, false
	}
	if trimmed[index] != '.' && trimmed[index] != ')' {
		return Token{}, false
	}
	if trimmed[index+1] != ' ' && trimmed[index+1] != '\t' {
		return Token{}, false
	}

	return Token{
		Type:     OLIST,
		Literal:  strings.TrimSpace(trimmed[index+1:]),
		Position: shiftedPosition(pos, leading),
	}, true
}

func lexInlineToken(input string, index int, pos Position) (Token, int, bool) {
	if token, consumed, ok := lexLinkOrImageToken(input, index, pos); ok {
		return token, consumed, true
	}
	if token, consumed, ok := lexDelimitedToken(input, index, pos, "~~", "~~", "", STRIKETHROUGH); ok {
		return token, consumed, true
	}
	if token, consumed, ok := lexDelimitedToken(input, index, pos, "**", "**", "", BOLD); ok {
		return token, consumed, true
	}
	if token, consumed, ok := lexDelimitedToken(input, index, pos, "__", "__", "", BOLD); ok {
		return token, consumed, true
	}
	if token, consumed, ok := lexDelimitedToken(input, index, pos, "`", "`", "", CODE_INLINE); ok {
		return token, consumed, true
	}
	if token, consumed, ok := lexDelimitedToken(input, index, pos, "*", "*", "", ITALIC); ok {
		return token, consumed, true
	}
	if token, consumed, ok := lexDelimitedToken(input, index, pos, "_", "_", "", ITALIC); ok {
		return token, consumed, true
	}

	return Token{}, 0, false
}

func lexLinkOrImageToken(input string, index int, pos Position) (Token, int, bool) {
	isImage := strings.HasPrefix(input[index:], "![")
	if !isImage && !strings.HasPrefix(input[index:], "[") {
		return Token{}, 0, false
	}

	openLength := 1
	tokenType := LINK
	if isImage {
		openLength = 2
		tokenType = IMAGE
	}

	contentStart := index + openLength
	textEnd := strings.IndexByte(input[contentStart:], ']')
	if textEnd < 0 {
		return Token{}, 0, false
	}

	if contentStart+textEnd+1 >= len(input) || input[contentStart+textEnd+1] != '(' {
		return Token{}, 0, false
	}

	text := input[contentStart : contentStart+textEnd]
	urlStart := contentStart + textEnd + 2
	urlEnd := strings.IndexByte(input[urlStart:], ')')
	if urlEnd < 0 {
		return Token{}, 0, false
	}

	url := input[urlStart : urlStart+urlEnd]
	consumed := urlStart + urlEnd + 1 - index
	return Token{
		Type:     tokenType,
		Literal:  text,
		Position: shiftPosition(pos, index),
		Title:    text,
		URL:      url,
	}, consumed, true
}

func lexDelimitedToken(input string, index int, pos Position, open string, close string, separator string, tokenType TokenType) (Token, int, bool) {
	if !strings.HasPrefix(input[index:], open) {
		return Token{}, 0, false
	}

	contentStart := index + len(open)
	if separator == "" {
		contentEnd := strings.Index(input[contentStart:], close)
		if contentEnd < 0 {
			return Token{}, 0, false
		}

		literal := input[contentStart : contentStart+contentEnd]
		consumed := contentStart + contentEnd + len(close) - index
		return Token{
			Type:     tokenType,
			Literal:  literal,
			Position: shiftPosition(pos, index),
		}, consumed, true
	}

	textEnd := strings.Index(input[contentStart:], separator)
	if textEnd < 0 {
		return Token{}, 0, false
	}

	text := input[contentStart : contentStart+textEnd]
	urlStart := contentStart + textEnd + len(separator)
	urlEnd := strings.Index(input[urlStart:], close)
	if urlEnd < 0 {
		return Token{}, 0, false
	}

	url := input[urlStart : urlStart+urlEnd]
	consumed := urlStart + urlEnd + len(close) - index
	return Token{
		Type:     tokenType,
		Literal:  text,
		Position: shiftPosition(pos, index),
		Title:    text,
		URL:      url,
	}, consumed, true
}

func isInlineMarker(input string, index int) bool {
	if index >= len(input) {
		return false
	}

	switch input[index] {
	case '!', '[', '*', '_', '~', '`':
		return true
	default:
		return false
	}
}

func isHorizontalRule(line string) bool {
	if len(line) < 3 {
		return false
	}

	compact := strings.Builder{}
	for i := 0; i < len(line); i++ {
		if line[i] == ' ' || line[i] == '\t' {
			continue
		}
		compact.WriteByte(line[i])
	}

	value := compact.String()
	if len(value) < 3 {
		return false
	}

	marker := value[0]
	if marker != '-' && marker != '*' && marker != '_' {
		return false
	}
	for i := 1; i < len(value); i++ {
		if value[i] != marker {
			return false
		}
	}

	return true
}

func shiftPosition(pos Position, offset int) Position {
	shifted := pos
	shifted.Offset += offset
	shifted.Column += offset
	return shifted
}

func shiftedPosition(pos Position, offset int) Position {
	return shiftPosition(pos, offset)
}

func leadingWhitespaceCount(line string) int {
	count := 0
	for count < len(line) {
		if line[count] != ' ' && line[count] != '\t' {
			break
		}
		count++
	}
	return count
}