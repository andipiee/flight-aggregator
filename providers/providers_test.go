package providers

import (
	"context"
	"flight-aggregator/models"
	"testing"
	"time"
)

func TestGarudaProvider_FetchFlights(t *testing.T) {
	prov := &GarudaProvider{}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	resp, err := prov.FetchFlights(ctx, models.SearchRequest{Origin: "CGK", Destination: "DPS", DepartureDate: "2025-12-15"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp) == 0 {
		t.Error("expected flights, got none")
	}
}

func TestAirAsiaProvider_FetchFlights(t *testing.T) {
	prov := &AirAsiaProvider{}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, err := prov.FetchFlights(ctx, models.SearchRequest{Origin: "CGK", Destination: "DPS", DepartureDate: "2025-12-15"})
	// AirAsia simulates 10% failure, so just check for no panic
	if err != nil && err.Error() != "AirAsia API Service Unavailable (503)" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBatikAirProvider_FetchFlights(t *testing.T) {
	prov := &BatikAirProvider{}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, err := prov.FetchFlights(ctx, models.SearchRequest{Origin: "CGK", Destination: "DPS", DepartureDate: "2025-12-15"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLionAirProvider_FetchFlights(t *testing.T) {
	prov := &LionAirProvider{}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, err := prov.FetchFlights(ctx, models.SearchRequest{Origin: "CGK", Destination: "DPS", DepartureDate: "2025-12-15"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSimulateDelay_Cancel(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()
	err := simulateDelay(ctx, 10, 20)
	if err == nil {
		t.Error("expected error due to context cancel, got nil")
	}
}
