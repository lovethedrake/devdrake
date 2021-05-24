package docker

import "github.com/lovethedrake/go-drake/config"

type adHocPipeline struct {
	jobs []config.PipelineJob
}

// NewAdHocPipeline constructs a new config.Pipeline from the provided
// config.Jobs.
func NewAdHocPipeline(jobs []config.Job) config.Pipeline {
	pipeline := &adHocPipeline{
		jobs: make([]config.PipelineJob, len(jobs)),
	}
	var previousJob config.PipelineJob
	for i, job := range jobs {
		pipeline.jobs[i] = &adHocPipelineJob{
			job:        job,
			dependency: previousJob,
		}
		previousJob = pipeline.jobs[i]
	}
	return pipeline
}

func (a *adHocPipeline) Name() string {
	return ""
}

func (a *adHocPipeline) Jobs() []config.PipelineJob {
	return a.jobs
}

func (a *adHocPipeline) Triggers() []config.PipelineTrigger {
	return []config.PipelineTrigger{}
}

type adHocPipelineJob struct {
	job        config.Job
	dependency config.PipelineJob
}

func (a *adHocPipelineJob) Job() config.Job {
	return a.job
}

func (a *adHocPipelineJob) Dependencies() []config.PipelineJob {
	if a.dependency == nil {
		return nil
	}
	return []config.PipelineJob{a.dependency}
}
