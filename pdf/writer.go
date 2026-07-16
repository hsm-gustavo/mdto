package pdf

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"strconv"
	"strings"
)

// Writer serializa as páginas coletadas pelo Canvas em um PDF simples.
type Writer struct {
	pageWidth  float64
	pageHeight float64
	pages      []pageData
}

type pageData struct {
	content     string
	annotations []PageAnnotation
	images      []PageImage
}

type pageAllocation struct {
	pageObject       int
	contentObject    int
	annotationObjects []int
	imageObjects     []int
	images           []PageImage
	annotations      []PageAnnotation
}

// NewWriter cria um writer com as dimensões informadas.
func NewWriter(pageWidth, pageHeight float64) *Writer {
	return &Writer{
		pageWidth:  pageWidth,
		pageHeight: pageHeight,
	}
}

// AddPage adiciona o stream de conteúdo de uma página e seus metadados.
func (w *Writer) AddPage(content string, annotations []PageAnnotation, images []PageImage) {
	pageAnnotations := make([]PageAnnotation, len(annotations))
	copy(pageAnnotations, annotations)
	pageImages := make([]PageImage, len(images))
	copy(pageImages, images)
	w.pages = append(w.pages, pageData{content: content, annotations: pageAnnotations, images: pageImages})
}

// String gera o PDF completo em formato textual/bytes compatíveis.
func (w *Writer) String() string {
	if len(w.pages) == 0 {
		w.pages = []pageData{{}}
	}

	allocations, totalObjects := w.allocateObjectNumbers()
	objects := make([]string, totalObjects)

	objects[0] = "<< /Type /Catalog /Pages 2 0 R >>"
	objects[1] = fmt.Sprintf("<< /Type /Pages /Kids [%s] /Count %d >>", pageRefsFromAllocations(allocations), len(w.pages))
	objects[2] = "<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>"
	objects[3] = "<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica-Bold >>"
	objects[4] = "<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica-Oblique >>"
	objects[5] = "<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica-BoldOblique >>"
	objects[6] = "<< /Type /Font /Subtype /Type1 /BaseFont /Courier >>"

	for i, allocation := range allocations {
		pageObject := fmt.Sprintf(
			"<< /Type /Page /Parent 2 0 R /MediaBox [0 0 %s %s] /Contents %d 0 R /Resources << /Font << /F1 3 0 R /F2 4 0 R /F3 5 0 R /F4 6 0 R /F5 7 0 R >>",
			formatFloat(w.pageWidth),
			formatFloat(w.pageHeight),
			allocation.contentObject,
		)

		if len(allocation.imageObjects) > 0 {
			pageObject += fmt.Sprintf(" /XObject << %s >>", pageXObjectResources(allocation.images, allocation.imageObjects))
		}
		if len(allocation.annotationObjects) > 0 {
			pageObject += fmt.Sprintf(" /Annots [%s]", joinObjectRefs(allocation.annotationObjects))
		}
		pageObject += " >>"
		objects[allocation.pageObject-1] = pageObject

		objects[allocation.contentObject-1] = fmt.Sprintf("<< /Length %d >>\nstream\n%s\nendstream", len(w.pages[i].content), w.pages[i].content)

		for j, imageObjectNumber := range allocation.imageObjects {
			objects[imageObjectNumber-1] = imageObjectString(allocation.images[j])
		}

		for j, annotationObjectNumber := range allocation.annotationObjects {
			annotation := allocation.annotations[j]
			objects[annotationObjectNumber-1] = fmt.Sprintf(
				"<< /Type /Annot /Subtype /Link /Rect [%s %s %s %s] /Border [0 0 0] /A << /S /URI /URI (%s) >> >>",
				formatFloat(annotation.X1),
				formatFloat(annotation.Y1),
				formatFloat(annotation.X2),
				formatFloat(annotation.Y2),
				escapePDFString(annotation.URL),
			)
		}
	}

	var buffer bytes.Buffer
	buffer.WriteString("%PDF-1.4\n")

	offsets := make([]int, len(objects))
	for i, object := range objects {
		offsets[i] = buffer.Len()
		buffer.WriteString(fmt.Sprintf("%d 0 obj\n%s\nendobj\n", i+1, object))
	}

	xrefStart := buffer.Len()
	buffer.WriteString(fmt.Sprintf("xref\n0 %d\n", len(objects)+1))
	buffer.WriteString("0000000000 65535 f \n")
	for _, offset := range offsets {
		buffer.WriteString(fmt.Sprintf("%010d 00000 n \n", offset))
	}

	buffer.WriteString(fmt.Sprintf("trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF", len(objects)+1, xrefStart))
	return buffer.String()
}

// Bytes retorna o PDF gerado como bytes.
func (w *Writer) Bytes() []byte {
	return []byte(w.String())
}

func (w *Writer) allocateObjectNumbers() ([]pageAllocation, int) {
	allocations := make([]pageAllocation, len(w.pages))
	nextObjectNumber := 8

	for i, page := range w.pages {
		allocations[i].pageObject = nextObjectNumber
		nextObjectNumber++
		allocations[i].contentObject = nextObjectNumber
		nextObjectNumber++
		allocations[i].images = page.images
		allocations[i].annotations = page.annotations

		allocations[i].imageObjects = make([]int, len(page.images))
		for j := range page.images {
			allocations[i].imageObjects[j] = nextObjectNumber
			nextObjectNumber++
		}

		allocations[i].annotationObjects = make([]int, len(page.annotations))
		for j := range page.annotations {
			allocations[i].annotationObjects[j] = nextObjectNumber
			nextObjectNumber++
		}
	}

	return allocations, nextObjectNumber - 1
}

func pageRefsFromAllocations(allocations []pageAllocation) string {
	refs := make([]string, 0, len(allocations))
	for _, allocation := range allocations {
		refs = append(refs, strconv.Itoa(allocation.pageObject)+" 0 R")
	}
	return strings.Join(refs, " ")
}

func pageXObjectResources(images []PageImage, objectNumbers []int) string {
	parts := make([]string, 0, len(images))
	for i, image := range images {
		parts = append(parts, fmt.Sprintf("/%s %d 0 R", image.Name, objectNumbers[i]))
	}
	return strings.Join(parts, " ")
}

func imageObjectString(pageImage PageImage) string {
	encoded := encodeJPEG(pageImage.Image)
	return fmt.Sprintf(
		"<< /Type /XObject /Subtype /Image /Width %d /Height %d /ColorSpace /DeviceRGB /BitsPerComponent 8 /Filter /DCTDecode /Length %d >>\nstream\n%s\nendstream",
		pageImage.Image.Bounds().Dx(),
		pageImage.Image.Bounds().Dy(),
		len(encoded),
		string(encoded),
	)
}

func encodeJPEG(img image.Image) []byte {
	if img == nil {
		return nil
	}

	bounds := img.Bounds()
	opaque := image.NewRGBA(bounds)
	draw.Draw(opaque, bounds, &image.Uniform{C: color.White}, image.Point{}, draw.Src)
	draw.Draw(opaque, bounds, img, bounds.Min, draw.Over)

	var buffer bytes.Buffer
	_ = jpeg.Encode(&buffer, opaque, &jpeg.Options{Quality: 90})
	return buffer.Bytes()
}

func joinObjectRefs(refs []int) string {
	parts := make([]string, 0, len(refs))
	for _, ref := range refs {
		parts = append(parts, strconv.Itoa(ref)+" 0 R")
	}
	return strings.Join(parts, " ")
}

func formatFloat(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}

func escapePDFString(text string) string {
	var builder strings.Builder
	for _, r := range text {
		switch r {
		case '\\', '(', ')':
			builder.WriteByte('\\')
			builder.WriteRune(r)
		case '\n', '\r':
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
