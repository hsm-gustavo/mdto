package main

import (
	"bufio"
	"log"
	"os"

	"github.com/hsm-gustavo/md-to-html-go/renderer"
)

func ReadMarkdownFile(path string) (string, error) {
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
	markdownContent, err := ReadMarkdownFile("./examples/EXAMPLE.md")
	if err != nil {
		log.Fatal(err)
	}

	r := renderer.NewHTMLRenderer()

	html := r.RenderMarkdown(markdownContent)

	os.WriteFile("./examples/EXAMPLE.html", []byte(html), 0644)
	log.Println("HTML gerado com sucesso em ./examples/EXAMPLE.html")
}