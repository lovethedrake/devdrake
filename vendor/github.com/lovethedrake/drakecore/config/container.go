package config

// Container is a public interface for container configuration.
type Container interface {
	// Name returns the container's name
	Name() string
	// Image returns the name of the OCI image used by the container
	Image() string
	// Environment returns container-specific environment variables
	Environment() []string
	// WorkingDirectory returns the container's working directory
	WorkingDirectory() string
	// Command returns the command that should be run in the container
	Command() string
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
}

type container struct {
	ContainerName           string   `json:"name"`
	Img                     string   `json:"image"`
	Env                     []string `json:"environment"`
	WorkDir                 string   `json:"workingDirectory"`
	Cmd                     string   `json:"command"`
	IsTTY                   bool     `json:"tty"`
	IsPrivileged            bool     `json:"privileged"`
	ShouldMountDockerSocket bool     `json:"mountDockerSocket"`
	SrcMountPath            string   `json:"sourceMountPath"`
}

func (c *container) Name() string {
	return c.ContainerName
}

func (c *container) Image() string {
	return c.Img
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

func (c *container) Command() string {
	return c.Cmd
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
