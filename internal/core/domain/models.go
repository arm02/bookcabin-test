package domain

import "time"

type SearchCriteria struct {
	Origin        string        `json:"origin"`
	Destination   string        `json:"destination"`
	DepartureDate string        `json:"departureDate"`
	ReturnDate    *string       `json:"returnDate"`
	Passengers    int           `json:"passengers"`
	CabinClass    string        `json:"cabinClass"`
	Filters       FilterOptions `json:"filters"`
	SortBy        string        `json:"sortBy"`
}

type FilterOptions struct {
	MaxPrice    *float64 `json:"maxPrice"`
	MinPrice    *float64 `json:"minPrice"`
	MaxStops    *int     `json:"maxStops"`
	Airlines    []string `json:"airlines"`
	MinDuration *int     `json:"minDurationMinutes"`
	MaxDuration *int     `json:"maxDurationMinutes"`
	MinDepTime  *string  `json:"minDepTime"`
	MaxDepTime  *string  `json:"maxDepTime"`
	MinArrTime  *string  `json:"minArrTime"`
	MaxArrTime  *string  `json:"maxArrTime"`
}

type UnifiedFlight struct {
	ID             string       `json:"id"`
	Provider       string       `json:"provider"`
	Airline        AirlineInfo  `json:"airline"`
	FlightNumber   string       `json:"flight_number"`
	Departure      FlightPoint  `json:"departure"`
	Arrival        FlightPoint  `json:"arrival"`
	Duration       DurationInfo `json:"duration"`
	Stops          int          `json:"stops"`
	Price          PriceInfo    `json:"price"`
	AvailableSeats int          `json:"available_seats"`
	CabinClass     string       `json:"cabin_class"`
	Aircraft       string       `json:"aircraft,omitempty"`
	Amenities      []string     `json:"amenities,omitempty"`
	Baggage        BaggageInfo  `json:"baggage,omitempty"`
	Score          float64      `json:"best_value_score,omitempty"`
	IsValid        bool         `json:"-"`
}

type AirlineInfo struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

type FlightPoint struct {
	Airport   string    `json:"airport"`
	City      string    `json:"city"`
	Datetime  string    `json:"datetime"`
	Timestamp int64     `json:"timestamp"`
	TimeOfDay time.Time `json:"time_of_day"`
}

type DurationInfo struct {
	TotalMinutes int    `json:"total_minutes"`
	Formatted    string `json:"formatted"`
}

type BaggageInfo struct {
	CarryOn string `json:"carry_on"`
	Checked string `json:"checked"`
}

type PriceInfo struct {
	Amount          float64 `json:"amount"`
	FormattedAmount string  `json:"formatted_amount"`
	Currency        string  `json:"currency"`
}

type SearchResponse struct {
	SearchCriteria SearchCriteria   `json:"search_criteria"`
	Metadata       ResponseMetadata `json:"metadata"`
	Flights        []UnifiedFlight  `json:"flights"`
}

type ResponseMetadata struct {
	TotalResults       int   `json:"total_results"`
	ProvidersQueried   int   `json:"providers_queried"`
	ProvidersSucceeded int   `json:"providers_succeeded"`
	ProvidersFailed    int   `json:"providers_failed"`
	SearchTimeMs       int64 `json:"search_time_ms"`
	CacheHit           bool  `json:"cache_hit"`
}
