package config

import "github.com/pkg/errors"

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
	name         string
	Selector     *pipelineSelector `json:"criteria"`
	PipelineJobs []*pipelineJob    `json:"jobs"`
}

func (p *pipeline) resolveJobs(jobs map[string]*job) error {
	pipelineJobs := map[string]*pipelineJob{}
	for _, plj := range p.PipelineJobs {
		if _, ok := pipelineJobs[plj.Name]; ok {
			return errors.Errorf(
				"pipeline %q contains the job %q more than once; this is not permitted",
				p.name,
				plj.Name,
			)
		}
		if err := plj.resolveJobAndDependencies(jobs, pipelineJobs); err != nil {
			return errors.Wrapf(err, "error resolving jobs for pipeline %q", p.name)
		}
		pipelineJobs[plj.Name] = plj
	}
	return nil
}

func (p *pipeline) Name() string {
	return p.name
}

func (p *pipeline) Matches(branch, tag string) (bool, error) {
	// If no criteria are specified, the default is to NOT match
	if p.Selector == nil {
		return false, nil
	}
	return p.Selector.matches(branch, tag)
}

func (p *pipeline) Jobs() []PipelineJob {
	pipelineJobsIfaces := make([]PipelineJob, len(p.PipelineJobs))
	for i := range p.PipelineJobs {
		pipelineJobsIfaces[i] = p.PipelineJobs[i]
	}
	return pipelineJobsIfaces
}
