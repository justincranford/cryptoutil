// Package cmdsplit tokenizes a command string into argv, honoring double-quoted
// segments so embedded spaces (e.g. python -c "import a; print(a)") survive intact.
package cmdsplit

import "strings"

// Split tokenizes cmd on whitespace, treating double-quoted spans as single tokens.
func Split(cmd string) []string {
	var (
		tokens   []string
		current  strings.Builder
		inQuotes bool
		hasToken bool
	)

	flush := func() {
		if hasToken {
			tokens = append(tokens, current.String())
			current.Reset()

			hasToken = false
		}
	}

	for _, r := range cmd {
		switch {
		case r == '"':
			inQuotes = !inQuotes
			hasToken = true
		case r == ' ' && !inQuotes:
			flush()
		default:
			current.WriteRune(r)

			hasToken = true
		}
	}

	flush()

	return tokens
}
