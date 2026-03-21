package example_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	example "github.com/selfshop-dev/tpl-seed"
)

func TestSummarize(t *testing.T) {
	t.Parallel()

	testCases := [...]struct {
		name   string
		values []int
		want   int
	}{
		{name: "single value", values: []int{5}, want: 5},
		{name: "multiple values", values: []int{1, 2, 3}, want: 6},
		{name: "empty", values: []int{}, want: 0},
		{name: "negative values", values: []int{-1, -2, -3}, want: -6},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			// defer ctrl.Finish() — not needed, registered via t.Cleanup automatically
			mock := example.NewMockCalculator(ctrl)
			// Expect Sum to be called once per value — accumulating into result.
			result := 0
			for _, v := range tc.values {
				prev := result
				result += v
				mock.EXPECT().Sum(prev, v).Return(result)
			}

			got := example.Summarize(mock, tc.values...)
			assert.Equal(t, tc.want, got)
		})
	}
}
