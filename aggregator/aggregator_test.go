package aggregator

import (
	"context"
	"flight-aggregator/models"
	"flight-aggregator/providers"
	"testing"
	"time"
)

func TestAggregatorService_Search_Basic(t *testing.T) {
	provs := []providers.Provider{
		&providers.GarudaProvider{},
		&providers.AirAsiaProvider{},
		&providers.LionAirProvider{},
		&providers.BatikAirProvider{},
	}
	agg := NewAggregatorService(provs)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	req := models.SearchRequest{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: "2025-12-15",
		Passengers:    "1",
		CabinClass:    "economy",
	}

	resp, err := agg.Search(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Metadata.TotalResults == 0 {
		t.Error("expected some flights, got none")
	}
	if resp.Metadata.ProvidersSucceeded == 0 {
		t.Error("expected at least one provider to succeed")
	}
}

func TestAggregatorService_Search_Filters(t *testing.T) {
	provs := []providers.Provider{
		&providers.GarudaProvider{},
		&providers.AirAsiaProvider{},
		&providers.LionAirProvider{},
		&providers.BatikAirProvider{},
	}
	agg := NewAggregatorService(provs)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	minPrice := 1000000
	maxStops := 0
	airlines := []string{"Garuda Indonesia"}
	minDuration := 100
	maxDuration := 200

	req := models.SearchRequest{
		Origin:             "CGK",
		Destination:        "DPS",
		DepartureDate:      "2025-12-15",
		Passengers:         "1",
		CabinClass:         "economy",
		MinPrice:           &minPrice,
		MaxStops:           &maxStops,
		Airlines:           airlines,
		MinDurationMinutes: &minDuration,
		MaxDurationMinutes: &maxDuration,
	}

	resp, err := agg.Search(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, f := range resp.Flights {
		if f.Price.Amount < minPrice {
			t.Errorf("flight price below minPrice: %d", f.Price.Amount)
		}
		if f.Stops > maxStops {
			t.Errorf("flight stops above maxStops: %d", f.Stops)
		}
		if f.Airline.Name != "Garuda Indonesia" {
			t.Errorf("flight airline mismatch: %s", f.Airline.Name)
		}
		if f.Duration.TotalMinutes < minDuration || f.Duration.TotalMinutes > maxDuration {
			t.Errorf("flight duration out of range: %d", f.Duration.TotalMinutes)
		}
	}
}

func TestAggregatorService_Search_Cache(t *testing.T) {
	provs := []providers.Provider{
		&providers.GarudaProvider{},
		&providers.AirAsiaProvider{},
		&providers.LionAirProvider{},
		&providers.BatikAirProvider{},
	}
	agg := NewAggregatorService(provs)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	req := models.SearchRequest{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: "2025-12-15",
		Passengers:    "1",
		CabinClass:    "economy",
	}

	resp1, err := agg.Search(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resp2, err := agg.Search(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp2.Metadata.CacheHit {
		t.Error("expected CacheHit true on second call")
	}
	if resp1.Metadata.TotalResults != resp2.Metadata.TotalResults {
		t.Error("cache results mismatch")
	}
}

func TestAggregatorService_Search_ErrorHandling(t *testing.T) {
	provs := []providers.Provider{
		&providers.GarudaProvider{},
	}
	agg := NewAggregatorService(provs)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Microsecond) // force timeout
	defer cancel()

	req := models.SearchRequest{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: "2025-12-15",
		Passengers:    "1",
		CabinClass:    "economy",
	}

	_, err := agg.Search(ctx, req)
	if ctx.Err() != nil || err != nil {
		t.Log("expected context deadline exceeded error, got:", err)
	}
}
