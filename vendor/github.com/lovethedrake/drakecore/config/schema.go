package config

// nolint: lll
var jsonSchemaBytes = []byte(`
{
	"$schema": "http://json-schema.org/draft-07/schema#",
	"$id": "github.com/lovethedrake/drakecore/config.schema.json",
	
	"definitions": {

		"semanticVersion": {
      "type": "string",
			"pattern": "^v[0-9]+(\\.[0-9]+)*$"
		},

		"identifier": {
			"type": "string",
			"pattern": "^\\w[\\w-]*$"
		},

		"job": {
			"type": "object",
			"description": "A single job",
			"required": ["primaryContainer"],
			"additionalProperties": false,
			"properties": {
				"primaryContainer": {
					"allOf": [{ "$ref": "#/definitions/container" }],
					"description": "The main OCI container that implements the job"
				},
				"sidecarContainers": {
					"type": "array",
					"description": "OCI containers that play a supporting role in the job",
					"items": { "$ref": "#/definitions/container" }
				},
				"sourceMountMode": {
					"type": "string",
					"description": "The mode to use if/when mounting source code into any of the job's containers",
					"enum": [ "RO", "COPY", "RW" ]
				},
				"osFamily": {
					"type": "string",
					"description": "The OS family (linux or windows) of ALL of the job's containers",
					"enum": [ "linux", "windows" ]
				},
				"cpuArch": {
					"type": "string",
					"description": "The CPU architecture of ALL of the job's containers"
				}
			}
		},

		"container": {
			"type": "object",
			"description": "A single OCI container",
			"required": ["name", "image"],
			"additionalProperties": false,
			"properties": {
				"name": { 
					"allOf": [{ "$ref": "#/definitions/identifier" }],
					"description": "A name for the OCI container; unique to the job"
				},
				"image": {
					"type": "string",
					"description": "URL for the OCI image on which to base the container"
				},
				"imagePullPolicy": {
					"type": "string",
					"description": "Pull policy for the OCI image on which to base the container",
					"enum": [ "IfNotPresent", "Always" ]
				},
				"environment": {
					"type": "object",
					"description": "A map of key/value pairs to be exposed as environment variables within the container",
					"additionalProperties": {
						"type": "string"
					}
				},
				"workingDirectory": {
					"type": "string",
					"description": "The working directory for the container's main process"
				},
				"command": {
					"type": "array",
					"description": "Override the container's ENTRYPOINT",
					"items": {
						"type": "string"
					}
				},
				"args": {
					"type": "array",
					"description": "Override the container's CMD",
					"items": {
						"type": "string"
					}
				},
				"tty": {
					"type": "boolean",
					"description": "Whether to provision a pseudo-TTY for the container"
				},
				"privileged": {
					"type": "boolean",
					"description": "Whether to the container should be privileged"
				},
				"mountDockerSocket": {
					"type": "boolean",
					"description": "Whether to the host's Docker socket should be mounted into the container"
				},
				"sourceMountPath": {
					"type": "string",
					"description": "Where within the container to mount project source code"
				},
				"sharedStorageMountPath": {
					"type": "string",
					"description": "Where within the container to mount shared storage"
				},
				"resources": { "$ref": "#/definitions/resources" }
			}
		},

		"resources": {
			"type": "object",
			"description": "Resource constraints for a container",
			"additionalProperties": false,
			"properties": {
				"cpu": { "$ref": "#/definitions/cpu" },
				"memory": { "$ref": "#/definitions/memory" }
			}
		},

		"cpu": {
			"type": "object",
			"description": "CPU constraints for a container",
			"additionalProperties": false,
			"properties": {
				"requestedMillicores": {
					"type": "integer",
					"description": "The requested number of millicores for use by the container",
					"minimum": 1
				},
				"maxMillicores": {
					"type": "integer",
					"description": "The maximum number of millicores that may be used by the container",
					"minimum": 1
				}
			}
		},

		"memory": {
			"type": "object",
			"description": "Memory constraints for a container",
			"additionalProperties": false,
			"properties": {
				"requestedMegabytes": {
					"type": "integer",
					"description": "The requested amount of memory in megabytes for use by the container",
					"minimum": 4
				},
				"maxMegabytes": {
					"type": "integer",
					"description": "The maximum amount of memory in megabytes that may be used by the container",
					"minimum": 4
				}
			}
		},

		"pipeline": {
			"type": "object",
			"description": "A single pipeline",
			"additionalProperties": false,
			"properties": {
				"triggers": {
					"type": "array",
					"description": "The triggers that might cause the pipeline to execute",
					"items": { "$ref": "#/definitions/pipelineTrigger" }
				},
				"jobs": {
					"type": "array",
					"description": "The jobs that make up this pipeline",
					"items": { "$ref": "#/definitions/pipelineJob" }
				}
			}
		},

		"pipelineTrigger": {
			"type": "object",
			"description": "A single pipeline trigger",
			"required": ["specUri", "specVersion"],
			"additionalProperties": false,
			"properties": {
				"specUri": {
					"type": "string",
					"description": "A reference to a third-party specification with which this trigger complies"
				},
				"specVersion": { 
					"allOf": [{ "$ref": "#/definitions/semanticVersion" }],
					"description": "The revision of the third-party specification with which this trigger complies"
				},
				"config": {
					"type": "object",
					"description": "Trigger-specific configuration"
				}
			}
		},

		"pipelineJob": {
			"type": "object",
			"description": "A single element of the pipeline",
			"required": ["name"],
			"additionalProperties": false,
			"properties": {
				"name": { 
					"allOf": [{ "$ref": "#/definitions/identifier" }],
					"description": "A reference to a job from the jobs map"
				},
				"dependencies": {
					"type": "array",
					"description": "References to jobs that are prerequisites for this job",
					"items": { "$ref": "#/definitions/identifier" }
				}
			}
		}

	},

  "title": "Config",
	"type": "object",
	"required": ["specUri", "specVersion"],
	"additionalProperties": false,
  "properties": {
    "specUri": {
      "type": "string",
			"description": "A reference to the specification with which this configuration complies",
			"enum": [ "github.com/lovethedrake/drakespec" ]
		},
		"specVersion": {
			"type": "string",
			"description": "The revision of the specification with which this configuration complies",
			"enum": [ "v0.6.0" ]
		},
		"snippets": {
			"type": "object"
		},
		"jobs": {
      "type": "object",
			"description": "A map of jobs indexed by unique names",
			"additionalProperties": false,
			"patternProperties": {
				"^\\w[\\w-]*$": { "$ref": "#/definitions/job" }
			}
		},
		"pipelines": {
      "type": "object",
			"description": "A map of pipelines indexed by unique names",
			"additionalProperties": false,
			"patternProperties": {
				"^\\w[\\w-]*$": { "$ref": "#/definitions/pipeline" }
			}
    }
	}
}
`)
