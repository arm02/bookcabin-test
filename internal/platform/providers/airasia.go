package providers

import (
	"bookcabin-test/internal/core/domain"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type AirAsiaProvider struct{}

func (a *AirAsiaProvider) Name() string { return "AirAsia" }

func (a *AirAsiaProvider) Search(c domain.SearchCriteria) ([]domain.UnifiedFlight, error) {
	var success bool
	const maxRetries = 3
	var lastErr error

	for i := 0; i < maxRetries; i++ {
		time.Sleep(time.Duration(rand.Intn(100)+50) * time.Millisecond)
		if rand.Float32() < 0.9 {
			success = true
			break
		}
		lastErr = fmt.Errorf("AirAsia: failed attempt %d", i+1)
	}

	if !success {
		return nil, fmt.Errorf("AirAsia timed out after %d retries: %w", maxRetries, lastErr)
	}

	file := filepath.Join("mock", "airasia_search_response.json")
	rawData, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	type AAFlightRaw struct {
		Code      string  `json:"flight_code"`
		Airline   string  `json:"airline"`
		From      string  `json:"from_airport"`
		To        string  `json:"to_airport"`
		Dep       string  `json:"depart_time"`
		Arr       string  `json:"arrive_time"`
		DurHour   float64 `json:"duration_hours"`
		Direct    bool    `json:"direct_flight"`
		StopsData []struct {
			AirPort  string `json:"airport"`
			WaitTime int    `json:"wait_time_minutes"`
		} `json:"stops,omitempty"`
		Price       float64 `json:"price_idr"`
		Seats       int     `json:"seats"`
		CabinClass  string  `json:"cabin_class"`
		BaggageNote string  `json:"baggage_note"`
	}
	type AAResp struct {
		Flights []AAFlightRaw `json:"flights"`
	}

	var resp AAResp
	if err := json.Unmarshal(rawData, &resp); err != nil {
		return nil, fmt.Errorf("airasia: unmarshal error: %w", err)
	}

	var results []domain.UnifiedFlight
	for _, f := range resp.Flights {
		depTime, errDep := time.Parse(time.RFC3339, f.Dep)
		arrTime, errArr := time.Parse(time.RFC3339, f.Arr)

		if errDep != nil || errArr != nil {
			fmt.Printf("AirAsia Time Parsing Error: Dep: %v, Arr: %v\n", errDep, errArr)
			continue
		}
		isValid := arrTime.After(depTime)

		dur := CalculateDuration(depTime, arrTime)

		stops := 0
		if !f.Direct {
			stops = len(f.StopsData)
		}

		baggageInfo := domain.BaggageInfo{}
		parts := strings.Split(f.BaggageNote, ",")

		if len(parts) >= 1 {
			baggageInfo.CarryOn = strings.TrimSpace(parts[0])
		}
		if len(parts) >= 2 {
			baggageInfo.Checked = strings.TrimSpace(parts[1])
		}
		results = append(results, domain.UnifiedFlight{
			ID:       f.Code + "_QZ",
			Provider: "AirAsia",
			IsValid:  isValid,
			Airline: domain.AirlineInfo{
				Name: f.Airline,
				Code: "QZ",
			},
			FlightNumber: f.Code,
			Stops:        stops,
			Departure: domain.FlightPoint{
				Airport:   f.From,
				City:      "Jakarta",
				Datetime:  f.Dep,
				Timestamp: depTime.Unix(),
				TimeOfDay: depTime,
			},
			Arrival: domain.FlightPoint{
				Airport:   f.To,
				City:      "Denpasar",
				Datetime:  f.Arr,
				Timestamp: arrTime.Unix(),
				TimeOfDay: arrTime,
			},
			Duration: domain.DurationInfo{
				TotalMinutes: dur,
				Formatted:    fmt.Sprintf("%dh %dm", dur/60, dur%60),
			},
			Price: domain.PriceInfo{
				Amount:          f.Price,
				FormattedAmount: FormatIDR(f.Price),
				Currency:        "IDR",
			},
			AvailableSeats: f.Seats,
			CabinClass:     "economy",
			Aircraft:       "",
			Amenities:      []string{},
			Baggage:        baggageInfo,
		})
	}
	return results, nil
}
