package config

// Resources is a public interface for container resource configuration.
type Resources interface {
	// CPU returns container CPU configuration
	CPU() CPU
	// Memory returns container memory configuration
	Memory() Memory
}

type resources struct {
	// Getting creative with spelling to avoid field/func name clashes
	CPUU    *cpu    `json:"cpu"`
	Memorie *memory `json:"memory"`
}

func (r *resources) CPU() CPU {
	if r.CPUU == nil {
		return &cpu{}
	}
	return r.CPUU
}

func (r *resources) Memory() Memory {
	if r.Memorie == nil {
		return &memory{}
	}
	return r.Memorie
}

// CPU is a public interface for container CPU configuration.
type CPU interface {
	// RequestedMillicores returns the requested number of CPU millicores for use
	// by the container.
	RequestedMillicores() int
	// MaxMillicores returns the maximum number of CPU millicores usable by
	// the container.
	MaxMillicores() int
}

type cpu struct {
	RequestedMillicorez *int `json:"requestedMillicores"`
	MaxMillicorez       *int `json:"maxMillicores"`
}

func (c *cpu) RequestedMillicores() int {
	if c.RequestedMillicorez == nil {
		return 100
	}
	return *c.RequestedMillicorez
}

func (c *cpu) MaxMillicores() int {
	if c.MaxMillicorez == nil {
		return 200
	}
	return *c.MaxMillicorez
}

// Memory is a public interface for container memory configuration.
type Memory interface {
	// RequestedMegabytes returns the requested amount of memory (in megabytes)
	// for use by the container.
	RequestedMegabytes() int
	// MaxMegabytes returns the maximum amount of memory (in megabytes) usable
	// by the container.
	MaxMegabytes() int
}

type memory struct {
	RequestedMegabytez *int `json:"requestedMegabytes"`
	MaxMegabytez       *int `json:"maxMegabytes"`
}

func (m *memory) RequestedMegabytes() int {
	if m.RequestedMegabytez == nil {
		return 128
	}
	return *m.RequestedMegabytez
}

func (m *memory) MaxMegabytes() int {
	if m.MaxMegabytez == nil {
		return 256
	}
	return *m.MaxMegabytez
}
