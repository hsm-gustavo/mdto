package main

import (
	"bufio"
	"log"
	"os"

	"github.com/hsm-gustavo/mdto/renderer"
)

func readMarkdownFile(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)
	var content string
	for scanner.Scan() {
		content += scanner.Text() + "\n"
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return content, nil
}

func main() {
	markdownContent, err := readMarkdownFile("./examples/EXAMPLE.md")
	if err != nil {
		log.Fatal(err)
	}

	r := renderer.NewHTMLRenderer()

  pdf := renderer.NewPDFRenderer()

	html := r.RenderMarkdown(markdownContent)
  pdfContent := pdf.RenderMarkdown(markdownContent)
	os.WriteFile("./examples/EXAMPLE.html", []byte(html), 0644)
	os.WriteFile("./examples/EXAMPLE.pdf", []byte(pdfContent), 0644)
	log.Println("HTML gerado com sucesso em ./examples/EXAMPLE.html")
	log.Println("PDF gerado com sucesso em ./examples/EXAMPLE.pdf")
}
