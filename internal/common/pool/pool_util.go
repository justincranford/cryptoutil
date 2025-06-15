package pool

func CloseAllNotNil[T any](keyGenPools []*ValueGenPool[T]) {
	for _, keyGenPool := range keyGenPools {
		CloseNotNil(keyGenPool)
	}
}

func CloseNotNil[T any](keyGenPool *ValueGenPool[T]) {
	if keyGenPool != nil {
		keyGenPool.Cancel()
	}
}
