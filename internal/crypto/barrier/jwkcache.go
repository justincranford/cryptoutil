package barrier

import (
	"fmt"
	"sync"

	"github.com/google/uuid"
	googleUuid "github.com/google/uuid"
	lru "github.com/hashicorp/golang-lru"
	"github.com/lestrrat-go/jwx/v3/jwk"
)

type JWKCache struct {
	cache          *lru.Cache
	latest         JWKCacheEntry
	mu             sync.RWMutex
	loadLatestFunc func() (*JWKCacheEntry, error)
	loadFunc       func(uuid googleUuid.UUID) (*jwk.Key, error)
	storeFunc      func(uuid googleUuid.UUID, jwk jwk.Key, parentUuid googleUuid.UUID) error
}

type JWKCacheEntry struct {
	key   uuid.UUID // uuid.Time() only supports UUID versions 1, 2, 6, or 7
	value jwk.Key   // copy by value
}

var nilEntry = JWKCacheEntry{key: uuid.Nil, value: nil}

func NewJWKCache(size int, loadLatestFunc func() (*JWKCacheEntry, error), loadFunc func(uuid googleUuid.UUID) (*jwk.Key, error), storeFunc func(uuid googleUuid.UUID, jwk jwk.Key, parentUuid googleUuid.UUID) error) (*JWKCache, error) {
	cache, err := lru.New(size)
	if err != nil {
		return nil, fmt.Errorf("failed to create LRU cache: %w", err)
	}
	jwkCache := &JWKCache{cache: cache, loadLatestFunc: loadLatestFunc, loadFunc: loadFunc, storeFunc: storeFunc}
	return jwkCache, nil
}

func (jwkCache *JWKCache) Shutdown() {
	jwkCache.Purge()
}

func (jwkCache *JWKCache) GetLatest() (jwk.Key, error) {
	jwkCache.mu.RLock()
	defer jwkCache.mu.RUnlock()
	if jwkCache.latest.key == uuid.Nil {
		latest, err := jwkCache.loadLatestFunc() // get from database
		if err != nil {
			return nil, fmt.Errorf("failed to load latest: %w", err)
		}
		jwkCache.cache.Add(jwkCache.latest.key, jwkCache.latest.value)
		jwkCache.latest = *latest
	}
	return jwkCache.latest.value, nil
}

func (jwkCache *JWKCache) Get(key uuid.UUID) (jwk.Key, error) {
	if key == uuid.Nil { // guard against zero time
		return nil, fmt.Errorf("get nil key not supported")
	} else if key == uuid.Max { // guard against max time
		return nil, fmt.Errorf("get max key not supported")
	}
	jwkCache.mu.RLock()
	value, ok := jwkCache.cache.Get(key) // Get from LRU cache
	jwkCache.mu.RUnlock()
	if !ok {
		jwkCache.mu.Lock()
		defer jwkCache.mu.Unlock()
		value, err := jwkCache.loadFunc(key) // get from database
		if err != nil {
			return nil, fmt.Errorf("key not found in cache or database: %w", err)
		}
		jwkCache.cache.Add(key, value)
		if key.Time() > jwkCache.latest.key.Time() {
			jwkCache.latest.key = key
			jwkCache.latest.value = *value
		}
	}
	return value.(jwk.Key), nil
}

func (jwkCache *JWKCache) Put(key uuid.UUID, value jwk.Key, parentUuid googleUuid.UUID) error {
	if key == uuid.Nil { // guard against zero time
		return fmt.Errorf("put nil key not supported")
	} else if key == uuid.Max { // guard against max time
		return fmt.Errorf("put max key not supported")
	}
	jwkCache.mu.Lock()
	defer jwkCache.mu.Unlock()
	err := jwkCache.storeFunc(key, value, parentUuid) // put in database
	if err != nil {
		return fmt.Errorf("failed to put key in database: %w", err)
	}
	jwkCache.cache.Add(key, value)
	if key.Time() > jwkCache.latest.key.Time() {
		jwkCache.latest.key = key
		jwkCache.latest.value = value
	}
	return nil
}

func (jwkCache *JWKCache) Remove(key uuid.UUID) error {
	jwkCache.mu.Lock()
	defer jwkCache.mu.Unlock()
	jwkCache.cache.Remove(key)
	if jwkCache.latest.key == key {
		var err error
		latest, err := jwkCache.loadLatestFunc()
		if err != nil {
			jwkCache.latest = nilEntry
			return fmt.Errorf("failed to load latest: %w", err)
		}
		jwkCache.latest = *latest
		if latest.key != uuid.Nil {
			jwkCache.cache.Add(jwkCache.latest.key, jwkCache.latest.value)
		}
	}
	return nil
}

func (jwkCache *JWKCache) Purge() {
	jwkCache.mu.Lock()
	defer jwkCache.mu.Unlock()
	newCache, _ := lru.New(jwkCache.cache.Len())
	jwkCache.cache = newCache
	jwkCache.latest.key = uuid.Nil
	jwkCache.latest.value = nil
}
