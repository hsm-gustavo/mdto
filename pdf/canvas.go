package pdf

import (
	"fmt"
	"image"
	"strings"
	"unicode"
)

const (
	DefaultPageWidth  = 595.28
	DefaultPageHeight = 841.89
	DefaultMargin     = 72.0
	DefaultFontSize   = 12.0
	DefaultLineHeight = 16.0

	FontRegular     = "F1"
	FontBold        = "F2"
	FontItalic      = "F3"
	FontBoldItalic  = "F4"
	FontMonospace   = "F5"
)

// TextStyle descreve a aparência de um trecho inline.
type TextStyle struct {
	Font      string
	Color     string
	Underline bool
	Strike    bool
	LinkURL   string
}

// TextRun representa um fragmento de texto com estilo.
type TextRun struct {
	Text  string
	Style TextStyle
}

// Canvas guarda o estado de layout de uma renderização simples em PDF.
type Canvas struct {
	PageWidth  float64
	PageHeight float64
	Margin     float64
	FontSize   float64
	LineHeight float64

	pages      []string
	pageAnnots [][]PageAnnotation
	pageImages [][]PageImage
	current    strings.Builder
	cursorY    float64
	started    bool
	hasContent bool
}

type PageAnnotation struct {
	X1  float64
	Y1  float64
	X2  float64
	Y2  float64
	URL string
}

type PageImage struct {
	Name         string
	X            float64
	Y            float64
	Width        float64
	Height       float64
	Image        image.Image
}

// NewCanvas cria um canvas com dimensões A4 e margens padrão.
func NewCanvas() *Canvas {
	return &Canvas{
		PageWidth:  DefaultPageWidth,
		PageHeight: DefaultPageHeight,
		Margin:     DefaultMargin,
		FontSize:   DefaultFontSize,
		LineHeight: DefaultLineHeight,
	}
}

// StartPage abre uma nova página e reseta o cursor vertical.
func (c *Canvas) StartPage() {
	if c.started {
		c.flushPage()
	}

	c.current = strings.Builder{}
	c.cursorY = c.PageHeight - c.Margin
	c.started = true
	c.hasContent = false
	if len(c.pageAnnots) == len(c.pages) {
		c.pageAnnots = append(c.pageAnnots, nil)
	}
	if len(c.pageImages) == len(c.pages) {
		c.pageImages = append(c.pageImages, nil)
	}
}

// Pages devolve o conteúdo já fechado de todas as páginas.
func (c *Canvas) Pages() []string {
	c.flushPage()
	pages := make([]string, len(c.pages))
	copy(pages, c.pages)
	return pages
}

// WriteRuns escreve texto com estilo, respeitando quebra de linha e paginação.
func (c *Canvas) WriteRuns(runs []TextRun, indent float64, fontSize float64) {
	c.WriteHangingRuns(runs, indent, indent, fontSize)
}

// WriteHangingRuns escreve texto com indentação inicial e indentação de continuação.
func (c *Canvas) WriteHangingRuns(runs []TextRun, firstIndent float64, restIndent float64, fontSize float64) {
	tokens := tokenizeRuns(runs)
	if len(tokens) == 0 {
		c.WriteSpacer(c.LineHeight)
		return
	}

	lines := wrapTokens(tokens, firstIndent, restIndent, fontSize, c.PageWidth, c.Margin)
	for _, line := range lines {
		c.writeStyledLine(line.Tokens, fontSize, line.FirstIndent, line.RestIndent)
	}
}

// WritePreformattedText escreve texto preservando quebras de linha.
func (c *Canvas) WritePreformattedText(text string, indent float64, fontSize float64) {
	if text == "" {
		c.WriteSpacer(c.LineHeight)
		return
	}

	for _, line := range strings.Split(text, "\n") {
		c.writeStyledLine([]textToken{{Text: line, Style: TextStyle{Font: FontMonospace, Color: "0 0 0"}}}, fontSize, indent, indent)
	}
}

// WriteRule desenha uma linha horizontal simples.
func (c *Canvas) WriteRule(indent float64) {
	c.ensurePage()
	c.ensureSpace(c.LineHeight)

	x1 := c.Margin + indent
	x2 := c.PageWidth - c.Margin
	y := c.cursorY
	if x2 < x1 {
		x2 = x1 + 10
	}

	c.current.WriteString(fmt.Sprintf("q 0.5 w %.2f %.2f m %.2f %.2f l S Q\n", x1, y, x2, y))
	c.hasContent = true
	c.cursorY -= c.LineHeight
}

// DrawImage desenha uma imagem rasterizada e registra seus metadados para o writer.
func (c *Canvas) DrawImage(img image.Image, indent float64, maxWidth float64) {
	if img == nil {
		return
	}

	bounds := img.Bounds()
	width := float64(bounds.Dx())
	height := float64(bounds.Dy())
	if width <= 0 || height <= 0 {
		return
	}

	availableWidth := c.PageWidth - (2 * c.Margin) - indent
	if maxWidth > 0 && maxWidth < availableWidth {
		availableWidth = maxWidth
	}
	if availableWidth < 40 {
		availableWidth = 40
	}

	displayWidth := width
	if displayWidth > availableWidth {
		displayWidth = availableWidth
	}
	displayHeight := (displayWidth / width) * height

	required := displayHeight + c.LineHeight
	c.ensurePage()
	c.ensureSpace(required)

	x := c.Margin + indent
	y := c.cursorY - displayHeight
	name := c.nextImageName()
	c.current.WriteString(fmt.Sprintf("q %.2f 0 0 %.2f %.2f %.2f cm /%s Do Q\n", displayWidth, displayHeight, x, y, name))
	c.hasContent = true
	c.addPageImage(PageImage{
		Name:   name,
		X:      x,
		Y:      y,
		Width:  displayWidth,
		Height: displayHeight,
		Image:  img,
	})
	c.cursorY = y - (c.LineHeight / 2)
	if c.cursorY < c.Margin {
		c.StartPage()
	}
}

func (c *Canvas) nextImageName() string {
	if len(c.pageImages) == 0 {
		return "Im1"
	}
	return fmt.Sprintf("Im%d", len(c.pageImages[len(c.pageImages)-1])+1)
}

// WriteSpacer avança o cursor vertical sem desenhar nada.
func (c *Canvas) WriteSpacer(points float64) {
	c.ensurePage()
	c.cursorY -= points
	if c.cursorY < c.Margin {
		c.StartPage()
	}
}

func (c *Canvas) writeStyledLine(line []textToken, fontSize float64, firstIndent float64, restIndent float64) {
	if len(line) == 0 {
		c.WriteSpacer(c.LineHeight)
		return
	}

	required := maxFloat(c.LineHeight, fontSize*1.4) + 2
	c.ensurePage()
	c.ensureSpace(required)

	x := c.Margin + firstIndent
	y := c.cursorY
	c.current.WriteString("BT\n")
	c.current.WriteString(fmt.Sprintf("%.2f %.2f Td\n", x, y))

	currentX := x
	for _, run := range line {
		style := normalizeTextStyle(run.Style)
		if run.Text == "" {
			continue
		}

		if isWhitespaceOnly(run.Text) && currentX == x {
			continue
		}

		c.current.WriteString(fmt.Sprintf("/%s %.2f Tf\n", style.Font, fontSize))
		c.current.WriteString(fmt.Sprintf("%s rg\n", style.Color))
		c.current.WriteString(fmt.Sprintf("(%s) Tj\n", escapePDFText(run.Text)))

		width := measureText(run.Text, fontSize, style.Font)
		if style.LinkURL != "" {
			c.addLinkAnnotation(currentX, y, currentX+width, y+fontSize*1.1, style.LinkURL)
		}
		if style.Strike {
			lineY := y - (fontSize * 0.18)
			if style.Strike {
				lineY = y + (fontSize * 0.3)
			}
			c.current.WriteString(fmt.Sprintf("q 0.8 w %.2f %.2f m %.2f %.2f l S Q\n", currentX, lineY, currentX+width, lineY))
		}

		currentX += width
	}

	c.current.WriteString("ET\n")
	c.hasContent = true
	c.cursorY -= maxFloat(c.LineHeight, fontSize*1.4)
	if c.cursorY < c.Margin {
		c.StartPage()
	}
}

func (c *Canvas) ensurePage() {
	if !c.started {
		c.StartPage()
	}
}

func (c *Canvas) ensureSpace(required float64) {
	if c.cursorY-required < c.Margin {
		c.StartPage()
	}
}

func (c *Canvas) flushPage() {
	if !c.started {
		return
	}

	if c.hasContent {
		c.pages = append(c.pages, c.current.String())
		if len(c.pageAnnots) == len(c.pages)-1 {
			c.pageAnnots = append(c.pageAnnots, nil)
		}
		if len(c.pageImages) == len(c.pages)-1 {
			c.pageImages = append(c.pageImages, nil)
		}
	}

	c.started = false
	c.hasContent = false
}

func (c *Canvas) addLinkAnnotation(x1, y1, x2, y2 float64, url string) {
	if url == "" {
		return
	}
	if len(c.pageAnnots) == 0 {
		c.pageAnnots = append(c.pageAnnots, nil)
	}
	index := len(c.pageAnnots) - 1
	c.pageAnnots[index] = append(c.pageAnnots[index], PageAnnotation{
		X1:  x1,
		Y1:  y1 - 1,
		X2:  x2,
		Y2:  y2,
		URL: url,
	})
}

func (c *Canvas) addPageImage(image PageImage) {
	if len(c.pageImages) == 0 {
		c.pageImages = append(c.pageImages, nil)
	}
	index := len(c.pageImages) - 1
	c.pageImages[index] = append(c.pageImages[index], image)
}

// PageAnnotations devolve as anotações coletadas por página.
func (c *Canvas) PageAnnotations() [][]PageAnnotation {
	c.flushPage()
	annotations := make([][]PageAnnotation, len(c.pageAnnots))
	for i, pageAnnotations := range c.pageAnnots {
		copied := make([]PageAnnotation, len(pageAnnotations))
		copy(copied, pageAnnotations)
		annotations[i] = copied
	}
	return annotations
}

// PageImages devolve as imagens coletadas por página.
func (c *Canvas) PageImages() [][]PageImage {
	c.flushPage()
	images := make([][]PageImage, len(c.pageImages))
	for i, pageImages := range c.pageImages {
		copied := make([]PageImage, len(pageImages))
		copy(copied, pageImages)
		images[i] = copied
	}
	return images
}

type textToken struct {
	Text  string
	Style TextStyle
}

type textLine struct {
	Tokens    []textToken
	FirstIndent float64
	RestIndent  float64
}

func tokenizeRuns(runs []TextRun) []textToken {
	tokens := make([]textToken, 0)
	for _, run := range runs {
		style := normalizeTextStyle(run.Style)
		for _, fragment := range splitInlineText(run.Text) {
			if fragment == "" {
				continue
			}
			tokens = append(tokens, textToken{Text: fragment, Style: style})
		}
	}
	return tokens
}

func wrapTokens(tokens []textToken, firstIndent float64, restIndent float64, fontSize float64, pageWidth float64, margin float64) []textLine {
	lines := make([]textLine, 0)
	current := textLine{FirstIndent: firstIndent, RestIndent: restIndent}
	currentWidth := 0.0
	limit := availableWidth(pageWidth, margin, firstIndent)

	for _, token := range tokens {
		if isWhitespaceOnly(token.Text) {
			if len(current.Tokens) == 0 {
				continue
			}
			current.Tokens = append(current.Tokens, token)
			currentWidth += measureText(token.Text, fontSize, token.Style.Font)
			continue
		}

		width := measureText(token.Text, fontSize, token.Style.Font)
		if currentWidth > 0 && currentWidth+width > limit {
			current.Tokens = trimTrailingWhitespace(current.Tokens)
			if len(current.Tokens) > 0 {
				lines = append(lines, current)
			}
			current = textLine{FirstIndent: restIndent, RestIndent: restIndent}
			currentWidth = 0
			limit = availableWidth(pageWidth, margin, restIndent)
		}

		current.Tokens = append(current.Tokens, token)
		currentWidth += width
	}

	current.Tokens = trimTrailingWhitespace(current.Tokens)
	if len(current.Tokens) > 0 {
		lines = append(lines, current)
	}

	if len(lines) == 0 {
		lines = append(lines, textLine{FirstIndent: firstIndent, RestIndent: restIndent})
	}

	return lines
}

func trimTrailingWhitespace(tokens []textToken) []textToken {
	end := len(tokens)
	for end > 0 && isWhitespaceOnly(tokens[end-1].Text) {
		end--
	}
	trimmed := make([]textToken, end)
	copy(trimmed, tokens[:end])
	return trimmed
}

func availableWidth(pageWidth float64, margin float64, indent float64) float64 {
	width := pageWidth - (2 * margin) - indent
	if width < 40 {
		width = 40
	}
	return width
}

func splitInlineText(text string) []string {
	if text == "" {
		return nil
	}

	parts := make([]string, 0)
	var current strings.Builder
	var currentSpace bool
	for _, r := range text {
		space := unicode.IsSpace(r)
		if current.Len() == 0 {
			currentSpace = space
			current.WriteRune(r)
			continue
		}

		if space == currentSpace {
			current.WriteRune(r)
			continue
		}

		parts = append(parts, current.String())
		current.Reset()
		currentSpace = space
		current.WriteRune(r)
	}

	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts
}

func normalizeTextStyle(style TextStyle) TextStyle {
	if style.Font == "" {
		style.Font = FontRegular
	}
	if style.Color == "" {
		style.Color = "0 0 0"
	}
	return style
}

func isWhitespaceOnly(text string) bool {
	for _, r := range text {
		if !unicode.IsSpace(r) {
			return false
		}
	}
	return text != ""
}

func measureText(text string, fontSize float64, font string) float64 {
	runeCount := float64(len([]rune(text)))
	factor := 0.55
	switch font {
	case FontBold:
		factor = 0.57
	case FontItalic:
		factor = 0.54
	case FontBoldItalic:
		factor = 0.58
	case FontMonospace:
		factor = 0.6
	}
	if isWhitespaceOnly(text) {
		factor = 0.3
	}
	return runeCount * fontSize * factor
}

func escapePDFText(text string) string {
	var builder strings.Builder
	for _, r := range text {
		switch r {
		case '\\', '(', ')':
			builder.WriteByte('\\')
			builder.WriteRune(r)
		case '\n', '\r', '\t':
			builder.WriteByte(' ')
		default:
			if r <= 255 {
				builder.WriteByte(byte(r))
			} else {
				builder.WriteByte('?')
			}
		}
	}
	return builder.String()
}

func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
