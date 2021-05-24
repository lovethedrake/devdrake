package config

// PipelineJob is a public interface for a job in a pipeline.
type PipelineJob interface {
	Job() Job
	Dependencies() []PipelineJob
}

// pipelineJob is a composition of a job and its dependencies within a pipeline.
type pipelineJob struct {
	job          Job
	dependencies []PipelineJob
}

func (p *pipelineJob) Job() Job {
	return p.job
}

func (p *pipelineJob) Dependencies() []PipelineJob {
	// We don't want any alterations a caller may make to the slice we return to
	// affect the pipelineJobs's own dependencies slice, which we'd like to treat
	// as immutable, so we return a COPY of that slice.
	dependencies := make([]PipelineJob, len(p.dependencies))
	copy(dependencies, p.dependencies)
	return dependencies
}
