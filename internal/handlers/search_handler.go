package handlers

import (
	"bookcabin-test/internal/core/domain"
	"bookcabin-test/internal/core/services"
	"encoding/json"
	"log"
	"net/http"
)

type SearchHandlers struct {
	AggregatorService *services.Aggregator
}

func NewSearchHandlers(svc *services.Aggregator) *SearchHandlers {
	return &SearchHandlers{AggregatorService: svc}
}

func (s *SearchHandlers) SearchFlight(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var criteria domain.SearchCriteria
	if err := json.NewDecoder(r.Body).Decode(&criteria); err != nil {
		http.Error(w, "Bad Request: Invalid JSON or format - "+err.Error(), http.StatusBadRequest)
		return
	}

	if criteria.Origin == "" || criteria.Destination == "" || criteria.DepartureDate == "" {
		http.Error(w, "Bad Request: Origin, Destination, and DepartureDate are required.", http.StatusBadRequest)
		return
	}

	if criteria.SortBy == "" {
		criteria.SortBy = "best_value"
	}

	resp := s.AggregatorService.SearchFlights(criteria)

	if len(resp.Flights) == 0 && resp.Metadata.ProvidersFailed > 0 {
		log.Printf("Warning: %d providers failed.", resp.Metadata.ProvidersFailed)
	}

	json.NewEncoder(w).Encode(resp)
}
