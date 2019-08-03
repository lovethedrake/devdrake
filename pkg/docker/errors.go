package docker

import "fmt"

type errJobExitedNonZero struct {
	Job      string
	ExitCode int64
}

func (e *errJobExitedNonZero) Error() string {
	return fmt.Sprintf(
		`job "%s" failed with non-zero exit code %d`,
		e.Job,
		e.ExitCode,
	)
}

type multiError struct {
	errs []error
}

func (m *multiError) Error() string {
	str := fmt.Sprintf("%d errors encountered: ", len(m.errs))
	for i, err := range m.errs {
		str = fmt.Sprintf("%s\n%d. %s", str, i, err.Error())
	}
	return str
}
