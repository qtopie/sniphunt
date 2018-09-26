package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

var (
	inputFile   string
	archiveFile string
	searchDir   string
)

func loadSnippet() {
	data, err := ioutil.ReadFile(inputFile)
	if err != nil {
		panic(err)
	}

	snippet = string(data)
}

func visit(path string, f os.FileInfo, err error) error {
	fmt.Printf("Visited: %s\n", path)
	return nil
}

func main() {
	flag.StringVar(&inputFile, "input", "", "input code snippet file")
	flag.StringVar(&archiveFile, "archive", "", "archive")
	flag.Parse()

	loadSnippet()

	if true {
		walkArchive(archiveFile)
	} else {
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
