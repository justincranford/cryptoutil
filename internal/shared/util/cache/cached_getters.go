// Copyright (c) 2025 Justin Cranford
//
//

// Package cache provides caching utilities for getter functions.
package cache

import "sync"

// GetCached retrieves a value using sync.Once for cached execution or direct call for non-cached.
func GetCached(cached bool, syncOnce *sync.Once, getterFunction func() any) any {
	var value any

	if cached {
		syncOnce.Do(func() {
			value = getterFunction()
		})

		return value
	}

	syncOnce.Do(func() {
		value = getterFunction()
	})

	return value
}

// GetCachedWithError retrieves a value and error using sync.Once for cached execution or direct call for non-cached.
func GetCachedWithError(cached bool, syncOnce *sync.Once, getterFunction func() (any, error)) (any, error) {
	var value any

	var err error

	if cached {
		syncOnce.Do(func() {
			value, err = getterFunction()
		})

		return value, err
	}

	syncOnce.Do(func() {
		value, err = getterFunction()
	})

	return value, err
}
