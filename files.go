package gasx

import (
	"os"
	"strings"
	"go/build"
	"path/filepath"
	
	"github.com/visualfc/fastmod"
)

type File struct {
	Path       string
	Extension  string
	IsExternal bool
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
func GasFilesCustomDir(directory string, extensions []string) ([]File, error) {
	return parseModDir(directory, extensions, []string{})
}

func parseModDir(root string, extensions, already []string) ([]File, error) {
	for _, alreadyDir := range already {
		if alreadyDir == root {
			return []File{}, nil
		}
	}

	files, err := getGasFilesBody(root, extensions)
	if err != nil {
		return files, nil
	}

	already = append(already, root)

	pkg, err := fastmod.LoadPackage(root, &build.Default)
	if err != nil {
		return files, err
	}

	for _, nodeValue := range pkg.NodeMap {
		newFiles, err := parseModDir(nodeValue.ModDir(), extensions, already)
		if err != nil {
			return files, err
		}
		files = append(files, newFiles...)
	}

	return files, nil
}

func getGasFilesBody(root string, extensions []string) ([]File, error) {
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
