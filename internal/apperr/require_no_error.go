package apperr

import "fmt"

func RequireNoError(err error, message string) {
	if err != nil {
		panic(fmt.Sprintf(message+": %v", err))
	}
}
