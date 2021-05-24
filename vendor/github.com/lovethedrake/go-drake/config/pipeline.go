package config

// Pipeline is a public interface for pipeline configuration.
type Pipeline interface {
	// Name returns the pipeline's name
	Name() string
	// Triggers returns all the triggers that can trigger the pipeline
	Triggers() []PipelineTrigger
	// Jobs returns all the jobs that comprise the pipeline
	Jobs() []PipelineJob
}

type pipeline struct {
	name     string
	triggers []PipelineTrigger
	jobs     []PipelineJob
}

func (p *pipeline) Name() string {
	return p.name
}

func (p *pipeline) Triggers() []PipelineTrigger {
	// We don't want any alterations a caller may make to the slice we return to
	// affect the pipelines's own triggers slice, which we'd like to treat as
	// immutable, so we return a COPY of that slice.
	triggers := make([]PipelineTrigger, len(p.triggers))
	copy(triggers, p.triggers)
	return triggers
}

func (p *pipeline) Jobs() []PipelineJob {
	// We don't want any alterations a caller may make to the slice we return to
	// affect the pipelines's own jobs slice, which we'd like to treat as
	// immutable, so we return a COPY of that slice.
	jobs := make([]PipelineJob, len(p.jobs))
	copy(jobs, p.jobs)
	return jobs
}
