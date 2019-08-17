package gasx

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
)

var tRgxp = regexp.MustCompile(`\$([a-zA-Z0-9]*){((.|\s)*?)[^\\]}\$`)

// ParseFiles parse and compile GOS files
func (builder *Builder) ParseFiles(files []File) error {
	for _, fileInfo := range files {
		filename := strings.TrimSuffix(fileInfo.Path, "."+fileInfo.Extension) + "_gas.go"

		// TODO: Add hook here

		fileBytes, err := ioutil.ReadFile(fileInfo.Path)
		if err != nil {
			return fmt.Errorf("error while opening %s: \n%s", fileInfo.Path, err.Error())
		}
		fileBody := string(fileBytes)

		parsedFileBody, err := builder.CompileFile(fileInfo, fileBody)
		if err != nil {
			return err
		}

		osFile, err := os.Create(filename)
		if err != nil {
			if err == os.ErrPermission {
				fmt.Println("Run: \"chmod 777 -R $GOPATH/pkg/mod\"")
			}

			return err
		}

		_, err = osFile.Write([]byte(parsedFileBody))
		if err != nil {
			return err
		}

		err = osFile.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

// ParseFile compile GOS file to pure golang
func (builder *Builder) CompileFile(fileInfo File, fileBody string) (string, error) {
	var lenDiff int
	matches := tRgxp.FindAllStringSubmatchIndex(fileBody, -1)
	for _, match := range matches {
		n := func(i int) int {
			return i + lenDiff
		}

		var (
			blockStart = n(match[0])
			blockEnd   = n(match[1])

			nameStart = n(match[2])
			nameEnd   = n(match[3])
			name      = fileBody[nameStart:nameEnd]

			valueStart = n(match[4])
			valueEnd   = n(match[5])
			value      = fileBody[valueStart:valueEnd]
		)

		newVal, err := builder.RenderBlock(&BlockInfo{
			Name:      string(name),
			Value:     strings.TrimSpace(value),
			FileInfo:  fileInfo,
			FileBytes: fileBody,
		})
		if err != nil {
			return "", fmt.Errorf("error while rendering block in %s (name: %s, valS: %d, valE: %d): \n%s", fileInfo.Path, name, valueStart, valueEnd, err.Error())
		}

		lenDiff += len(newVal) - len(fileBody[blockStart:blockEnd])

		fileBody = fileBody[:blockStart] + newVal + fileBody[blockEnd:]
	}

	return fileBody, nil
}
