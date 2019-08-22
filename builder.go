package gasx

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"path/filepath"

	"github.com/fatih/color"
	copyPkg "github.com/otiai10/copy"
)

// Builder GOS files builder
type Builder struct {
	// BlockCompilers pipeline of special blocks compilers
	BlockCompilers []BlockCompiler
}

// BlockInfo information about special block
type BlockInfo struct {
	// Name File name
	Name string

	// Value special block value
	Value string

	// FileInfo isExternal, file path, extension, e.t.c.
	FileInfo File

	// FileBytes full GOS file value
	FileBytes string
}

// BlockCompiler node for render pipeline
type BlockCompiler func(*BlockInfo) (string, error)

// RenderBlock compile special block by render pipeline
func (builder *Builder) RenderBlock(block *BlockInfo) (string, error) {
	for i, renderer := range builder.BlockCompilers {
		newVal, err := renderer(block)
		if err != nil {
			return "", fmt.Errorf("error in block renderer %d: \n%s", i, err.Error())
		}

		block.Value = newVal
	}

	return block.Value, nil
}

// Log println message with "Builder" prefix
func Log(msg string) {
	fmt.Println(color.BlueString("Builder:") + " " + msg)
}

// Must compact "if err != nil"
func Must(err error) {
	if err != nil {
		ErrorMsg(err.Error())
	}
}

// ErrorMsg print message with ERROR tag
func ErrorMsg(msg string) {
	fmt.Println(color.RedString("ERROR") + ": " + msg)
	panic(msg)
}

// RunCommand execute command
func RunCommand(command string) {
	cmd := exec.Command("sh", "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	Must(cmd.Run())
}

// NewFile create new file
func NewFile(path, body string) {
	Must(ioutil.WriteFile(path, []byte(body), 0644))
}

// DeleteFile delete file
func DeleteFile(path string) {
	Must(os.Remove(path))
}

// CopyFile copy file from pathA to file in pathB
func CopyFile(pathA, pathB string) {
	file, err := ioutil.ReadFile(pathA)
	Must(err)

	err = os.Remove(pathB)
	if err != nil && !os.IsNotExist(err) {
		Must(err)
	}

	NewFile(pathB, string(file))
}

// ClearDir reacreate directory or create if it doesn't exists
func ClearDir(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		Must(os.Mkdir(dir, os.ModePerm))
	} else {
		Must(os.RemoveAll(dir))
	}
}

// CopyDir copy dirA to dirB
func CopyDir(dirA, dirB string) {
	Must(copyPkg.Copy(dirA, dirB))
}

// Exists return true if file exisits
func Exists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// InArrayString return is string in array of strings
func InArrayString(a string, arr []string) bool {
	for _, el := range arr {
		if el == a {
			return true
		}
	}

	return false
}

// FilesByPattern return files path matched by mattern in directory
func FilesByPattern(dir string, pattern string) ([]string, error) {
	var matched []string
	pattern = dir+"/"+pattern
	err := filepath.Walk(dir+"/", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("cannot access path %q: %s\n", path, err.Error())
		}

		if info.IsDir() {
			return nil
		}

		pathIsMatched, err := filepath.Match(pattern, path)
		if err != nil {
			return fmt.Errorf("error matching path: %s", err.Error())
		}

		if pathIsMatched {
			matched = append(matched, path)
		}
		
		return nil
	})
	if err != nil {
		return []string{}, fmt.Errorf("error walking the path %q: %s\n", dir, err.Error())
	}

	return matched, nil
}

func UniteFilesByPaths(paths []string) (string, error) {
	outFile := strings.Builder{}
	for _, path := range paths {
		fileFromPath, err := ioutil.ReadFile(path)
		if err != nil {
			return "", fmt.Errorf("cannot open file: \"%s\": %s", path, err.Error())
		}

		outFile.Write(fileFromPath)
	}
	return outFile.String(), nil
}
