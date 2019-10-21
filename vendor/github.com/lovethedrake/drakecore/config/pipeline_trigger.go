package config

import "encoding/json"

// PipelineTrigger is the public interface for a pipeline trigger.
type PipelineTrigger interface {
	SpecURI() string
	SpecVersion() string
	Config() []byte
}

type pipelineTrigger struct {
	// Hey... we've got to get creative with spelling so we don't have a field
	// and a function with the same name...
	SpeckURI     string          `json:"specUri"`
	SpeckVersion string          `json:"specVersion"`
	Cfg          json.RawMessage `json:"config"`
}

func (p *pipelineTrigger) SpecURI() string {
	return p.SpeckURI
}

func (p *pipelineTrigger) SpecVersion() string {
	return p.SpeckVersion
}

func (p *pipelineTrigger) Config() []byte {
	return p.Cfg
}
