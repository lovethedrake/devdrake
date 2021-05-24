package main

import (
	"path/filepath"

	docker "github.com/docker/docker/client"
	"github.com/lovethedrake/go-drake/config"
	libDocker "github.com/lovethedrake/mallard/pkg/docker"
	"github.com/lovethedrake/mallard/pkg/signals"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func run(c *cli.Context) error {
	// This context will automatically be canceled on SIGINT or SIGTERM.
	ctx := signals.Context()
	configFile := c.GlobalString(flagFile)
	secretsFile := c.String(flagSecretsFile)
	debugOnly := c.Bool(flagDebug)
	maxConcurrency := c.Int(flagConcurrency)
	absConfigFilePath, err := filepath.Abs(configFile)
	if err != nil {
		return err
	}
	sourcePath := filepath.Dir(absConfigFilePath)
	dockerClient, err := docker.NewClientWithOpts(docker.FromEnv)
	if err != nil {
		return errors.Wrap(err, "error building Docker client")
	}

	cfg, err := config.NewConfigFromFile(configFile)
	if err != nil {
		return err
	}
	secrets, err := secretsFromFile(secretsFile)
	if err != nil {
		return err
	}

	// TODO: Should pass the stream that we want output to go to-- stdout
	executor := libDocker.NewExecutor(sourcePath, dockerClient, debugOnly)
	if c.Bool(flagPipeline) {
		if len(c.Args()) == 0 {
			return errors.New("no pipeline was specified for execution")
		}
		if len(c.Args()) > 1 {
			return errors.New("only one pipeline may be executed at a time")
		}
		var pipelines []config.Pipeline
		pipelines, err = cfg.Pipelines(c.Args()...)
		if err != nil {
			return err
		}
		return executor.ExecutePipeline(ctx, pipelines[0], secrets, maxConcurrency)
	}
	if len(c.Args()) == 0 {
		return errors.New("no jobs were specified for execution")
	}
	var jobs []config.Job
	jobs, err = cfg.Jobs(c.Args()...)
	if err != nil {
		return err
	}
	return executor.ExecutePipeline(
		ctx,
		libDocker.NewAdHocPipeline(jobs),
		secrets,
		maxConcurrency,
	)
}
