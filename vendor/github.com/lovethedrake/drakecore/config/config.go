package config

import (
	"io/ioutil"
	"sort"

	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
)

// Config is a public interface for the root of the Drake configuration tree.
type Config interface {
	// AllJobs returns a list of all Jobs
	AllJobs() []Job
	// Jobs returns an ordered list of Jobs given the provided jobNames.
	Jobs(jobNames []string) ([]Job, error)
	// AllPipelines returns a list of all Pipelines
	AllPipelines() []Pipeline
	// Pipelines returns an ordered list of Pipelines given the provided
	// pipelineNames.
	Pipelines(pipelineNames []string) ([]Pipeline, error)
}

// config represents the root of the Drake configuration tree.
type config struct {
	Jobz      map[string]*job      `json:"jobs"`
	Pipelinez map[string]*pipeline `json:"pipelines"`
	jobs      []Job
	pipelines []Pipeline
}

// NewConfigFromFile loads configuration from the specified path and returns it.
func NewConfigFromFile(configFilePath string) (Config, error) {
	configFileBytes, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return nil,
			errors.Wrapf(err, "error reading config file %s", configFilePath)
	}
	config := &config{}
	if err := yaml.Unmarshal(configFileBytes, config); err != nil {
		return nil,
			errors.Wrapf(err, "error unmarshalling config file %s", configFilePath)
	}
	// Step through all jobs to add their name attribute, which is inferred
	// from the keys in the config.Jobs map. While we're at it, create a list
	// of all jobs.
	config.jobs = make([]Job, len(config.Jobz))
	i := 0
	for jobName, job := range config.Jobz {
		job.name = jobName
		config.jobs[i] = job
		i++
	}
	// Sort the list of all jobs lexically
	sort.Slice(config.jobs, func(a, b int) bool {
		return config.jobs[a].Name() < config.jobs[b].Name()
	})
	// Step through all pipelines to add their name attribute, which is inferred
	// from the keys in the config.Pipelines map. Also resolve jobs referenced
	// by each pipeline. While we're at it, create a list of all pipelines.
	config.pipelines = make([]Pipeline, len(config.Pipelinez))
	i = 0
	for pipelineName, pipeline := range config.Pipelinez {
		pipeline.name = pipelineName
		if err := pipeline.resolveJobs(config.Jobz); err != nil {
			return nil, err
		}
		config.pipelines[i] = pipeline
		i++
	}
	// Sort the list of all pipelines lexically
	sort.Slice(config.pipelines, func(a, b int) bool {
		return config.pipelines[a].Name() < config.pipelines[b].Name()
	})
	return config, nil
}

func (c *config) AllJobs() []Job {
	// We don't want any alterations a caller may make to the slice we return to
	// affect config's own jobs slice, which we'd like to treat as immutable, so
	// we return a COPY of that slice.
	jobs := make([]Job, len(c.jobs))
	copy(jobs, c.jobs)
	return jobs
}

func (c *config) Jobs(jobNames []string) ([]Job, error) {
	jobs := []Job{}
	for _, jobName := range jobNames {
		job, ok := c.Jobz[jobName]
		if !ok {
			return nil,
				errors.Errorf("job \"%s\" not found", jobName)
		}
		jobs = append(jobs, job)
	}
	return jobs, nil
}

func (c *config) AllPipelines() []Pipeline {
	// We don't want any alterations a caller may make to the slice we return to
	// affect config's own pipelines slice, which we'd like to treat as immutable,
	// so we return a COPY of that slice.
	pipelines := make([]Pipeline, len(c.pipelines))
	copy(pipelines, c.pipelines)
	return pipelines
}

func (c *config) Pipelines(pipelineNames []string) ([]Pipeline, error) {
	pipelines := []Pipeline{}
	for _, pipelineName := range pipelineNames {
		pipeline, ok := c.Pipelinez[pipelineName]
		if !ok {
			return nil,
				errors.Errorf("pipeline \"%s\" not found", pipelineName)
		}
		pipelines = append(pipelines, pipeline)
	}
	return pipelines, nil
}
