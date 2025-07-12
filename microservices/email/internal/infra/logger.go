package infra

import (
	"io"
	"log"
	"os"
)

func NewLogger(path string) *log.Logger {
	var writers []io.Writer

	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		log.Printf("logger: cannot open %s: %v (fallback to stdout only)", path, err)
	} else {
		writers = append(writers, file)
	}

	writers = append(writers, os.Stdout)

	multi := io.MultiWriter(writers...)
	return log.New(multi, "", log.LstdFlags)
}
