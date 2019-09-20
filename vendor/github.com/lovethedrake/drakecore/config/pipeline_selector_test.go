package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMatches(t *testing.T) {
	testCases := []struct {
		name       string
		pipeline   *pipeline
		assertions func(*testing.T, *pipeline)
	}{
		{
			name:     "a pipeline that can only be manually executed",
			pipeline: &pipeline{},
			assertions: func(t *testing.T, pipeline *pipeline) {
				// Looks like a PR
				matches, err := pipeline.Matches("", "")
				require.NoError(t, err)
				require.False(t, matches)
				// Looks like a merge to master
				matches, err = pipeline.Matches("master", "")
				require.NoError(t, err)
				require.False(t, matches)
				// Looks like a release
				matches, err = pipeline.Matches("", "v0.0.1")
				require.NoError(t, err)
				require.False(t, matches)
			},
		},
		{
			name: "a pipeline that tests PRs",
			pipeline: &pipeline{
				selector: &pipelineSelector{
					BranchSelector: &refSelector{
						BlacklistedRefs: []string{"master"},
					},
				},
			},
			assertions: func(t *testing.T, pipeline *pipeline) {
				// Looks like a PR
				matches, err := pipeline.Matches("", "")
				require.NoError(t, err)
				require.True(t, matches)
				// Looks like a merge to master
				matches, err = pipeline.Matches("master", "")
				require.NoError(t, err)
				require.False(t, matches)
				// Looks like a release
				matches, err = pipeline.Matches("", "v0.0.1")
				require.NoError(t, err)
				require.False(t, matches)
			},
		},
		{
			name: "a pipeline that tests the master branch",
			pipeline: &pipeline{
				selector: &pipelineSelector{
					BranchSelector: &refSelector{
						WhitelistedRefs: []string{"master"},
					},
				},
			},
			assertions: func(t *testing.T, pipeline *pipeline) {
				// Looks like a PR
				matches, err := pipeline.Matches("", "")
				require.NoError(t, err)
				require.False(t, matches)
				// Looks like a merge to master
				matches, err = pipeline.Matches("master", "")
				require.NoError(t, err)
				require.True(t, matches)
				// Looks like a release
				matches, err = pipeline.Matches("", "v0.0.1")
				require.NoError(t, err)
				require.False(t, matches)
			},
		},
		{
			name: "a pipeline for executing a release",
			pipeline: &pipeline{
				selector: &pipelineSelector{
					TagSelector: &refSelector{
						WhitelistedRefs: []string{`/v[0-9]+(\.[0-9]+)*(\-.+)?/`},
					},
				},
			},
			assertions: func(t *testing.T, pipeline *pipeline) {
				// Looks like a PR
				matches, err := pipeline.Matches("", "")
				require.NoError(t, err)
				require.False(t, matches)
				// Looks like a merge to master
				matches, err = pipeline.Matches("master", "")
				require.NoError(t, err)
				require.False(t, matches)
				// Looks like a release
				matches, err = pipeline.Matches("", "v0.0.1")
				require.NoError(t, err)
				require.True(t, matches)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.assertions(t, testCase.pipeline)
		})
	}
}
