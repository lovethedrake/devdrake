package config

import "github.com/pkg/errors"

// Pipeline is a public interface for pipeline configuration.
type Pipeline interface {
	Name() string
	Matches(branch, tag string) (bool, error)
	Jobs() [][]Job
}

type pipeline struct {
	name     string
	Selector *pipelineSelector `json:"criteria"`
	Stages   []*pipelineStage  `json:"stages"`
	jobs     [][]*job
}

type pipelineStage struct {
	Jobs []string `json:"jobs"`
}

func (p *pipeline) resolveJobs(jobs map[string]*job) error {
	p.jobs = make([][]*job, len(p.Stages))
	for i, stage := range p.Stages {
		p.jobs[i] = make([]*job, len(stage.Jobs))
		for j, jobName := range stage.Jobs {
			job, ok := jobs[jobName]
			if !ok {
				return errors.Errorf(
					"pipeline \"%s\" stage %d (zero-indexed) depends on undefined "+
						"job \"%s\"",
					p.name,
					i,
					jobName,
				)
			}
			p.jobs[i][j] = job
		}
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
	return p.Selector.Matches(branch, tag)
}

func (p *pipeline) Jobs() [][]Job {
	jobsIfaces := make([][]Job, len(p.jobs))
	for i, jobs := range p.jobs {
		jobsIfaces[i] = make([]Job, len(jobs))
		for j, job := range jobs {
			jobsIfaces[i][j] = job
		}
	}
	return jobsIfaces
}
