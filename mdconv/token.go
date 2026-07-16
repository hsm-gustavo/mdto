package mdconv

type TokenType string

const (
	// Blocos de texto

	// um parágrafo de texto comum
	PARAGRAPH TokenType = "PARAGRAPH"

	// títulos ou cabeçalhos (#)
	HEADING TokenType = "HEADING"

	// blocos de código (``` ``` ou 4 espaços)
	CODE_BLOCK TokenType = "CODE_BLOCK"

	// blocos de citação (>)
	BLOCKQUOTE TokenType = "BLOCKQUOTE"

	// listas não ordenadas (*, - ou +)
	ULIST TokenType = "ULIST"

	// listas ordenadas (1. , 2. )
	OLIST TokenType = "OLIST"

	// item de lista interno usado pela AST
	LIST_ITEM TokenType = "LIST_ITEM"

	// linha horizontal (---, *** ou ___)
	HR TokenType = "HR"

	// Elementos de texto

	// texto puro sem nenhuma formatação
	TEXT TokenType = "TEXT"

	// ênfase ou itálico (_ _ ou * *)
	ITALIC TokenType = "ITALIC"

	// ênfase forte ou negrito (** ** ou __ __)
	BOLD TokenType = "BOLD"

	// corte (~~ ~~)
	STRIKETHROUGH TokenType = "STRIKETHROUGH"

	// código embutido no texto (``)
	CODE_INLINE TokenType = "CODE_INLINE"

	// link no formato [texto descritivo](url)
	LINK TokenType = "LINK"

	// imagem no formato ![texto alternativo](url)
	IMAGE TokenType = "IMAGE"

	// Fim de arquivo
	EOF TokenType = "EOF"
)

type Position struct {
	Offset int // posição de leitura
	Line int // número da linha
	Column int // número da coluna
}
type Token struct {
	Type     TokenType // tipo do token
	Literal  string    // valor literal do token
	Position Position  // posição do token no texto de entrada
	Level    int       // nível de heading
	Language string    // linguagem do bloco de código
	URL      string    // endereço de destino para link e image
	Title    string    // texto descritivo ou alternativo para link e image
}