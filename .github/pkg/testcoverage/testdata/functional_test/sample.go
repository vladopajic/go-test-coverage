//go:build sample
// +build sample

package sample

import "fmt"

func funcHas100PercentCoverage() {
	fmt.Printf("")
}

func zeroCoverageButSuppressedAtFuncLevel() { // coverage-ignore
	println("not covered 1")

	println("not covered 2")
}

func partialCoverage() {

	println("covered")

	if false { // coverage-ignore
		println("not covered 1")
	}

	println("covered")

	{ // coverage-ignore
		println("not covered 2")
	}

	println("covered")
}

func suppressFuncLevelWitNestedBlock() { // coverage-ignore
	if true {
		println("not covered 1")
	} else {
		println("not covered 2")
	}
}
