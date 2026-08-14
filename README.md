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

## Parsing Markdown

Use `mdto.Parse` when you need access to the parsed AST before rendering it:

```go
nodes := mdto.Parse("# Hello, World!")
```

`Parse` returns the document's top-level `[]*mdto.Node`. A separate `Document`
wrapper is intentionally not used yet because the AST has no document-level metadata.

## Command line

Convert a Markdown file to HTML on standard output:

```sh
mdconv README.md > README.html
```

When no input file is provided, `mdconv` reads from standard input:

```sh
cat README.md | mdconv
```

Use `-o` to write directly to a file. HTML is selected by default; use `ast` to
print the parsed AST as formatted JSON:

```sh
mdconv README.md -o README.html
mdconv --to html README.md
mdconv --to ast README.md
```

## Plain text

Use `mdto.Text` to render Markdown as readable plain text. It preserves image alt
text and code block contents while removing Markdown formatting:

```go
text := mdto.Text("# Hello\n\nThis is **bold**.")
```

## TODO

- Add support for tables and other advanced markdown features.
