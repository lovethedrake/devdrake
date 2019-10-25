package config

type SourceMountMode string

const (
	SourceMountModeReadOnly  SourceMountMode = "RO"
	SourceMountModeCopy      SourceMountMode = "COPY"
	SourceMountModeReadWrite SourceMountMode = "RW"
)

// Job is a public interface for job configuration.
type Job interface {
	// Name returns the job's name
	Name() string
	// Containers returns this job's containers
	Containers() []Container
	// Returns the job's SourceMountMode
	SourceMountMode() SourceMountMode
}

type job struct {
	name            string
	containers      []Container
	sourceMountMode SourceMountMode
}

func (j *job) Name() string {
	return j.name
}

func (j *job) Containers() []Container {
	// We don't want any alterations a caller may make to the slice we return to
	// affect the job's own containers slice, which we'd like to treat as
	// immutable, so we return a COPY of that slice.
	containers := make([]Container, len(j.containers))
	copy(containers, j.containers)
	return containers
}

func (j *job) SourceMountMode() SourceMountMode {
	return j.sourceMountMode
}
