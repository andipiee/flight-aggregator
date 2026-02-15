package models

// SearchRequest represents the incoming search parameters[cite: 30].
type SearchRequest struct {
	Origin        string  `json:"origin"`
	Destination   string  `json:"destination"`
	DepartureDate string  `json:"departure_date"` // format: YYYY-MM-DD
	ReturnDate    *string `json:"returnDate,omitempty"`
	Passengers    string  `json:"passengers"`
	CabinClass    string  `json:"cabinClass"`

	// --- Advanced Filters ---
	MinPrice            *int     `json:"min_price,omitempty"`
	MaxPrice            *int     `json:"max_price,omitempty"`
	MinStops            *int     `json:"min_stops,omitempty"`
	MaxStops            *int     `json:"max_stops,omitempty"`
	DepartureTimeStart  *string  `json:"departure_time_start,omitempty"` // "HH:MM"
	DepartureTimeEnd    *string  `json:"departure_time_end,omitempty"`   // "HH:MM"
	ArrivalTimeStart    *string  `json:"arrival_time_start,omitempty"`   // "HH:MM"
	ArrivalTimeEnd      *string  `json:"arrival_time_end,omitempty"`     // "HH:MM"
	Airlines            []string `json:"airlines,omitempty"`
	MinDurationMinutes  *int     `json:"min_duration_minutes,omitempty"`
	MaxDurationMinutes  *int     `json:"max_duration_minutes,omitempty"`
	SortBy              *string  `json:"sort_by,omitempty"`
}

// SearchResponse matches the expected_result.json structure[cite: 50].
type SearchResponse struct {
	SearchCriteria SearchRequest `json:"search_criteria"`
	Metadata       Metadata      `json:"metadata"`
	Flights        []Flight      `json:"flights"`
}

type Metadata struct {
	TotalResults       int   `json:"total_results"`
	ProvidersQueried   int   `json:"providers_queried"`
	ProvidersSucceeded int   `json:"providers_succeeded"`
	ProvidersFailed    int   `json:"providers_failed"`
	SearchTimeMs       int64 `json:"search_time_ms"`
	CacheHit           bool  `json:"cache_hit"`
}

type Flight struct {
	ID             string    `json:"id"`
	Provider       string    `json:"provider"`
	Airline        Airline   `json:"airline"`
	FlightNumber   string    `json:"flight_number"`
	Departure      Event     `json:"departure"`
	Arrival        Event     `json:"arrival"`
	Duration       Duration  `json:"duration"`
	Stops          int       `json:"stops"`
	Price          Price     `json:"price"`
	AvailableSeats int       `json:"available_seats"`
	CabinClass     string    `json:"cabin_class"`
	Aircraft       *string   `json:"aircraft"`
	Amenities      []string  `json:"amenities"`
	Baggage        Baggage   `json:"baggage"`
}

type Airline struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

type Event struct {
	Airport   string `json:"airport"`
	City      string `json:"city"`
	Datetime  string `json:"datetime"`
	Timestamp int64  `json:"timestamp"`
}

type Duration struct {
	TotalMinutes int    `json:"total_minutes"`
	Formatted    string `json:"formatted"`
}

type Price struct {
	Amount   int    `json:"amount"`
	Currency string `json:"currency"`
}

type Baggage struct {
	CarryOn string `json:"carry_on"`
	Checked string `json:"checked"`
}