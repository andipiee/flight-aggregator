package aggregator

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"time"

	"flight-aggregator/models"
	"flight-aggregator/providers"
)

type AggregatorService struct {
	providers []providers.Provider
}

func NewAggregatorService(p []providers.Provider) *AggregatorService {
	return &AggregatorService{providers: p}
}

func (s *AggregatorService) Search(ctx context.Context, req models.SearchRequest) (models.SearchResponse, error) {
	start := time.Now()
	key := cacheKey(req)
	resp, found := aggCache.Get(key)
	if found {
		resp.Metadata.SearchTimeMs = time.Since(start).Milliseconds()
		resp.Metadata.CacheHit = true
		return resp, nil
	}


	// Check for context timeout before starting provider calls
	if ctx.Err() != nil {
		return models.SearchResponse{
			SearchCriteria: req,
			Metadata: models.Metadata{
				TotalResults:       0,
				ProvidersQueried:   len(s.providers),
				ProvidersSucceeded: 0,
				ProvidersFailed:    len(s.providers),
				SearchTimeMs:       time.Since(start).Milliseconds(),
				CacheHit:           false,
			},
			Flights: nil,
		}, ctx.Err()
	}

	results, successCount, err := s.fetchFromProviders(ctx, req)
	if err != nil {
		return models.SearchResponse{
			SearchCriteria: req,
			Metadata: models.Metadata{
				TotalResults:       0,
				ProvidersQueried:   len(s.providers),
				ProvidersSucceeded: successCount,
				ProvidersFailed:    len(s.providers) - successCount,
				SearchTimeMs:       time.Since(start).Milliseconds(),
				CacheHit:           false,
			},
			Flights: nil,
		}, err
	}

	// Check for context timeout after provider calls
	if ctx.Err() != nil {
		return models.SearchResponse{
			SearchCriteria: req,
			Metadata: models.Metadata{
				TotalResults:       0,
				ProvidersQueried:   len(s.providers),
				ProvidersSucceeded: successCount,
				ProvidersFailed:    len(s.providers) - successCount,
				SearchTimeMs:       time.Since(start).Milliseconds(),
				CacheHit:           false,
			},
			Flights: nil,
			}, ctx.Err()
	}

	filtered, err := s.filterFlights(results, req)
	if err != nil {
		return models.SearchResponse{
			SearchCriteria: req,
			Metadata: models.Metadata{
				TotalResults:       0,
				ProvidersQueried:   len(s.providers),
				ProvidersSucceeded: successCount,
				ProvidersFailed:    len(s.providers) - successCount,
				SearchTimeMs:       time.Since(start).Milliseconds(),
				CacheHit:           false,
			},
			Flights: nil,
		}, err
	}

	unique, err := s.comparePrices(filtered)
	if err != nil {
		return models.SearchResponse{
			SearchCriteria: req,
			Metadata: models.Metadata{
				TotalResults:       0,
				ProvidersQueried:   len(s.providers),
				ProvidersSucceeded: successCount,
				ProvidersFailed:    len(s.providers) - successCount,
				SearchTimeMs:       time.Since(start).Milliseconds(),
				CacheHit:           false,
			},
			Flights: nil,
		}, err
	}

	err = s.calcDurations(unique)
	if err != nil {
		return models.SearchResponse{
			SearchCriteria: req,
			Metadata: models.Metadata{
				TotalResults:       0,
				ProvidersQueried:   len(s.providers),
				ProvidersSucceeded: successCount,
				ProvidersFailed:    len(s.providers) - successCount,
				SearchTimeMs:       time.Since(start).Milliseconds(),
				CacheHit:           false,
			},
			Flights: nil,
		}, err
	}

	err = s.rankFlights(unique)
	if err != nil {
		return models.SearchResponse{
			SearchCriteria: req,
			Metadata: models.Metadata{
				TotalResults:       0,
				ProvidersQueried:   len(s.providers),
				ProvidersSucceeded: successCount,
				ProvidersFailed:    len(s.providers) - successCount,
				SearchTimeMs:       time.Since(start).Milliseconds(),
				CacheHit:           false,
			},
			Flights: nil,
		}, err
	}

	sorted, err := s.sortFlights(unique, req)
	if err != nil {
		return models.SearchResponse{
			SearchCriteria: req,
			Metadata: models.Metadata{
				TotalResults:       0,
				ProvidersQueried:   len(s.providers),
				ProvidersSucceeded: successCount,
				ProvidersFailed:    len(s.providers) - successCount,
				SearchTimeMs:       time.Since(start).Milliseconds(),
				CacheHit:           false,
			},
			Flights: nil,
		}, err
	}

	resp = models.SearchResponse{
		SearchCriteria: req,
		Metadata: models.Metadata{
			TotalResults:       len(sorted),
			ProvidersQueried:   len(s.providers),
			ProvidersSucceeded: successCount,
			ProvidersFailed:    len(s.providers) - successCount,
			SearchTimeMs:       time.Since(start).Milliseconds(),
			CacheHit:           false,
		},
		Flights: sorted,
	}
	aggCache.Set(key, resp)
	return resp, nil
}

// Concurrent provider calls
func (s *AggregatorService) fetchFromProviders(ctx context.Context, req models.SearchRequest) ([]models.Flight, int, error) {
	resultsChan := make(chan []models.Flight, len(s.providers))
	var successCount int
	var failedCount int

	done := make(chan struct{})

	for _, p := range s.providers {
		go func(prov providers.Provider) {
			var flights []models.Flight
			var err error
			retries := 2
			for i := 0; i <= retries; i++ {
				flights, err = prov.FetchFlights(ctx, req)
				if err == nil {
					break
				}
				time.Sleep(time.Duration(100*(i+1)) * time.Millisecond)
			}
			if err == nil {
				resultsChan <- flights
				successCount++
			} else {
				failedCount++
			}
			done <- struct{}{}
		}(p)
	}

	// Wait for all goroutines
	for i := 0; i < len(s.providers); i++ {
		<-done
	}
	close(resultsChan)

	var allFlights []models.Flight
	for fList := range resultsChan {
		allFlights = append(allFlights, fList...)
	}

	return allFlights, successCount, nil
}

// Filtering
func (s *AggregatorService) filterFlights(flights []models.Flight, req models.SearchRequest) ([]models.Flight, error) {
	var filtered []models.Flight
	for _, f := range flights {
		if f.Departure.Airport != req.Origin || f.Arrival.Airport != req.Destination {
			continue
		}
		if req.MinPrice != nil && f.Price.Amount < *req.MinPrice {
			continue
		}
		if req.MaxPrice != nil && f.Price.Amount > *req.MaxPrice {
			continue
		}
		if req.MinStops != nil && f.Stops < *req.MinStops {
			continue
		}
		if req.MaxStops != nil && f.Stops > *req.MaxStops {
			continue
		}
		if req.DepartureTimeStart != nil && req.DepartureTimeEnd != nil {
			depTime := time.Unix(f.Departure.Timestamp, 0).Format("15:04")
			if depTime < *req.DepartureTimeStart || depTime > *req.DepartureTimeEnd {
				continue
			}
		}
		if req.ArrivalTimeStart != nil && req.ArrivalTimeEnd != nil {
			arrTime := time.Unix(f.Arrival.Timestamp, 0).Format("15:04")
			if arrTime < *req.ArrivalTimeStart || arrTime > *req.ArrivalTimeEnd {
				continue
			}
		}
		if req.Airlines != nil && len(req.Airlines) > 0 {
			found := false
			for _, airline := range req.Airlines {
				if f.Airline.Name == airline || f.Airline.Code == airline {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}
		if req.MinDurationMinutes != nil && f.Duration.TotalMinutes < *req.MinDurationMinutes {
			continue
		}
		if req.MaxDurationMinutes != nil && f.Duration.TotalMinutes > *req.MaxDurationMinutes {
			continue
		}
		filtered = append(filtered, f)
	}
	return filtered, nil
}

// Price comparison (deduplicate by flight number and departure time, keep lowest price)
func (s *AggregatorService) comparePrices(flights []models.Flight) ([]models.Flight, error) {
	flightMap := make(map[string]models.Flight)
	for _, f := range flights {
		key := f.FlightNumber + "|" + strconv.FormatInt(f.Departure.Timestamp, 10)
		if existing, ok := flightMap[key]; ok {
			if f.Price.Amount < existing.Price.Amount {
				flightMap[key] = f
			}
		} else {
			flightMap[key] = f
		}
	}
	unique := make([]models.Flight, 0, len(flightMap))
	for _, f := range flightMap {
		unique = append(unique, f)
	}
	return unique, nil
}

// 4. Calculate total trip duration including layovers
func (s *AggregatorService) calcDurations(flights []models.Flight) error {
	for i := range flights {
		f := &flights[i]
		if f.Duration.TotalMinutes == 0 && f.Departure.Timestamp > 0 && f.Arrival.Timestamp > 0 {
			dur := int((f.Arrival.Timestamp - f.Departure.Timestamp) / 60)
			f.Duration.TotalMinutes = dur
			f.Duration.Formatted = fmt.Sprintf("%dh %dm", dur/60, dur%60)
		}
	}
	return nil
}

// Ranking (best value)
func (s *AggregatorService) rankFlights(flights []models.Flight) error {
	bestValue := func(f models.Flight) int {
		return f.Price.Amount + f.Stops*100000 + f.Duration.TotalMinutes*100 + int(f.Departure.Timestamp/3600)
	}
	sort.Slice(flights, func(i, j int) bool {
		return bestValue(flights[i]) < bestValue(flights[j])
	})
	return nil
}

// 6. Sorting (by user request)
func (s *AggregatorService) sortFlights(flights []models.Flight, req models.SearchRequest) ([]models.Flight, error) {
	if req.SortBy == nil || *req.SortBy == "" || len(flights) < 2 {
		return flights, nil
	}
	switch *req.SortBy {
	case "price_asc":
		sort.Slice(flights, func(i, j int) bool {
			return flights[i].Price.Amount < flights[j].Price.Amount
		})
	case "price_desc":
		sort.Slice(flights, func(i, j int) bool {
			return flights[i].Price.Amount > flights[j].Price.Amount
		})
	case "duration_asc":
		sort.Slice(flights, func(i, j int) bool {
			return flights[i].Duration.TotalMinutes < flights[j].Duration.TotalMinutes
		})
	case "duration_desc":
		sort.Slice(flights, func(i, j int) bool {
			return flights[i].Duration.TotalMinutes > flights[j].Duration.TotalMinutes
		})
	case "departure_asc":
		sort.Slice(flights, func(i, j int) bool {
			return flights[i].Departure.Timestamp < flights[j].Departure.Timestamp
		})
	case "departure_desc":
		sort.Slice(flights, func(i, j int) bool {
			return flights[i].Departure.Timestamp > flights[j].Departure.Timestamp
		})
	case "arrival_asc":
		sort.Slice(flights, func(i, j int) bool {
			return flights[i].Arrival.Timestamp < flights[j].Arrival.Timestamp
		})
	case "arrival_desc":
		sort.Slice(flights, func(i, j int) bool {
			return flights[i].Arrival.Timestamp > flights[j].Arrival.Timestamp
		})
	}
	return flights, nil
}
