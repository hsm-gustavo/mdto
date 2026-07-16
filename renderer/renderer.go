package renderer

import "github.com/hsm-gustavo/md-to-html-go/mdconv"

// Renderer define o contrato básico para qualquer conversor baseado na AST do markdown.
type Renderer interface {
	Render(nodes []*mdconv.Node) string
}