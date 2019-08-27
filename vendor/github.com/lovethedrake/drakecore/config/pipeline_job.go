package config

import "github.com/pkg/errors"

// PipelineJob is a public interface for a job in a pipeline.
type PipelineJob interface {
	Job() Job
	Dependencies() []PipelineJob
}

// pipelineJob is a composition of a job and its dependencies within a pipeline.
type pipelineJob struct {
	Name         string   `json:"name"`
	Dependenciez []string `json:"dependencies"`
	job          *job
	dependencies []PipelineJob
}

func (p *pipelineJob) resolveJobAndDependencies(
	jobs map[string]*job,
	pipelineJobs map[string]*pipelineJob,
) error {
	var ok bool
	if p.job, ok = jobs[p.Name]; !ok {
		return errors.Errorf("job %q is undefined", p.Name)
	}
	p.dependencies = make([]PipelineJob, len(p.Dependenciez))
	for i, dependency := range p.Dependenciez {
		if p.dependencies[i], ok = pipelineJobs[dependency]; !ok {
			if _, ok := jobs[dependency]; !ok {
				return errors.Errorf(
					"job %q depends on undefined job %q",
					p.Name,
					dependency,
				)
			}
			return errors.Errorf(
				"job %q depends on job %q, which is defined, but does not precede %q "+
					"in this pipeline",
				p.Name,
				dependency,
				p.Name,
			)
		}
	}
	return nil
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
