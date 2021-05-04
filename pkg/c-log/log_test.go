package c_log

import (
	"log"
	"testing"
)

func TestCLogInit(t *testing.T) {
	cf := CLogInit(&CLogOptions{
		Flag: log.Lshortfile | log.Ltime,
		Path: "test.log",
	})
	defer cf()
}
