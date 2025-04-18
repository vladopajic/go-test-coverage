package logger

import (
	"bytes"

	"github.com/rs/zerolog"
)

//nolint:gochecknoglobals // relax
var (
	buffer bytes.Buffer
	L      zerolog.Logger
)

func Init() { // coverage-ignore
	L = zerolog.New(&buffer).With().Logger()
}

func Destruct() {
	L = zerolog.Logger{}
	buffer = bytes.Buffer{}
}

func Bytes() []byte {
	return buffer.Bytes()
}
