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

// OSFamily represents the supported kernels-- linux and windows
type OSFamily string

const (
	// OSFamilyLinux represents the linux family of operating systems
	OSFamilyLinux OSFamily = "linux"
	// OSFamilyWindows represents the windows family of operating systems
	OSFamilyWindows OSFamily = "windows"
)

// CPUArch represents CPU architecture
type CPUArch string

const (
	// CPUArchAMD64 represents amd64 CPU architecture
	CPUArchAMD64 CPUArch = "amd64"

	// Note that there are a lot of different CPU architectures supported by
	// OCI container runtimes and it was a conscious choice to only enumerate the
	// default here and not attempt to enumerate all of them. Users of go-drake
	// should use string literals cast as CPUArch to reference alternative
	// architectures.
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
	// OSFamily returns the job's OSFamily
	OSFamily() OSFamily
	// CPUArch returns the job's CPU architecture
	CPUArch() CPUArch
	// TimeoutSeconds returns the maximum number of seconds to wait for the job to complete.
	// An empty (zero) value disables the timeout.
	TimeoutSeconds() int64
}

var _ Job = &job{}

type job struct {
	name              string
	primaryContainer  Container
	sidecarContainers []Container
	sourceMountMode   SourceMountMode
	osFamily          OSFamily
	cpuArch           CPUArch
	timeoutSeconds    int64
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

func (j *job) OSFamily() OSFamily {
	return j.osFamily
}

func (j *job) CPUArch() CPUArch {
	return j.cpuArch
}

func (j *job) TimeoutSeconds() int64 {
	return j.timeoutSeconds
}
