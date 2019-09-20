package config

// Pipeline is a public interface for pipeline configuration.
type Pipeline interface {
	// Name returns the pipeline's name
	Name() string
	// Matches tages branch and/or tag names as criteria to determine if the
	// pipeline is eligible for execution
	Matches(branch, tag string) (bool, error)
	// Jobs returns all the jobs that comprise the pipeline
	Jobs() []PipelineJob
}

type pipeline struct {
	name     string
	selector *pipelineSelector
	jobs     []PipelineJob
}

func (p *pipeline) Name() string {
	return p.name
}

func (p *pipeline) Matches(branch, tag string) (bool, error) {
	// If no criteria are specified, the default is to NOT match
	if p.selector == nil {
		return false, nil
	}
	return p.selector.matches(branch, tag)
}

func (p *pipeline) Jobs() []PipelineJob {
	// We don't want any alterations a caller may make to the slice we return to
	// affect the pipelines 's own jobs slice, which we'd like to treat as
	// immutable, so we return a COPY of that slice.
	jobs := make([]PipelineJob, len(p.jobs))
	copy(jobs, p.jobs)
	return jobs
}
