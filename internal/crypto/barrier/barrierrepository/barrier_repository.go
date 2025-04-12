package barrierrepository

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	cryptoutilJose "cryptoutil/internal/crypto/jose"
	cryptoutilTelemetry "cryptoutil/internal/telemetry"

	googleUuid "github.com/google/uuid"
	lru "github.com/hashicorp/golang-lru"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"

	"go.opentelemetry.io/otel/attribute"
	metricApi "go.opentelemetry.io/otel/metric"
	traceApi "go.opentelemetry.io/otel/trace"
)

type Repository struct {
	name             string
	telemetryService *cryptoutilTelemetry.Service
	cacheSize        int
	cache            *lru.Cache
	latestJwk        joseJwk.Key
	mu               sync.RWMutex
	loadLatestFunc   func() (joseJwk.Key, error)
	loadFunc         func(kid googleUuid.UUID) (joseJwk.Key, error)
	storeFunc        func(jwk joseJwk.Key) error
	deleteFunc       func(kid googleUuid.UUID) error
	observations     Observations
}

type Observations struct {
	tracer                 traceApi.Tracer
	meter                  metricApi.Meter
	histogramWaitGetLatest metricApi.Int64Histogram
	histogramWaitGet       metricApi.Int64Histogram
	histogramWaitPut       metricApi.Int64Histogram
	histogramWaitRemove    metricApi.Int64Histogram
	histogramWaitPurge     metricApi.Int64Histogram
}

func New(name string, telemetryService *cryptoutilTelemetry.Service, cacheSize int, loadLatestFunc func() (joseJwk.Key, error), loadFunc func(kid googleUuid.UUID) (joseJwk.Key, error), storeFunc func(jwk joseJwk.Key) error, removeFunc func(kid googleUuid.UUID) (joseJwk.Key, error)) (*Repository, error) {
	repositoryNameAttribute := attribute.String("repository.name", name)

	tracer := telemetryService.TracesProvider.Tracer("barrierrepository", traceApi.WithInstrumentationAttributes(repositoryNameAttribute))
	_, span := tracer.Start(context.Background(), "NewJWKCache")
	defer span.End()

	meter := telemetryService.MetricsProvider.Meter("barrierrepository", metricApi.WithInstrumentationAttributes(repositoryNameAttribute))

	cache, err := lru.New(cacheSize)
	if err != nil {
		return nil, fmt.Errorf("failed to create LRU cache: %w", err)
	}

	histogramWaitGetLatest, err1 := meter.Int64Histogram("cache.request.getlatest")
	histogramWaitGet, err2 := meter.Int64Histogram("cache.request.get")
	histogramWaitPut, err3 := meter.Int64Histogram("cache.request.put")
	histogramWaitRemove, err4 := meter.Int64Histogram("cache.request.remove")
	histogramWaitPurge, err5 := meter.Int64Histogram("cache.request.purge")
	if err1 != nil || err2 != nil || err3 != nil || err4 != nil || err5 != nil {
		return nil, fmt.Errorf("failed to create Int64Histograms: %w", errors.Join(err1, err2, err3, err4, err5))
	}

	jwkCache := Repository{
		name:             name,
		telemetryService: telemetryService,
		cacheSize:        cacheSize,
		cache:            cache,
		loadLatestFunc:   loadLatestFunc,
		loadFunc:         loadFunc,
		storeFunc:        storeFunc,
		observations: Observations{
			tracer:                 tracer,
			meter:                  meter,
			histogramWaitGetLatest: histogramWaitGetLatest,
			histogramWaitGet:       histogramWaitGet,
			histogramWaitPut:       histogramWaitPut,
			histogramWaitRemove:    histogramWaitRemove,
			histogramWaitPurge:     histogramWaitPurge,
		},
	}
	return &jwkCache, nil
}

func (jwkCache *Repository) Shutdown() error {
	_, span := jwkCache.observations.tracer.Start(context.Background(), "Shutdown")
	defer span.End()

	return jwkCache.Purge()
}

func (jwkCache *Repository) GetLatest() (joseJwk.Key, error) {
	ctx, span := jwkCache.observations.tracer.Start(context.Background(), "GetLatest")
	defer span.End()

	waitStart := time.Now().UTC()
	jwkCache.mu.RLock()
	jwkCache.observations.histogramWaitGetLatest.Record(ctx, int64(time.Now().UTC().Sub(waitStart)))
	defer jwkCache.mu.RUnlock()

	if jwkCache.latestJwk != nil {
		return jwkCache.latestJwk, nil
	}
	latestJwk, err := jwkCache.loadLatestFunc() // get from database
	if err != nil {
		return nil, fmt.Errorf("failed to load latest from database: %w", err)
	} else if latestJwk == nil {
		return nil, nil
	}
	kidUuid, err := cryptoutilJose.ExtractKidUuid(latestJwk)
	if err != nil {
		return nil, fmt.Errorf("failed to extract kid uuid: %w", err)
	}
	jwkCache.latestJwk = latestJwk
	jwkCache.cache.Add(kidUuid, latestJwk)

	return latestJwk, nil
}

func (jwkCache *Repository) Get(kid googleUuid.UUID) (joseJwk.Key, error) {
	ctx, span := jwkCache.observations.tracer.Start(context.Background(), "Get")
	defer span.End()

	if kid == googleUuid.Nil { // guard against zero time
		return nil, fmt.Errorf("get nil key not supported")
	} else if kid == googleUuid.Max { // guard against max time
		return nil, fmt.Errorf("get max key not supported")
	}
	waitStart := time.Now().UTC()
	jwkCache.mu.Lock()
	jwkCache.observations.histogramWaitGet.Record(ctx, int64(time.Now().UTC().Sub(waitStart)))
	defer jwkCache.mu.Unlock()
	cachedJwk, ok := jwkCache.cache.Get(kid) // Get from LRU cache
	if !ok {
		var err error
		databaseJwk, err := jwkCache.loadFunc(kid)
		if err != nil {
			return nil, fmt.Errorf("failed to load from database: %w", err)
		}

		if jwkCache.latestJwk == nil {
			// no latestJwk in memory, so database value is assumed to be the latest
			jwkCache.latestJwk = databaseJwk
		} else {
			// update latestJwk if retrieved value is newer
			latestKidUuid, err := cryptoutilJose.ExtractKidUuid(jwkCache.latestJwk)
			if err != nil {
				return nil, fmt.Errorf("failed to extract kid uuid: %w", err)
			}
			if kid.Time() > latestKidUuid.Time() {
				jwkCache.latestJwk = databaseJwk
			}
		}

		jwkCache.cache.Add(kid, databaseJwk)
		return databaseJwk, nil
	}
	castedJwk, ok := cachedJwk.(joseJwk.Key)
	if !ok {
		return nil, fmt.Errorf("type assertion to joseJwk.Key failed")
	}
	return castedJwk, nil
}

func (jwkCache *Repository) Put(jwk joseJwk.Key) error {
	ctx, span := jwkCache.observations.tracer.Start(context.Background(), "Put")
	defer span.End()

	jwkKid, err := cryptoutilJose.ExtractKidUuid(jwk)
	if err != nil {
		return fmt.Errorf("failed to get jwk kid: %w", err)
	}

	waitStart := time.Now().UTC()
	jwkCache.mu.Lock()
	jwkCache.observations.histogramWaitPut.Record(ctx, int64(time.Now().UTC().Sub(waitStart)))
	defer jwkCache.mu.Unlock()
	err = jwkCache.storeFunc(jwk)
	if err != nil {
		return fmt.Errorf("failed to put key in database: %w", err)
	}

	if jwkCache.latestJwk == nil {
		jwkCache.latestJwk = jwk
	} else {
		// update latestJwk if added value is newer
		latestKidUuid, err := cryptoutilJose.ExtractKidUuid(jwkCache.latestJwk)
		if err != nil {
			return fmt.Errorf("failed to extract kid uuid: %w", err)
		}
		if jwkKid.Time() > latestKidUuid.Time() {
			jwkCache.latestJwk = jwk
		}
	}

	jwkCache.cache.Add(jwkKid, jwk)
	return nil
}

func (jwkCache *Repository) Remove(kidUuid googleUuid.UUID) error {
	ctx, span := jwkCache.observations.tracer.Start(context.Background(), "Remove")
	defer span.End()

	waitStart := time.Now().UTC()
	jwkCache.mu.Lock()
	jwkCache.observations.histogramWaitRemove.Record(ctx, int64(time.Now().UTC().Sub(waitStart)))
	defer jwkCache.mu.Unlock()

	latestKidUuid, err := cryptoutilJose.ExtractKidUuid(jwkCache.latestJwk)
	if err != nil {
		return fmt.Errorf("failed to extract kid uuid: %w", err)
	}

	err = jwkCache.deleteFunc(kidUuid)
	if err != nil {
		return fmt.Errorf("failed to delete jwk: %w", err)
	}
	jwkCache.cache.Remove(kidUuid)

	if latestKidUuid == kidUuid {
		jwkCache.latestJwk = nil

		// try loading next latest from database
		latest, err := jwkCache.loadLatestFunc()
		if err != nil {
			// there are no entries remaining in the DB, so latest doesn't needs updating in memory
			return nil
		}
		latestKidUuid, err = cryptoutilJose.ExtractKidUuid(latest)
		if err != nil {
			return fmt.Errorf("failed to extract kid uuid: %w", err)
		}
		jwkCache.cache.Add(latestKidUuid, latest)
		jwkCache.latestJwk = latest
	}

	return nil
}

func (jwkCache *Repository) Purge() error {
	ctx, span := jwkCache.observations.tracer.Start(context.Background(), "Purge")
	defer span.End()

	waitStart := time.Now().UTC()
	jwkCache.mu.Lock()
	jwkCache.observations.histogramWaitPurge.Record(ctx, int64(time.Now().UTC().Sub(waitStart)))
	defer jwkCache.mu.Unlock()

	newCache, err := lru.New(jwkCache.cacheSize)
	if err != nil {
		return fmt.Errorf("failed to purge cache: %w", err)
	}
	jwkCache.cache = newCache
	jwkCache.latestJwk = nil
	return nil
}
