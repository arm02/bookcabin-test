package services

import (
	"bookcabin-test/internal/core/domain"
	"bookcabin-test/internal/platform/providers"
	"sync"
	"time"
)

const CacheExpiration = 60 * time.Second
const filterTimeLayout = "15:04"

type CachedResponse struct {
	Flights   []domain.UnifiedFlight
	Timestamp time.Time
}

type Aggregator struct {
	Providers   []providers.ProviderInterface
	FlightCache *sync.Map
}

func NewAggregator(providers []providers.ProviderInterface) *Aggregator {
	return &Aggregator{
		Providers:   providers,
		FlightCache: &sync.Map{},
	}
}

func (a *Aggregator) SearchFlights(criteria domain.SearchCriteria) domain.SearchResponse {
	start := time.Now()
	cacheKey := criteriaHash(criteria)

	if cachedVal, ok := a.FlightCache.Load(cacheKey); ok {
		cached := cachedVal.(CachedResponse)
		if time.Since(cached.Timestamp) < CacheExpiration {
			filteredFlights := filterFlights(cached.Flights, criteria)
			scoredFlights := calculateBestValue(filteredFlights)
			sortedFlights := sortFlights(scoredFlights, criteria.SortBy)

			return domain.SearchResponse{
				SearchCriteria: criteria,
				Flights:        sortedFlights,
				Metadata: domain.ResponseMetadata{
					TotalResults:       len(sortedFlights),
					ProvidersQueried:   len(a.Providers),
					ProvidersSucceeded: len(a.Providers),
					ProvidersFailed:    0,
					SearchTimeMs:       time.Since(start).Milliseconds(),
					CacheHit:           true,
				},
			}
		} else {
			a.FlightCache.Delete(cacheKey)
		}
	}

	flights, providersSucceeded := a.fetchInParallel(criteria)

	a.FlightCache.Store(cacheKey, CachedResponse{
		Flights:   flights,
		Timestamp: time.Now(),
	})

	filteredFlights := filterFlights(flights, criteria)
	scoredFlights := calculateBestValue(filteredFlights)
	sortedFlights := sortFlights(scoredFlights, criteria.SortBy)

	providersQueried := len(a.Providers)
	providersFailed := providersQueried - providersSucceeded

	return domain.SearchResponse{
		SearchCriteria: criteria,
		Flights:        sortedFlights,
		Metadata: domain.ResponseMetadata{
			TotalResults:       len(sortedFlights),
			ProvidersQueried:   providersQueried,
			ProvidersSucceeded: providersSucceeded,
			ProvidersFailed:    providersFailed,
			SearchTimeMs:       time.Since(start).Milliseconds(),
			CacheHit:           false,
		},
	}
}
