package pool

func CloseAll[T any](keyGenPools []*ValueGenPool[T]) {
	for _, keyGenPool := range keyGenPools {
		Close(keyGenPool)
	}
}

func Close[T any](keyGenPool *ValueGenPool[T]) {
	if keyGenPool != nil {
		keyGenPool.Cancel()
	}
}
