package gasx

import (
	"os"
	"fmt"
	"regexp"
	"strings"
	"io/ioutil"
)

var tRgxp = regexp.MustCompile(`\$([a-zA-Z0-9]*){((.|\s)*?)[^\\]}\$`)

func (builder *Builder) ParseFiles(files []File) error {
	for _, fileInfo := range files {
		filename := strings.TrimSuffix(fileInfo.Path, "."+fileInfo.Extension) + "_gox.go"

		// TODO: Add hook here
		
		fileBytes, err := ioutil.ReadFile(fileInfo.Path)
		if err != nil {
			return fmt.Errorf("error while opening %s: \n%s", fileInfo.Path, err.Error())
		}
		fileBody := string(fileBytes)

		matches := tRgxp.FindAllStringSubmatchIndex(fileBody, -1)
		for _, match := range matches {
			var (
				blockStart = match[0]
				blockEnd = match[1]

				nameStart = match[2]
				nameEnd = match[3]
				name = fileBody[nameStart:nameEnd]

				valueStart = match[4]
				valueEnd = match[5]
				value = strings.TrimSpace(fileBody[valueStart:valueEnd])
			)

			newVal, err := builder.RenderBlock(&BlockInfo{
				Name: string(name),
				Value: string(value),
				FileInfo: fileInfo,
				FileBytes: fileBody,
			})
			if err != nil {
				return fmt.Errorf("error while rendering block in %s (name: %s, valS: %d, valE: %d): \n%s", fileInfo.Path, name, valueStart, valueEnd, err.Error())
			}

			fileBody = fileBody[:blockStart] + newVal + fileBody[:blockEnd]
		}

		osFile, err := os.Create(filename)
		if err != nil {
			if err == os.ErrPermission {
				fmt.Println("Run: \"chmod 777 -R $GOPATH/pkg/mod\"")
			}

			return err
		}

		_, err = osFile.Write([]byte(fileBytes))
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
