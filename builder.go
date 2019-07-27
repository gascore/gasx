package gasx

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"net/http"

	"github.com/fatih/color"
	copyPkg "github.com/otiai10/copy"
)

type Builder struct {
	LockFile *LockFile
	BlockCompilers []BlockCompiler
}

type BlockInfo struct {
	Name      string
	Value     string
	FileInfo  File
	FileBytes string
	LockFile  *LockFile
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

func Log(msg string) {
	fmt.Println(color.BlueString("Builder:") + " " + msg)
}

func Must(err error) {
	if err != nil {
		ErrorMsg(err.Error())
	}
}

func ErrorMsg(msg string) {
	fmt.Println(color.RedString("ERROR") + ": " + msg)
	panic(msg)
}

func RunCommand(command string) {
	cmd := exec.Command("sh", "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	Must(cmd.Run())
}

func NewFile(path, body string) {
	Must(ioutil.WriteFile(path, []byte(body), 0644))
}

func CopyFile(pathA, pathB string) {
	file, err := ioutil.ReadFile(pathA)
	Must(err)

	err = os.Remove(pathB)
	if err != nil && !os.IsNotExist(err) {
		Must(err)
	}

	NewFile(pathB, string(file))
}

func ClearDir(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		Must(os.Mkdir(dir, os.ModePerm))
	} else {
		Must(os.RemoveAll(dir))
	}
}

func CopyDir(dirA, dirB string) {
	Must(copyPkg.Copy(dirA, dirB))
}

func ServeDir(port, dir string) *http.Server {
	srv := &http.Server{Addr: port, Handler: http.FileServer(http.Dir(dir))}
	
	go func() {
		Must(srv.ListenAndServe())
	}()

	return srv
}
