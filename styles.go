package gasx

import (
	"os"
	"fmt"
	"strings"
	"io/ioutil"
	"go/build"
	"github.com/visualfc/fastmod"
)

// GrepStyles grep styles files path in deps by their styles.gas
func GrepStyles() ([]string, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return []string{}, err
	}

	return GrepStylesCustom(currentDir+"/app", make(map[string]bool))
}

// GrepStylesCustom grep styles files path in deps (for custom dir) by their styles.gas
func GrepStylesCustom(dir string, already map[string]bool) ([]string, error) {
	var stylesOut []string

	stylesGasFile, err := ioutil.ReadFile(dir+"/styles.gas")
	if err != nil && !os.IsNotExist(err) {
		return []string{}, fmt.Errorf("error opening \"styles.gas\": \"%s\"", err.Error())
	}

	patterns := strings.Split(string(stylesGasFile), "\n")
	if len(patterns) > 0 && patterns[0] == "" || patterns[0] == " " {
		patterns = []string{}
	}
	
	for _, pattern := range patterns {
		paths, err := FilesByPattern(dir, pattern)
		if err != nil {
			return []string{}, err
		}

		stylesOut = append(stylesOut, paths...)
	}

	pkg, err := fastmod.LoadPackage(dir, &build.Default)
	if err != nil {
		return []string{}, fmt.Errorf("invalid package: \"%s\"", err.Error())
	}
	
	for _, nodeValue := range pkg.NodeMap {
		pkgDir := nodeValue.ModDir()
		if already[pkgDir] {
			continue
		}
		already[pkgDir] = true

		newStyles, err := GrepStylesCustom(pkgDir, already)
		if err != nil {
			return []string{}, err
		}

		stylesOut = append(stylesOut, newStyles...)
	}

	return stylesOut, nil
}

