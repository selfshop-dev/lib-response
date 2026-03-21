package example

// Calc is the default implementation of Calculator.
type Calc struct{}

// Sum returns the sum of a and b.
func (c Calc) Sum(a, b int) int { return a + b }

// Diff returns the difference of a and b.
func (c Calc) Diff(a, b int) int { return a - b }
