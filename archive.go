package main

import (
	"archive/zip"
	"io/ioutil"
	"path/filepath"
)

func walkArchive(archiveFile string) error {
	reader, err := zip.OpenReader(archiveFile)
	if err != nil {
		return err
	}

	for _, file := range reader.File {

		extension := filepath.Ext(file.Name)
		if extension != ".java" {
			continue
		}

		fileReader, err := file.Open()
		if err != nil {
			return err
		}
		defer fileReader.Close()

		s, err := ioutil.ReadAll(fileReader)
		if err != nil {
			return err
		}

		similarity := caclDistance(snippet, string(s))
		similarSnippets[file.Name] = similarity
	}

	return nil
}
