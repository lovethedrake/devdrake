package config

import "encoding/json"

// PipelineTrigger is the public interface for a pipeline trigger.
type PipelineTrigger interface {
	Spec() PipelineTriggerSpec
	Config() []byte
}

type pipelineTrigger struct {
	// Hey... we've got to get creative with spelling so we don't have a field
	// and a function with the same name...
	Speck *pipelineTriggerSpec `json:"spec"`
	Cfg   json.RawMessage      `json:"config"`
}

// PipelineTriggerSpec is the public interface for a pipeline trigger spec.
type PipelineTriggerSpec interface {
	URI() string
	Version() string
}

type pipelineTriggerSpec struct {
	// Hey... we've got to get creative with spelling so we don't have a field
	// and a function with the same name...
	URII string `json:"uri"`
	Vrzn string `json:"version"`
}

func (p *pipelineTrigger) Spec() PipelineTriggerSpec {
	return p.Speck
}

func (p *pipelineTrigger) Config() []byte {
	return p.Cfg
}

func (p *pipelineTriggerSpec) URI() string {
	return p.URII
}

func (p *pipelineTriggerSpec) Version() string {
	return p.Vrzn
}
