package renderer

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	_ "github.com/gen2brain/avif"
	"github.com/hsm-gustavo/mdto/mdconv"
	pdfdoc "github.com/hsm-gustavo/mdto/pdf"
)

// PDFRenderer converte a AST do Markdown em um PDF simples com fluxo de texto.
type PDFRenderer struct{}

// NewPDFRenderer cria uma instância reutilizável de renderer PDF.
func NewPDFRenderer() *PDFRenderer {
	return &PDFRenderer{}
}

// RenderMarkdown executa a pipeline completa: tokeniza, faz parse e renderiza em PDF.
func (r *PDFRenderer) RenderMarkdown(input string) string {
	lexer := mdconv.NewLexer(input)
	parser := mdconv.NewParser(lexer.Tokenize())
	return r.Render(parser.Parse())
}

// Render converte uma lista de nós em um PDF textual válido.
func (r *PDFRenderer) Render(nodes []*mdconv.Node) string {
	canvas := pdfdoc.NewCanvas()
	canvas.StartPage()

	for _, node := range nodes {
		r.renderNode(canvas, node, 0, defaultInlineStyle())
	}

	writer := pdfdoc.NewWriter(canvas.PageWidth, canvas.PageHeight)
	pages := canvas.Pages()
	annotations := canvas.PageAnnotations()
	images := canvas.PageImages()
	for i, page := range pages {
		var pageAnnotations []pdfdoc.PageAnnotation
		if i < len(annotations) {
			pageAnnotations = annotations[i]
		}
		var pageImages []pdfdoc.PageImage
		if i < len(images) {
			pageImages = images[i]
		}
		writer.AddPage(page, pageAnnotations, pageImages)
	}

	return writer.String()
}

func (r *PDFRenderer) renderNode(canvas *pdfdoc.Canvas, node *mdconv.Node, indent float64, baseStyle pdfdoc.TextStyle) {
	if node == nil {
		return
	}

	switch node.Type {
	case mdconv.HEADING:
		rendered := r.renderInlineRuns(node.Children, node.Value, pdfdoc.TextStyle{Font: pdfdoc.FontBold, Color: baseStyle.Color}, false, false, false, false)
		size := headingFontSize(node.Level)
		canvas.WriteRuns(rendered, indent, size)
		canvas.WriteSpacer(size / 3)
	case mdconv.PARAGRAPH:
		r.renderFlowChildren(canvas, node.Children, node.Value, indent, baseStyle)
		canvas.WriteSpacer(canvas.LineHeight / 2)
	case mdconv.CODE_BLOCK:
		canvas.WriteSpacer(canvas.LineHeight / 4)
		canvas.WritePreformattedText(node.Value, indent+12, 11)
		canvas.WriteSpacer(canvas.LineHeight / 2)
	case mdconv.BLOCKQUOTE:
		blockquoteStyle := pdfdoc.TextStyle{Font: pdfdoc.FontItalic, Color: baseStyle.Color}
		r.renderFlowChildren(canvas, node.Children, node.Value, indent+12, blockquoteStyle)
		canvas.WriteSpacer(canvas.LineHeight / 2)
	case mdconv.ULIST:
		r.renderList(canvas, node, indent, false, baseStyle)
		canvas.WriteSpacer(canvas.LineHeight / 4)
	case mdconv.OLIST:
		r.renderList(canvas, node, indent, true, baseStyle)
		canvas.WriteSpacer(canvas.LineHeight / 4)
	case mdconv.HR:
		canvas.WriteRule(indent)
		canvas.WriteSpacer(canvas.LineHeight / 2)
	case mdconv.LIST_ITEM:
		r.renderListItem(canvas, node, indent, 1, false, baseStyle)
	case mdconv.IMAGE:
		r.renderImageNode(canvas, node, indent, baseStyle)
	default:
		rendered := r.renderInlineRuns(node.Children, node.Value, baseStyle, false, false, false, false)
		if len(rendered) > 0 {
			canvas.WriteRuns(rendered, indent, canvas.FontSize)
		}
	}
}

func (r *PDFRenderer) renderFlowChildren(canvas *pdfdoc.Canvas, children []*mdconv.Node, fallback string, indent float64, baseStyle pdfdoc.TextStyle) {
	if len(children) == 0 {
		rendered := r.renderInlineRuns(nil, fallback, baseStyle, false, false, false, false)
		if len(rendered) > 0 {
			canvas.WriteRuns(rendered, indent, canvas.FontSize)
		}
		return
	}

	textRuns := make([]pdfdoc.TextRun, 0)
	flushText := func() {
		if len(textRuns) == 0 {
			return
		}
		canvas.WriteRuns(textRuns, indent, canvas.FontSize)
		textRuns = nil
	}

	for _, child := range children {
		if child == nil {
			continue
		}

		switch child.Type {
		case mdconv.IMAGE:
			flushText()
			r.renderImageNode(canvas, child, indent, baseStyle)
			canvas.WriteSpacer(canvas.LineHeight / 3)
		case mdconv.ULIST:
			flushText()
			r.renderList(canvas, child, indent, false, baseStyle)
		case mdconv.OLIST:
			flushText()
			r.renderList(canvas, child, indent, true, baseStyle)
		default:
			textRuns = append(textRuns, r.renderInlineRuns([]*mdconv.Node{child}, child.Value, baseStyle, false, false, false, false)...)
		}
	}

	flushText()
}

func (r *PDFRenderer) renderList(canvas *pdfdoc.Canvas, node *mdconv.Node, indent float64, ordered bool, baseStyle pdfdoc.TextStyle) {
	for i, item := range node.Children {
		r.renderListItem(canvas, item, indent, i+1, ordered, baseStyle)
	}
}

func (r *PDFRenderer) renderListItem(canvas *pdfdoc.Canvas, node *mdconv.Node, indent float64, index int, ordered bool, baseStyle pdfdoc.TextStyle) {
	if node == nil {
		return
	}

	prefix := "- "
	if ordered {
		prefix = strconv.Itoa(index) + ". "
	}

	contentRuns := make([]pdfdoc.TextRun, 0)
	for _, child := range node.Children {
		switch child.Type {
		case mdconv.ULIST:
			r.renderList(canvas, child, indent+24, false, baseStyle)
		case mdconv.OLIST:
			r.renderList(canvas, child, indent+24, true, baseStyle)
		default:
			contentRuns = append(contentRuns, r.renderInlineRuns([]*mdconv.Node{child}, child.Value, baseStyle, false, false, false, false)...)
		}
	}

	lineRuns := append([]pdfdoc.TextRun{{Text: prefix, Style: baseStyle}}, contentRuns...)
	prefixWidth := float64(len([]rune(prefix))) * canvas.FontSize * 0.55
	canvas.WriteHangingRuns(lineRuns, indent, indent+prefixWidth+6, canvas.FontSize)
}

func (r *PDFRenderer) renderInlineRuns(children []*mdconv.Node, fallback string, baseStyle pdfdoc.TextStyle, bold bool, italic bool, code bool, strike bool) []pdfdoc.TextRun {
	if len(children) == 0 {
		if fallback == "" {
			return nil
		}
		style := resolveInlineStyle(baseStyle, bold, italic, code, strike)
		return []pdfdoc.TextRun{{Text: fallback, Style: style}}
	}

	runs := make([]pdfdoc.TextRun, 0)
	for _, child := range children {
		runs = append(runs, r.inlineNodeRuns(child, baseStyle, bold, italic, code, strike)...)
	}

	return runs
}

func (r *PDFRenderer) inlineNodeRuns(node *mdconv.Node, baseStyle pdfdoc.TextStyle, bold bool, italic bool, code bool, strike bool) []pdfdoc.TextRun {
	if node == nil {
		return nil
	}

	switch node.Type {
	case mdconv.TEXT:
		style := resolveInlineStyle(baseStyle, bold, italic, code, strike)
		return []pdfdoc.TextRun{{Text: node.Value, Style: style}}
	case mdconv.CODE_INLINE:
		style := resolveInlineStyle(baseStyle, false, false, true, false)
		return []pdfdoc.TextRun{{Text: node.Value, Style: style}}
	case mdconv.BOLD:
		return r.renderInlineRuns(node.Children, node.Value, baseStyle, true, italic, code, strike)
	case mdconv.ITALIC:
		return r.renderInlineRuns(node.Children, node.Value, baseStyle, bold, true, code, strike)
	case mdconv.STRIKETHROUGH:
		return r.renderInlineRuns(node.Children, node.Value, baseStyle, bold, italic, code, true)
	case mdconv.LINK:
		linkStyle := resolveInlineStyle(baseStyle, bold, italic, code, strike)
		linkStyle.Color = "0 0 1"
		linkStyle.LinkURL = node.URL
		if len(node.Children) > 0 {
			return r.renderInlineRuns(node.Children, node.Title, linkStyle, bold, italic, code, strike)
		}
		text := node.Title
		if text == "" {
			text = node.URL
		}
		if text == "" {
			return nil
		}
		return []pdfdoc.TextRun{{Text: text, Style: linkStyle}}
	case mdconv.IMAGE:
		return nil
	case mdconv.ULIST, mdconv.OLIST:
		return nil
	default:
		return r.renderInlineRuns(node.Children, node.Value, baseStyle, bold, italic, code, strike)
	}
}

func (r *PDFRenderer) renderImageNode(canvas *pdfdoc.Canvas, node *mdconv.Node, indent float64, baseStyle pdfdoc.TextStyle) {
	if node == nil || node.URL == "" {
		return
	}

	altText := node.Title
	if altText == "" {
		altText = node.Value
	}
	if altText == "" {
		altText = node.URL
	}
	if altText == "" {
		altText = "imagem"
	}

	img, err := r.loadImage(node.URL)
	if err != nil {
		placeholder := resolveInlineStyle(baseStyle, false, false, false, false)
		placeholder.Color = "0.35 0.35 0.35"
		canvas.WriteRuns([]pdfdoc.TextRun{{Text: "[" + altText + "]", Style: placeholder}}, indent, canvas.FontSize)
		return
	}
	if img == nil {
		placeholder := resolveInlineStyle(baseStyle, false, false, false, false)
		placeholder.Color = "0.35 0.35 0.35"
		canvas.WriteRuns([]pdfdoc.TextRun{{Text: "[" + altText + "]", Style: placeholder}}, indent, canvas.FontSize)
		return
	}

	maxWidth := canvas.PageWidth - (2 * canvas.Margin) - indent
	canvas.DrawImage(img, indent, maxWidth)
}

func (r *PDFRenderer) loadImage(source string) (image.Image, error) {
	if source == "" {
		return nil, nil
	}

	if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
		client := &http.Client{Timeout: 15 * time.Second}
		response, err := client.Get(source)
		if err != nil {
			return nil, err
		}
		defer response.Body.Close()
		if response.StatusCode < 200 || response.StatusCode >= 300 {
			return nil, fmt.Errorf("unexpected status %s for %s", response.Status, source)
		}
		img, _, err := image.Decode(response.Body)
		return img, err
	}

	pathCandidates := candidatePaths(source)
	var lastErr error
	for _, path := range pathCandidates {
		file, err := os.Open(path)
		if err != nil {
			lastErr = err
			continue
		}
		img, _, decodeErr := image.Decode(file)
		_ = file.Close()
		if decodeErr != nil {
			lastErr = decodeErr
			continue
		}
		return img, nil
	}

	if lastErr == nil {
		lastErr = fmt.Errorf("unable to resolve image path: %s", source)
	}
	return nil, lastErr
}

func candidatePaths(source string) []string {
	paths := make([]string, 0, 4)
	cleanSource := strings.TrimPrefix(source, "file://")
	paths = append(paths, cleanSource)

	if cwd, err := os.Getwd(); err == nil {
		paths = append(paths, filepath.Join(cwd, cleanSource))
		paths = append(paths, filepath.Join(cwd, "examples", cleanSource))
		paths = append(paths, filepath.Join(filepath.Dir(cwd), cleanSource))
	}

	if strings.HasPrefix(cleanSource, "./") {
		trimmed := strings.TrimPrefix(cleanSource, "./")
		paths = append(paths, trimmed)
		if cwd, err := os.Getwd(); err == nil {
			paths = append(paths, filepath.Join(cwd, trimmed))
			paths = append(paths, filepath.Join(cwd, "examples", trimmed))
		}
	}

	seen := make(map[string]struct{}, len(paths))
	unique := make([]string, 0, len(paths))
	for _, path := range paths {
		if path == "" {
			continue
		}
		if _, ok := seen[path]; ok {
			continue
		}
		seen[path] = struct{}{}
		unique = append(unique, path)
	}
	return unique
}

func resolveInlineStyle(base pdfdoc.TextStyle, bold bool, italic bool, code bool, strike bool) pdfdoc.TextStyle {
	style := base
	if code {
		style.Font = pdfdoc.FontMonospace
	} else {
		style.Font = composeFont(style.Font, bold, italic)
	}
	if style.Font == "" {
		style.Font = pdfdoc.FontRegular
	}
	if style.Color == "" {
		style.Color = "0 0 0"
	}
	style.Strike = strike
	return style
}

func composeFont(baseFont string, bold bool, italic bool) string {
	switch baseFont {
	case pdfdoc.FontBold:
		bold = true
	case pdfdoc.FontItalic:
		italic = true
	case pdfdoc.FontBoldItalic:
		bold = true
		italic = true
	case pdfdoc.FontMonospace:
		return pdfdoc.FontMonospace
	}

	switch {
	case bold && italic:
		return pdfdoc.FontBoldItalic
	case bold:
		return pdfdoc.FontBold
	case italic:
		return pdfdoc.FontItalic
	default:
		return pdfdoc.FontRegular
	}
}

func defaultInlineStyle() pdfdoc.TextStyle {
	return pdfdoc.TextStyle{Font: pdfdoc.FontRegular, Color: "0 0 0"}
}

func headingFontSize(level int) float64 {
	switch {
	case level <= 1:
		return 22
	case level == 2:
		return 20
	case level == 3:
		return 18
	case level == 4:
		return 16
	case level == 5:
		return 14
	default:
		return 13
	}
}
