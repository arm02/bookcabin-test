package services

import (
	"bookcabin-test/internal/core/domain"
	"bookcabin-test/internal/platform/providers"
	"log"
	"sync"
)

func (a *Aggregator) fetchInParallel(criteria domain.SearchCriteria) ([]domain.UnifiedFlight, int) {
	var wg sync.WaitGroup
	resultsChan := make(chan []domain.UnifiedFlight, len(a.Providers))
	statusChan := make(chan bool, len(a.Providers))

	for _, p := range a.Providers {
		wg.Add(1)
		go func(provider providers.ProviderInterface) {
			defer wg.Done()
			res, err := provider.Search(criteria)

			if err != nil {
				log.Printf("worker goroutine failed for %s: %v", provider.Name(), err)
				statusChan <- false
				return
			}

			statusChan <- true
			var validFlights []domain.UnifiedFlight
			for _, f := range res {
				if f.IsValid {
					validFlights = append(validFlights, f)
				}
			}
			resultsChan <- validFlights
		}(p)
	}

	wg.Wait()
	close(resultsChan)
	close(statusChan)

	var allFlights []domain.UnifiedFlight
	for res := range resultsChan {
		allFlights = append(allFlights, res...)
	}

	successCount := 0
	for s := range statusChan {
		if s {
			successCount++
		}
	}

	return allFlights, successCount
}
