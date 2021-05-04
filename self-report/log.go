package self_report

import (
	"log"
	"os"
)

type CloseFunc func() error

func LogInit(logPath string) (CloseFunc, error) {
	log.SetFlags(log.Lshortfile)

	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND, 0755)
	if err != nil {
		return nil, err
	}
	log.SetOutput(file)

	return file.Close, nil
}
