package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"flight-aggregator/aggregator"
	"flight-aggregator/models"
	"flight-aggregator/providers"
)

func main() {
	// Initialize simulated providers
	provs := []providers.Provider{
		&providers.GarudaProvider{},
		&providers.AirAsiaProvider{},
		&providers.LionAirProvider{},
		&providers.BatikAirProvider{},
	}

	aggService := aggregator.NewAggregatorService(provs)

	// Initialize all required and optional search variables
	origin := "CGK"
	destination := "DPS"
	departureDate := "2025-12-15"
	passengers := "1"
	cabinClass := "economy"

	// Optional filters
	minPrice := 500000
	maxPrice := 1500000
	minStops := 0
	maxStops := 1
	departureTimeStart := "05:00"
	departureTimeEnd := "20:00"
	arrivalTimeStart := "07:00"
	arrivalTimeEnd := "23:00"
	airlines := []string{"Garuda Indonesia", "AirAsia", "Lion Air", "Batik Air"}
	minDuration := 60
	maxDuration := 300

	// Build request
	req := models.SearchRequest{
		Origin:             origin,
		Destination:        destination,
		DepartureDate:      departureDate,
		Passengers:         passengers,
		CabinClass:         cabinClass,
		MinPrice:           &minPrice,
		MaxPrice:           &maxPrice,
		MinStops:           &minStops,
		MaxStops:           &maxStops,
		DepartureTimeStart: &departureTimeStart,
		DepartureTimeEnd:   &departureTimeEnd,
		ArrivalTimeStart:   &arrivalTimeStart,
		ArrivalTimeEnd:     &arrivalTimeEnd,
		Airlines:           airlines,
		MinDurationMinutes: &minDuration,
		MaxDurationMinutes: &maxDuration,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	
	response, err := aggService.Search(ctx, req)
	if err != nil {
		log.Fatalf("Calling Aggregate Search got Error : %v", err)
	}

	// Print results
	b, _ := json.MarshalIndent(response, "", "  ")
	log.Println(string(b))

	log.Println("Performing second search with same parameters to test caching...")
	response2, err := aggService.Search(ctx, req)
	if err != nil {
		log.Fatalf("Calling Aggregate Search got Error : %v", err)
	}

	// Print results
	b2, _ := json.MarshalIndent(response2, "", "  ")
	log.Println(string(b2))
}
