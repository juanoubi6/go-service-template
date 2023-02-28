package utils

import (
	"io"
	"os"
	"path/filepath"
)

func GetJSONFileFromDir(path, filename string) []byte {
	ext := filepath.Ext(filename)
	if ext == "" {
		filename += ".json"
	}
	fPath := filepath.Join(path, filename)
	f, err := os.Open(fPath)
	if err != nil {
		panic(err)
	}

	content, err := io.ReadAll(f)
	if err != nil {
		panic(err)
	}

	return content
}

func GetJSONFileContent(file string) []byte {
	return GetJSONFileFromDir("./testFiles/", file)
}
