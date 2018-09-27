package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var (
	inputFile   string
	archiveFile string
	searchDir   string
)

func loadSnippet() {
	data, err := ioutil.ReadFile(inputFile)
	if err != nil {
		log.Fatal(err)
	}

	snippet = string(data)
}

func visit(path string, f os.FileInfo, err error) error {
	if f.IsDir() || !strings.HasSuffix(path, ".java") {
		return nil
	}

	fmt.Printf("Visited: %s\n", path)

	s, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	similarity := caclDistance(snippet, string(s))
	similarSnippets[path] = similarity

	return nil
}

func main() {
	flag.StringVar(&inputFile, "input", "", "input code snippet file")
	flag.StringVar(&archiveFile, "archive", "", "archive")
	flag.StringVar(&searchDir, "dir", "", "search dir")
	flag.Parse()

	loadSnippet()

	if archiveFile != "" {
		walkArchive(archiveFile)
	}

	if searchDir != "" {
		log.Println("walking")
		err := filepath.Walk(searchDir, visit)
		if err != nil {
			log.Fatal(err)
		}
	}

	var closestPath string
	var closestRate float64 = -1

	for k, v := range similarSnippets {
		if closestRate == -1 || v < closestRate {
			closestPath = k
			closestRate = v
		}
	}

	fmt.Println("Clothest", closestPath, closestRate)
}
