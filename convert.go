package mdto

import "github.com/hsm-gustavo/mdto/renderer"

/*
Converte o conteúdo Markdown em HTML usando a pipeline completa: tokenização, parsing e renderização.
*/
func HTML(markdownContent string) string {
	r := renderer.NewHTMLRenderer()
	html := r.RenderMarkdown(markdownContent)
	return html
}

func PDF(markdownContent string) string {
  r := renderer.NewPDFRenderer()
  pdf := r.RenderMarkdown(markdownContent)
  return pdf
}
