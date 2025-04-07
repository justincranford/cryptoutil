package barriercache

import (
	"fmt"
	"sync"

	googleUuid "github.com/google/uuid"
	lru "github.com/hashicorp/golang-lru"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
)

type Cache struct {
	cache          *lru.Cache
	latest         Entry
	mu             sync.RWMutex
	loadLatestFunc func() (*Entry, error)
	loadFunc       func(uuid googleUuid.UUID) (joseJwk.Key, error)
	storeFunc      func(uuid googleUuid.UUID, jwk joseJwk.Key, parentUuid googleUuid.UUID) error
}

type Entry struct {
	Key   googleUuid.UUID // googleUuid.Time() only supports UUID versions 1, 2, 6, or 7
	Value joseJwk.Key     // copy by value
}

var nilEntry = Entry{Key: googleUuid.Nil, Value: nil}

func NewJWKCache(size int, loadLatestFunc func() (*Entry, error), loadFunc func(uuid googleUuid.UUID) (joseJwk.Key, error), storeFunc func(uuid googleUuid.UUID, jwk joseJwk.Key, parentUuid googleUuid.UUID) error) (*Cache, error) {
	cache, err := lru.New(size)
	if err != nil {
		return nil, fmt.Errorf("failed to create LRU cache: %w", err)
	}
	jwkCache := &Cache{cache: cache, loadLatestFunc: loadLatestFunc, loadFunc: loadFunc, storeFunc: storeFunc}
	return jwkCache, nil
}

func (jwkCache *Cache) Shutdown() {
	jwkCache.Purge()
}

func (jwkCache *Cache) GetLatest() (joseJwk.Key, error) {
	jwkCache.mu.RLock()
	defer jwkCache.mu.RUnlock()
	if jwkCache.latest.Key == googleUuid.Nil {
		latest, err := jwkCache.loadLatestFunc() // get from database
		if err != nil {
			return nil, fmt.Errorf("failed to load latest: %w", err)
		}
		jwkCache.cache.Add(jwkCache.latest.Key, jwkCache.latest.Value)
		jwkCache.latest = *latest
	}
	return jwkCache.latest.Value, nil
}

func (jwkCache *Cache) Get(key googleUuid.UUID) (joseJwk.Key, error) {
	if key == googleUuid.Nil { // guard against zero time
		return nil, fmt.Errorf("get nil key not supported")
	} else if key == googleUuid.Max { // guard against max time
		return nil, fmt.Errorf("get max key not supported")
	}
	jwkCache.mu.RLock()
	value, ok := jwkCache.cache.Get(key) // Get from LRU cache
	jwkCache.mu.RUnlock()
	if !ok {
		jwkCache.mu.Lock()
		defer jwkCache.mu.Unlock()
		var err error
		value, err = jwkCache.loadFunc(key) // get from database
		if err != nil {
			return nil, fmt.Errorf("key not found in cache or database: %w", err)
		}
		jwkCache.cache.Add(key, value)
		if key.Time() > jwkCache.latest.Key.Time() {
			jwkCache.latest.Key = key
			jwkCache.latest.Value = value.(joseJwk.Key)
		}
	}
	return value.(joseJwk.Key), nil
}

func (jwkCache *Cache) Put(key googleUuid.UUID, value joseJwk.Key, parentUuid googleUuid.UUID) error {
	if key == googleUuid.Nil { // guard against zero time
		return fmt.Errorf("put nil key not supported")
	} else if key == googleUuid.Max { // guard against max time
		return fmt.Errorf("put max key not supported")
	}
	jwkCache.mu.Lock()
	defer jwkCache.mu.Unlock()
	err := jwkCache.storeFunc(key, value, parentUuid) // put in database
	if err != nil {
		return fmt.Errorf("failed to put key in database: %w", err)
	}
	jwkCache.cache.Add(key, value)
	if key.Time() > jwkCache.latest.Key.Time() {
		jwkCache.latest.Key = key
		jwkCache.latest.Value = value
	}
	return nil
}

func (jwkCache *Cache) Remove(key googleUuid.UUID) error {
	jwkCache.mu.Lock()
	defer jwkCache.mu.Unlock()
	jwkCache.cache.Remove(key)
	if jwkCache.latest.Key == key {
		var err error
		latest, err := jwkCache.loadLatestFunc()
		if err != nil {
			jwkCache.latest = nilEntry
			return fmt.Errorf("failed to load latest: %w", err)
		}
		jwkCache.latest = *latest
		if latest.Key != googleUuid.Nil {
			jwkCache.cache.Add(jwkCache.latest.Key, jwkCache.latest.Value)
		}
	}
	return nil
}

func (jwkCache *Cache) Purge() {
	jwkCache.mu.Lock()
	defer jwkCache.mu.Unlock()
	newCache, _ := lru.New(jwkCache.cache.Len())
	jwkCache.cache = newCache
	jwkCache.latest.Key = googleUuid.Nil
	jwkCache.latest.Value = nil
}
