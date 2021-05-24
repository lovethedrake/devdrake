package docker

import (
	"context"

	docker "github.com/docker/docker/client"
	"github.com/lovethedrake/go-drake/config"
	"github.com/technosophos/moniker"
)

// Executor is the public interface for the CLI executor
type Executor interface {
	ExecutePipeline(
		ctx context.Context,
		pipeline config.Pipeline,
		secrets map[string]string,
		maxConcurrency int,
	) error
}

type executor struct {
	sourcePath   string
	namer        moniker.Namer
	dockerClient *docker.Client
	debugOnly    bool
}

// NewExecutor returns an executor suitable for use with local development
func NewExecutor(
	sourcePath string,
	dockerClient *docker.Client,
	debugOnly bool,
) Executor {
	return &executor{
		sourcePath:   sourcePath,
		namer:        moniker.New(),
		dockerClient: dockerClient,
		debugOnly:    debugOnly,
	}
}
