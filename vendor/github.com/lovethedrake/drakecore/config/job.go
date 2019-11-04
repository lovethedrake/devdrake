package config

// SourceMountMode represents the different methods for mounting source code
// into an OCI container
type SourceMountMode string

const (
	// SourceMountModeReadOnly represents source code mounted in read-only fashion
	SourceMountModeReadOnly SourceMountMode = "RO"
	// SourceMountModeCopy represents source code mounted as a writable copy
	SourceMountModeCopy SourceMountMode = "COPY"
	// SourceMountModeReadWrite represents source code mounted in a writeable
	// fashion
	SourceMountModeReadWrite SourceMountMode = "RW"
)

// Job is a public interface for job configuration.
type Job interface {
	// Name returns the job's name
	Name() string
	// PrimaryContainer returns this job's primary container
	PrimaryContainer() Container
	// SidecarContainers returns this job's sidecar containers
	SidecarContainers() []Container
	// SourceMountMode returns the job's SourceMountMode
	SourceMountMode() SourceMountMode
}

type job struct {
	name              string
	primaryContainer  Container
	sidecarContainers []Container
	sourceMountMode   SourceMountMode
}

func (j *job) Name() string {
	return j.name
}

func (j *job) PrimaryContainer() Container {
	return j.primaryContainer
}

func (j *job) SidecarContainers() []Container {
	// We don't want any alterations a caller may make to the slice we return to
	// affect the job's own containers slice, which we'd like to treat as
	// immutable, so we return a COPY of that slice.
	sidecarContainers := make([]Container, len(j.sidecarContainers))
	copy(sidecarContainers, j.sidecarContainers)
	return sidecarContainers
}

func (j *job) SourceMountMode() SourceMountMode {
	return j.sourceMountMode
}
