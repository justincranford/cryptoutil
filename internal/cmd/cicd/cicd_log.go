// Copyright (c) 2025 Justin Cranford
//
//

package cicd

import (
	"fmt"
	"os"
	"time"

	cryptoutilMagic "cryptoutil/internal/common/magic"
)

type LogUtil struct {
	startTime time.Time
}

func NewLogUtil(operation string) *LogUtil {
	start := time.Now().UTC()
	fmt.Fprintf(os.Stderr, "[CICD] start=%s\n", start.Format(cryptoutilMagic.TimeFormat))

	return &LogUtil{startTime: start}
}

func (l *LogUtil) Log(message string) {
	now := time.Now().UTC()
	fmt.Fprintf(os.Stderr, "[CICD] dur=%v now=%s: %s\n", now.Sub(l.startTime), now.Format(cryptoutilMagic.TimeFormat), message)
}
