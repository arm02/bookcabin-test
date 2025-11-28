package services

import (
	"bookcabin-test/internal/core/domain"
	"sort"
)

func calculateBestValue(flights []domain.UnifiedFlight) []domain.UnifiedFlight {
	for i := range flights {
		f := &flights[i]
		score := (f.Price.Amount / 1000.0) + (float64(f.Duration.TotalMinutes) * 5.0) + (float64(f.Stops) * 5000.0)
		f.Score = score
	}
	return flights
}

func sortFlights(flights []domain.UnifiedFlight, sortType string) []domain.UnifiedFlight {
	sort.Slice(flights, func(i, j int) bool {
		switch sortType {
		case "price_asc":
			return flights[i].Price.Amount < flights[j].Price.Amount
		case "price_desc":
			return flights[i].Price.Amount > flights[j].Price.Amount
		case "duration_asc":
			return flights[i].Duration.TotalMinutes < flights[j].Duration.TotalMinutes
		case "duration_desc":
			return flights[i].Duration.TotalMinutes > flights[j].Duration.TotalMinutes
		case "dep_time_asc":
			return flights[i].Departure.Timestamp < flights[j].Departure.Timestamp
		case "arr_time_asc":
			return flights[i].Arrival.Timestamp < flights[j].Arrival.Timestamp
		case "best_value":
			fallthrough
		default:
			return flights[i].Score < flights[j].Score
		}
	})
	return flights
}
