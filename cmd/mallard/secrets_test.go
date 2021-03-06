package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolveEnvVars(t *testing.T) {
	testCases := []struct {
		name           string
		setup          func()
		input          string
		expectedResult string
	}{
		{
			name:           "no substitutions",
			input:          "foobar",
			expectedResult: "foobar",
		},
		{
			name:           "failed substitution",
			input:          "${SUB}bar",
			expectedResult: "bar",
		},
		{
			name: "one substitution",
			setup: func() {
				os.Setenv("SUB", "foo")
			},
			input:          "${SUB}bar",
			expectedResult: "foobar",
		},
		{
			name:           "same substitution more than once",
			input:          "${SUB}${SUB}",
			expectedResult: "foofoo",
		},
		{
			name: "multiple substitutions",
			setup: func() {
				os.Setenv("SUB1", "foo")
				os.Setenv("SUB2", "bar")
			},
			input:          "${SUB1}${SUB2}",
			expectedResult: "foobar",
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if testCase.setup != nil {
				testCase.setup()
			}
			require.Equal(t, testCase.expectedResult, resolveEnvVars(testCase.input))
		})
	}
}
