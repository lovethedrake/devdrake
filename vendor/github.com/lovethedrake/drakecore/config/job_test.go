package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestJobName(t *testing.T) {
	job := &job{
		name: "foo",
	}
	require.Equal(t, job.name, job.Name())
}

func TestJobContainers(t *testing.T) {
	job := &job{
		containers: []Container{
			&container{
				ContainerName: "foo",
			},
			&container{
				ContainerName: "bar",
			},
		},
	}
	require.Equal(t, job.containers, job.Containers())
}
