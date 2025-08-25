package main

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

func Test_WorkingPull(t *testing.T) {
	test := []struct {
		name            string
		numbers         []int
		workers         int
		ctx             context.Context
		expectedResults []Result
		expectedError   []Error
		expectFuncError bool
	}{
		{
			name:    "SuccessCase",
			numbers: []int{1, 2, 3},
			workers: 2,
			expectedResults: []Result{
				{
					Input:  1,
					Output: int64(1),
				},
				{
					Input:  2,
					Output: int64(2),
				},
				{
					Input:  3,
					Output: int64(6),
				},
			},
			expectedError:   []Error{},
			expectFuncError: false,
		},
		{
			name:    "BigIntError",
			numbers: []int{1, 2, 3, 500},
			workers: 2,
			expectedResults: []Result{
				{
					Input:  1,
					Output: int64(1),
				},
				{
					Input:  2,
					Output: int64(2),
				},
				{
					Input:  3,
					Output: int64(6),
				},
			},
			expectedError: []Error{
				{
					WorkerID: 1,
					Error:    fmt.Errorf("factorial 500 too large for int64"),
				},
			},
			expectFuncError: false,
		},
		{
			name:            "EmptyInput",
			numbers:         []int{},
			workers:         2,
			ctx:             context.Background(),
			expectedResults: []Result{},
			expectedError:   []Error{},
			expectFuncError: true,
		},
		{
			name:    "NegativeNumber",
			numbers: []int{-1, 2, 3},
			workers: 2,
			ctx:     context.Background(),
			expectedResults: []Result{
				{Input: 2, Output: 2},
				{Input: 3, Output: 6},
			},
			expectedError: []Error{
				{Error: fmt.Errorf("negative number -1")},
			},
			expectFuncError: false,
		},
		//{
		//	name:            "TimeoutCase",
		//	numbers:         []int{1, 2, 3},
		//	workers:         2,
		//	ctx:             context.WithTimeout(context.Background(), time.Microsecond * 100),
		//	expectedResults: []Result{},
		//	expectedError: []Error{
		//		{Error: context.DeadlineExceeded},
		//		{Error: context.DeadlineExceeded},
		//	},
		//	expectFuncError: false,
		//},
	}

	for _, tc := range test {
		t.Run(tc.name, func(t *testing.T) {
			var wg sync.WaitGroup
			ctx := tc.ctx
			if ctx == nil {
				ctx = context.Background()
			}
			results, errResult, err := processFactorials(ctx, tc.numbers, tc.workers, &wg)

			if tc.expectFuncError {
				assert.Error(t, err, "expected function error in test %s", tc.name)
				assert.Equal(t, "numbers is empty", err.Error(), "unexpected error message in test %s", tc.name)
			} else {
				assert.NoError(t, err, "unexpected function error in test %s", tc.name)
			}

			assert.Equal(t, len(tc.expectedResults), len(results), "unexpected number of results in test %s", tc.name)
			for i, expected := range tc.expectedResults {
				assert.Less(t, i, len(results), "not enough results in test %s", tc.name)
				assert.Equal(t, expected.Input, results[i].Input, "Input mismatch at index %d in test %s", i, tc.name)
				assert.Equal(t, expected.Output, results[i].Output, "Output mismatch at index %d in test %s", i, tc.name)
			}

			assert.Equal(t, len(tc.expectedError), len(errResult), "unexpected number of errors in test %s", tc.name)
			for i, expected := range tc.expectedError {
				assert.Less(t, i, len(errResult), "not enough errors in test %s", tc.name)
				assert.Equal(t, expected.Error.Error(), errResult[i].Error.Error(), "Error message mismatch at index %d in test %s", i, tc.name)
			}
		})
	}
}
