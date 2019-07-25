package gasx

import (
	"go/build"
	"os"
	"path/filepath"
	"strings"

	"github.com/visualfc/fastmod"
)

type File struct {
	Path       string
	Extension  string
	IsExternal bool
}

// GasFiles find files for builder in current directory
func GasFiles(extensions []string, buildExternal bool) ([]File, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return []File{}, err
	}

	return GasFilesCustomDir(currentDir+"/app/", extensions, buildExternal)
}

// GasFilesCustomDir find files for builder in directory
func GasFilesCustomDir(directory string, extensions []string, buildExternal bool) ([]File, error) {
	if buildExternal {
		return parseModDir(directory, extensions, []string{}, false)
	}

	files, err := getGasFilesBody(directory, extensions, false)
	if err != nil {
		return files, nil
	}

	return files, nil
}

func parseModDir(root string, extensions, already []string, isExternal bool) ([]File, error) {
	for _, alreadyDir := range already {
		if alreadyDir == root {
			return []File{}, nil
		}
	}

	files, err := getGasFilesBody(root, extensions, isExternal)
	if err != nil {
		return files, nil
	}

	already = append(already, root)

	pkg, err := fastmod.LoadPackage(root, &build.Default)
	if err != nil {
		return files, err
	}

	for _, nodeValue := range pkg.NodeMap {
		newFiles, err := parseModDir(nodeValue.ModDir(), extensions, already, true)
		if err != nil {
			return files, err
		}
		files = append(files, newFiles...)
	}

	return files, nil
}

func getGasFilesBody(root string, extensions []string, isExternal bool) ([]File, error) {
	var files []File

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		for _, ext := range extensions {
			if strings.HasSuffix(path, "."+ext) {
				files = append(files, File{Path: path, Extension: ext, IsExternal: isExternal})
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
