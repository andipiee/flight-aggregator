# Flight Aggregator

A Go-based flight search aggregator that queries multiple airline providers, applies advanced filtering, ranking, and caching, and returns the best flight options for a given search request.

## Project Structure

```
flight-aggregator/
├── go.mod, go.sum           # Go module files
├── main.go                  # Entry point for running the aggregator
├── readme.MD                # Project documentation
├── aggregator/              # Aggregator service logic and tests
│   ├── aggregator.go        # Main aggregator implementation
│   ├── aggregator_test.go   # Unit tests for aggregator
│   ├── cache.go             # In-memory cache implementation
├── mock_data/               # Mock flight data for providers
│   ├── airasia_search_response.json
│   ├── batik_air_search_response.json
│   ├── expected_result.json
│   ├── garuda_indonesia_search_response.json
│   ├── lion_air_search_response.json
├── models/                  # Data models
│   └── models.go            # Structs for requests, responses, flights, etc.
├── providers/               # Provider interfaces and implementations
│   ├── providers.go         # Provider logic and mock data reading
│   └── providers_test.go    # Unit tests for providers
```

## How to Run

1. **Install Go (>= 1.18)**
2. Clone the repository and navigate to the project root.
3. Update the Go package
```sh
go mod tidy
```
4. Run the aggregator:

```sh
# From project root
go run main.go
```

This will execute a sample search and print results, including a cache hit demonstration.

## How to Run Tests

Run all tests with coverage:

```sh
go test ./aggregator ./providers -cover
```

You can also run tests for individual packages:

```sh
go test ./providers -cover
go test ./aggregator -cover
```

## Approach

This project implements a robust flight search aggregator in Go, designed for extensibility and reliability:

- **Provider Abstraction:** Each airline provider is implemented as a Go interface, allowing easy addition of new providers and uniform querying.
- **Concurrent Calls:** Provider queries are executed concurrently, improving performance and reducing latency.
- **Advanced Filtering:** The aggregator supports filtering by price, stops, airlines, departure/arrival time, and duration, giving users granular control over search results.
- **Deduplication & Price Comparison:** Flights from different providers are deduplicated and the best price is selected for each unique flight.
- **Ranking & Sorting:** Results are ranked by a "best value" score (combining price and convenience) and can be sorted by price, duration, or time.
- **In-Memory Caching:** Search results are cached in-memory with TTL and size limits, reducing repeated provider calls and improving response time.
- **Graceful Error Handling:** If a provider fails, results from other providers are still returned. Timeout and context cancellation are handled cleanly.
- **Test Coverage:** The codebase is covered by unit tests for both providers and aggregator logic, ensuring reliability and maintainability.

This approach balances performance, flexibility, and reliability, making it suitable for both demonstration and production use cases.

## Assumptions & Notes

- **Providers use mock data** from the `mock_data/` directory. No real API calls are made.
- **Caching** is in-memory, production-ready (TTL, size limit, FIFO eviction).
- **Filtering** supports price, stops, airlines, departure/arrival time, and duration.
- **Ranking** is based on a combination of price, stops, duration, and convenience.
- **Error handling**: If a provider fails, results from other providers are still returned. Timeout and context cancellation are handled gracefully.
- **Tests**: >65% coverage for both providers and aggregator logic.
- **Run from project root** to ensure mock data is found by tests and providers.

---

For any questions or improvements, see the code comments or reach out to the project maintainer.
