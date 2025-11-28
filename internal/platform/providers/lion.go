package providers

import (
	"bookcabin-test/internal/core/domain"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"time"
)

type LionAirProvider struct{}

func (l *LionAirProvider) Name() string { return "Lion Air" }

func (l *LionAirProvider) Search(c domain.SearchCriteria) ([]domain.UnifiedFlight, error) {
	time.Sleep(time.Duration(rand.Intn(100)+100) * time.Millisecond)

	file := filepath.Join("mock", "lion_air_search_response.json")
	rawData, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	type LionFlightRaw struct {
		ID      string `json:"id"`
		Carrier struct {
			Name,
			Iata string
		} `json:"carrier"`
		Route struct {
			From, To struct {
				Code,
				Name,
				City string
			}
		} `json:"route"`
		Schedule struct {
			Departure         string `json:"departure"`
			DepartureTimezone string `json:"departure_timezone"`
			Arrival           string `json:"arrival"`
			ArrivalTimezone   string `json:"arrival_timezone"`
		} `json:"schedule"`
		StopCount  int  `json:"stop_count"`
		FlightTime int  `json:"flight_time"`
		IsDirect   bool `json:"is_direct"`
		Pricing    struct {
			Total    float64 `json:"total"`
			Currency string  `json:"currency"`
			FareType string  `json:"fare_type"`
		} `json:"pricing"`
		Seats     int    `json:"seats_left"`
		PlaneType string `json:"plane_type"`
		Services  struct {
			WifiAvailable bool `json:"wifi_available"`
			MealsIncluded bool `json:"meals_included"`
			Baggage       struct {
				Cabin,
				Hold string
			} `json:"baggage_allowance"`
		} `json:"services"`
	}
	type LionResp struct {
		Data struct {
			Flights []LionFlightRaw `json:"available_flights"`
		} `json:"data"`
	}

	var resp LionResp
	if err := json.Unmarshal(rawData, &resp); err != nil {
		return nil, fmt.Errorf("lion: unmarshal error: %w", err)
	}

	var results []domain.UnifiedFlight
	layout := "2006-01-02T15:04:05"
	for _, f := range resp.Data.Flights {
		depLoc := LocationWIB
		if f.Schedule.DepartureTimezone == "Asia/Makassar" {
			depLoc = LocationWITA
		}
		arrLoc := LocationWIB
		if f.Schedule.ArrivalTimezone == "Asia/Makassar" {
			arrLoc = LocationWITA
		}

		depTime, errDep := time.ParseInLocation(layout, f.Schedule.Departure, depLoc)
		arrTime, errArr := time.ParseInLocation(layout, f.Schedule.Arrival, arrLoc)

		if errDep != nil || errArr != nil {
			fmt.Printf("Lion Time Parsing Error: Dep: %v, Arr: %v\n", errDep, errArr)
			continue
		}

		isValid := arrTime.After(depTime)

		dur := CalculateDuration(depTime, arrTime)

		results = append(results, domain.UnifiedFlight{
			ID:       f.ID + "_JT",
			Provider: "Lion Air",
			IsValid:  isValid,
			Airline: domain.AirlineInfo{
				Name: f.Carrier.Name,
				Code: f.Carrier.Iata,
			},
			FlightNumber: f.ID,
			Stops:        f.StopCount,
			Departure: domain.FlightPoint{
				Airport:   f.Route.From.Code,
				City:      f.Route.From.City,
				Datetime:  depTime.Format(time.RFC3339),
				Timestamp: depTime.Unix(),
				TimeOfDay: depTime,
			},
			Arrival: domain.FlightPoint{
				Airport:   f.Route.To.Code,
				City:      f.Route.To.City,
				Datetime:  arrTime.Format(time.RFC3339),
				Timestamp: arrTime.Unix(),
				TimeOfDay: depTime,
			},
			Duration: domain.DurationInfo{
				TotalMinutes: dur,
				Formatted:    fmt.Sprintf("%dh %dm", dur/60, dur%60),
			},
			Price: domain.PriceInfo{
				Amount:          f.Pricing.Total,
				FormattedAmount: FormatIDR(f.Pricing.Total),
				Currency:        f.Pricing.Currency,
			},
			AvailableSeats: f.Seats, CabinClass: "economy",
			Aircraft: f.PlaneType,
			Baggage: domain.BaggageInfo{
				CarryOn: f.Services.Baggage.Cabin,
				Checked: f.Services.Baggage.Hold,
			},
		})
	}
	return results, nil
}
