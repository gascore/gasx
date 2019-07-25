package gasx

import (
	"io/ioutil"
	"os"
	"time"

	"github.com/liamylian/jsontime"
)

var jsonWithTime = jsontime.ConfigWithCustomTimeFormat

type LockFile struct {
	fileName      string
	BuildExternal bool

	Body map[string]string `json:"body"`
	Date time.Time         `json:"last" time_format:"2006-01-02 15:04:05"`
}

func GetLockFile(fileName string, ignoreExternal bool) (*LockFile, error) {
	file, err := ioutil.ReadFile(fileName)
	if err != nil {
		if os.IsNotExist(err) {
			return &LockFile{Body: make(map[string]string), BuildExternal: true, fileName: fileName}, nil
		}

		return nil, err
	}

	var lock LockFile
	err = jsonWithTime.Unmarshal(file, &lock)
	if err != nil {
		return nil, err
	}

	lock.fileName = fileName
	lock.BuildExternal = !ignoreExternal || lock.Date.After(time.Now().Add(24*time.Hour))

	return &lock, nil
}

func (l *LockFile) Save() error {
	l.Date = time.Now()

	lockJSON, err := jsonWithTime.Marshal(&l)
	if err != nil {
		return err
	}

	if exists(l.fileName) {
		err := os.Remove(l.fileName)
		if err != nil {
			return err
		}
	}

	lockFile, err := os.Create(l.fileName)
	if err != nil {
		return err
	}

	_, err = lockFile.Write(lockJSON)
	if err != nil {
		return err
	}

	err = lockFile.Close()
	if err != nil {
		return err
	}

	return nil
}

func exists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}
