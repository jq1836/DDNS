# DDNS Client

A robust, configurable Dynamic DNS (DDNS) client written in Go with support for DuckDNS.

## Features

- 🦆 **DuckDNS Support**: Built specifically for DuckDNS with their simple API
- 🔁 **Generic Retry/Timeout Strategy**: Applies to any operation, not just HTTP requests
- ⚙️ **Configuration via JSON or Environment Variables**: Flexible configuration options
- 🛡️ **Robust Error Handling**: Comprehensive retry logic with exponential backoff
- 🧪 **Fully Tested**: Complete test coverage for all components
- 📦 **Clean Architecture**: Separation of concerns with interfaces and dependency injection

## Quick Start

1. **Clone the repository:**
   ```bash
   git clone <repository-url>
   cd ddns
   ```

2. **Configure the application:**
   ```bash
   cp config.example.json config.json
   # Edit config.json with your DuckDNS domain and token
   ```

3. **Run the application:**
   ```bash
   go run main.go
   ```

## DuckDNS Setup

1. **Get your DuckDNS token:**
   - Go to [DuckDNS.org](https://www.duckdns.org)
   - Sign in with your preferred account
   - Your token will be displayed on the main page

2. **Create a subdomain:**
   - Enter your desired subdomain name
   - Click "add domain"

3. **Configure the client:**
   ```bash
   export DDNS_DOMAIN=yoursubdomain.duckdns.org
   export DDNS_API_KEY=your-duckdns-token
   ```

## Usage

### Running the Client

```bash
go run main.go
```

### Using as a Library

```go
package main

import (
    "context"
    "ddns/ddns"
    "ddns/providers"
)

func main() {
    // Create provider
    factory := providers.NewFactory()
    config := ddns.Config{
        Provider: "duckdns",
        APIKey:   "your-duckdns-token",
        Domain:   "yoursubdomain.duckdns.org",
        TTL:      300,
    }
    
    provider, err := factory.CreateProvider(config)
    if err != nil {
        panic(err)
    }
    
    // Create service
    service := ddns.NewService(provider, config)
    
    // Update IP
    resp, err := service.UpdateIP(context.Background())
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Update result: %s\n", resp.Message)
}
```

## Generic Executor Usage

The executor package provides a generic retry/timeout strategy that can be applied to any operation:

### Basic Usage

```go
import "ddns/executor"

// Define a task
task := func(ctx context.Context) (string, error) {
    // Your operation here
    return "result", nil
}

// Execute with default strategies
exec := executor.NewExecutor()
result, err := executor.ExecuteSimple(exec, context.Background(), task)
```

### Custom Strategies

```go
// Exponential backoff with custom settings
retryStrategy := executor.NewExponentialBackoffStrategy(5, time.Second, 2.0)

// Progressive timeout (increases with each attempt)
timeoutStrategy := executor.NewProgressiveTimeoutStrategy(
    time.Second,    // base timeout
    1.5,           // multiplier
    10*time.Second, // max timeout
)

exec := executor.NewExecutor(
    executor.WithRetryStrategy(retryStrategy),
    executor.WithTimeoutStrategy(timeoutStrategy),
)
```

### Real-World Examples

```go
// Database operation with retries
dbTask := func(ctx context.Context) (*sql.Rows, error) {
    return db.QueryContext(ctx, "SELECT * FROM users")
}

rows, err := executor.ExecuteSimple(exec, ctx, dbTask)

// HTTP API call with conditional retry
apiTask := func(ctx context.Context) (*APIResponse, error) {
    return makeAPICall(ctx)
}

// Only retry on specific errors
customRetry := executor.NewConditionalRetryStrategy(
    3, time.Second,
    func(attempt int, err error) bool {
        return err != nil && isRetryableError(err)
    },
    nil,
)

exec := executor.NewExecutor(executor.WithRetryStrategy(customRetry))
response, err := executor.ExecuteSimple(exec, ctx, apiTask)
```

## Adding New DNS Providers

1. **Implement the Provider interface:**

```go
type MyProvider struct {
    apiKey string
}

func (p *MyProvider) UpdateRecord(ctx context.Context, req ddns.UpdateRequest) (*ddns.UpdateResponse, error) {
    // Implementation
}

func (p *MyProvider) GetCurrentRecord(ctx context.Context, domain, recordType string) (string, error) {
    // Implementation
}

func (p *MyProvider) ValidateCredentials(ctx context.Context) error {
    // Implementation
}

func (p *MyProvider) GetProviderName() string {
    return "myprovider"
}
```

2. **Add to the factory:**

```go
// In providers/factory.go
case "myduckdnsprovider":
    return NewMyDuckDNSProvider(config.APIKey), nil
```

3. **Add configuration validation:**

```go
// In providers/factory.go ValidateProviderConfig method
case "myduckdnsprovider":
    if config.APIKey == "" {
        return fmt.Errorf("myduckdnsprovider requires API key")
    }
    return nil
```

## Testing

Run all tests:
```bash
go test ./...
```

Run specific package tests:
```bash
go test ./ddns -v
go test ./executor -v
go test ./config -v
```

## Configuration Reference

### Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `DDNS_DOMAIN` | Domain to update | - | ✅ |
| `DDNS_API_KEY` | DNS provider API key | - | ✅ |
| `DDNS_PROVIDER` | DNS provider name | `duckdns` | ❌ |
| `DDNS_UPDATE_INTERVAL` | Check interval | `5m` | ❌ |
| `HTTP_TIMEOUT` | HTTP request timeout | `30s` | ❌ |
| `HTTP_MAX_RETRIES` | Max retry attempts | `3` | ❌ |
| `HTTP_RETRY_DELAY` | Delay between retries | `1s` | ❌ |
| `HTTP_USER_AGENT` | HTTP User-Agent | `ddns-client/1.0` | ❌ |

### Provider-Specific Configuration

#### DuckDNS
- `DDNS_API_KEY`: Your DuckDNS token from the dashboard
- `DDNS_DOMAIN`: Your subdomain (e.g., `yourname.duckdns.org`)

## Docker Support

```dockerfile
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o ddns-client .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/ddns-client .

ENV DDNS_DOMAIN=yoursubdomain.duckdns.org
ENV DDNS_API_KEY=your_duckdns_token
ENV DDNS_PROVIDER=duckdns

CMD ["./ddns-client"]
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for your changes
4. Ensure all tests pass
5. Submit a pull request

## License

MIT License - see LICENSE file for details.
