package docker

import (
	"context"
	"fmt"
	"sync"

	volumetypes "github.com/docker/docker/api/types/volume"
	"github.com/lovethedrake/drakecore/config"
	"github.com/pkg/errors"
)

func (e *executor) ExecutePipeline(
	ctx context.Context,
	pipeline config.Pipeline,
	secrets []string,
	maxConcurrency int,
) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if e.debugOnly {
		jobs := pipeline.Jobs()
		jobNames := make([]string, len(jobs))
		for i, job := range jobs {
			jobNames[i] = job.Job().Name()
		}
		fmt.Printf("would execute jobs: %q\n", jobNames)
		return nil
	}

	// Build a set of all the image names required for this pipeline.
	imageNames := map[string]struct{}{}
	for _, job := range pipeline.Jobs() {
		imageNames[job.Job().PrimaryContainer().Image()] = struct{}{}
		for _, container := range job.Job().SidecarContainers() {
			imageNames[container.Image()] = struct{}{}
		}
	}
	// Pull all the images before executing anything.
	// TODO: We need to make this conditionally pull images only if:
	// 1. They're not present
	// 2. The image pull policy is "Always"
	if err := e.pullImages(ctx, imageNames); err != nil {
		return err
	}

	// If ANY of the pipeline's jobs' containers mount shared storage, we need to
	// create a volume.
	var pipelineNeedsSharedStorage bool
jobsLoop:
	for _, pipelineJob := range pipeline.Jobs() {
		if pipelineJob.Job().PrimaryContainer().SharedStorageMountPath() != "" {
			pipelineNeedsSharedStorage = true
			break jobsLoop
		}
		for _, sidecarContainer := range pipelineJob.Job().SidecarContainers() {
			if sidecarContainer.SharedStorageMountPath() != "" {
				pipelineNeedsSharedStorage = true
				break jobsLoop
			}
		}
	}

	pipelineExecutionName :=
		fmt.Sprintf("%s-%s", e.namer.NameSep("-"), pipeline.Name())

	var sharedStorageVolumeName string
	if pipelineNeedsSharedStorage {
		sharedStorageVolumeName =
			fmt.Sprintf("%s-shared-storage", pipelineExecutionName)
		if _, err := e.dockerClient.VolumeCreate(
			ctx,
			volumetypes.VolumesCreateBody{
				Name: sharedStorageVolumeName,
			},
		); err != nil {
			return errors.Wrapf(
				err,
				"error creating shared storage volume for pipeline %q",
				pipeline.Name(),
			)
		}
		defer e.forceRemoveVolumes(ctx, sharedStorageVolumeName)
	}

	if _, ok := pipeline.(*adHocPipeline); ok {
		fmt.Println("====> executing ad hoc pipeline <====")
	} else {
		fmt.Printf("====> executing pipeline %q <====\n", pipeline.Name())
	}

	jobs := pipeline.Jobs()

	// We'll put jobs on this channel when they're 100% ready to execute
	jobsCh := make(chan config.PipelineJob)

	// Build a map of channels that lets the job scheduler subscribe to the
	// completion of each job's dependencies. (A given dependency is complete if
	// its channel is closed.)
	dependencyChs := map[string]chan struct{}{}
	for _, job := range jobs {
		dependencyChs[job.Job().Name()] = make(chan struct{})
	}

	// We'll cancel this context if a job fails and we don't want to start any
	// new ones that may be pending. This does NOT mean we cancel jobs that are
	// already in-progress.
	pendingJobsCtx, cancelPendingJobs := context.WithCancel(ctx)
	defer cancelPendingJobs()

	errCh := make(chan error)

	// Start a goroutine to coordinate each job. This doesn't automatically run
	// it; rather it waits for all the job's dependencies to be filled before
	// attempting to schedule it for execution.
	schedulersWg := &sync.WaitGroup{}
	for _, j := range jobs {
		job := j
		schedulersWg.Add(1)
		go func() {
			defer schedulersWg.Done()
			// Wait for the job's dependencies to complete
			for _, dependency := range job.Dependencies() {
				select {
				case <-dependencyChs[dependency.Job().Name()]:
					// Continue to wait for the next dependency
				case <-pendingJobsCtx.Done():
					// Pending jobs were canceled; abort
					errCh <- &errPendingJobCanceled{job: job.Job().Name()}
					return
				case <-ctx.Done():
					// Everything was canceled; abort
					errCh <- &errPendingJobCanceled{job: job.Job().Name()}
					return
				}
			}
			// Schedule the job
			select {
			case jobsCh <- job: // Schedule the job
				// Done
			case <-pendingJobsCtx.Done():
				// Pending jobs were canceled; abort
				errCh <- &errPendingJobCanceled{job: job.Job().Name()}
				return
			case <-ctx.Done():
				// Everything was canceled; abort
				errCh <- &errPendingJobCanceled{job: job.Job().Name()}
				return
			}
		}()
	}

	// When all scheduler goroutines have stopped, there is nothing left to
	// put on the jobsCh. Close it so the executor goroutines can finish up.
	go func() {
		schedulersWg.Wait()
		close(jobsCh)
	}()

	// Fan out to maxConcurrency goroutines for actually executing jobs
	executorsWg := &sync.WaitGroup{}
	for i := 0; i < maxConcurrency; i++ {
		executorsWg.Add(1)
		go func() {
			defer executorsWg.Done()
			for {
				select {
				case job, ok := <-jobsCh:
					if !ok {
						return // The jobsCh was closed
					}
					if err := e.executeJob(
						ctx,
						secrets,
						fmt.Sprintf("%s-%s", pipelineExecutionName, job.Job().Name()),
						e.sourcePath,
						sharedStorageVolumeName,
						job.Job(),
					); err != nil {
						// This errCh write isn't in a select because we don't want it to be
						// interruptable since we never want to lose an error message. And
						// we know the goroutine that is collecting errors is also not
						// interruptable and won't stop listening until all the executor
						// goroutines return, so this is ok.
						errCh <- err
					} else {
						// Unblock anything that's waiting for this job to complete
						close(dependencyChs[job.Job().Name()])
					}
				case <-pendingJobsCtx.Done():
					return // Pending jobs were canceled; abort
				case <-ctx.Done():
					return // Everything was canceled; abort
				}
			}
		}()
	}

	// Convert executorsWg to a channel so we can use it in selects
	allSchedulersAndExecutorsDoneCh := make(chan struct{})
	go func() {
		schedulersWg.Wait()
		executorsWg.Wait()
		close(allSchedulersAndExecutorsDoneCh)
	}()

	// Collect errors from all the executors until they have all completed
	errs := []error{}
errLoop:
	for {
		// Note this select isn't interruptable by canceled contexts because we
		// never want to lose an error message. We know this will inevitably unblock
		// when all the executor goroutines conclude-- which they WILL since those
		// are interruptable.
		select {
		case err := <-errCh:
			if err != nil {
				errs = append(errs, err)
				// Once we've had any error, we know the pipeline is failed. We can
				// let jobs already in-progress continue executing, but we don't want
				// to start any new ones. We can signal that by closing this context.
				cancelPendingJobs()
			}
		case <-allSchedulersAndExecutorsDoneCh:
			break errLoop
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

func (e *executor) forceRemoveVolumes(
	ctx context.Context,
	volumeNames ...string,
) {
	for _, volumeName := range volumeNames {
		if err := e.dockerClient.VolumeRemove(ctx, volumeName, true); err != nil {
			// TODO: Maybe this isn't the best way to deal with this
			fmt.Printf(`error removing volume "%s": %s`, volumeName, err)
		}
	}
}
