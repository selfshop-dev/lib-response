package example_test

import (
	"fmt"

	example "github.com/selfshop-dev/tpl-seed"
)

// ExampleCalc_Sum demonstrates basic usage and appears in package documentation.
func ExampleCalc_Sum() {
	c := example.Calc{}
	fmt.Println(c.Sum(2, 3))
	// Output: 5
}

// ExampleSummarize demonstrates summing multiple values and appears in package documentation.
func ExampleSummarize() {
	c := example.Calc{}
	fmt.Println(example.Summarize(c, 1, 2, 3, 4))
	// Output: 10
}
