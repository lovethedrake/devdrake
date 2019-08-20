package docker

import "fmt"

type errJobExitedNonZero struct {
	job      string
	exitCode int64
}

func (e *errJobExitedNonZero) Error() string {
	return fmt.Sprintf(
		`job "%s" failed with non-zero exit code %d`,
		e.job,
		e.exitCode,
	)
}

type multiError struct {
	errs []error
}

func (m *multiError) Error() string {
	str := fmt.Sprintf("%d errors encountered: \n", len(m.errs))
	for i, err := range m.errs {
		str = fmt.Sprintf("%s\n  %d. %s", str, i+1, err.Error())
	}
	return str
}

type errPendingJobCanceled struct {
	job string
}

func (e *errPendingJobCanceled) Error() string {
	return fmt.Sprintf("pending job %q canceled", e.job)
}

type errInProgressJobAborted struct {
	job string
}

func (e *errInProgressJobAborted) Error() string {
	return fmt.Sprintf("in-progress job %q aborted", e.job)
}
