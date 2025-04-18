package logger

import (
	"bytes"

	"github.com/rs/zerolog"
)

//nolint:gochecknoglobals // relax
var (
	Buffer bytes.Buffer
	L      zerolog.Logger
)

//nolint:gochecknoinits // relax
func init() {
	L = zerolog.New(&Buffer).With().Logger()
}
