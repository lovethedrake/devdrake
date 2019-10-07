package docker

import (
	"context"
	"fmt"
	"io"
	"os"

	dockerTypes "github.com/docker/docker/api/types"
	dockerContainer "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/lovethedrake/drakecore/config"
	"github.com/mattn/go-shellwords"
	"github.com/pkg/errors"
)

func (e *executor) executeJob(
	ctx context.Context,
	secrets []string,
	jobExecutionName string,
	sourcePath string,
	job config.Job,
) error {
	if len(job.Containers()) == 0 {
		return nil
	}

	containerIDs := make([]string, len(job.Containers()))

	// Ensure cleanup of all containers
	defer e.forceRemoveContainers(context.Background(), containerIDs...)

	fmt.Printf("----> executing job %q <----\n", job.Name())

	var networkContainerID, lastContainerID string
	var lastContainer config.Container
	// Create and start all containers-- except the last one-- that one we will
	// only create, then we will set ourselves up to capture its output and exit
	// code before we start it.
	for i, container := range job.Containers() {
		containerID, err := e.createContainer(
			ctx,
			secrets,
			jobExecutionName,
			sourcePath,
			networkContainerID,
			container,
		)
		if err != nil {
			return errors.Wrapf(
				err,
				"error creating container %q for job %q",
				container.Name(),
				job.Name(),
			)
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
			if err := e.dockerClient.ContainerStart(
				ctx,
				containerID,
				dockerTypes.ContainerStartOptions{},
			); err != nil {
				return errors.Wrapf(
					err,
					"error starting container %q for job %q",
					container.Name(),
					job.Name(),
				)
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
	containerAttachResp, err := e.dockerClient.ContainerAttach(
		ctx,
		lastContainerID,
		dockerTypes.ContainerAttachOptions{
			Stream: true,
			Stdout: true,
			Stderr: true,
		},
	)
	if err != nil {
		return errors.Wrapf(
			err,
			"error attaching to container %q for job %q",
			lastContainer.Name(),
			job.Name(),
		)
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
				"error processing output from container %q for job %q: %s\n",
				lastContainer.Name(),
				job.Name(),
				err,
			)
		}
	}()
	// Finally, start the last container
	if err := e.dockerClient.ContainerStart(
		ctx,
		lastContainerID,
		dockerTypes.ContainerStartOptions{},
	); err != nil {
		return errors.Wrapf(
			err,
			"error starting container %q for job %q",
			lastContainer.Name(),
			job.Name(),
		)
	}
	select {
	case containerWaitResp := <-containerWaitRespCh:
		if containerWaitResp.StatusCode != 0 {
			// The command executed inside the container exited non-zero
			return &errJobExitedNonZero{
				job:      job.Name(),
				exitCode: containerWaitResp.StatusCode,
			}
		}
	case err := <-containerWaitErrCh:
		if err == ctx.Err() {
			return &errInProgressJobAborted{job: job.Name()}
		}
		return errors.Wrapf(
			err,
			"error waiting for completion of container %q for job %q",
			lastContainer.Name(),
			job.Name(),
		)
	case <-ctx.Done():
		return &errInProgressJobAborted{job: job.Name()}
	}
	return nil
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
		cmd, err := shellwords.Parse(container.Command())
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
				"error creating container %q",
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
