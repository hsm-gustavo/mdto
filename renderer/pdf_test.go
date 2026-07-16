package renderer

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRenderMarkdownToPDF(t *testing.T) {
	r := NewPDFRenderer()
	imagePath := writeTestImage(t)
	pdf := r.RenderMarkdown("# Titulo\n\nTexto com **negrito** e _italico_ e [link](https://example.com) e ![Imagem](" + imagePath + ")\n\n```go\nfmt.Println(\"hi\")\n```")

	if !strings.HasPrefix(pdf, "%PDF-1.4") {
		t.Fatalf("pdf deveria começar com o cabeçalho PDF, recebeu: %q", pdf[:min(16, len(pdf))])
	}

	if !strings.Contains(pdf, "Titulo") {
		t.Fatalf("pdf deveria conter o título renderizado")
	}

	if !strings.Contains(pdf, "Texto") || !strings.Contains(pdf, "com") {
		t.Fatalf("pdf deveria conter o corpo do parágrafo")
	}

	if !strings.Contains(pdf, "negrito") || !strings.Contains(pdf, "italico") || !strings.Contains(pdf, "link") {
		t.Fatalf("pdf deveria conter o texto inline renderizado")
	}

	if !strings.Contains(pdf, "/Subtype /Image") {
		t.Fatalf("pdf deveria embutir a imagem como XObject")
	}

	if !strings.Contains(pdf, "/F2") || !strings.Contains(pdf, "/F3") || !strings.Contains(pdf, "/F5") {
		t.Fatalf("pdf deveria usar fontes diferentes para ênfase, itálico e código")
	}

	if !strings.Contains(pdf, "0 0 1 rg") {
		t.Fatalf("pdf deveria colorir links")
	}

	if !strings.Contains(pdf, "/Subtype /Link") || !strings.Contains(pdf, "/A << /S /URI") {
		t.Fatalf("pdf deveria criar anotacoes clicaveis para links")
	}

	if !strings.Contains(pdf, "/Type /Catalog") || !strings.Contains(pdf, "/Type /Pages") {
		t.Fatalf("pdf deveria conter a estrutura básica de objetos")
	}
}

func TestRenderMarkdownToPDFFlowsAcrossPages(t *testing.T) {
	r := NewPDFRenderer()
	paragraph := "Um paragrafo longo para forcar quebra de pagina. "
	input := strings.Repeat(paragraph, 120)
	pdf := r.RenderMarkdown(input)

	if strings.Count(pdf, "/MediaBox") < 2 {
		t.Fatalf("pdf deveria gerar mais de uma pagina para conteudo longo")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func writeTestImage(t *testing.T) string {
	t.Helper()

	img := image.NewRGBA(image.Rect(0, 0, 24, 24))
	for y := 0; y < 24; y++ {
		for x := 0; x < 24; x++ {
			img.Set(x, y, color.RGBA{R: 30, G: 180, B: 90, A: 255})
		}
	}

	path := filepath.Join(t.TempDir(), "test-image.png")
	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("nao foi possivel criar imagem de teste: %v", err)
	}
	defer file.Close()

	if err := png.Encode(file, img); err != nil {
		t.Fatalf("nao foi possivel codificar imagem de teste: %v", err)
	}

	return path
}
