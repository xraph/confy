package confy

import (
	"time"

	configcore "github.com/xraph/confy/internal"
	errors "github.com/xraph/go-utils/errs"
	logger "github.com/xraph/go-utils/log"
	"github.com/xraph/go-utils/metrics"
)

// =============================================================================
// CONSTRUCTOR OPTIONS
// =============================================================================

// Option configures a Confy instance.
type Option func(*Config)

// WithDefaultSources sets the default configuration sources.
func WithDefaultSources(sources []SourceConfig) Option {
	return func(c *Config) {
		c.DefaultSources = sources
	}
}

// WithWatchInterval sets the watch interval for configuration changes.
func WithWatchInterval(interval time.Duration) Option {
	return func(c *Config) {
		c.WatchInterval = interval
	}
}

// WithValidationMode sets the validation mode.
func WithValidationMode(mode ValidationMode) Option {
	return func(c *Config) {
		c.ValidationMode = mode
	}
}

// WithSecretsEnabled enables or disables secrets management.
func WithSecretsEnabled(enabled bool) Option {
	return func(c *Config) {
		c.SecretsEnabled = enabled
	}
}

// WithCacheEnabled enables or disables caching.
func WithCacheEnabled(enabled bool) Option {
	return func(c *Config) {
		c.CacheEnabled = enabled
	}
}

// WithReloadOnChange enables or disables automatic reload on configuration changes.
func WithReloadOnChange(enabled bool) Option {
	return func(c *Config) {
		c.ReloadOnChange = enabled
	}
}

// WithErrorRetryCount sets the number of retry attempts on errors.
func WithErrorRetryCount(count int) Option {
	return func(c *Config) {
		c.ErrorRetryCount = count
	}
}

// WithErrorRetryDelay sets the delay between retry attempts.
func WithErrorRetryDelay(delay time.Duration) Option {
	return func(c *Config) {
		c.ErrorRetryDelay = delay
	}
}

// WithMetricsEnabled enables or disables metrics collection.
func WithMetricsEnabled(enabled bool) Option {
	return func(c *Config) {
		c.MetricsEnabled = enabled
	}
}

// WithLogger sets the logger instance.
func WithLogger(log logger.Logger) Option {
	return func(c *Config) {
		c.Logger = log
	}
}

// WithMetrics sets the metrics instance.
func WithMetrics(m metrics.Metrics) Option {
	return func(c *Config) {
		c.Metrics = m
	}
}

// WithErrorHandler sets the error handler.
func WithErrorHandler(handler errors.ErrorHandler) Option {
	return func(c *Config) {
		c.ErrorHandler = handler
	}
}

// =============================================================================
// GET OPTIONS
// =============================================================================

// WithDefault sets a default value.
func WithDefault(value any) configcore.GetOption {
	return func(opts *configcore.GetOptions) {
		opts.Default = value
	}
}

// WithRequired marks the key as required.
func WithRequired() configcore.GetOption {
	return func(opts *configcore.GetOptions) {
		opts.Required = true
	}
}

// WithValidator adds a validation function.
func WithValidator(fn func(any) error) configcore.GetOption {
	return func(opts *configcore.GetOptions) {
		opts.Validator = fn
	}
}

// WithTransform adds a transformation function.
func WithTransform(fn func(any) any) configcore.GetOption {
	return func(opts *configcore.GetOptions) {
		opts.Transform = fn
	}
}

// WithOnMissing sets a callback for missing keys.
func WithOnMissing(fn func(string) any) configcore.GetOption {
	return func(opts *configcore.GetOptions) {
		opts.OnMissing = fn
	}
}

// AllowEmpty allows empty values.
func AllowEmpty() configcore.GetOption {
	return func(opts *configcore.GetOptions) {
		opts.AllowEmpty = true
	}
}

// WithCacheKey sets a custom cache key.
func WithCacheKey(key string) configcore.GetOption {
	return func(opts *configcore.GetOptions) {
		opts.CacheKey = key
	}
}
