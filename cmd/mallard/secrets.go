package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
)

var secretValFromEnvVarRegex = regexp.MustCompile(`\$\{(\w+)\}`)

func secretsFromFile(secretsFilePath string) (map[string]string, error) {
	secrets := map[string]string{}
	if _, err := os.Stat(secretsFilePath); os.IsNotExist(err) {
		return secrets, nil
	}
	secretsFile, err := os.Open(secretsFilePath)
	if err != nil {
		return nil,
			errors.Wrapf(err, "error opening secrets file %s", secretsFilePath)
	}
	defer secretsFile.Close()
	secretBytes, err := ioutil.ReadAll(secretsFile)
	if err != nil {
		return nil,
			errors.Wrapf(err, "error reading secrets file %s", secretsFilePath)
	}
	if strings.HasSuffix(secretsFilePath, ".yaml") ||
		strings.HasSuffix(secretsFilePath, ".yml") {
		if secretBytes, err = yaml.YAMLToJSON(secretBytes); err != nil {
			return secrets, errors.Wrapf(
				err,
				"error converting secrets file %s to JSON",
				secretsFilePath,
			)
		}
	}
	if err := json.Unmarshal(secretBytes, &secrets); err != nil {
		return nil,
			errors.Wrapf(err, "error parsing secrets file %s", secretsFilePath)
	}
	for k, v := range secrets {
		secrets[k] = resolveEnvVars(v)
	}
	return secrets, nil
}

func resolveEnvVars(val string) string {
	for {
		matches := secretValFromEnvVarRegex.FindStringSubmatch(val)
		if len(matches) != 2 {
			break
		}
		val = strings.Replace(
			val,
			matches[0],
			os.Getenv(matches[1]),
			1,
		)
	}
	return val
}
