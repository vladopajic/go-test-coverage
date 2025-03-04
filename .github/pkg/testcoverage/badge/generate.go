package badge

import (
	"strconv"

	"github.com/narqo/go-badge"
)

const (
	ContentType = "image/svg+xml"

	label = "coverage"
)

func Generate(coverage int) ([]byte, error) {
	return badge.RenderBytes( //nolint:wrapcheck // error should never happen
		label,
		strconv.Itoa(coverage)+"%",
		badge.Color(Color(coverage)),
	)
}

func Color(coverage int) string {
	//nolint:mnd // relax
	switch {
	case coverage >= 100:
		return "#44cc11" // strong green
	case coverage >= 90:
		return "#97ca00" // light green
	case coverage >= 80:
		return "#dfb317" // yellow
	case coverage >= 70:
		return "#fa7739" // orange
	case coverage >= 50:
		return "#e05d44" // light red
	default:
		return "#cb2431" // strong red
	}
}
