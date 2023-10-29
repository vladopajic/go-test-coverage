package badge_test

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/vladopajic/go-test-coverage/v2/pkg/testcoverage/badge"
)

func Test_Generate(t *testing.T) {
	t.Parallel()

	for i := 0; i <= 100; i++ {
		i := i
		c := strconv.Itoa(i) + "%"
		t.Run(c, func(t *testing.T) {
			t.Parallel()

			svg, err := Generate(i)
			assert.NoError(t, err)

			svgStr := string(svg)
			assert.NotEmpty(t, svgStr)
			assert.Contains(t, svgStr, ">"+c+"<")
			assert.Contains(t, svgStr, Color(i))
		})
	}
}

func Test_Color(t *testing.T) {
	t.Parallel()

	colors := make(map[string]struct{})

	{ // Assert that there are 5 colors for coverage [0-101]
		for i := 0; i <= 101; i++ {
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
