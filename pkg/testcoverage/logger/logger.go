package logger

import (
	"bytes"
	"sync"

	"github.com/rs/zerolog"
)

//nolint:gochecknoglobals // relax
var (
	buffer bytes.Buffer
	lock   sync.Mutex
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
	lock.Lock()
	defer lock.Unlock()

	return buffer.Bytes()
}
