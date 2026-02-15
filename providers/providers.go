package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"flight-aggregator/models"
)

// Provider interface ensures uniform calling for any airline.
type Provider interface {
	Name() string
	FetchFlights(ctx context.Context, req models.SearchRequest) ([]models.Flight, error)
}

// Base helper to simulate network latency.
func simulateDelay(ctx context.Context, minMs, maxMs int) error {
	delay := time.Duration(rand.Intn(maxMs-minMs)+minMs) * time.Millisecond
	select {
	case <-time.After(delay):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func readMockData(filename string, target interface{}) error {
	// Try local mock_data first
	path := filepath.Join("mock_data", filename)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Try parent directory mock_data
		path = filepath.Join("..", "mock_data", filename)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, target)
}

// --- GARUDA INDONESIA --- //
type GarudaProvider struct{}

func (g *GarudaProvider) Name() string { return "Garuda Indonesia" }
func (g *GarudaProvider) FetchFlights(ctx context.Context, req models.SearchRequest) ([]models.Flight, error) {
	if err := simulateDelay(ctx, 50, 100); err != nil {
		return nil, err
	}

	var mock struct {
		Flights []struct {
			FlightId string                               `json:"flight_id"`
			Airline  string                               `json:"airline"`
			AirlineC string                               `json:"airline_code"`
			Dep      struct{ Airport, City, Time string } `json:"departure"`
			Arr      struct{ Airport, City, Time string } `json:"arrival"`
			DurMins  int                                  `json:"duration_minutes"`
			Stops    int                                  `json:"stops"`
			Aircraft string                               `json:"aircraft"`
			Price    struct{ Amount int }                 `json:"price"`
			Seats    int                                  `json:"available_seats"`
			Baggage  struct {
				CarryOn int
				Checked int
			} `json:"baggage"`
		} `json:"flights"`
	}

	if err := readMockData("garuda_indonesia_search_response.json", &mock); err != nil {
		return nil, err
	}

	var results []models.Flight
	for _, f := range mock.Flights {
		depT, _ := time.Parse(time.RFC3339, f.Dep.Time)
		arrT, _ := time.Parse(time.RFC3339, f.Arr.Time)

		results = append(results, models.Flight{
			ID: fmt.Sprintf("%s_Garuda", f.FlightId), Provider: g.Name(),
			Airline:      models.Airline{Name: f.Airline, Code: f.AirlineC},
			FlightNumber: f.FlightId,
			Departure:    models.Event{Airport: f.Dep.Airport, City: f.Dep.City, Datetime: f.Dep.Time, Timestamp: depT.Unix()},
			Arrival:      models.Event{Airport: f.Arr.Airport, City: f.Arr.City, Datetime: f.Arr.Time, Timestamp: arrT.Unix()},
			Duration:     models.Duration{TotalMinutes: f.DurMins, Formatted: fmt.Sprintf("%dh %dm", f.DurMins/60, f.DurMins%60)},
			Stops:        f.Stops, AvailableSeats: f.Seats,
			Price:   models.Price{Amount: f.Price.Amount, Currency: "IDR"},
			Baggage: models.Baggage{CarryOn: fmt.Sprintf("%d piece(s)", f.Baggage.CarryOn), Checked: fmt.Sprintf("%d piece(s)", f.Baggage.Checked)},
		})
	}
	return results, nil
}

// --- AIRASIA --- //
type AirAsiaProvider struct{}

func (a *AirAsiaProvider) Name() string { return "AirAsia" }
func (a *AirAsiaProvider) FetchFlights(ctx context.Context, req models.SearchRequest) ([]models.Flight, error) {
	if err := simulateDelay(ctx, 50, 150); err != nil {
		return nil, err
	}

	// Simulate 10% failure rate
	if rand.Intn(10) == 0 {
		return nil, fmt.Errorf("AirAsia API Service Unavailable (503)")
	}

	var mock struct {
		Flights []struct {
			Code   string  `json:"flight_code"`
			Dep    string  `json:"depart_time"`
			Arr    string  `json:"arrive_time"`
			From   string  `json:"from_airport"`
			To     string  `json:"to_airport"`
			Dur    float64 `json:"duration_hours"`
			Direct bool    `json:"direct_flight"`
			Price  int     `json:"price_idr"`
			Seats  int     `json:"seats"`
			Bag    string  `json:"baggage_note"`
		} `json:"flights"`
	}

	if err := readMockData("airasia_search_response.json", &mock); err != nil {
		return nil, err
	}

	var results []models.Flight
	for _, f := range mock.Flights {
		depT, _ := time.Parse(time.RFC3339, f.Dep)
		arrT, _ := time.Parse(time.RFC3339, f.Arr)
		mins := int(f.Dur * 60)
		stops := 0
		if !f.Direct {
			stops = 1
		}

		results = append(results, models.Flight{
			ID: fmt.Sprintf("%s_AirAsia", f.Code), Provider: a.Name(),
			Airline:      models.Airline{Name: "AirAsia", Code: strings.TrimRight(f.Code, "0123456789")},
			FlightNumber: f.Code,
			Departure:    models.Event{Airport: f.From, Datetime: f.Dep, Timestamp: depT.Unix()},
			Arrival:      models.Event{Airport: f.To, Datetime: f.Arr, Timestamp: arrT.Unix()},
			Duration:     models.Duration{TotalMinutes: mins, Formatted: fmt.Sprintf("%dh %dm", mins/60, mins%60)},
			Stops:        stops, Price: models.Price{Amount: f.Price, Currency: "IDR"}, AvailableSeats: f.Seats,
			Baggage: models.Baggage{CarryOn: "Included", Checked: f.Bag},
		})
	}
	return results, nil
}

// --- Batik Air --- //
type BatikAirProvider struct{}

func (b *BatikAirProvider) Name() string { return "Batik Air" }
func (b *BatikAirProvider) FetchFlights(ctx context.Context, req models.SearchRequest) ([]models.Flight, error) {
	if err := simulateDelay(ctx, 200, 400); err != nil {
		return nil, err
	}

	var mock struct {
		Flights []struct {
			FlightId string                               `json:"flight_id"`
			Airline  string                               `json:"airline"`
			AirlineC string                               `json:"airline_code"`
			Dep      struct{ Airport, City, Time string } `json:"departure"`
			Arr      struct{ Airport, City, Time string } `json:"arrival"`
			DurMins  int                                  `json:"duration_minutes"`
			Stops    int                                  `json:"stops"`
			Aircraft string                               `json:"aircraft"`
			Price    struct{ Amount int }                 `json:"price"`
			Seats    int                                  `json:"available_seats"`
			Baggage  struct {
				CarryOn int
				Checked int
			} `json:"baggage"`
		} `json:"flights"`
	}

	if err := readMockData("batik_air_search_response.json", &mock); err != nil {
		return nil, err
	}

	var results []models.Flight
	for _, f := range mock.Flights {
		depT, _ := time.Parse(time.RFC3339, f.Dep.Time)
		arrT, _ := time.Parse(time.RFC3339, f.Arr.Time)

		results = append(results, models.Flight{
			ID: fmt.Sprintf("%s_Batik", f.FlightId), Provider: b.Name(),
			Airline:      models.Airline{Name: f.Airline, Code: f.AirlineC},
			FlightNumber: f.FlightId,
			Departure:    models.Event{Airport: f.Dep.Airport, City: f.Dep.City, Datetime: f.Dep.Time, Timestamp: depT.Unix()},
			Arrival:      models.Event{Airport: f.Arr.Airport, City: f.Arr.City, Datetime: f.Arr.Time, Timestamp: arrT.Unix()},
			Duration:     models.Duration{TotalMinutes: f.DurMins, Formatted: fmt.Sprintf("%dh %dm", f.DurMins/60, f.DurMins%60)},
			Stops:        f.Stops, AvailableSeats: f.Seats,
			Price:   models.Price{Amount: f.Price.Amount, Currency: "IDR"},
			Baggage: models.Baggage{CarryOn: fmt.Sprintf("%d piece(s)", f.Baggage.CarryOn), Checked: fmt.Sprintf("%d piece(s)", f.Baggage.Checked)},
		})
	}
	return results, nil
}

// --- Lion Air --- //
type LionAirProvider struct{}

func (l *LionAirProvider) Name() string { return "Lion Air" }
func (l *LionAirProvider) FetchFlights(ctx context.Context, req models.SearchRequest) ([]models.Flight, error) {
	if err := simulateDelay(ctx, 100, 200); err != nil {
		return nil, err
	}

	var mock struct {
		Flights []struct {
			FlightId string                               `json:"flight_id"`
			Airline  string                               `json:"airline"`
			AirlineC string                               `json:"airline_code"`
			Dep      struct{ Airport, City, Time string } `json:"departure"`
			Arr      struct{ Airport, City, Time string } `json:"arrival"`
			DurMins  int                                  `json:"duration_minutes"`
			Stops    int                                  `json:"stops"`
			Aircraft string                               `json:"aircraft"`
			Price    struct{ Amount int }                 `json:"price"`
			Seats    int                                  `json:"available_seats"`
			Baggage  struct {
				CarryOn int
				Checked int
			} `json:"baggage"`
		} `json:"flights"`
	}

	if err := readMockData("lion_air_search_response.json", &mock); err != nil {
		return nil, err
	}

	var results []models.Flight
	for _, f := range mock.Flights {
		depT, _ := time.Parse(time.RFC3339, f.Dep.Time)
		arrT, _ := time.Parse(time.RFC3339, f.Arr.Time)

		results = append(results, models.Flight{
			ID: fmt.Sprintf("%s_Lion", f.FlightId), Provider: l.Name(),
			Airline:      models.Airline{Name: f.Airline, Code: f.AirlineC},
			FlightNumber: f.FlightId,
			Departure:    models.Event{Airport: f.Dep.Airport, City: f.Dep.City, Datetime: f.Dep.Time, Timestamp: depT.Unix()},
			Arrival:      models.Event{Airport: f.Arr.Airport, City: f.Arr.City, Datetime: f.Arr.Time, Timestamp: arrT.Unix()},
			Duration:     models.Duration{TotalMinutes: f.DurMins, Formatted: fmt.Sprintf("%dh %dm", f.DurMins/60, f.DurMins%60)},
			Stops:        f.Stops, AvailableSeats: f.Seats,
			Price:   models.Price{Amount: f.Price.Amount, Currency: "IDR"},
			Baggage: models.Baggage{CarryOn: fmt.Sprintf("%d piece(s)", f.Baggage.CarryOn), Checked: fmt.Sprintf("%d piece(s)", f.Baggage.Checked)},
		})
	}
	return results, nil
}
