package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMatches(t *testing.T) {

	// A pipeline that can only be manually executed
	pl := &pipeline{}
	// Looks like a PR
	matches, err := pl.Matches("", "")
	require.NoError(t, err)
	require.False(t, matches)
	// Looks like a merge to master
	matches, err = pl.Matches("master", "")
	require.NoError(t, err)
	require.False(t, matches)
	// Looks like a release
	matches, err = pl.Matches("", "v0.0.1")
	require.NoError(t, err)
	require.False(t, matches)

	// A pipeline that tests PRS
	pl = &pipeline{
		Selector: &pipelineSelector{
			BranchSelector: &refSelector{
				BlacklistedRefs: []string{"master"},
			},
		},
	}
	// Looks like a PR
	matches, err = pl.Matches("", "")
	require.NoError(t, err)
	require.True(t, matches)
	// Looks like a merge to master
	matches, err = pl.Matches("master", "")
	require.NoError(t, err)
	require.False(t, matches)
	// Looks like a release
	matches, err = pl.Matches("", "v0.0.1")
	require.NoError(t, err)
	require.False(t, matches)

	// A pipeline that tests the master branch
	pl = &pipeline{
		Selector: &pipelineSelector{
			BranchSelector: &refSelector{
				WhitelistedRefs: []string{"master"},
			},
		},
	}
	// Looks like a PR
	matches, err = pl.Matches("", "")
	require.NoError(t, err)
	require.False(t, matches)
	// Looks like a merge to master
	matches, err = pl.Matches("master", "")
	require.NoError(t, err)
	require.True(t, matches)
	// Looks like a release
	matches, err = pl.Matches("", "v0.0.1")
	require.NoError(t, err)
	require.False(t, matches)

	// A pipeline for executing a release
	pl = &pipeline{
		Selector: &pipelineSelector{
			TagSelector: &refSelector{
				WhitelistedRefs: []string{`/v[0-9]+(\.[0-9]+)*(\-.+)?/`},
			},
		},
	}
	// Looks like a PR
	matches, err = pl.Matches("", "")
	require.NoError(t, err)
	require.False(t, matches)
	// Looks like a merge to master
	matches, err = pl.Matches("master", "")
	require.NoError(t, err)
	require.False(t, matches)
	// Looks like a release
	matches, err = pl.Matches("", "v0.0.1")
	require.NoError(t, err)
	require.True(t, matches)
}
