// Copyright (c) 2025 Justin Cranford
//
//

package pool

func CancelAllNotNil[T any](keyGenPools []*ValueGenPool[T]) {
	for _, keyGenPool := range keyGenPools {
		CancelNotNil(keyGenPool)
	}
}

func CancelNotNil[T any](keyGenPool *ValueGenPool[T]) {
	if keyGenPool != nil {
		keyGenPool.Cancel()
	}
}
