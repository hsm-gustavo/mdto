package mdconv

type Reader struct {
	input string

	pos     int
	readPos int

	ch byte

	line   int
	column int
}

func NewReader(input string) *Reader {
	r := &Reader{
		input: input,
		line:  1,
	}
	r.Next()

	return r
}

/* 
Se readPos for maior que o tamanho do texto significa que chegamos no fim do arquivo.
Caso contrário, ch = proximo char, pos = readPos, readPos vai ser a próxima posição.
Se for quebra de linha, aumente o número de linhas e resete a coluna, caso contrário, aumente a coluna.
*/
func (r *Reader) Next() {
	if r.readPos >= len(r.input) {
		r.ch = 0
		return
	}

	r.ch = r.input[r.readPos]
	r.pos = r.readPos
	r.readPos++

	if r.ch == '\n' {
		r.line++
		r.column = 0
	} else {
		r.column++
	}
}

/* 
Lê o caractere atual
*/
func (r *Reader) Current() byte {
	return r.ch
}

/* 
readPos sempre é o índice do próximo caractere.
Se readPos for maior que o tamanho do texto significa que chegamos no fim do arquivo.
Caso contrário, retornamos o próximo caractere.
*/
func (r *Reader) Peek() byte {
	if r.readPos >= len(r.input) {
		return 0
	}

	return r.input[r.readPos]
}

/* 
Ver qual o n-ésimo caractere após o atual.
*/
func (r *Reader) PeekN(n int) byte {
	pos := r.readPos + n - 1

	if pos >= len(r.input) {
		return 0
	}

	return r.input[pos]
}

func (r *Reader) EOF() bool {
	return r.ch == 0
}

func (r *Reader) Position() Position {
	return Position{
		Offset: r.pos,
		Line:   r.line,
		Column: r.column,
	}
}