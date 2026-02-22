// Copyright (c) 2025 Justin Cranford
//
//

package poll_test

import (
"context"
"errors"
"testing"
"time"

cryptoutilSharedUtilPoll "cryptoutil/internal/shared/util/poll"

"github.com/stretchr/testify/require"
)

func TestUntil(t *testing.T) {
t.Parallel()

tests := []struct {
name        string
timeout     time.Duration
interval    time.Duration
condFn      func() cryptoutilSharedUtilPoll.ConditionFunc
ctxFn       func() context.Context
wantErr     bool
errContains string
}{
{
name:     "immediate success",
timeout:  1 * time.Second,
interval: 10 * time.Millisecond,
condFn: func() cryptoutilSharedUtilPoll.ConditionFunc {
return func(_ context.Context) (bool, error) {
return true, nil
}
},
ctxFn:   func() context.Context { return context.Background() },
wantErr: false,
},
{
name:     "success after retries",
timeout:  1 * time.Second,
interval: 10 * time.Millisecond,
condFn: func() cryptoutilSharedUtilPoll.ConditionFunc {
callCount := 0
return func(_ context.Context) (bool, error) {
callCount++
return callCount >= 3, nil
}
},
ctxFn:   func() context.Context { return context.Background() },
wantErr: false,
},
{
name:     "timeout exceeded",
timeout:  50 * time.Millisecond,
interval: 20 * time.Millisecond,
condFn: func() cryptoutilSharedUtilPoll.ConditionFunc {
return func(_ context.Context) (bool, error) {
return false, nil
}
},
ctxFn:       func() context.Context { return context.Background() },
wantErr:     true,
errContains: "poll timed out",
},
{
name:     "fatal error stops polling",
timeout:  1 * time.Second,
interval: 10 * time.Millisecond,
condFn: func() cryptoutilSharedUtilPoll.ConditionFunc {
return func(_ context.Context) (bool, error) {
return false, errors.New("fatal check error")
}
},
ctxFn:       func() context.Context { return context.Background() },
wantErr:     true,
errContains: "poll condition failed",
},
{
name:     "context canceled",
timeout:  5 * time.Second,
interval: 10 * time.Millisecond,
condFn: func() cryptoutilSharedUtilPoll.ConditionFunc {
return func(_ context.Context) (bool, error) {
return false, nil
}
},
ctxFn: func() context.Context {
ctx, cancel := context.WithCancel(context.Background())
cancel()
return ctx
},
wantErr:     true,
errContains: "poll canceled",
},
}

for _, tc := range tests {
t.Run(tc.name, func(t *testing.T) {
t.Parallel()

err := cryptoutilSharedUtilPoll.Until(tc.ctxFn(), tc.timeout, tc.interval, tc.condFn())

if tc.wantErr {
require.Error(t, err)
require.Contains(t, err.Error(), tc.errContains)
} else {
require.NoError(t, err)
}
})
}
}
