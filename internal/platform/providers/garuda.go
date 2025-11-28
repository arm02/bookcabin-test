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

type GarudaProvider struct{}

func (g *GarudaProvider) Name() string { return "Garuda Indonesia" }

func (g *GarudaProvider) Search(c domain.SearchCriteria) ([]domain.UnifiedFlight, error) {
	time.Sleep(time.Duration(rand.Intn(50)+50) * time.Millisecond)

	file := filepath.Join("mock", "garuda_indonesia_search_response.json")
	rawData, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	type GarudaFlightRaw struct {
		ID      string                               `json:"flight_id"`
		Airline string                               `json:"airline"`
		Code    string                               `json:"airline_code"`
		Dep     struct{ Airport, City, Time string } `json:"departure"`
		Arr     struct{ Airport, City, Time string } `json:"arrival"`
		Stops   int                                  `json:"stops"`
		Price   struct {
			Amount   float64
			Currency string
		} `json:"price"`
		Seats    int `json:"available_seats"`
		Segments []struct {
			Dep struct{ Time string } `json:"departure"`
			Arr struct{ Time string } `json:"arrival"`
		} `json:"segments"`
		Aircraft  string   `json:"aircraft"`
		Amenities []string `json:"amenities"`
		Baggage   struct {
			CarryOn int `json:"carry_on"`
			Checked int `json:"checked"`
		} `json:"baggage"`
	}
	type GarudaResp struct {
		Flights []GarudaFlightRaw `json:"flights"`
	}

	var resp GarudaResp
	if err := json.Unmarshal(rawData, &resp); err != nil {
		return nil, fmt.Errorf("garuda: unmarshal error: %w", err)
	}

	var results []domain.UnifiedFlight
	for _, f := range resp.Flights {
		depTime, errDep := time.Parse(time.RFC3339, f.Dep.Time)
		arrTime, errArr := time.Parse(time.RFC3339, f.Arr.Time)

		if f.Stops > 0 && len(f.Segments) >= 2 {
			depTime, _ = time.Parse(time.RFC3339, f.Segments[0].Dep.Time)
			arrTime, _ = time.Parse(time.RFC3339, f.Segments[len(f.Segments)-1].Arr.Time)
		}

		if errDep != nil || errArr != nil {
			fmt.Printf("Garuda Time Parsing Error: Dep: %v, Arr: %v\n", errDep, errArr)
			continue
		}
		isValid := arrTime.After(depTime)

		dur := CalculateDuration(depTime, arrTime)

		results = append(results, domain.UnifiedFlight{
			ID:       f.ID + "_GA",
			Provider: "Garuda Indonesia",
			IsValid:  isValid,
			Airline: domain.AirlineInfo{
				Name: f.Airline,
				Code: f.Code,
			},
			FlightNumber: f.ID,
			Stops:        f.Stops,
			Departure: domain.FlightPoint{
				Airport:   f.Dep.Airport,
				City:      f.Dep.City,
				Datetime:  depTime.Format(time.RFC3339),
				Timestamp: depTime.Unix(),
				TimeOfDay: depTime,
			},
			Arrival: domain.FlightPoint{
				Airport:   f.Arr.Airport,
				City:      f.Arr.City,
				Datetime:  arrTime.Format(time.RFC3339),
				Timestamp: arrTime.Unix(),
				TimeOfDay: arrTime,
			},
			Duration: domain.DurationInfo{
				TotalMinutes: dur,
				Formatted:    fmt.Sprintf("%dh %dm", dur/60, dur%60),
			},
			Price: domain.PriceInfo{
				Amount:          f.Price.Amount,
				FormattedAmount: FormatIDR(f.Price.Amount),
				Currency:        f.Price.Currency,
			},
			AvailableSeats: f.Seats,
			CabinClass:     "economy",
			Aircraft:       f.Aircraft,
			Amenities:      f.Amenities,
			Baggage: domain.BaggageInfo{
				CarryOn: fmt.Sprintf("%d piece(s)", f.Baggage.CarryOn),
				Checked: fmt.Sprintf("%d piece(s)", f.Baggage.Checked),
			},
		})
	}
	return results, nil
}
