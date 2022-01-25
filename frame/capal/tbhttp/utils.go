package tbhttp

import "strings"

func concat(substrs ...string) string {
	builder := new(strings.Builder)
	for _, substr := range substrs {
		builder.WriteString(substr)
	}
	return builder.String()
}
