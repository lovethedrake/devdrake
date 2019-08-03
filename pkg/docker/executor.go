package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	dockerTypes "github.com/docker/docker/api/types"
	dockerContainer "github.com/docker/docker/api/types/container"
	docker "github.com/docker/docker/client"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/lovethedrake/drakecore/config"
	"github.com/mattn/go-shellwords"
	"github.com/pkg/errors"
	"github.com/technosophos/moniker"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

// Executor is the public interface for the CLI executor
type Executor interface {
	ExecuteJobs(
		ctx context.Context,
		configFile string,
		secretsFile string,
		sourcePath string,
		jobNames []string,
		debugOnly bool,
		concurrencyEnabled bool,
	) error
	ExecutePipelines(
		ctx context.Context,
		configFile string,
		secretsFile string,
		sourcePath string,
		pipelineNames []string,
		debugOnly bool,
		concurrencyEnabled bool,
	) error
}

type executor struct {
	namer        moniker.Namer
	dockerClient *docker.Client
}

// NewExecutor returns an executor suitable for use with local development
func NewExecutor(dockerClient *docker.Client) Executor {
	return &executor{
		namer:        moniker.New(),
		dockerClient: dockerClient,
	}
}

func (e *executor) ExecuteJobs(
	ctx context.Context,
	configFile string,
	secretsFile string,
	sourcePath string,
	jobNames []string,
	debugOnly bool,
	concurrencyEnabled bool,
) error {
	config, err := config.NewConfigFromFile(configFile)
	if err != nil {
		return err
	}
	secrets, err := secretsFromFile(secretsFile)
	if err != nil {
		return err
	}
	jobs, err := config.Jobs(jobNames)
	if err != nil {
		return err
	}
	if debugOnly {
		fmt.Printf("would execute jobs: %s\n", jobNames)
		return nil
	}

	imageNames := map[string]struct{}{}
	for _, job := range jobs {
		for _, container := range job.Containers() {
			imageNames[container.Image()] = struct{}{}
		}
	}
	for imageName := range imageNames {
		fmt.Printf("~~~~> pulling image \"%s\" <~~~~\n", imageName)
		reader, err := e.dockerClient.ImagePull(
			ctx,
			imageName,
			dockerTypes.ImagePullOptions{},
		)
		if err != nil {
			return err
		}
		defer reader.Close()
		dec := json.NewDecoder(reader)
		for {
			var message jsonmessage.JSONMessage
			if err := dec.Decode(&message); err != nil {
				if err == io.EOF {
					break
				}
				return err
			}
			fmt.Println(message.Status)
		}
	}

	executionName := e.namer.NameSep("-")
	errCh := make(chan error)
	var runningJobs int
	for _, job := range jobs {
		jobExecutionName := fmt.Sprintf("%s-%s", executionName, job.Name())
		runningJobs++
		go e.executeJob(
			ctx,
			secrets,
			jobExecutionName,
			sourcePath,
			job,
			errCh,
		)
		if !concurrencyEnabled {
			// If concurrency isn't enabled, wait for a potential error. If it's nil,
			// move on. If it's not, return the error.
			if err := <-errCh; err != nil {
				return err
			}
			runningJobs--
		}
	}
	// If concurrency isn't enabled and we haven't already encountered an error,
	// then we're not going to. We're done!
	if !concurrencyEnabled {
		return nil
	}
	// Wait for all the jobs to finish.
	errs := []error{}
	for err := range errCh {
		if err != nil {
			errs = append(errs, err)
		}
		runningJobs--
		if runningJobs == 0 {
			break
		}
	}
	if len(errs) > 1 {
		return &multiError{errs: errs}
	}
	if len(errs) == 1 {
		return errs[0]
	}
	return nil
}

func (e *executor) ExecutePipelines(
	ctx context.Context,
	configFile string,
	secretsFile string,
	sourcePath string,
	pipelineNames []string,
	debugOnly bool,
	concurrencyEnabled bool,
) error {
	config, err := config.NewConfigFromFile(configFile)
	if err != nil {
		return err
	}
	secrets, err := secretsFromFile(secretsFile)
	if err != nil {
		return err
	}
	pipelines, err := config.Pipelines(pipelineNames)
	if err != nil {
		return err
	}
	if debugOnly {
		fmt.Println("would execute:")
		for _, pipeline := range pipelines {
			jobs := make([][]string, len(pipeline.Jobs()))
			for i, stageJobs := range pipeline.Jobs() {
				jobs[i] = make([]string, len(stageJobs))
				for j, job := range stageJobs {
					jobs[i][j] = job.Name()
				}
			}
			fmt.Printf("  %s jobs: %s\n", pipeline.Name(), jobs)
		}
		return nil
	}

	imageNames := map[string]struct{}{}
	for _, pipeline := range pipelines {
		for _, stageJobs := range pipeline.Jobs() {
			for _, job := range stageJobs {
				for _, container := range job.Containers() {
					imageNames[container.Image()] = struct{}{}
				}
			}
		}
	}
	for imageName := range imageNames {
		fmt.Printf("~~~~> pulling image \"%s\" <~~~~\n", imageName)
		reader, err := e.dockerClient.ImagePull(
			ctx,
			imageName,
			dockerTypes.ImagePullOptions{},
		)
		if err != nil {
			return err
		}
		defer reader.Close()
		dec := json.NewDecoder(reader)
		for {
			var message jsonmessage.JSONMessage
			if err := dec.Decode(&message); err != nil {
				if err == io.EOF {
					break
				}
				return err
			}
			fmt.Println(message.Status)
		}
	}

	executionName := e.namer.NameSep("-")
	for _, pipeline := range pipelines {
		fmt.Printf("====> executing pipeline \"%s\" <====\n", pipeline.Name())
		pipelineExecutionName :=
			fmt.Sprintf("%s-%s", executionName, pipeline.Name())
		for i, stageJobs := range pipeline.Jobs() {
			fmt.Printf("====> executing stage %d <====\n", i)
			stageExecutionName :=
				fmt.Sprintf("%s-stage%d", pipelineExecutionName, i)
			errCh := make(chan error)
			var runningJobs int
			for _, job := range stageJobs {
				jobExecutionName :=
					fmt.Sprintf("%s-%s", stageExecutionName, job.Name())
				runningJobs++
				go e.executeJob(
					ctx,
					secrets,
					jobExecutionName,
					sourcePath,
					job,
					errCh,
				)
				// If concurrency isn't enabled, wait for a potential error. If it's
				// nil, move on. If it's not, return the error.
				if !concurrencyEnabled {
					if err := <-errCh; err != nil {
						return err
					}
					runningJobs--
				}
			}
			// If concurrency is enabled, wait for all the jobs to finish.
			if concurrencyEnabled {
				errs := []error{}
				for err := range errCh {
					if err != nil {
						errs = append(errs, err)
					}
					runningJobs--
					if runningJobs == 0 {
						break
					}
				}
				if len(errs) > 1 {
					return &multiError{errs: errs}
				}
				if len(errs) == 1 {
					return errs[0]
				}
			}
		}
	}
	return nil
}

func (e *executor) executeJob(
	ctx context.Context,
	secrets []string,
	jobExecutionName string,
	sourcePath string,
	job config.Job,
	errCh chan<- error,
) {
	var err error
	containerIDs := make([]string, len(job.Containers()))

	// Ensure cleanup of all containers
	defer func() {
		e.forceRemoveContainers(context.Background(), containerIDs...)
		errCh <- err
	}()

	if len(job.Containers()) == 0 {
		return
	}

	fmt.Printf("----> executing job \"%s\" <----\n", job.Name())

	var networkContainerID, lastContainerID string
	var lastContainer config.Container
	// Create and start all containers-- except the last one-- that one we will
	// only create, then we will set ourselves up to capture its output and exit
	// code before we start it.
	for i, container := range job.Containers() {
		var containerID string
		if containerID, err = e.createContainer(
			ctx,
			secrets,
			jobExecutionName,
			sourcePath,
			networkContainerID,
			container,
		); err != nil {
			err = errors.Wrapf(
				err,
				"error creating container \"%s\" for job \"%s\"",
				container.Name(),
				job.Name(),
			)
			return
		}
		containerIDs[i] = containerID
		if i == 0 {
			networkContainerID = containerID
		}
		if i == len(containerIDs)-1 {
			lastContainerID = containerID
			lastContainer = container
		} else {
			// Start all but the last container
			if err = e.dockerClient.ContainerStart(
				ctx,
				containerID,
				dockerTypes.ContainerStartOptions{},
			); err != nil {
				err = errors.Wrapf(
					err,
					"error starting container \"%s\" for job \"%s\"",
					container.Name(),
					job.Name(),
				)
				return
			}
		}
	}
	// Establish channels to use for waiting for the last container to exit
	containerWaitRespCh, containerWaitErrCh := e.dockerClient.ContainerWait(
		ctx,
		lastContainerID,
		dockerContainer.WaitConditionNextExit,
	)
	// Attach to the last container to see its output
	var containerAttachResp dockerTypes.HijackedResponse
	if containerAttachResp, err = e.dockerClient.ContainerAttach(
		ctx,
		lastContainerID,
		dockerTypes.ContainerAttachOptions{
			Stream: true,
			Stdout: true,
			Stderr: true,
		},
	); err != nil {
		err = errors.Wrapf(
			err,
			"error attaching to container \"%s\" for job \"%s\"",
			lastContainer.Name(),
			job.Name(),
		)
		return
	}
	// Concurrently deal with the output from the last container
	go func() {
		defer containerAttachResp.Close()
		var gerr error
		stdOutWriter := prefixingWriter(
			job.Name(),
			lastContainer.Name(),
			os.Stdout,
		)
		if lastContainer.TTY() {
			_, gerr = io.Copy(stdOutWriter, containerAttachResp.Reader)
		} else {
			stdErrWriter := prefixingWriter(
				job.Name(),
				lastContainer.Name(),
				os.Stderr,
			)
			_, gerr = stdcopy.StdCopy(
				stdOutWriter,
				stdErrWriter,
				containerAttachResp.Reader,
			)
		}
		if gerr != nil {
			fmt.Printf(
				"error processing output from container \"%s\" for job \"%s\": %s\n",
				lastContainer.Name(),
				job.Name(),
				err,
			)
		}
	}()
	// Finally, start the last container
	if err = e.dockerClient.ContainerStart(
		ctx,
		lastContainerID,
		dockerTypes.ContainerStartOptions{},
	); err != nil {
		err = errors.Wrapf(
			err,
			"error starting container \"%s\" for job \"%s\"",
			lastContainer.Name(),
			job.Name(),
		)
		return
	}
	select {
	case containerWaitResp := <-containerWaitRespCh:
		if containerWaitResp.StatusCode != 0 {
			// The command executed inside the container exited non-zero
			err = &errJobExitedNonZero{
				Job:      job.Name(),
				ExitCode: containerWaitResp.StatusCode,
			}
			return
		}
	case err = <-containerWaitErrCh:
		err = errors.Wrapf(
			err,
			"error waiting for completion of container \"%s\" for job \"%s\"",
			lastContainer.Name(),
			job.Name(),
		)
		return
	case <-ctx.Done():
	}
}

// createContainer creates a container for the given execution and job,
// taking source path, any established networking, and container-specific
// configuration into account. It returns the newly created container's ID. It
// does not start the container.
func (e *executor) createContainer(
	ctx context.Context,
	secrets []string,
	jobExecutionName string,
	sourcePath string,
	networkContainerID string,
	container config.Container,
) (string, error) {

	sha := "unknown"
	var branch string

	// TODO: We should probably move this somewhere else
	workDir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	repo, err := git.PlainOpen(workDir)
	if err != nil && err != git.ErrRepositoryNotExists {
		return "", err
	}
	if repo != nil {
		ref, rerr := repo.Head()
		if rerr != nil && rerr != plumbing.ErrReferenceNotFound {
			return "", rerr
		}
		if ref != nil {
			sha = ref.Hash().String()
			branch = ref.Name().Short()
		}
	}
	// TODO: End "we should probably move this somewhere else"

	env := make([]string, len(secrets))
	copy(env, secrets)
	env = append(env, fmt.Sprintf("DRAKE_SHA1=%s", sha))
	env = append(env, fmt.Sprintf("DRAKE_BRANCH=%s", branch))
	env = append(env, "DRAKE_TAG=")

	containerConfig := &dockerContainer.Config{
		Image:        container.Image(),
		Env:          append(env, container.Environment()...),
		WorkingDir:   container.WorkingDirectory(),
		Tty:          container.TTY(),
		AttachStdout: true,
		AttachStderr: true,
	}
	if container.Command() != "" {
		var cmd []string
		cmd, err = shellwords.Parse(container.Command())
		if err != nil {
			return "", errors.Wrap(err, "error parsing container command")
		}
		containerConfig.Cmd = cmd
	}
	hostConfig := &dockerContainer.HostConfig{
		Privileged: container.Privileged(),
	}
	if networkContainerID != "" {
		hostConfig.NetworkMode = dockerContainer.NetworkMode(
			fmt.Sprintf("container:%s", networkContainerID),
		)
	}
	if container.MountDockerSocket() {
		hostConfig.Binds = []string{"/var/run/docker.sock:/var/run/docker.sock"}
	}
	if container.SourceMountPath() != "" {
		hostConfig.Binds = append(
			hostConfig.Binds,
			fmt.Sprintf("%s:%s", sourcePath, container.SourceMountPath()),
		)
	}
	fullContainerName := fmt.Sprintf(
		"%s-%s",
		jobExecutionName,
		container.Name(),
	)
	containerCreateResp, err := e.dockerClient.ContainerCreate(
		ctx,
		containerConfig,
		hostConfig,
		nil,
		fullContainerName,
	)
	if err != nil {
		return "",
			errors.Wrapf(
				err,
				"error creating container \"%s\"",
				fullContainerName,
			)
	}
	return containerCreateResp.ID, nil
}

func (e *executor) forceRemoveContainers(
	ctx context.Context,
	containerIDs ...string,
) {
	for _, containerID := range containerIDs {
		if err := e.dockerClient.ContainerRemove(
			ctx,
			containerID,
			dockerTypes.ContainerRemoveOptions{
				Force: true,
			},
		); err != nil {
			// TODO: Maybe this isn't the best way to deal with this
			fmt.Printf(`error removing container "%s": %s`, containerID, err)
		}
	}
}
