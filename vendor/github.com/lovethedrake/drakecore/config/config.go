package config

import (
	"encoding/json"
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
	Jobs(jobNames ...string) ([]Job, error)
	// AllPipelines returns a list of all Pipelines
	AllPipelines() []Pipeline
	// Pipelines returns an ordered list of Pipelines given the provided
	// pipelineNames.
	Pipelines(pipelineNames ...string) ([]Pipeline, error)
}

// config represents the root of the Drake configuration tree.
type config struct {
	jobs            []Job
	jobsByName      map[string]Job
	pipelines       []Pipeline
	pipelinesByName map[string]Pipeline
}

// NewConfigFromFile loads configuration from the specified path and returns it.
func NewConfigFromFile(configFilePath string) (Config, error) {
	configFileBytes, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return nil,
			errors.Wrapf(err, "error reading config file %s", configFilePath)
	}
	cfg, err := NewConfigFromYAML(configFileBytes)
	if err != nil {
		return nil,
			errors.Wrapf(err, "error unmarshalling config file %s", configFilePath)
	}
	return cfg, nil
}

// NewConfigFromYAML loads configuration from the specified YAML bytes.
func NewConfigFromYAML(yamlBytes []byte) (Config, error) {
	config := &config{}
	err := yaml.Unmarshal(yamlBytes, config)
	return config, err
}

func (c *config) UnmarshalJSON(data []byte) error {
	// We have a lot of work to do to turn flat JSON into a rich object graph.
	// We'll use these "flat" one-off types to facilitate this process.
	type flatJob struct {
		Containers []*container `json:"containers"`
	}
	type flatPipelineJob struct {
		Name         string   `json:"name"`
		Dependencies []string `json:"dependencies"`
	}
	type flatPipeline struct {
		Triggers []*pipelineTrigger `json:"triggers"`
		Jobs     []*flatPipelineJob `json:"jobs"`
	}
	type flatConfig struct {
		Jobs      map[string]*flatJob      `json:"jobs"`
		Pipelines map[string]*flatPipeline `json:"pipelines"`
	}
	flatCfg := flatConfig{}
	if err := json.Unmarshal(data, &flatCfg); err != nil {
		return err
	}
	// Step through all flatJobs to populate a real job for each. While we're at
	// it, create both a slice and a map of all jobs.
	c.jobs = make([]Job, len(flatCfg.Jobs))
	c.jobsByName = map[string]Job{}
	i := 0
	for jobName, flatJob := range flatCfg.Jobs {
		job := &job{
			name:       jobName,
			containers: make([]Container, len(flatJob.Containers)),
		}
		for j, container := range flatJob.Containers {
			job.containers[j] = container
		}
		c.jobs[i] = job
		c.jobsByName[job.name] = job
		i++
	}
	// Sort the slice of all jobs lexically
	sort.Slice(
		c.jobs,
		func(a, b int) bool {
			return c.jobs[a].Name() < c.jobs[b].Name()
		},
	)
	// Step through all flatPipelines to populate a real pipeline for each. While
	// we're at it, create both a slice and a map of all pipelines.
	c.pipelines = make([]Pipeline, len(flatCfg.Pipelines))
	c.pipelinesByName = map[string]Pipeline{}
	i = 0
	for pipelineName, flatPipeline := range flatCfg.Pipelines {
		pipeline := &pipeline{
			name:     pipelineName,
			triggers: make([]PipelineTrigger, len(flatPipeline.Triggers)),
			jobs:     make([]PipelineJob, len(flatPipeline.Jobs)),
		}
		// Step through all the triggers (implementations) and add to an slice of
		// Triggers (interfaces).
		for j, trigger := range flatPipeline.Triggers {
			pipeline.triggers[j] = trigger
		}
		// Step through all flatPipelineJobs to populate a real pipelineJob for
		// each.
		pipelineJobs := map[string]PipelineJob{}
		for j, flatPipelineJob := range flatPipeline.Jobs {
			if _, ok := pipelineJobs[flatPipelineJob.Name]; ok {
				return errors.Errorf(
					"pipeline %q references the job %q more than once; this is not "+
						"permitted",
					pipeline.name,
					flatPipelineJob.Name,
				)
			}
			pipelineJob := &pipelineJob{
				dependencies: make([]PipelineJob, len(flatPipelineJob.Dependencies)),
			}
			var ok bool
			if pipelineJob.job, ok = c.jobsByName[flatPipelineJob.Name]; !ok {
				return errors.Errorf(
					"pipeline %q references undefined job %q",
					pipeline.name,
					flatPipelineJob.Name,
				)
			}
			for h, dependencyName := range flatPipelineJob.Dependencies {
				if _, ok = c.jobsByName[dependencyName]; !ok {
					return errors.Errorf(
						"job %q of pipeline %q depends on undefined job %q",
						flatPipelineJob.Name,
						pipeline.name,
						dependencyName,
					)
				}
				if pipelineJob.dependencies[h], ok = pipelineJobs[dependencyName]; !ok {
					return errors.Errorf(
						"job %q of pipeline %q depends on job %q, which is defined, but "+
							"does not precede %q in this pipeline",
						flatPipelineJob.Name,
						pipeline.name,
						dependencyName,
						flatPipelineJob.Name,
					)
				}
			}
			pipeline.jobs[j] = pipelineJob
			pipelineJobs[flatPipelineJob.Name] = pipelineJob
		}
		c.pipelines[i] = pipeline
		c.pipelinesByName[pipeline.name] = pipeline
		i++
	}
	// Sort the slice of all pipelines lexically
	sort.Slice(
		c.pipelines,
		func(a, b int) bool {
			return c.pipelines[a].Name() < c.pipelines[b].Name()
		},
	)
	return nil
}

func (c *config) AllJobs() []Job {
	// We don't want any alterations a caller may make to the slice we return to
	// affect config's own jobs slice, which we'd like to treat as immutable, so
	// we return a COPY of that slice.
	jobs := make([]Job, len(c.jobs))
	copy(jobs, c.jobs)
	return jobs
}

func (c *config) Jobs(jobNames ...string) ([]Job, error) {
	jobs := []Job{}
	for _, jobName := range jobNames {
		job, ok := c.jobsByName[jobName]
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

func (c *config) Pipelines(pipelineNames ...string) ([]Pipeline, error) {
	pipelines := []Pipeline{}
	for _, pipelineName := range pipelineNames {
		pipeline, ok := c.pipelinesByName[pipelineName]
		if !ok {
			return nil,
				errors.Errorf("pipeline \"%s\" not found", pipelineName)
		}
		pipelines = append(pipelines, pipeline)
	}
	return pipelines, nil
}
