package docker

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnvironmentMapToSlice(t *testing.T) {
	env := map[string]string{
		"FOO": "bar",
		"BAZ": "qux",
	}
	vars := environmentMapToSlice(env)
	assert.Equal(t, []string{"BAZ=qux", "FOO=bar"}, vars)
}
