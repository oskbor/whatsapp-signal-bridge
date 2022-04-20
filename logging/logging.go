package logging

import (
	"io"
	"os"

	"github.com/rs/zerolog"
)

type LevelWriter struct {
	io.Writer
	ErrorWriter io.Writer
}

func (lw *LevelWriter) WriteLevel(l zerolog.Level, p []byte) (n int, err error) {
	var w io.Writer
	w = lw
	if l > zerolog.InfoLevel {
		w = lw.ErrorWriter
	}
	return w.Write(p)
}

func DefaultLogger(component string) *zerolog.Logger {
	logger := zerolog.New(LevelWriter{
		Writer:      os.Stdout,
		ErrorWriter: os.Stderr,
	}).With().Str("component", component).Logger()
	return &logger
}
