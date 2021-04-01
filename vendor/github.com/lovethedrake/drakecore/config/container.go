package config

type ImagePullPolicy string

const (
	ImagePullPolicyIfNotPresent ImagePullPolicy = "IfNotPresent"
	ImagePullPolicyAlways       ImagePullPolicy = "Always"
)

// Container is a public interface for container configuration.
type Container interface {
	// Name returns the container's name
	Name() string
	// Image returns the name of the OCI image used by the container
	Image() string
	// ImagePullPolicy returns the name of the pull policy, which indicates
	// whether to settle for existing images (if they already exist) or attempt
	// to refresh them by re-pulling
	ImagePullPolicy() ImagePullPolicy
	// Environment returns container-specific environment variables
	Environment() []string
	// WorkingDirectory returns the container's working directory
	WorkingDirectory() string
	// Command returns the command (entrypoint) that should be run in the
	// container
	Command() []string
	// Args returns the arguments for the command
	Args() []string
	// TTY returns an indicator of whether the container should use TTY or not
	TTY() bool
	// Privileged returns an indicator of whether the container should be
	// privileged
	Privileged() bool
	// MountDockerSocket returns an indicator of whether the container should
	// mount the Docker socket or not
	MountDockerSocket() bool
	// SourceMountPath returns a path to where project source should be mounted
	// into the container
	SourceMountPath() string
	// SharedStorageMountPath returns a path to where shared storage should be
	// mounted into the container
	SharedStorageMountPath() string
	// Returns the resource configuration (CPU and memory) for the container
	Resources() Resources
}

type container struct {
	ContainerName           string          `json:"name"`
	Img                     string          `json:"image"`
	ImgPullPolicy           ImagePullPolicy `json:"imagePullPolicy"`
	Env                     []string        `json:"environment"`
	WorkDir                 string          `json:"workingDirectory"`
	Cmd                     []string        `json:"command"`
	Arguments               []string        `json:"args"`
	IsTTY                   bool            `json:"tty"`
	IsPrivileged            bool            `json:"privileged"`
	ShouldMountDockerSocket bool            `json:"mountDockerSocket"`
	SrcMountPath            string          `json:"sourceMountPath"`
	SharedStrgMountPath     string          `json:"sharedStorageMountPath"`
	Resourcez               *resources      `json:"resources"`
}

func (c *container) Name() string {
	return c.ContainerName
}

func (c *container) Image() string {
	return c.Img
}

func (c *container) ImagePullPolicy() ImagePullPolicy {
	if c.ImgPullPolicy == "" {
		return ImagePullPolicyIfNotPresent
	}
	return c.ImgPullPolicy
}

func (c *container) Environment() []string {
	// We don't want any alterations a caller may make to the slice we return to
	// affect the containers's own Env slice, which we'd like to treat as
	// immutable, so we return a COPY of that slice.
	env := make([]string, len(c.Env))
	copy(env, c.Env)
	return env
}

func (c *container) WorkingDirectory() string {
	return c.WorkDir
}

func (c *container) Command() []string {
	// We don't want any alterations a caller may make to the slice we return to
	// affect the containers's own Cmd slice, which we'd like to treat as
	// immutable, so we return a COPY of that slice.
	cmd := make([]string, len(c.Cmd))
	copy(cmd, c.Cmd)
	return cmd
}

func (c *container) Args() []string {
	// We don't want any alterations a caller may make to the slice we return to
	// affect the containers's own Arguments slice, which we'd like to treat as
	// immutable, so we return a COPY of that slice.
	args := make([]string, len(c.Arguments))
	copy(args, c.Arguments)
	return args
}

func (c *container) TTY() bool {
	return c.IsTTY
}

func (c *container) Privileged() bool {
	return c.IsPrivileged
}

func (c *container) MountDockerSocket() bool {
	return c.ShouldMountDockerSocket
}

func (c *container) SourceMountPath() string {
	return c.SrcMountPath
}

func (c *container) SharedStorageMountPath() string {
	return c.SharedStrgMountPath
}

func (c *container) Resources() Resources {
	if c.Resourcez == nil {
		return &resources{}
	}
	return c.Resourcez
}
