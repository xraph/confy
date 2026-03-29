# Confy Variadic Options API

This document demonstrates the new functional options pattern for creating Confy instances.

## Quick Start

### New Pattern (Recommended)

```go
import (
    "time"
    "github.com/xraph/confy"
)

// Create with multiple options
cfg := confy.New(
    confy.WithWatchInterval(30 * time.Second),
    confy.WithSecretsEnabled(true),
    confy.WithMetricsEnabled(true),
    confy.WithErrorRetryCount(5),
    confy.WithErrorRetryDelay(2 * time.Second),
)

// Minimal configuration
cfg := confy.New()  // Uses all defaults

// With specific settings
cfg := confy.New(
    confy.WithValidationMode(confy.ValidationModeStrict),
)
```

### Legacy Pattern (Backward Compatible)

```go
// Using Config struct
cfg := confy.NewFromConfig(confy.Config{
    WatchInterval:   30 * time.Second,
    SecretsEnabled:  true,
    MetricsEnabled:  true,
    ErrorRetryCount: 5,
    ErrorRetryDelay: 2 * time.Second,
})
```

## Available Options

### Core Configuration

- `WithWatchInterval(time.Duration)` - Set watch interval for configuration changes
- `WithValidationMode(ValidationMode)` - Set validation mode (Strict, Permissive, etc.)
- `WithSecretsEnabled(bool)` - Enable/disable secrets management
- `WithCacheEnabled(bool)` - Enable/disable caching
- `WithReloadOnChange(bool)` - Enable/disable automatic reload on changes
- `WithMetricsEnabled(bool)` - Enable/disable metrics collection

### Error Handling

- `WithErrorRetryCount(int)` - Set number of retry attempts on errors
- `WithErrorRetryDelay(time.Duration)` - Set delay between retry attempts

### Dependencies

- `WithLogger(logger.Logger)` - Set custom logger instance
- `WithMetrics(metrics.Metrics)` - Set custom metrics instance
- `WithErrorHandler(errors.ErrorHandler)` - Set custom error handler

### Sources

- `WithDefaultSources([]SourceConfig)` - Set default configuration sources

## Usage Examples

### Development Configuration

```go
cfg := confy.New(
    confy.WithWatchInterval(5 * time.Second),     // Fast reload
    confy.WithValidationMode(confy.ValidationModePermissive),
    confy.WithMetricsEnabled(false),
)
```

### Production Configuration

```go
cfg := confy.New(
    confy.WithWatchInterval(30 * time.Second),
    confy.WithValidationMode(confy.ValidationModeStrict),
    confy.WithSecretsEnabled(true),
    confy.WithMetricsEnabled(true),
    confy.WithErrorRetryCount(10),
    confy.WithErrorRetryDelay(3 * time.Second),
)
```

### Composable Options

```go
// Define reusable option sets
var developmentOpts = []confy.Option{
    confy.WithWatchInterval(5 * time.Second),
    confy.WithValidationMode(confy.ValidationModePermissive),
}

var productionOpts = []confy.Option{
    confy.WithWatchInterval(30 * time.Second),
    confy.WithValidationMode(confy.ValidationModeStrict),
    confy.WithSecretsEnabled(true),
    confy.WithMetricsEnabled(true),
}

// Apply option sets
devCfg := confy.New(developmentOpts...)
prodCfg := confy.New(productionOpts...)

// Mix and match
customCfg := confy.New(append(
    developmentOpts,
    confy.WithLogger(myCustomLogger),
)...)
```

### With Dependencies

```go
import (
    "github.com/xraph/confy"
    "github.com/xraph/go-utils/log"
    "github.com/xraph/go-utils/metrics"
    "github.com/xraph/go-utils/errs"
)

logger := log.NewBeautifulLogger(log.WithLevel(log.LevelInfo))
metricsCollector := metrics.NewPrometheusMetrics()
errorHandler := errs.NewDefaultErrorHandler()

cfg := confy.New(
    confy.WithLogger(logger),
    confy.WithMetrics(metricsCollector),
    confy.WithErrorHandler(errorHandler),
    confy.WithMetricsEnabled(true),
)
```

## Benefits of Variadic Options

1. **Type Safety**: Compile-time validation of options
2. **Discoverability**: IDE autocomplete shows all available options
3. **Flexibility**: Mix and match options as needed
4. **Readability**: Self-documenting code
5. **Backward Compatibility**: Legacy `NewFromConfig` still works
6. **Composability**: Create reusable option sets
7. **Default Values**: Omit options to use sensible defaults

## Migration Guide

### Before (Config Struct)

```go
cfg := confy.New(confy.Config{
    WatchInterval:   30 * time.Second,
    SecretsEnabled:  true,
    MetricsEnabled:  true,
    Logger:          myLogger,
})
```

### After (Variadic Options)

```go
cfg := confy.New(
    confy.WithWatchInterval(30 * time.Second),
    confy.WithSecretsEnabled(true),
    confy.WithMetricsEnabled(true),
    confy.WithLogger(myLogger),
)
```

Or keep using the old pattern with `NewFromConfig`:

```go
cfg := confy.NewFromConfig(confy.Config{
    WatchInterval:   30 * time.Second,
    SecretsEnabled:  true,
    MetricsEnabled:  true,
    Logger:          myLogger,
})
```

## Best Practices

1. **Use Variadic Options for New Code**: The functional options pattern is more idiomatic and flexible
2. **Group Related Options**: Organize options logically for better readability
3. **Create Option Sets**: For different environments (dev, staging, prod)
4. **Document Custom Configurations**: When creating reusable option sets
5. **Leverage Defaults**: Only specify options that differ from defaults

## Testing

The options are fully tested:

```go
// Test individual options
func TestWithWatchInterval(t *testing.T) {
    cfg := &Config{}
    WithWatchInterval(10 * time.Second)(cfg)
    // Assert cfg.WatchInterval == 10 * time.Second
}

// Test composition
func TestNew_WithOptions(t *testing.T) {
    c := confy.New(
        confy.WithWatchInterval(20*time.Second),
        confy.WithSecretsEnabled(true),
    )
    // Assert instance created correctly
}
```

See `config_opts_test.go` for comprehensive test examples.

