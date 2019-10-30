package config

import (
	"strings"
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
			name: "undefined specUri field",
			yamlBytes: []byte(`
specVersion: v0.1.0
jobs:
  foo:
    containers:
    - name: demo
      image: debian:stretch
      command: ["echo"]
      args: ["foo"]`,
			),
			assertions: func(t *testing.T, _ Config, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "specUri is required")
			},
		},

		{
			name: "unsupported spec in specUri field",
			yamlBytes: []byte(`
specUri: github.com/lovethedrake/bogus-spec
specVersion: v0.1.0
jobs:
  foo:
    containers:
    - name: demo
      image: debian:stretch
      command: ["echo"]
      args: ["foo"]`,
			),
			assertions: func(t *testing.T, _ Config, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "specUri must be one of the following")
				require.Contains(t, err.Error(), "github.com/lovethedrake/drakespec")
			},
		},

		{
			name: "undefined specVersion field",
			yamlBytes: []byte(`
specUri: github.com/lovethedrake/drakespec
jobs:
  foo:
    containers:
    - name: demo
      image: debian:stretch
      command: ["echo"]
      args: ["foo"]`,
			),
			assertions: func(t *testing.T, _ Config, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "specVersion is required")
			},
		},

		{
			name: "invalid specVersion field",
			yamlBytes: []byte(`
specUri: github.com/lovethedrake/drakespec
specVersion: bogus
jobs:
  foo:
    containers:
    - name: demo
      image: debian:stretch
      command: ["echo"]
      args: ["foo"]`,
			),
			assertions: func(t *testing.T, _ Config, err error) {
				require.Error(t, err)
				require.Contains(
					t,
					err.Error(),
					"specVersion must be one of the following",
				)
				require.Contains(t, err.Error(), "v0.1.0")
			},
		},

		// TODO: krancour: We assume all pre-GA / pre-v1.0.0 revisions of the
		// DrakeSpec may contain breaking changes. As such, the JSON schema we're
		// using to validate configuration currently enumerates ONE specific
		// revision of the DrakeSpec as permissible in the specVersion field. When
		// this changes in the future, the following chunk of commented code will be
		// relevant again for testing how DrakeCore handles configuration that
		// claims compliance with an unsupported revision of the DrakeSpec.
		// 		{
		// 			name: "unsupported specVersion",
		// 			yamlBytes: []byte(`
		// specUri: github.com/lovethedrake/drakespec
		// specVersion: v1.0.0`,
		// 			),
		// 			assertions: func(t *testing.T, _ Config, err error) {
		// 				require.Error(t, err)
		// 				require.Contains(t, err.Error(), "v1.0.0")
		// 				require.Contains(t, err.Error(), "is not a supported version")
		// 			},
		// 		},

		{
			name: "undefined job",
			yamlBytes: []byte(`
specUri: github.com/lovethedrake/drakespec
specVersion: v0.1.0
jobs:
  bar:
    containers:
    - name: demo
      image: debian:stretch
      command: ["echo"]
      args: ["bar"]
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
specUri: github.com/lovethedrake/drakespec
specVersion: v0.1.0
jobs:
  bar:
    containers:
    - name: demo
      image: debian:stretch
      command: ["echo"]
      args: ["bar"]
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
			name: "pipeline references job with sourceMountMode RW",
			yamlBytes: []byte(`
specUri: github.com/lovethedrake/drakespec
specVersion: v0.1.0
jobs:
  foo:
    containers:
    - name: demo
      image: debian:stretch
      command: ["echo"]
      args: ["foo"]
    sourceMountMode: RW
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
					`pipeline "foobar" illegally references job "foo" with `+
						`sourceMountMode "RW"`,
				)
			},
		},

		{
			name: "job dependency does not precede job in pipeline",
			yamlBytes: []byte(`
specUri: github.com/lovethedrake/drakespec
specVersion: v0.1.0
snippets:
  baseDemoContainer: &baseDemoContainer
    name: demo
    image: debian:stretch
jobs:
  foo:
    containers:
    - <<: *baseDemoContainer
      command: ["echo"]
      args: ["foo"]
  bar:
    containers:
    - <<: *baseDemoContainer
      command: ["echo"]
      args: ["bar"]
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
specUri: github.com/lovethedrake/drakespec
specVersion: v0.1.0
jobs:
  foo:
    containers:
    - name: demo
      image: debian:stretch
      command: ["echo"]
      args: ["foo"]
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
specUri: github.com/lovethedrake/drakespec
specVersion: v0.1.0
jobs:
  foo:
    containers:
    - name: demo
      image: debian:stretch
      command: ["echo"]
      args: ["foo"]
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
specUri: github.com/lovethedrake/drakespec
specVersion: v0.1.0
snippets:
  baseDemoContainer: &baseDemoContainer
    name: demo
    image: debian:stretch
jobs:
  foo:
    containers:
    - <<: *baseDemoContainer
      command: ["echo"]
      args: ["foo"]
  bar:
    containers:
    - <<: *baseDemoContainer
      command: ["echo"]
      args: ["bar"]
    sourceMountMode: COPY
pipelines:
  foobar:
    triggers:
    - specUri: github.com/lovethedrake/drakespec-github
      specVersion: v1.0.0
      config:
        branches:
          only:
          - /.*/
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
				// Check that job "foo" has the correct default sourceMountMode (RO)
				require.Equal(t, SourceMountModeReadOnly, jobs[0].SourceMountMode())
				// Check that job "bar" overrides the default sourceMountMode
				require.Equal(t, SourceMountModeCopy, jobs[1].SourceMountMode())

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
				require.Len(t, foobarPipeline.Jobs(), 2)
				require.Equal(t, fooJob, foobarPipeline.Jobs()[0].Job())
				require.Empty(t, foobarPipeline.Jobs()[0].Dependencies())
				require.Equal(t, barJob, foobarPipeline.Jobs()[1].Job())
				require.Len(t, foobarPipeline.Jobs()[1].Dependencies(), 1)
				require.Equal(
					t,
					fooJob,
					foobarPipeline.Jobs()[1].Dependencies()[0].Job(),
				)

				barfooPipeline := pipelines[1]
				require.Len(t, barfooPipeline.Jobs(), 2)
				require.Equal(t, barJob, barfooPipeline.Jobs()[0].Job())
				require.Empty(t, barfooPipeline.Jobs()[0].Dependencies())
				require.Equal(t, fooJob, barfooPipeline.Jobs()[1].Job())
				require.Len(t, barfooPipeline.Jobs()[1].Dependencies(), 1)
				require.Equal(
					t,
					barJob,
					barfooPipeline.Jobs()[1].Dependencies()[0].Job(),
				)

				// Check that we got our triggers for the foobar pipeline
				require.Len(t, foobarPipeline.Triggers(), 1)
				trigger := foobarPipeline.Triggers()[0]
				require.Equal(
					t,
					"github.com/lovethedrake/drakespec-github",
					trigger.SpecURI(),
				)
				require.Equal(
					t,
					"v1.0.0",
					trigger.SpecVersion(),
				)
				require.NotEmpty(t, trigger.Config())
				require.True(
					t,
					strings.HasPrefix(string(trigger.Config()), "{"),
				)
				require.True(
					t,
					strings.HasSuffix(string(trigger.Config()), "}"),
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
				require.Len(t, jobs, 1)
				require.Equal(t, "foo", jobs[0].Name())
			},
		},

		{
			name:        "get multiple jobs",
			jobsToFetch: []string{"foo", "bar"},
			assertions: func(t *testing.T, jobs []Job, err error) {
				require.NoError(t, err)
				require.Len(t, jobs, 2)
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
				require.Len(t, pipelines, 1)
				require.Equal(t, "foo", pipelines[0].Name())
			},
		},

		{
			name:             "get multiple pipelines",
			pipelinesToFetch: []string{"foo", "bar"},
			assertions: func(t *testing.T, pipelines []Pipeline, err error) {
				require.NoError(t, err)
				require.Len(t, pipelines, 2)
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
