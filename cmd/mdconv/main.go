package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/hsm-gustavo/mdto"
	"github.com/hsm-gustavo/mdto/renderer"
)

type options struct {
	inputPath  string
	outputPath string
	format     string
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	options, err := parseOptions(args)
	if err != nil {
		fmt.Fprintf(stderr, "mdconv: %v\n", err)
		printUsage(stderr)
		return 2
	}
	if err := validateFormat(options.format); err != nil {
		fmt.Fprintf(stderr, "mdconv: %v\n", err)
		return 2
	}

	input := stdin
	if options.inputPath != "" {
		file, err := os.Open(options.inputPath)
		if err != nil {
			fmt.Fprintf(stderr, "mdconv: read %s: %v\n", options.inputPath, err)
			return 1
		}
		defer file.Close()
		input = file
	}

	content, err := io.ReadAll(input)
	if err != nil {
		fmt.Fprintf(stderr, "mdconv: read input: %v\n", err)
		return 1
	}

	output, err := render(string(content), options.format)
	if err != nil {
		fmt.Fprintf(stderr, "mdconv: %v\n", err)
		return 2
	}

	if options.outputPath == "" {
		if _, err := fmt.Fprint(stdout, output); err != nil {
			fmt.Fprintf(stderr, "mdconv: write stdout: %v\n", err)
			return 1
		}
		return 0
	}

	file, err := os.Create(options.outputPath)
	if err != nil {
		fmt.Fprintf(stderr, "mdconv: create %s: %v\n", options.outputPath, err)
		return 1
	}
	if _, err := fmt.Fprint(file, output); err != nil {
		file.Close()
		fmt.Fprintf(stderr, "mdconv: write %s: %v\n", options.outputPath, err)
		return 1
	}
	if err := file.Close(); err != nil {
		fmt.Fprintf(stderr, "mdconv: close %s: %v\n", options.outputPath, err)
		return 1
	}

	return 0
}

func parseOptions(args []string) (options, error) {
	options := options{format: "html"}

	for index := 0; index < len(args); index++ {
		argument := args[index]
		switch {
		case argument == "-o":
			index++
			if index == len(args) {
				return options, fmt.Errorf("-o requires a file path")
			}
			options.outputPath = args[index]
		case strings.HasPrefix(argument, "-o="):
			options.outputPath = strings.TrimPrefix(argument, "-o=")
		case argument == "--to":
			index++
			if index == len(args) {
				return options, fmt.Errorf("--to requires a format")
			}
			options.format = args[index]
		case strings.HasPrefix(argument, "--to="):
			options.format = strings.TrimPrefix(argument, "--to=")
		case strings.HasPrefix(argument, "-"):
			return options, fmt.Errorf("unknown option %q", argument)
		case options.inputPath != "":
			return options, fmt.Errorf("only one input file is supported")
		default:
			options.inputPath = argument
		}
	}

	return options, nil
}

func render(input, format string) (string, error) {
	switch format {
	case "html":
		return mdto.HTML(input), nil
	case "ast":
		return renderer.NewJSONRenderer().Render(mdto.Parse(input))
	default:
		return "", fmt.Errorf("unsupported output format %q", format)
	}
}

func validateFormat(format string) error {
	if format != "html" && format != "ast" {
		return fmt.Errorf("unsupported output format %q", format)
	}
	return nil
}

func printUsage(writer io.Writer) {
	fmt.Fprintln(writer, "usage: mdconv [--to html|ast] [-o output.html] [input.md]")
}
