package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewConfigFromYAML(t *testing.T) {
	testCases := []struct {
		name       string
		yamlBytes  []byte
		assertions func(*testing.T, Config, error)
	}{
		{
			name: "undefined job",
			yamlBytes: []byte(`
version: v1.0.0
pipelines:
  foobar:
    jobs:
    - name: foo`,
			),
			assertions: func(t *testing.T, _ Config, err error) {
				require.Error(t, err)
				require.Contains(
					t,
					err.Error(),
					`pipeline "foobar" references undefined job "foo"`,
				)
			},
		},

		{
			name: "undefined job dependency",
			yamlBytes: []byte(`
version: v1.0.0
jobs:
  bar:
    containers:
    - name: demo
      image: debian:stretch
      command: echo bar  
pipelines:
  foobar:
    jobs:
    - name: bar
      dependencies:
      - foo`,
			),
			assertions: func(t *testing.T, _ Config, err error) {
				require.Error(t, err)
				require.Contains(
					t,
					err.Error(),
					`job "bar" of pipeline "foobar" depends on undefined job "foo"`,
				)
			},
		},

		{
			name: "job dependency does not precede job in pipeline",
			yamlBytes: []byte(`
version: v1.0.0
baseDemoContainer: &baseDemoContainer
  name: demo
  image: debian:stretch
jobs:
  foo:
    containers:
    - <<: *baseDemoContainer
      command: echo foo
  bar:
    containers:
    - <<: *baseDemoContainer
    command: echo bar
pipelines:
  foobar:
    jobs:
    - name: bar
      dependencies:
      - foo`,
			),
			assertions: func(t *testing.T, _ Config, err error) {
				require.Error(t, err)
				require.Contains(
					t,
					err.Error(),
					`job "bar" of pipeline "foobar" depends on job "foo", which is `+
						`defined, but does not precede "bar" in this pipeline`,
				)
			},
		},

		{
			name: "job depends on itself",
			yamlBytes: []byte(`
version: v1.0.0
jobs:
  foo:
    containers:
    - name: demo
      image: debian:stretch
      command: echo foo
pipelines:
  foobar:
    jobs:
    - name: foo
      dependencies:
      - foo`,
			),
			assertions: func(t *testing.T, _ Config, err error) {
				require.Error(t, err)
				require.Contains(
					t,
					err.Error(),
					`job "foo" of pipeline "foobar" depends on job "foo", which is `+
						`defined, but does not precede "foo" in this pipeline`,
				)
			},
		},

		{
			name: "job appears in pipeline more than once",
			yamlBytes: []byte(`
version: v1.0.0
jobs:
  foo:
    containers:
    - name: demo
      image: debian:stretch
      command: echo foo
pipelines:
  foobar:
    jobs:
    - name: foo
    - name: foo`,
			),
			assertions: func(t *testing.T, _ Config, err error) {
				require.Error(t, err)
				require.Contains(
					t,
					err.Error(),
					`pipeline "foobar" references the job "foo" more than once; this `+
						`is not permitted`,
				)
			},
		},

		{
			name: "valid config",
			yamlBytes: []byte(`
version: v1.0.0
baseDemoContainer: &baseDemoContainer
  name: demo
  image: debian:stretch
jobs:
  foo:
    containers:
    - <<: *baseDemoContainer
      command: echo foo
  bar:
    containers:
    - <<: *baseDemoContainer
    command: echo bar
pipelines:
  foobar:
    jobs:
    - name: foo
    - name: bar
      dependencies:
      - foo
  barfoo:
    jobs:
    - name: bar
    - name: foo
      dependencies:
      - bar`,
			),
			assertions: func(t *testing.T, cfg Config, err error) {
				require.NoError(t, err)
				require.NotNil(t, cfg)
				jobs := cfg.AllJobs()
				// We got the expected number of jobs
				require.Len(t, jobs, 2)
				// They appear in the expected order (lexical, by name)
				require.Equal(t, "bar", jobs[0].Name())
				require.Equal(t, "foo", jobs[1].Name())
				// We can retrieve jobs by name(s)
				jobs, err = cfg.Jobs("foo", "bar")
				require.NoError(t, err)
				// We got the expected number of jobs
				require.Len(t, jobs, 2)
				// In the expected order
				require.Equal(t, "foo", jobs[0].Name())
				require.Equal(t, "bar", jobs[1].Name())

				pipelines := cfg.AllPipelines()
				// We got the expected number of pipelines
				require.Len(t, pipelines, 2)
				// They appear in the expected order (lexical, by name)
				require.Equal(t, "barfoo", pipelines[0].Name())
				require.Equal(t, "foobar", pipelines[1].Name())
				// We can retrieve pipelines by name(s)
				pipelines, err = cfg.Pipelines("foobar", "barfoo")
				require.NoError(t, err)
				// We got the expected number of pipelines
				require.Len(t, pipelines, 2)
				// In the expected order
				require.Equal(t, "foobar", pipelines[0].Name())
				require.Equal(t, "barfoo", pipelines[1].Name())

				// Check that we unmarshaled from "flat" JSON/YAML into the rich object
				// graph we expect
				fooJob := jobs[0]
				barJob := jobs[1]

				foobarPipeline := pipelines[0]
				require.Equal(t, 2, len(foobarPipeline.Jobs()))
				require.Equal(t, fooJob, foobarPipeline.Jobs()[0].Job())
				require.Empty(t, foobarPipeline.Jobs()[0].Dependencies())
				require.Equal(t, barJob, foobarPipeline.Jobs()[1].Job())
				require.Equal(t, 1, len(foobarPipeline.Jobs()[1].Dependencies()))
				require.Equal(
					t,
					fooJob,
					foobarPipeline.Jobs()[1].Dependencies()[0].Job(),
				)

				barfooPipeline := pipelines[1]
				require.Equal(t, 2, len(barfooPipeline.Jobs()))
				require.Equal(t, barJob, barfooPipeline.Jobs()[0].Job())
				require.Empty(t, barfooPipeline.Jobs()[0].Dependencies())
				require.Equal(t, fooJob, barfooPipeline.Jobs()[1].Job())
				require.Equal(t, 1, len(barfooPipeline.Jobs()[1].Dependencies()))
				require.Equal(
					t,
					barJob,
					barfooPipeline.Jobs()[1].Dependencies()[0].Job(),
				)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			cfg, err := NewConfigFromYAML(testCase.yamlBytes)
			testCase.assertions(t, cfg, err)
		})
	}
}

func TestJobs(t *testing.T) {
	cfg := &config{
		jobsByName: map[string]Job{
			"foo": &job{
				name: "foo",
			},
			"bar": &job{
				name: "bar",
			},
		},
	}
	testCases := []struct {
		name        string
		jobsToFetch []string
		assertions  func(*testing.T, []Job, error)
	}{
		{
			name:        "get job that doesn't exist",
			jobsToFetch: []string{"bat"},
			assertions: func(t *testing.T, _ []Job, err error) {
				require.Error(t, err)
			},
		},

		{
			name:        "get single job",
			jobsToFetch: []string{"foo"},
			assertions: func(t *testing.T, jobs []Job, err error) {
				require.NoError(t, err)
				require.Equal(t, 1, len(jobs))
				require.Equal(t, "foo", jobs[0].Name())
			},
		},

		{
			name:        "get multiple jobs",
			jobsToFetch: []string{"foo", "bar"},
			assertions: func(t *testing.T, jobs []Job, err error) {
				require.NoError(t, err)
				require.Equal(t, 2, len(jobs))
				require.Equal(t, "foo", jobs[0].Name())
				require.Equal(t, "bar", jobs[1].Name())
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			jobs, err := cfg.Jobs(testCase.jobsToFetch...)
			testCase.assertions(t, jobs, err)
		})
	}
}

func TestPipelines(t *testing.T) {
	cfg := &config{
		pipelinesByName: map[string]Pipeline{
			"foo": &pipeline{
				name: "foo",
			},
			"bar": &pipeline{
				name: "bar",
			},
		},
	}
	testCases := []struct {
		name             string
		pipelinesToFetch []string
		assertions       func(*testing.T, []Pipeline, error)
	}{
		{
			name:             "get pipeline that doesn't exist",
			pipelinesToFetch: []string{"bat"},
			assertions: func(t *testing.T, _ []Pipeline, err error) {
				require.Error(t, err)
			},
		},

		{
			name:             "get single pipeline",
			pipelinesToFetch: []string{"foo"},
			assertions: func(t *testing.T, pipelines []Pipeline, err error) {
				require.NoError(t, err)
				require.Equal(t, 1, len(pipelines))
				require.Equal(t, "foo", pipelines[0].Name())
			},
		},

		{
			name:             "get multiple pipelines",
			pipelinesToFetch: []string{"foo", "bar"},
			assertions: func(t *testing.T, pipelines []Pipeline, err error) {
				require.NoError(t, err)
				require.Equal(t, 2, len(pipelines))
				require.Equal(t, "foo", pipelines[0].Name())
				require.Equal(t, "bar", pipelines[1].Name())
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			pipelines, err := cfg.Pipelines(testCase.pipelinesToFetch...)
			testCase.assertions(t, pipelines, err)
		})
	}
}
