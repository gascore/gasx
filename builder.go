package gasx

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/fatih/color"
	copyPkg "github.com/otiai10/copy"
)

type Builder struct {
	BlockCompilers []BlockCompiler
}

type BlockInfo struct {
	Name      string
	Value     string
	FileInfo  File
	FileBytes string
}

type BlockCompiler func(*BlockInfo) (string, error)

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

func (builder *Builder) Log(msg string) {
	fmt.Println(color.BlueString("Builder:") + " " + msg)
}

func (builder *Builder) Must(err error) {
	if err != nil {
		builder.ErrorMsg(err.Error())
	}
}

func (builder *Builder) ErrorMsg(msg string) {
	fmt.Println(color.RedString("ERROR") + ": " + msg)
	panic(msg)
}

func (builder *Builder) RunCommand(command string) {
	cmd := exec.Command("sh", "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	builder.Must(cmd.Run())
}

func (builder *Builder) NewFile(path, body string) {
	builder.Must(ioutil.WriteFile(path, []byte(body), 0644))
}

func (builder *Builder) CopyFile(pathA, pathB string) {
	file, err := ioutil.ReadFile(pathA)
	builder.Must(err)

	err = os.Remove(pathB)
	if err != nil && !os.IsNotExist(err) {
		builder.Must(err)
	}

	builder.NewFile(pathB, string(file))
}

func (builder *Builder) ClearDir(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		builder.Must(os.Mkdir(dir, os.ModePerm))
	} else {
		builder.Must(os.RemoveAll(dir))
	}
}

func (builder *Builder) CopyDir(dirA, dirB string) {
	builder.Must(copyPkg.Copy(dirA, dirB))
}
