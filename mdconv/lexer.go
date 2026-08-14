package mdconv

import (
	"strings"
	"unicode"
)

type Lexer struct {
	reader  *Reader
	pending *line
}

type line struct {
	content  string
	position Position
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
	for {
		current, ok := l.nextLine()
		if !ok {
			return Token{
				Type:     EOF,
				Position: l.reader.Position(),
			}
		}

		if strings.TrimSpace(current.content) == "" {
			continue
		}

		if token, ok := l.lexCodeBlock(current.content, current.position); ok {
			return token
		}
		if token, ok := l.lexHeading(current.content, current.position); ok {
			return token
		}
		if token, ok := l.lexHorizontalRule(current.content, current.position); ok {
			return token
		}
		if token, ok := l.lexBlockQuote(current.content, current.position); ok {
			return token
		}
		if token, ok := l.lexUnorderedList(current.content, current.position); ok {
			return token
		}
		if token, ok := l.lexOrderedList(current.content, current.position); ok {
			return token
		}
		if token, ok := l.lexSetextHeading(current); ok {
			return token
		}
		if token, ok := l.lexTable(current); ok {
			return token
		}

		return l.lexParagraph(current)
	}
}

func (l *Lexer) nextLine() (line, bool) {
	if l.pending != nil {
		current := *l.pending
		l.pending = nil
		return current, true
	}
	if l.reader.EOF() {
		return line{}, false
	}

	content, position := l.readLine()
	return line{content: content, position: position}, true
}

func (l *Lexer) lexParagraph(first line) Token {
	lines := []string{strings.TrimSpace(first.content)}

	for {
		next, ok := l.nextLine()
		if !ok || strings.TrimSpace(next.content) == "" {
			break
		}
		if l.isBlockStart(next) {
			l.pending = &next
			break
		}

		lines = append(lines, strings.TrimSpace(next.content))
	}

	return Token{
		Type:     PARAGRAPH,
		Literal:  strings.Join(lines, "\n"),
		Position: first.position,
	}
}

func (l *Lexer) isBlockStart(current line) bool {
	trimmed := strings.TrimSpace(current.content)
	if strings.HasPrefix(trimmed, "```") {
		return true
	}
	if _, ok := l.lexHeading(current.content, current.position); ok {
		return true
	}
	if _, ok := l.lexHorizontalRule(current.content, current.position); ok {
		return true
	}
	if _, ok := blockQuoteContent(current.content); ok {
		return true
	}
	if _, ok := l.lexUnorderedList(current.content, current.position); ok {
		return true
	}
	if _, ok := l.lexOrderedList(current.content, current.position); ok {
		return true
	}
	if isTableRow(current.content) {
		return true
	}
	return false
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
	content, ok := blockQuoteContent(line)
	if !ok {
		return Token{}, false
	}

	lines := []string{content}
	for {
		next, ok := l.nextLine()
		if !ok {
			break
		}

		content, isQuote := blockQuoteContent(next.content)
		if !isQuote {
			if strings.TrimSpace(next.content) != "" {
				l.pending = &next
			}
			break
		}
		lines = append(lines, content)
	}

	return Token{
		Type:     BLOCKQUOTE,
		Literal:  strings.Join(lines, "\n"),
		Position: pos,
	}, true
}

func blockQuoteContent(line string) (string, bool) {
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, ">") {
		return "", false
	}
	return strings.TrimSpace(trimmed[1:]), true
}

func (l *Lexer) lexSetextHeading(first line) (Token, bool) {
	next, ok := l.nextLine()
	if !ok {
		return Token{}, false
	}
	level, isSetext := setextHeadingLevel(next.content)
	if !isSetext {
		l.pending = &next
		return Token{}, false
	}

	return Token{
		Type:     HEADING,
		Literal:  strings.TrimSpace(first.content),
		Position: first.position,
		Level:    level,
	}, true
}

func setextHeadingLevel(line string) (int, bool) {
	trimmed := strings.TrimSpace(line)
	if len(trimmed) == 0 {
		return 0, false
	}
	marker := trimmed[0]
	if marker != '=' && marker != '-' {
		return 0, false
	}
	for index := 1; index < len(trimmed); index++ {
		if trimmed[index] != marker {
			return 0, false
		}
	}
	if marker == '=' {
		return 1, true
	}
	return 2, true
}

func (l *Lexer) lexTable(first line) (Token, bool) {
	header := tableCells(first.content)
	if len(header) == 0 {
		return Token{}, false
	}

	separator, ok := l.nextLine()
	if !ok {
		return Token{}, false
	}
	if !isTableSeparator(separator.content, len(header)) {
		l.pending = &separator
		return Token{}, false
	}

	rows := make([][]string, 0)
	for {
		next, ok := l.nextLine()
		if !ok {
			break
		}
		cells := tableCells(next.content)
		if len(cells) != len(header) {
			if strings.TrimSpace(next.content) != "" {
				l.pending = &next
			}
			break
		}
		rows = append(rows, cells)
	}

	return Token{Type: TABLE, Position: first.position, TableHeader: header, TableRows: rows}, true
}

func isTableRow(line string) bool {
	return len(tableCells(line)) > 0
}

func tableCells(line string) []string {
	trimmed := strings.TrimSpace(line)
	if !strings.Contains(trimmed, "|") {
		return nil
	}
	trimmed = strings.TrimPrefix(trimmed, "|")
	trimmed = strings.TrimSuffix(trimmed, "|")
	parts := strings.Split(trimmed, "|")
	if len(parts) < 2 {
		return nil
	}
	for index, part := range parts {
		parts[index] = strings.TrimSpace(part)
	}
	return parts
}

func isTableSeparator(line string, columnCount int) bool {
	cells := tableCells(line)
	if len(cells) != columnCount {
		return false
	}
	for _, cell := range cells {
		if len(cell) < 3 {
			return false
		}
		for index := 0; index < len(cell); index++ {
			if cell[index] != '-' {
				return false
			}
		}
	}
	return true
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
	if token, consumed, ok := lexEscapedToken(input, index, pos); ok {
		return token, consumed, true
	}
	if token, consumed, ok := lexAutolinkToken(input, index, pos); ok {
		return token, consumed, true
	}
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

func lexEscapedToken(input string, index int, pos Position) (Token, int, bool) {
	if input[index] != '\\' || index+1 >= len(input) || !isEscapable(input[index+1]) {
		return Token{}, 0, false
	}

	return Token{Type: TEXT, Literal: string(input[index+1]), Position: shiftPosition(pos, index)}, 2, true
}

func lexAutolinkToken(input string, index int, pos Position) (Token, int, bool) {
	if input[index] != '<' {
		return Token{}, 0, false
	}

	end := strings.IndexByte(input[index+1:], '>')
	if end < 0 {
		return Token{}, 0, false
	}
	url := input[index+1 : index+1+end]
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return Token{}, 0, false
	}

	return Token{
		Type:     AUTOLINK,
		Literal:  url,
		URL:      url,
		Position: shiftPosition(pos, index),
	}, end + 2, true
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

	rawDestination := strings.TrimSpace(input[urlStart : urlStart+urlEnd])
	url := rawDestination
	title := text
	if spaceIndex := strings.IndexFunc(rawDestination, unicode.IsSpace); spaceIndex >= 0 {
		url = strings.TrimSpace(rawDestination[:spaceIndex])
		title = strings.TrimSpace(rawDestination[spaceIndex:])
		title = strings.Trim(title, " \t\r\n\"")
	}
	if strings.HasPrefix(url, "<") && strings.HasSuffix(url, ">") {
		url = strings.TrimSuffix(strings.TrimPrefix(url, "<"), ">")
	}
	if title == "" {
		title = text
	}
	consumed := urlStart + urlEnd + 1 - index
	return Token{
		Type:     tokenType,
		Literal:  text,
		Position: shiftPosition(pos, index),
		Title:    title,
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
	case '!', '[', '*', '_', '~', '`', '\\', '<':
		return true
	default:
		return false
	}
}

func isEscapable(character byte) bool {
	return strings.ContainsRune("!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~", rune(character))
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
