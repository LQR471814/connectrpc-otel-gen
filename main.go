package main

import (
	"flag"
	"fmt"
	"go/parser"
	"go/token"
	"io"
	"log"
	"os"
	"path/filepath"
)

func processFile(filename string, input io.Reader) string {
	src, err := io.ReadAll(input)
	if err != nil {
		log.Fatal(err)
	}

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filename, string(src), parser.SkipObjectResolution)
	if err != nil {
		log.Fatal(err)
	}

	targets := parseTargets(file)
	if targets == nil {
		log.Fatal("could not find connectrpc client interface")
	}

	return generate(file, targets)
}

func processFilesRecursively(dir string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		log.Printf("failed to read directory '%s'\nerr: %v\n", dir, err)
		return
	}

	for _, e := range entries {
		path := filepath.Join(dir, e.Name())
		if e.IsDir() {
			processFilesRecursively(path)
			continue
		}
		if e.Name() != "api.connect.go" {
			continue
		}

		f, err := os.Open(path)
		if err != nil {
			log.Printf("failed to generate instrumentation for source '%s'\nerr: %v\n", dir, err)
			continue
		}
		generated := processFile(path, f)
		f.Close()

		err = os.WriteFile(filepath.Join(dir, "api.telemetry.go"), []byte(generated), 0600)
		if err != nil {
			log.Printf("failed to write generated code for source '%s'\nerr: %v\n", dir, err)
		}
		return
	}
}

func main() {
	flag.Parse()
	directories := flag.Args()

	if len(directories) == 0 {
		generated := processFile("STDIN", os.Stdin)
		fmt.Print(generated)
		return
	}

	for _, dir := range directories {
		processFilesRecursively(dir)
	}
}
