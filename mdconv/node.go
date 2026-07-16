package mdconv

type Node struct {
	Type TokenType

	// metadados do bloco
	Level int // nível do heading
	Language string // linguagem do bloco de código
	Children []*Node // filhos do nó

	// metadados de texto
	Value string // conteúdo textual bruto
	URL string // endereço de destino para link e image
	Title string // texto descritivo ou alternativo para link e image
}