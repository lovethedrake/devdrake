package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPipelineName(t *testing.T) {
	pipeline := &pipeline{
		name: "foo",
	}
	require.Equal(t, pipeline.name, pipeline.Name())
}

func TestPipelineJobs(t *testing.T) {
	pipeline := &pipeline{
		jobs: []PipelineJob{
			&pipelineJob{
				job: &job{
					name: "foo",
				},
			},
			&pipelineJob{
				job: &job{
					name: "foo",
				},
			},
		},
	}
	require.Equal(t, pipeline.jobs, pipeline.Jobs())
}
