package docker

import (
	"bufio"
	"fmt"
	"io"
)

func prefixingWriter(
	jobName string,
	containerName string,
	output io.Writer,
) io.Writer {
	pipeReader, pipeWriter := io.Pipe()
	scanner := bufio.NewScanner(pipeReader)
	scanner.Split(bufio.ScanLines)
	go func() {
		for scanner.Scan() {
			fmt.Fprintf(output, "[%s-%s] ", jobName, containerName)
			output.Write(scanner.Bytes()) // nolint: errcheck
			fmt.Fprint(output, "\n")
		}
	}()
	return pipeWriter
}
