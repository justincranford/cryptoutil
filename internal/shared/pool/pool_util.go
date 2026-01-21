// Copyright (c) 2025 Justin Cranford
//
//

package pool

// CancelAllNotNil cancels all non-nil pools in the provided slice.
func CancelAllNotNil[T any](keyGenPools []*ValueGenPool[T]) {
	for _, keyGenPool := range keyGenPools {
		CancelNotNil(keyGenPool)
	}
}

// CancelNotNil cancels the pool if it is not nil.
func CancelNotNil[T any](keyGenPool *ValueGenPool[T]) {
	if keyGenPool != nil {
		keyGenPool.Cancel()
	}
}
