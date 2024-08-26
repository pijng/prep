package main

import (
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const PrepDir = "prep_dumps"

func Dump(m map[string]string, filename string) {
	foundDir, err := resolveDir(os.TempDir(), PrepDir)
	if err != nil {
		panic(fmt.Sprintf("failed finding dir: %s", err))
	}

	file, err := os.Create(fmt.Sprintf("%s/%s", foundDir, filename))
	if err != nil {
		panic(fmt.Sprintf("failed creating file: %s", err))
	}
	defer file.Close()

	encoder := gob.NewEncoder(file)
	err = encoder.Encode(m)
	if err != nil {
		panic(fmt.Sprintf("failed encoding map: %s", err))
	}
}

func Restore(filename string) map[string]string {
	foundDir, err := resolveDir(os.TempDir(), PrepDir)
	if err != nil {
		panic(fmt.Sprintf("failed finding dir: %s", err))
	}

	path := fmt.Sprintf("%s/%s", foundDir, filename)

	var file *os.File
	file, err = os.Open(path)
	if os.IsNotExist(err) {
		file, err = os.Create(path)
		if err != nil {
			panic(fmt.Sprintf("failed opening file: %s", err))
		}
	}
	defer file.Close()

	decoder := gob.NewDecoder(file)
	var m map[string]string

	err = decoder.Decode(&m)
	if err != nil && !errors.Is(err, io.EOF) {
		panic(fmt.Sprintf("failed decoding map: %s", err))
	}

	return m
}

func Merge[K comparable, V any](maps ...map[K]V) map[K]V {
	merged := make(map[K]V)

	for _, m := range maps {
		for k, v := range m {
			merged[k] = v
		}
	}

	return merged
}

func resolveDir(root string, pattern string) (string, error) {
	var foundDir string

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() && strings.Contains(info.Name(), pattern) {
			foundDir = path
			return nil
		}

		return nil
	})

	if foundDir == "" {
		foundDir, err = os.MkdirTemp(root, pattern)
	}

	return foundDir, err
}
