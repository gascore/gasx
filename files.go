package gasx

import (
	"os"
	"path/filepath"
	"strings"
)

// File information about file
type File struct {
	Path      string
	Extension string
}

// GasFiles find files for builder in current directory
func GasFiles(extensions []string) ([]File, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return []File{}, err
	}

	return GasFilesCustomDir(currentDir+"/app/", extensions)
}

// GasFilesCustomDir find files for builder in directory
func GasFilesCustomDir(root string, extensions []string) ([]File, error) {
	var files []File
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		for _, ext := range extensions {
			if strings.HasSuffix(path, "."+ext) {
				files = append(files, File{Path: path, Extension: ext})
				return nil
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return files, err
}
