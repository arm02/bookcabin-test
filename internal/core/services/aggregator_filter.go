package services

import (
	"bookcabin-test/internal/core/domain"
	"time"
)

func timeOnly(t time.Time) time.Time {
	return time.Date(0, 1, 1, t.Hour(), t.Minute(), 0, 0, time.UTC)
}

func outOfRangeInt(value int, min, max *int) bool {
	return (min != nil && value < *min) || (max != nil && value > *max)
}

func outOfRangeFloat(value float64, min, max *float64) bool {
	return (min != nil && value < *min) || (max != nil && value > *max)
}

func parseFilterTime(s *string) *time.Time {
	if s == nil {
		return nil
	}
	t, err := time.Parse(filterTimeLayout, *s)
	if err != nil {
		return nil
	}
	tOnly := timeOnly(t)
	return &tOnly
}

func filterFlights(flights []domain.UnifiedFlight, opts domain.SearchCriteria) []domain.UnifiedFlight {
	var res []domain.UnifiedFlight

	minDepTime := parseFilterTime(opts.Filters.MinDepTime)
	maxDepTime := parseFilterTime(opts.Filters.MaxDepTime)
	minArrTime := parseFilterTime(opts.Filters.MinArrTime)
	maxArrTime := parseFilterTime(opts.Filters.MaxArrTime)

	allowedAirlines := map[string]struct{}{}
	for _, name := range opts.Filters.Airlines {
		allowedAirlines[name] = struct{}{}
	}

	for _, f := range flights {
		if f.Departure.Airport != opts.Origin || f.Arrival.Airport != opts.Destination {
			continue
		}
		if opts.CabinClass != "" && f.CabinClass != opts.CabinClass {
			continue
		}
		if opts.Passengers != 0 && f.AvailableSeats < opts.Passengers {
			continue
		}
		if f.Departure.TimeOfDay.Format("2006-01-02") != opts.DepartureDate {
			continue
		}

		if outOfRangeFloat(f.Price.Amount, opts.Filters.MinPrice, opts.Filters.MaxPrice) ||
			outOfRangeInt(f.Stops, nil, opts.Filters.MaxStops) ||
			outOfRangeInt(f.Duration.TotalMinutes, opts.Filters.MinDuration, opts.Filters.MaxDuration) {
			continue
		}

		if len(allowedAirlines) > 0 {
			if _, ok := allowedAirlines[f.Airline.Name]; !ok {
				continue
			}
		}

		depMinutes := f.Departure.TimeOfDay.Hour()*60 + f.Departure.TimeOfDay.Minute()
		arrMinutes := f.Arrival.TimeOfDay.Hour()*60 + f.Arrival.TimeOfDay.Minute()

		if (minDepTime != nil && depMinutes < minDepTime.Hour()*60+minDepTime.Minute()) ||
			(maxDepTime != nil && depMinutes > maxDepTime.Hour()*60+maxDepTime.Minute()) ||
			(minArrTime != nil && arrMinutes < minArrTime.Hour()*60+minArrTime.Minute()) ||
			(maxArrTime != nil && arrMinutes > maxArrTime.Hour()*60+maxArrTime.Minute()) {
			continue
		}

		res = append(res, f)
	}

	return res
}
