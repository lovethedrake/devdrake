package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPipelineJobJob(t *testing.T) {
	pipelineJob := &pipelineJob{
		job: &job{
			name: "foo",
		},
	}
	require.Equal(t, pipelineJob.job, pipelineJob.Job())
}

func TestPipelineJobDependencies(t *testing.T) {
	pipelineJob := &pipelineJob{
		dependencies: []PipelineJob{
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
	require.Equal(t, pipelineJob.dependencies, pipelineJob.Dependencies())
}
