package docker

import (
	"bufio"
	"os"

	"github.com/pkg/errors"
)

func secretsFromFile(secretsFilePath string) ([]string, error) {
	if _, err := os.Stat(secretsFilePath); os.IsNotExist(err) {
		return []string{}, nil
	}
	secretsFile, err := os.Open(secretsFilePath)
	if err != nil {
		return nil,
			errors.Wrapf(err, "error opening secrets file %s", secretsFilePath)
	}
	secrets := []string{}
	scanner := bufio.NewScanner(secretsFile)
	for scanner.Scan() {
		secrets = append(secrets, scanner.Text())
	}
	return secrets, nil
}
