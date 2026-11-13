package logger

import (
	"errors"
	"io"
	"strings"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
)

func RotateWriter(infoFile, errFile string) (infoWriter io.Writer, errWriter io.Writer, err error) {
	if infoWriter, err = rotateWriter(infoFile); err != nil {
		return
	}

	if errWriter, err = rotateWriter(errFile); err != nil {
		return
	}

	return
}

// ResolveWriters decides where the info and error logs go: two explicit writers
// are used as they are, a single one is a configuration error, and nil writers
// fall back to rotating files created from infoFile and errFile.
// Callers own the writers they pass in: the framework never closes them, and
// typed nil writers are not detected.
func ResolveWriters(infoWriter, errorWriter io.Writer, infoFile, errFile string) (io.Writer, io.Writer, error) {
	if infoWriter == nil && errorWriter == nil {
		fileInfoWriter, fileErrorWriter, err := RotateWriter(infoFile, errFile)
		if err != nil {
			return nil, nil, err
		}

		return fileInfoWriter, fileErrorWriter, nil
	}

	if infoWriter == nil || errorWriter == nil {
		return nil, nil, errors.New("logger: infoWriter and errorWriter must be provided together")
	}

	return infoWriter, errorWriter, nil
}

func rotateWriter(filename string) (io.Writer, error) {
	return rotatelogs.New(
		strings.Replace(filename, ".log", "", -1)+"-%Y%m%d%H.log", // 2022年1月1日12点 => filename-2022010112.log
		rotatelogs.WithLinkName(filename),
		rotatelogs.WithMaxAge(time.Hour*24*7),  // log max save seven days
		rotatelogs.WithRotationTime(time.Hour), // rotate once an hour
	)
}
