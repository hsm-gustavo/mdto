# mdto

[![Go Reference](https://pkg.go.dev/badge/github.com/hsm-gustavo/mdto)](https://pkg.go.dev/github.com/hsm-gustavo/mdto)

[![Go](https://github.com/hsm-gustavo/mdto/actions/workflows/go.yaml/badge.svg)](https://github.com/hsm-gustavo/mdto/actions/workflows/go.yaml)

A simple markdown converter that parses markdown into an AST and renders it into HTML. It supports basic markdown syntax, including headings, lists, links, images, code blocks, and more.

## Installation

```
go get github.com/hsm-gustavo/mdto
```

## Usage

```go
package main

import (
	"log"
	"os"

	"github.com/hsm-gustavo/mdto/renderer"
)

func main() {
	markdownContent := "# Hello, World!\n\nThis is a simple markdown file."

	r := renderer.NewHTMLRenderer()

	html := r.RenderMarkdown(markdownContent)

	os.WriteFile("./examples/EXAMPLE.html", []byte(html), 0644)
	log.Println("HTML gerado com sucesso em ./examples/EXAMPLE.html")
}
```

or using the high-level function:

```go
package main

import (
    "log"
    "os"

    "github.com/hsm-gustavo/mdto"
)

func main() {
    markdownContent := "# Hello, World!\n\nThis is a simple markdown file."

    html := mdto.HTML(markdownContent)

    os.WriteFile("./examples/EXAMPLE.html", []byte(html), 0644)
    log.Println("HTML gerado com sucesso em ./examples/EXAMPLE.html")
}
```

## TODO

- Add support for tables and other advanced markdown features.
