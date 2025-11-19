// Copyright (c) 2025 Justin Cranford
//
//

package apperr

import "fmt"

// RequireNoError If non-nil error, panic with value "%s: %v" using provided message and error.
func RequireNoError(err error, message string) {
	if err != nil {
		panic(fmt.Sprintf("%s: %v", message, err))
	}
}
