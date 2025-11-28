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

type BatikAirProvider struct{}

func (b *BatikAirProvider) Name() string { return "Batik Air" }

func (b *BatikAirProvider) Search(c domain.SearchCriteria) ([]domain.UnifiedFlight, error) {
	time.Sleep(time.Duration(rand.Intn(200)+200) * time.Millisecond)

	file := filepath.Join("mock", "batik_air_search_response.json")
	rawData, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	type BatikFlightRaw struct {
		Num        string `json:"flightNumber"`
		Name       string `json:"airlineName"`
		Iata       string `json:"airlineIATA"`
		Org        string `json:"origin"`
		Dst        string `json:"destination"`
		Dep        string `json:"departureDateTime"`
		Arr        string `json:"arrivalDateTime"`
		TravelTime string `json:"travelTime"`
		Stops      int    `json:"numberOfStops"`
		Fare       struct {
			BasePrice float64 `json:"basePrice"`
			Taxes     float64 `json:"taxes"`
			Total     float64 `json:"totalPrice"`
			Currency  string  `json:"currencyCode"`
			Class     string  `json:"class"`
		} `json:"fare"`
		Seats           int      `json:"seatsAvailable"`
		AirCraftModel   string   `json:"aircraftModel"`
		BaggageInfo     string   `json:"baggageInfo"`
		OnBoardServices []string `json:"onboardServices"`
	}
	type BatikResp struct {
		Results []BatikFlightRaw `json:"results"`
	}

	var resp BatikResp
	if err := json.Unmarshal(rawData, &resp); err != nil {
		return nil, fmt.Errorf("batik: unmarshal error: %w", err)
	}

	var results []domain.UnifiedFlight
	layout := "2006-01-02T15:04:05-0700"

	for _, f := range resp.Results {
		depTime, errDep := time.Parse(layout, f.Dep)
		arrTime, errArr := time.Parse(layout, f.Arr)

		if errDep != nil || errArr != nil {
			fmt.Printf("Batik Time Parsing Error: Dep: %v, Arr: %v\n", errDep, errArr)
			continue
		}
		isValid := arrTime.After(depTime)

		dur := CalculateDuration(depTime, arrTime)

		baggageParts := strings.Split(f.BaggageInfo, ",")
		carryOn, checked := "", ""
		if len(baggageParts) >= 1 {
			carryOn = strings.TrimSpace(baggageParts[0])
		}
		if len(baggageParts) >= 2 {
			checked = strings.TrimSpace(baggageParts[1])
		}

		results = append(results, domain.UnifiedFlight{
			ID:       f.Num + "_ID",
			Provider: "Batik Air",
			IsValid:  isValid,
			Airline: domain.AirlineInfo{
				Name: f.Name,
				Code: f.Iata,
			},
			FlightNumber: f.Num,
			Stops:        f.Stops,
			Departure: domain.FlightPoint{
				Airport:   f.Org,
				City:      "Jakarta",
				Datetime:  depTime.Format(time.RFC3339),
				Timestamp: depTime.Unix(),
				TimeOfDay: depTime,
			},
			Arrival: domain.FlightPoint{
				Airport:   f.Dst,
				City:      "Denpasar",
				Datetime:  arrTime.Format(time.RFC3339),
				Timestamp: arrTime.Unix(),
				TimeOfDay: arrTime,
			},
			Duration: domain.DurationInfo{
				TotalMinutes: dur,
				Formatted:    fmt.Sprintf("%dh %dm", dur/60, dur%60),
			},
			Price: domain.PriceInfo{
				Amount:          f.Fare.Total,
				FormattedAmount: FormatIDR(f.Fare.Total),
				Currency:        f.Fare.Currency,
			},
			AvailableSeats: f.Seats,
			CabinClass:     "economy",
			Aircraft:       f.AirCraftModel,
			Amenities:      f.OnBoardServices,
			Baggage: domain.BaggageInfo{
				CarryOn: carryOn,
				Checked: checked,
			},
		})
	}
	return results, nil
}
