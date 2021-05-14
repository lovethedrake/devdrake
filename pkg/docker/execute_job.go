package docker

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"sort"

	dockerTypes "github.com/docker/docker/api/types"
	dockerContainer "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/docker/pkg/sysinfo"
	"github.com/lovethedrake/devdrake/pkg/file"
	"github.com/lovethedrake/drakecore/config"
	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
)

const (
	millicoresPerCPU = 1000
)

func (e *executor) executeJob(
	ctx context.Context,
	secrets map[string]string,
	jobExecutionName string,
	sourcePath string,
	sharedStorageVolumeName string,
	job config.Job,
) error {
	primaryContainer := job.PrimaryContainer()
	// Make a copy of source in the working directory if that is what
	// job configuration says to do
	jobSrcPath := sourcePath
	if job.SourceMountMode() == config.SourceMountModeCopy {
		var jobNeedsSource bool
		if primaryContainer.SourceMountPath() != "" {
			jobNeedsSource = true
		}
		if !jobNeedsSource {
			for _, sidecarContainer := range job.SidecarContainers() {
				if sidecarContainer.SourceMountPath() != "" {
					jobNeedsSource = true
					break
				}
			}
		}
		if jobNeedsSource {
			homePath, err := homedir.Dir()
			if err != nil {
				return errors.Wrap(err, "error finding home directory")
			}
			// TODO: Move this into its own package? This probably won't be the last
			// time we need to find "devdrakeHome".
			jobPath := path.Join(homePath, ".devdrake", "jobs", jobExecutionName)
			if _, err := os.Stat(jobPath); err != nil {
				if !os.IsNotExist(err) {
					return errors.Wrapf(
						err,
						"error checkiong for existence of directory %s",
						jobPath,
					)
				}
				if err := os.MkdirAll(jobPath, 0755); err != nil {
					return errors.Wrapf(err, "error creating directory %s", jobPath)
				}
			}
			jobSrcPath = path.Join(jobPath, "src")
			defer os.RemoveAll(jobPath) // nolint: errcheck
			if err := file.CopyDir(sourcePath, jobSrcPath); err != nil {
				return errors.Wrapf(err, "error copying source to %s", jobSrcPath)
			}
		}
	}

	// Slice big enough for the primary container and all sidecars
	containerIDs := make([]string, 1+len(job.SidecarContainers()))

	// Ensure cleanup of all containers
	defer e.forceRemoveContainers(context.Background(), containerIDs...)

	fmt.Printf("----> executing job %q <----\n", job.Name())

	var networkContainerID string
	// Create and start all sidecar containers
	for i, sidecarContainer := range job.SidecarContainers() {
		sidecarContainerID, err := e.createContainer(
			ctx,
			secrets,
			jobExecutionName,
			jobSrcPath,
			job.SourceMountMode(),
			sharedStorageVolumeName,
			networkContainerID,
			job.OSFamily(),
			sidecarContainer,
		)
		if err != nil {
			return errors.Wrapf(
				err,
				"error creating sidecar container %q for job %q",
				sidecarContainer.Name(),
				job.Name(),
			)
		}
		containerIDs[i] = sidecarContainerID
		if i == 0 {
			networkContainerID = sidecarContainerID
		}
		if err := e.dockerClient.ContainerStart(
			ctx,
			sidecarContainerID,
			dockerTypes.ContainerStartOptions{},
		); err != nil {
			return errors.Wrapf(
				err,
				"error starting sidecar container %q for job %q",
				sidecarContainer.Name(),
				job.Name(),
			)
		}
	}
	// Create the primary container
	primaryContainerID, err := e.createContainer(
		ctx,
		secrets,
		jobExecutionName,
		jobSrcPath,
		job.SourceMountMode(),
		sharedStorageVolumeName,
		networkContainerID,
		job.OSFamily(),
		primaryContainer,
	)
	if err != nil {
		return errors.Wrapf(
			err,
			"error creating primary container %q for job %q",
			primaryContainer.Name(),
			job.Name(),
		)
	}
	containerIDs[len(containerIDs)-1] = primaryContainerID
	// Establish channels to use for waiting for the primary container to exit
	primaryContainerWaitRespCh, primaryContainerWaitErrCh :=
		e.dockerClient.ContainerWait(
			ctx,
			primaryContainerID,
			dockerContainer.WaitConditionNextExit,
		)
	// Attach to the primary container to see its output
	primaryContainerAttachResp, err := e.dockerClient.ContainerAttach(
		ctx,
		primaryContainerID,
		dockerTypes.ContainerAttachOptions{
			Stream: true,
			Stdout: true,
			Stderr: true,
		},
	)
	if err != nil {
		return errors.Wrapf(
			err,
			"error attaching to primary container %q for job %q",
			primaryContainer.Name(),
			job.Name(),
		)
	}
	// Concurrently deal with the output from the primary container
	go func() {
		defer primaryContainerAttachResp.Close()
		var gerr error
		stdOutWriter := prefixingWriter(
			job.Name(),
			primaryContainer.Name(),
			os.Stdout,
		)
		if primaryContainer.TTY() {
			_, gerr = io.Copy(stdOutWriter, primaryContainerAttachResp.Reader)
		} else {
			stdErrWriter := prefixingWriter(
				job.Name(),
				primaryContainer.Name(),
				os.Stderr,
			)
			_, gerr = stdcopy.StdCopy(
				stdOutWriter,
				stdErrWriter,
				primaryContainerAttachResp.Reader,
			)
		}
		if gerr != nil {
			fmt.Printf(
				"error processing output from primary container %q for job %q: %s\n",
				primaryContainer.Name(),
				job.Name(),
				err,
			)
		}
	}()
	// Finally, start the primary container
	if err := e.dockerClient.ContainerStart(
		ctx,
		primaryContainerID,
		dockerTypes.ContainerStartOptions{},
	); err != nil {
		return errors.Wrapf(
			err,
			"error starting primary container %q for job %q",
			primaryContainer.Name(),
			job.Name(),
		)
	}
	select {
	case primaryContainerWaitResp := <-primaryContainerWaitRespCh:
		if primaryContainerWaitResp.StatusCode != 0 {
			// The command executed inside the container exited non-zero
			return &errJobExitedNonZero{
				job:      job.Name(),
				exitCode: primaryContainerWaitResp.StatusCode,
			}
		}
	case err := <-primaryContainerWaitErrCh:
		if err == ctx.Err() {
			return &errInProgressJobAborted{job: job.Name()}
		}
		return errors.Wrapf(
			err,
			"error waiting for completion of primary container %q for job %q",
			primaryContainer.Name(),
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
	secrets map[string]string,
	jobExecutionName string,
	sourcePath string,
	sourceMountMode config.SourceMountMode,
	sharedStorageVolumeName string,
	networkContainerID string,
	osFamily config.OSFamily,
	container config.Container,
) (string, error) {
	env := secrets
	if env == nil {
		env = map[string]string{}
	}
	for k, v := range container.Environment() {
		env[k] = v
	}
	envSlice := config.EnvironmentMapToSlice(container.Environment())

	containerConfig := &dockerContainer.Config{
		Image:        container.Image(),
		Env:          envSlice,
		WorkingDir:   container.WorkingDirectory(),
		Tty:          container.TTY(),
		AttachStdout: true,
		AttachStderr: true,
	}
	cmd := container.Command()
	if len(cmd) > 0 {
		containerConfig.Entrypoint = cmd
	}
	args := container.Args()
	if len(args) > 0 {
		containerConfig.Cmd = args
	}

	// Work out resource limits... involves some math
	const (
		minMegabytes     = 4
		bytesPerMegabyte = 1000000
	)
	maxMegabytes := int64(container.Resources().Memory().MaxMegabytes())
	if maxMegabytes < minMegabytes {
		maxMegabytes = minMegabytes
	}
	memoryBytes := maxMegabytes * bytesPerMegabyte
	hostConfig := &dockerContainer.HostConfig{
		Resources: dockerContainer.Resources{
			Memory: memoryBytes,
		},
		Privileged: container.Privileged(),
	}
	requestedMillicores :=
		int64(container.Resources().CPU().RequestedMillicores())
	maxMillicores := int64(container.Resources().CPU().MaxMillicores())
	if osFamily == config.OSFamilyLinux {
		hostConfig.Resources.CPUShares =
			millicoresToSharesLinux(requestedMillicores)
		const quotaPeriod = 100000 // 100000 is equivalent to 100ms
		hostConfig.Resources.CPUPeriod = quotaPeriod
		hostConfig.Resources.CPUQuota =
			millicoresToCPUQuotaLinux(maxMillicores, quotaPeriod)
	} else if osFamily == config.OSFamilyWindows {
		hyperv := true // TODO: Don't assume hyperv
		hostConfig.Resources.CPUShares = millicoresToSharesWindows(
			requestedMillicores,
			hyperv,
		)
		hostConfig.Resources.CPUCount =
			millicoresToCPUCountWindows(maxMillicores, hyperv)
		hostConfig.Resources.CPUPercent =
			millicoresToCPUPercentWindows(maxMillicores, hyperv)
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
		containerSourceMountPath := container.SourceMountPath()
		if sourceMountMode == config.SourceMountModeReadOnly {
			containerSourceMountPath = fmt.Sprintf("%s:ro", containerSourceMountPath)
		}
		hostConfig.Binds = append(
			hostConfig.Binds,
			fmt.Sprintf("%s:%s", sourcePath, containerSourceMountPath),
		)
	}
	if container.SharedStorageMountPath() != "" {
		hostConfig.Binds = append(
			hostConfig.Binds,
			fmt.Sprintf(
				"%s:%s",
				sharedStorageVolumeName,
				container.SharedStorageMountPath(),
			),
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

func millicoresToSharesLinux(millicores int64) int64 {
	const (
		minShares    = 2
		sharesPerCPU = 1024
	)
	if millicores == 0 {
		// Return minimum share supported by the kernel
		return minShares
	}
	// Conceptually millicores / millicoresPerCPU) * sharesPerCPU, but factored to
	// improve rounding.
	shares := millicores * sharesPerCPU / millicoresPerCPU
	if shares <= minShares {
		return minShares
	}
	return shares
}

func millicoresToSharesWindows(millicores int64, hyperv bool) int64 {
	const (
		// nolint: lll
		// https://docs.microsoft.com/en-us/virtualization/windowscontainers/manage-containers/resource-controls
		minSharesProcess = 5000
		minSharesHyperV  = 10
		maxShares        = 10000
	)
	var minShares int64 = minSharesProcess
	if hyperv {
		minShares = minSharesHyperV
	}
	if millicores == 0 {
		// Return minimum share supported by the kernel
		return minShares
	}
	// Conceptually (millicores / millicoresPerCPU) * sharesPerCPU, but factored
	// to improve rounding.
	totalCPU := sysinfo.NumCPU()
	shares :=
		(millicores * (maxShares - minShares)) / int64(totalCPU) / millicoresPerCPU
	if shares < minShares {
		return minShares
	}
	return shares
}

func millicoresToCPUQuotaLinux(millicores int64, quotaPeriod int64) int64 {
	if millicores == 0 {
		return 0
	}
	const (
		// 100000 is equivalent to 100ms
		minQuotaPeriod = 1000
	)
	// Conceptually (millicores / millicoresPerCPU) * quotaPeriod, but factored
	// to improve rounding.
	cpuQuota := millicores * quotaPeriod / millicoresPerCPU
	if cpuQuota < minQuotaPeriod {
		return minQuotaPeriod
	}
	return cpuQuota
}

func millicoresToCPUCountWindows(millicores int64, hyperv bool) int64 {
	if !hyperv {
		return 0
	}
	return (millicores + 999) / 1000
}

func millicoresToCPUPercentWindows(millicores int64, hyperv bool) int64 {
	// Much of this algorithm was largely cribbed from Kubernetes, but it seems to
	// find a "cpuMaximum" that is a factor of 100 larger than what we expect the
	// cpuPercent to be. The algorithm is adjusted accordingly, with the original
	// lines commented in case we ever figure out exactly what was going on here.
	// In the meantime, this *appears* to produce correct values that work fine
	// with Docker. (Those that are off by a factor of 100 simply do NOT work.)
	// cpuMaximum := 10000 * millicores / int64(sysinfo.NumCPU()) / 1000
	cpuPercent := 10000 * millicores / int64(sysinfo.NumCPU()) / 1000 / 100
	if hyperv {
		cpuCount := millicoresToCPUCountWindows(millicores, hyperv)
		if cpuCount != 0 {
			// cpuMaximum = millicores / cpuCount * 10000 / 1000
			cpuPercent = millicores / cpuCount * 10000 / 1000 / 100
		}
	}
	// // ensure cpuMaximum is in range [1, 10000].
	// if cpuMaximum < 1 {
	// 	cpuMaximum = 1
	// } else if cpuMaximum > 10000 {
	// 	cpuMaximum = 10000
	// }
	// ensure cpuPercent is in range [1, 100].
	if cpuPercent < 1 {
		cpuPercent = 1
	} else if cpuPercent > 100 {
		cpuPercent = 100
	}
	// return cpuMaximum
	return cpuPercent
}

// environmentMapToSlice converts a map of environment variables to a slice of
// KEY=VALUE pairs.
func environmentMapToSlice(env map[string]string) []string {
	vars := make([]string, 0, len(env))
	for k, v := range env {
		vars = append(vars, fmt.Sprintf("%s=%s", k, v))
	}
	sort.Strings(vars)
	return vars
}
