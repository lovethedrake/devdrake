package docker

import "github.com/lovethedrake/drakecore/config"

type adHocPipeline struct {
	jobs []config.PipelineJob
}

func NewAdHocPipeline(jobs []config.Job) config.Pipeline {
	pipeline := &adHocPipeline{
		jobs: make([]config.PipelineJob, len(jobs)),
	}
	for i, job := range jobs {
		pipeline.jobs[i] = &adHocPipelineJob{
			job: job,
		}
	}
	return pipeline
}

func (a *adHocPipeline) Name() string {
	return ""
}

func (a *adHocPipeline) Matches(string, string) (bool, error) {
	return false, nil
}

func (a *adHocPipeline) Jobs() []config.PipelineJob {
	return a.jobs
}

type adHocPipelineJob struct {
	job config.Job
}

func (a *adHocPipelineJob) Job() config.Job {
	return a.job
}

func (a *adHocPipelineJob) Dependencies() []config.PipelineJob {
	return nil
}
