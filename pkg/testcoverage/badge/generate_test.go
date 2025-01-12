package badge_test

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/subhambhardwaj/go-test-coverage/v2/pkg/testcoverage/badge"
)

func Test_Generate(t *testing.T) {
	t.Parallel()

	t.Run("generate for all values [0-100]", func(t *testing.T) {
		t.Parallel()

		for i := range 101 {
			svg, err := Generate(i)
			assert.NoError(t, err)

			svgStr := string(svg)
			assert.Contains(t, svgStr, ">"+strconv.Itoa(i)+"%<")
			assert.Contains(t, svgStr, Color(i))
		}
	})

	t.Run("exact match", func(t *testing.T) {
		t.Parallel()

		//nolint:lll // relax
		const expected = `<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" width="109" height="20"><linearGradient id="smooth" x2="0" y2="100%"><stop offset="0" stop-color="#bbb" stop-opacity=".1"/><stop offset="1" stop-opacity=".1"/></linearGradient><mask id="round"><rect width="109" height="20" rx="3" fill="#fff"/></mask><g mask="url(#round)"><rect width="65" height="20" fill="#555"/><rect x="65" width="44" height="20" fill="#44cc11"/><rect width="109" height="20" fill="url(#smooth)"/></g><g fill="#fff" text-anchor="middle" font-family="DejaVu Sans,Verdana,Geneva,sans-serif" font-size="11"><text x="33.5" y="15" fill="#010101" fill-opacity=".3">coverage</text><text x="33.5" y="14">coverage</text><text x="86" y="15" fill="#010101" fill-opacity=".3">100%</text><text x="86" y="14">100%</text></g></svg>`

		svg, err := Generate(100)
		assert.NoError(t, err)
		assert.Equal(t, expected, string(svg))
	})
}

func Test_Color(t *testing.T) {
	t.Parallel()

	colors := make(map[string]struct{})

	{ // Assert that there are 5 colors for coverage [0-101]
		for i := range 101 {
			color := Color(i)
			colors[color] = struct{}{}
		}

		assert.Len(t, colors, 6)
	}

	{ // Assert valid color values
		isHexColor := func(color string) bool {
			return string(color[0]) == "#" && len(color) == 7
		}

		for color := range colors {
			assert.True(t, isHexColor(color))
		}
	}
}
