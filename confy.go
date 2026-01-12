package confy

//nolint:gosec // G115: All integer type conversions are intentional and value-controlled
// This file contains helper methods for type-safe configuration value retrieval.
// The integer conversions are safe because values come from application configuration sources.

import (
	"context"
	"fmt"
	"maps"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	configformats "github.com/xraph/confy/formats"
	configcore "github.com/xraph/confy/internal"
	errors "github.com/xraph/go-utils/errs"
	logger "github.com/xraph/go-utils/log"
	"github.com/xraph/go-utils/metrics"
)

// =============================================================================
// CONFY IMPLEMENTATION
// =============================================================================

// ConfyImpl implements the Confy interface for configuration management.
type ConfyImpl struct {
	sources         []ConfigSource
	registry        SourceRegistry
	loader          *configformats.Loader
	validator       *Validator
	watcher         *Watcher
	data            map[string]any
	watchCallbacks  map[string][]func(string, any)
	changeCallbacks []func(ConfigChange)
	mu              sync.RWMutex
	watchCtx        context.Context
	watchCancel     context.CancelFunc
	started         bool
	logger          logger.Logger
	metrics         metrics.Metrics
	errorHandler    errors.ErrorHandler
	secretsManager  configcore.SecretsManager
	converter       *configcore.TypeConverter
	merger          *configcore.MergeUtil
}

// Config contains configuration for creating a ConfyImpl instance.
type Config struct {
	DefaultSources  []SourceConfig      `json:"default_sources"   yaml:"default_sources"`
	WatchInterval   time.Duration       `json:"watch_interval"    yaml:"watch_interval"`
	ValidationMode  ValidationMode      `json:"validation_mode"   yaml:"validation_mode"`
	SecretsEnabled  bool                `json:"secrets_enabled"   yaml:"secrets_enabled"`
	CacheEnabled    bool                `json:"cache_enabled"     yaml:"cache_enabled"`
	ReloadOnChange  bool                `json:"reload_on_change"  yaml:"reload_on_change"`
	ErrorRetryCount int                 `json:"error_retry_count" yaml:"error_retry_count"`
	ErrorRetryDelay time.Duration       `json:"error_retry_delay" yaml:"error_retry_delay"`
	MetricsEnabled  bool                `json:"metrics_enabled"   yaml:"metrics_enabled"`
	Logger          logger.Logger       `json:"-"                 yaml:"-"`
	Metrics         metrics.Metrics     `json:"-"                 yaml:"-"`
	ErrorHandler    errors.ErrorHandler `json:"-"                 yaml:"-"`
}

// New creates a new ConfyImpl instance that implements the Confy interface.
func New(config Config) Confy {
	if config.WatchInterval == 0 {
		config.WatchInterval = 30 * time.Second
	}

	if config.ErrorRetryCount == 0 {
		config.ErrorRetryCount = 3
	}

	if config.ErrorRetryDelay == 0 {
		config.ErrorRetryDelay = 5 * time.Second
	}

	impl := &ConfyImpl{
		sources:         make([]ConfigSource, 0),
		data:            make(map[string]any),
		watchCallbacks:  make(map[string][]func(string, any)),
		changeCallbacks: make([]func(ConfigChange), 0),
		logger:          config.Logger,
		metrics:         config.Metrics,
		errorHandler:    config.ErrorHandler,
		converter:       configcore.NewTypeConverter(),
		merger:          configcore.NewMergeUtil(),
	}

	impl.registry = NewSourceRegistry(impl.logger)
	impl.loader = configformats.NewLoader(configformats.LoaderConfig{
		Logger:       impl.logger,
		Metrics:      impl.metrics,
		ErrorHandler: impl.errorHandler,
		RetryCount:   config.ErrorRetryCount,
		RetryDelay:   config.ErrorRetryDelay,
	})
	impl.validator = NewValidator(ValidatorConfig{
		Mode:         config.ValidationMode,
		Logger:       impl.logger,
		ErrorHandler: impl.errorHandler,
	})
	impl.watcher = NewWatcher(WatcherConfig{
		Interval:     config.WatchInterval,
		Logger:       impl.logger,
		Metrics:      impl.metrics,
		ErrorHandler: impl.errorHandler,
	})

	if config.SecretsEnabled {
		impl.secretsManager = NewSecretsManager(SecretsConfig{
			Logger:       impl.logger,
			ErrorHandler: impl.errorHandler,
		})
	}

	return impl
}

func (c *ConfyImpl) Name() string {
	return "confy"
}

func (c *ConfyImpl) SecretsManager() SecretsManager {
	return c.secretsManager
}

// =============================================================================
// SIMPLE API - VARIADIC DEFAULTS
// =============================================================================

// Get returns a configuration value.
func (c *ConfyImpl) Get(key string) any {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.getValue(key)
}

// GetString returns a string value with optional default.
func (c *ConfyImpl) GetString(key string, defaultValue ...string) string {
	value := c.Get(key)
	if value == nil {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return ""
	}
	return c.converter.ToString(value)
}

// GetInt returns an int value with optional default.
func (c *ConfyImpl) GetInt(key string, defaultValue ...int) int {
	value := c.Get(key)
	if value == nil {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return 0
	}

	result, err := c.converter.ToInt(value)
	if err != nil {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return 0
	}

	return result
}

// GetInt8 returns an int8 value with optional default.
func (c *ConfyImpl) GetInt8(key string, defaultValue ...int8) int8 {
	value := c.Get(key)
	if value == nil {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return 0
	}

	result, err := c.converter.ToInt8(value)
	if err != nil {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return 0
	}

	return result
}

// GetInt16 returns an int16 value with optional default.
func (c *ConfyImpl) GetInt16(key string, defaultValue ...int16) int16 {
	value := c.Get(key)
	if value == nil {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return 0
	}

	result, err := c.converter.ToInt16(value)
	if err != nil {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return 0
	}

	return result
}

// GetInt32 returns an int32 value with optional default.
func (c *ConfyImpl) GetInt32(key string, defaultValue ...int32) int32 {
	value := c.Get(key)
	if value == nil {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return 0
	}

	result, err := c.converter.ToInt32(value)
	if err != nil {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return 0
	}

	return result
}

// GetInt64 returns an int64 value with optional default.
func (c *ConfyImpl) GetInt64(key string, defaultValue ...int64) int64 {
	value := c.Get(key)
	if value == nil {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return 0
	}

	result, err := c.converter.ToInt64(value)
	if err != nil {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return 0
	}

	return result
}

// GetUint returns a uint value with optional default.
func (c *ConfyImpl) GetUint(key string, defaultValue ...uint) uint {
	value := c.Get(key)
	if value == nil {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return 0
	}

	result, err := c.converter.ToUint(value)
	if err != nil {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return 0
	}

	return result
}

// GetUint8 returns a uint8 value with optional default.
func (c *ConfyImpl) GetUint8(key string, defaultValue ...uint8) uint8 {
	value := c.Get(key)
	if value == nil {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return 0
	}

	result, err := c.converter.ToUint8(value)
	if err != nil {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return 0
	}

	return result
}

// GetUint16 returns a uint16 value with optional default.
func (c *ConfyImpl) GetUint16(key string, defaultValue ...uint16) uint16 {
	value := c.Get(key)
	if value == nil {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return 0
	}

	result, err := c.converter.ToUint16(value)
	if err != nil {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return 0
	}

	return result
}

// GetUint32 returns a uint32 value with optional default.
func (c *ConfyImpl) GetUint32(key string, defaultValue ...uint32) uint32 {
	value := c.Get(key)
	if value == nil {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return 0
	}

	result, err := c.converter.ToUint32(value)
	if err != nil {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return 0
	}
	return result
}

// GetUint64 returns a uint64 value with optional default.
func (c *ConfyImpl) GetUint64(key string, defaultValue ...uint64) uint64 {
	value := c.Get(key)
	if value == nil {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return 0
	}

	result, err := c.converter.ToUint64(value)
	if err != nil {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return 0
	}

	return result
}

// GetFloat32 returns a float32 value with optional default.
func (c *ConfyImpl) GetFloat32(key string, defaultValue ...float32) float32 {
	value := c.Get(key)
	if value == nil {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return 0
	}

	result, err := c.converter.ToFloat32(value)
	if err != nil {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return 0
	}

	return result
}

// GetFloat64 returns a float64 value with optional default.
func (c *ConfyImpl) GetFloat64(key string, defaultValue ...float64) float64 {
	value := c.Get(key)
	if value == nil {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return 0
	}

	result, err := c.converter.ToFloat64(value)
	if err != nil {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return 0
	}

	return result
}

// GetBool returns a bool value with optional default.
func (c *ConfyImpl) GetBool(key string, defaultValue ...bool) bool {
	value := c.Get(key)
	if value == nil {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return false
	}

	result, err := c.converter.ToBool(value)
	if err != nil {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return false
	}
	return result
}

// GetDuration returns a duration value with optional default.
func (c *ConfyImpl) GetDuration(key string, defaultValue ...time.Duration) time.Duration {
	value := c.Get(key)
	if value == nil {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return 0
	}

	result, err := c.converter.ToDuration(value)
	if err != nil {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return 0
	}
	return result
}

// GetTime returns a time value with optional default.
func (c *ConfyImpl) GetTime(key string, defaultValue ...time.Time) time.Time {
	value := c.Get(key)
	if value == nil {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return time.Time{}
	}

	result, err := c.converter.ToTime(value)
	if err != nil {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return time.Time{}
	}

	return result
}

// GetSizeInBytes returns size in bytes with optional default.
func (c *ConfyImpl) GetSizeInBytes(key string, defaultValue ...uint64) uint64 {
	value := c.Get(key)
	if value == nil {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return 0
	}

	result, err := c.converter.ToSizeInBytes(value)
	if err != nil {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return 0
	}

	return result
}

// GetStringSlice returns a string slice with optional default.
func (c *ConfyImpl) GetStringSlice(key string, defaultValue ...[]string) []string {
	value := c.Get(key)
	if value == nil {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return nil
	}

	result, err := c.converter.ToStringSlice(value)
	if err != nil {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return nil
	}
	return result
}

// GetIntSlice returns an int slice with optional default.
func (c *ConfyImpl) GetIntSlice(key string, defaultValue ...[]int) []int {
	value := c.Get(key)
	if value == nil {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return nil
	}

	result, err := c.converter.ToIntSlice(value)
	if err != nil {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return nil
	}
	return result
}

// GetInt64Slice returns an int64 slice with optional default.
func (c *ConfyImpl) GetInt64Slice(key string, defaultValue ...[]int64) []int64 {
	value := c.Get(key)
	if value == nil {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return nil
	}

	result, err := c.converter.ToInt64Slice(value)
	if err != nil {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return nil
	}
	return result
}

// GetFloat64Slice returns a float64 slice with optional default.
func (c *ConfyImpl) GetFloat64Slice(key string, defaultValue ...[]float64) []float64 {
	value := c.Get(key)
	if value == nil {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return nil
	}

	result, err := c.converter.ToFloat64Slice(value)
	if err != nil {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return nil
	}
	return result
}

// GetBoolSlice returns a bool slice with optional default.
func (c *ConfyImpl) GetBoolSlice(key string, defaultValue ...[]bool) []bool {
	value := c.Get(key)
	if value == nil {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return nil
	}

	result, err := c.converter.ToBoolSlice(value)
	if err != nil {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return nil
	}

	return result
}

// GetStringMap returns a string map with optional default.
func (c *ConfyImpl) GetStringMap(key string, defaultValue ...map[string]string) map[string]string {
	value := c.Get(key)
	if value == nil {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}

		return nil
	}

	switch v := value.(type) {
	case map[string]string:
		return v
	case map[string]any:
		result := make(map[string]string)
		for k, val := range v {
			result[k] = fmt.Sprintf("%v", val)
		}

		return result
	case map[any]any:
		result := make(map[string]string)
		for k, val := range v {
			result[fmt.Sprintf("%v", k)] = fmt.Sprintf("%v", val)
		}

		return result
	}

	if len(defaultValue) > 0 {
		return defaultValue[0]
	}

	return nil
}

// GetStringMapStringSlice returns a map of string slices with optional default.
func (c *ConfyImpl) GetStringMapStringSlice(key string, defaultValue ...map[string][]string) map[string][]string {
	value := c.Get(key)
	if value == nil {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}

		return nil
	}

	switch v := value.(type) {
	case map[string][]string:
		return v
	case map[string]any:
		result := make(map[string][]string)

		for k, val := range v {
			switch slice := val.(type) {
			case []string:
				result[k] = slice
			case []any:
				strSlice := make([]string, len(slice))
				for i, item := range slice {
					strSlice[i] = fmt.Sprintf("%v", item)
				}

				result[k] = strSlice
			case string:
				result[k] = []string{slice}
			}
		}

		return result
	}

	if len(defaultValue) > 0 {
		return defaultValue[0]
	}

	return nil
}

// =============================================================================
// ADVANCED API - FUNCTIONAL OPTIONS
// =============================================================================

// GetWithOptions returns a value with advanced options.
func (c *ConfyImpl) GetWithOptions(key string, opts ...configcore.GetOption) (any, error) {
	options := &configcore.GetOptions{}
	for _, opt := range opts {
		opt(options)
	}

	value := c.Get(key)

	// Handle missing key
	if value == nil {
		if options.Required {
			return nil, ErrConfigError(fmt.Sprintf("required key '%s' not found", key), nil)
		}

		if options.OnMissing != nil {
			value = options.OnMissing(key)
		} else if options.Default != nil {
			return options.Default, nil
		} else {
			return nil, nil
		}
	}

	// Transform
	if options.Transform != nil {
		value = options.Transform(value)
	}

	// Validate
	if options.Validator != nil {
		if err := options.Validator(value); err != nil {
			return nil, ErrConfigError(fmt.Sprintf("validation failed for key '%s'", key), err)
		}
	}

	return value, nil
}

// GetStringWithOptions returns a string with advanced options.
func (c *ConfyImpl) GetStringWithOptions(key string, opts ...configcore.GetOption) (string, error) {
	options := &configcore.GetOptions{}
	for _, opt := range opts {
		opt(options)
	}

	value := c.Get(key)

	// Handle missing key
	if value == nil {
		if options.Required {
			return "", ErrConfigError(fmt.Sprintf("required key '%s' not found", key), nil)
		}

		if options.OnMissing != nil {
			value = options.OnMissing(key)
		} else if options.Default != nil {
			value = options.Default
		} else {
			return "", nil
		}
	}

	// Transform
	if options.Transform != nil {
		value = options.Transform(value)
	}

	// Convert to string
	result := c.converter.ToString(value)

	// Check empty
	if !options.AllowEmpty && result == "" {
		if options.Required {
			return "", ErrConfigError(fmt.Sprintf("key '%s' is empty", key), nil)
		}

		if options.Default != nil {
			if defaultStr, ok := options.Default.(string); ok {
				result = defaultStr
			}
		}
	}

	// Validate
	if options.Validator != nil {
		if err := options.Validator(result); err != nil {
			return "", ErrConfigError(fmt.Sprintf("validation failed for key '%s'", key), err)
		}
	}

	return result, nil
}

// GetIntWithOptions returns an int with advanced options.
func (c *ConfyImpl) GetIntWithOptions(key string, opts ...configcore.GetOption) (int, error) {
	options := &configcore.GetOptions{}
	for _, opt := range opts {
		opt(options)
	}

	value := c.Get(key)

	// Handle missing key
	if value == nil {
		if options.Required {
			return 0, ErrConfigError(fmt.Sprintf("required key '%s' not found", key), nil)
		}

		if options.OnMissing != nil {
			value = options.OnMissing(key)
		} else if options.Default != nil {
			if defaultInt, ok := options.Default.(int); ok {
				return defaultInt, nil
			}
		} else {
			return 0, nil
		}
	}

	// Transform
	if options.Transform != nil {
		value = options.Transform(value)
	}

	// Convert to int
	result, err := c.converter.ToInt64(value)
	if err != nil {
		if options.Default != nil {
			if defaultInt, ok := options.Default.(int); ok {
				result = int64(defaultInt)
			}
		} else {
			return 0, ErrConfigError(fmt.Sprintf("failed to convert key '%s' to int", key), err)
		}
	}

	// Validate
	if options.Validator != nil {
		if err := options.Validator(int(result)); err != nil {
			return 0, ErrConfigError(fmt.Sprintf("validation failed for key '%s'", key), err)
		}
	}

	return int(result), nil
}

// GetBoolWithOptions returns a bool with advanced options.
func (c *ConfyImpl) GetBoolWithOptions(key string, opts ...configcore.GetOption) (bool, error) {
	options := &configcore.GetOptions{}
	for _, opt := range opts {
		opt(options)
	}

	value := c.Get(key)

	// Handle missing key
	if value == nil {
		if options.Required {
			return false, ErrConfigError(fmt.Sprintf("required key '%s' not found", key), nil)
		}

		if options.OnMissing != nil {
			value = options.OnMissing(key)
		} else if options.Default != nil {
			if defaultBool, ok := options.Default.(bool); ok {
				return defaultBool, nil
			}
		} else {
			return false, nil
		}
	}

	// Transform
	if options.Transform != nil {
		value = options.Transform(value)
	}

	// Convert to bool
	result, err := c.converter.ToBool(value)
	if err != nil {
		if options.Default != nil {
			if defaultBool, ok := options.Default.(bool); ok {
				result = defaultBool
			}
		} else {
			return false, ErrConfigError(fmt.Sprintf("failed to convert key '%s' to bool", key), err)
		}
	}

	// Validate
	if options.Validator != nil {
		if err := options.Validator(result); err != nil {
			return false, ErrConfigError(fmt.Sprintf("validation failed for key '%s'", key), err)
		}
	}

	return result, nil
}

// GetDurationWithOptions returns a duration with advanced options.
func (c *ConfyImpl) GetDurationWithOptions(key string, opts ...configcore.GetOption) (time.Duration, error) {
	options := &configcore.GetOptions{}
	for _, opt := range opts {
		opt(options)
	}

	value := c.Get(key)

	// Handle missing key
	if value == nil {
		if options.Required {
			return 0, ErrConfigError(fmt.Sprintf("required key '%s' not found", key), nil)
		}

		if options.OnMissing != nil {
			value = options.OnMissing(key)
		} else if options.Default != nil {
			if defaultDur, ok := options.Default.(time.Duration); ok {
				return defaultDur, nil
			}
		} else {
			return 0, nil
		}
	}

	// Transform
	if options.Transform != nil {
		value = options.Transform(value)
	}

	// Convert to duration
	var result time.Duration

	switch v := value.(type) {
	case time.Duration:
		result = v
	case string:
		var err error

		result, err = time.ParseDuration(v)
		if err != nil {
			if options.Default != nil {
				if defaultDur, ok := options.Default.(time.Duration); ok {
					result = defaultDur
				}
			} else {
				return 0, ErrConfigError(fmt.Sprintf("failed to parse duration for key '%s'", key), err)
			}
		}
	case int, int64:
		if intVal, err := c.converter.ToInt64(v); err == nil {
			result = time.Duration(intVal) * time.Second
		}
	default:
		if options.Default != nil {
			if defaultDur, ok := options.Default.(time.Duration); ok {
				result = defaultDur
			}
		}
	}

	// Validate
	if options.Validator != nil {
		if err := options.Validator(result); err != nil {
			return 0, ErrConfigError(fmt.Sprintf("validation failed for key '%s'", key), err)
		}
	}

	return result, nil
}

// LoadFrom loads configuration from multiple sources.
func (c *ConfyImpl) LoadFrom(sources ...ConfigSource) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.logger != nil {
		c.logger.Info("loading configuration from sources",
			logger.Int("source_count", len(sources)),
		)
	}

	for _, source := range sources {
		if err := c.registry.RegisterSource(source); err != nil {
			return ErrConfigError("failed to register source "+source.Name(), err)
		}

		c.sources = append(c.sources, source)
	}

	if err := c.loadAllSources(context.Background()); err != nil {
		return err
	}

	if err := c.validator.ValidateAll(c.data); err != nil {
		return ErrConfigError("configuration validation failed", err)
	}

	if c.metrics != nil {
		c.metrics.Counter("config.sources_loaded").Add(float64(len(sources)))
		c.metrics.Gauge("config.active_sources").Set(float64(len(c.sources)))
		c.metrics.Gauge("config.keys_count").Set(float64(len(c.data)))
	}

	return nil
}

// Watch starts watching for configuration changes.
func (c *ConfyImpl) Watch(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.started {
		return ErrLifecycleError("watch", errors.New("configuration already watching"))
	}

	c.watchCtx, c.watchCancel = context.WithCancel(ctx)

	for _, source := range c.sources {
		if source.IsWatchable() {
			if err := c.watcher.WatchSource(c.watchCtx, source, c.handleConfigChange); err != nil {
				if c.logger != nil {
					c.logger.Error("failed to start watching source",
						logger.String("source", source.Name()),
						logger.Error(err),
					)
				}
			}
		}
	}

	c.started = true

	if c.logger != nil {
		c.logger.Info("configuration started watching")
	}

	if c.metrics != nil {
		c.metrics.Counter("config.watch_started").Inc()
	}

	return nil
}

// Reload forces a reload of all configuration sources.
func (c *ConfyImpl) Reload() error {
	return c.ReloadContext(context.Background())
}

// ReloadContext forces a reload with context.
func (c *ConfyImpl) ReloadContext(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.logger != nil {
		c.logger.Info("reloading configuration from all sources")
	}

	startTime := time.Now()

	if err := c.loadAllSources(ctx); err != nil {
		return err
	}

	if err := c.validator.ValidateAll(c.data); err != nil {
		return ErrConfigError("configuration validation failed after reload", err)
	}

	c.notifyWatchCallbacks()

	if c.metrics != nil {
		c.metrics.Counter("config.reloads").Inc()
		c.metrics.Histogram("config.reload_duration").Observe(time.Since(startTime).Seconds())
	}

	return nil
}

// Validate validates the current configuration.
func (c *ConfyImpl) Validate() error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.validator.ValidateAll(c.data)
}

// Set sets a configuration value.
func (c *ConfyImpl) Set(key string, value any) {
	c.mu.Lock()
	defer c.mu.Unlock()

	oldValue := c.getValue(key)
	c.setValue(key, value)

	change := ConfigChange{
		Source:    "manager",
		Type:      ChangeTypeSet,
		Key:       key,
		OldValue:  oldValue,
		NewValue:  value,
		Timestamp: time.Now(),
	}
	c.notifyChangeCallbacks(change)
	c.notifyWatchCallbacks()
}

// =============================================================================
// BINDING METHODS
// =============================================================================

// Bind binds configuration to a struct.
func (c *ConfyImpl) Bind(key string, target any) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var data any
	if key == "" {
		data = c.data
	} else {
		data = c.getValue(key)
	}

	if data == nil {
		return ErrConfigError(fmt.Sprintf("no configuration found for key '%s'", key), nil)
	}

	return c.bindValue(data, target)
}

// BindWithDefault binds with a default value.
func (c *ConfyImpl) BindWithDefault(key string, target any, defaultValue any) error {
	return c.BindWithOptions(key, target, configcore.BindOptions{
		DefaultValue:   defaultValue,
		UseDefaults:    true,
		TagName:        "yaml",
		DeepMerge:      true,
		ErrorOnMissing: false,
	})
}

// BindWithOptions binds with flexible options.
func (c *ConfyImpl) BindWithOptions(key string, target any, options configcore.BindOptions) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var data any
	if key == "" {
		data = c.data
	} else {
		data = c.getValue(key)
	}

	// Convert struct defaultValue to map if needed (before checking if data is nil)
	if options.DefaultValue != nil {
		defaultVal := reflect.ValueOf(options.DefaultValue)
		if defaultVal.Kind() == reflect.Struct || (defaultVal.Kind() == reflect.Ptr && defaultVal.Elem().Kind() == reflect.Struct) {
			if converted, err := c.structToMap(options.DefaultValue, options.TagName); err == nil {
				// Replace DefaultValue with converted map for proper deep merge
				options.DefaultValue = converted
			} else {
				return ErrConfigError(fmt.Sprintf("failed to convert struct defaultValue: %v", err), nil)
			}
		}
	}

	if data == nil {
		if options.DefaultValue != nil {
			data = options.DefaultValue
		} else if options.UseDefaults {
			data = make(map[string]any)
		} else {
			if options.ErrorOnMissing {
				return ErrConfigError(fmt.Sprintf("no configuration found for key '%s'", key), nil)
			}

			data = make(map[string]any)
		}
	}

	return c.bindValueWithOptions(data, target, options)
}

// =============================================================================
// WATCH AND CHANGE CALLBACKS
// =============================================================================

// WatchWithCallback registers a callback for key changes.
func (c *ConfyImpl) WatchWithCallback(key string, callback func(string, any)) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.watchCallbacks[key] == nil {
		c.watchCallbacks[key] = make([]func(string, any), 0)
	}

	c.watchCallbacks[key] = append(c.watchCallbacks[key], callback)
}

// WatchChanges registers a callback for all changes.
func (c *ConfyImpl) WatchChanges(callback func(ConfigChange)) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.changeCallbacks = append(c.changeCallbacks, callback)
}

// =============================================================================
// METADATA AND INTROSPECTION
// =============================================================================

// GetSourceMetadata returns metadata for all sources.
func (c *ConfyImpl) GetSourceMetadata() map[string]*SourceMetadata {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.registry.GetAllMetadata()
}

// GetKeys returns all configuration keys.
func (c *ConfyImpl) GetKeys() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.getAllKeys(c.data, "")
}

// GetSection returns a configuration section.
func (c *ConfyImpl) GetSection(key string) map[string]any {
	value := c.Get(key)
	if value == nil {
		return nil
	}

	if section, ok := value.(map[string]any); ok {
		return section
	}

	return nil
}

// HasKey checks if a key exists.
func (c *ConfyImpl) HasKey(key string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.getValue(key) != nil
}

// IsSet checks if a key is set and not empty.
func (c *ConfyImpl) IsSet(key string) bool {
	value := c.Get(key)
	if value == nil {
		return false
	}

	switch v := value.(type) {
	case string:
		return v != ""
	case []any:
		return len(v) > 0
	case map[string]any:
		return len(v) > 0
	default:
		return true
	}
}

// Size returns the number of keys.
func (c *ConfyImpl) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.getAllKeys(c.data, ""))
}

// =============================================================================
// STRUCTURE OPERATIONS
// =============================================================================

// Sub returns a sub-configuration.
func (c *ConfyImpl) Sub(key string) Confy {
	subData := c.GetSection(key)
	if subData == nil {
		subData = make(map[string]any)
	}

	subManager := &ConfyImpl{
		data:            subData,
		watchCallbacks:  make(map[string][]func(string, any)),
		changeCallbacks: make([]func(ConfigChange), 0),
		logger:          c.logger,
		metrics:         c.metrics,
		errorHandler:    c.errorHandler,
	}

	subManager.registry = NewSourceRegistry(subManager.logger)
	subManager.validator = NewValidator(ValidatorConfig{
		Mode:         ValidationModePermissive,
		Logger:       subManager.logger,
		ErrorHandler: subManager.errorHandler,
	})

	return subManager
}

// MergeWith merges another Confy instance.
func (c *ConfyImpl) MergeWith(other Confy) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if otherImpl, ok := other.(*ConfyImpl); ok {
		otherImpl.mu.RLock()
		defer otherImpl.mu.RUnlock()

		c.mergeData(c.data, otherImpl.data)

		return nil
	}

	return errors.New("merge not supported for this Confy implementation")
}

// Clone creates a deep copy.
func (c *ConfyImpl) Clone() Confy {
	c.mu.RLock()
	defer c.mu.RUnlock()

	clonedData := c.merger.DeepCopy(c.data)

	cloned := &ConfyImpl{
		data:            clonedData,
		watchCallbacks:  make(map[string][]func(string, any)),
		changeCallbacks: make([]func(ConfigChange), 0),
		logger:          c.logger,
		metrics:         c.metrics,
		errorHandler:    c.errorHandler,
	}

	cloned.registry = NewSourceRegistry(cloned.logger)
	cloned.validator = NewValidator(ValidatorConfig{
		Mode:         ValidationModePermissive,
		Logger:       cloned.logger,
		ErrorHandler: cloned.errorHandler,
	})

	return cloned
}

// GetAllSettings returns all settings.
func (c *ConfyImpl) GetAllSettings() map[string]any {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.merger.DeepCopy(c.data)
}

// =============================================================================
// UTILITY METHODS
// =============================================================================

// Reset clears all configuration.
func (c *ConfyImpl) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data = make(map[string]any)
	c.watchCallbacks = make(map[string][]func(string, any))
	c.changeCallbacks = make([]func(ConfigChange), 0)

	if c.logger != nil {
		c.logger.Info("configuration reset")
	}

	if c.metrics != nil {
		c.metrics.Counter("config.reset").Inc()
		c.metrics.Gauge("config.keys_count").Set(0)
	}
}

// ExpandEnvVars expands environment variables.
func (c *ConfyImpl) ExpandEnvVars() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.expandEnvInMap(c.data)

	return nil
}

// SafeGet returns a value with type checking.
func (c *ConfyImpl) SafeGet(key string, expectedType reflect.Type) (any, error) {
	value := c.Get(key)
	if value == nil {
		return nil, fmt.Errorf("key '%s' not found", key)
	}

	valueType := reflect.TypeOf(value)
	if valueType != expectedType {
		return nil, fmt.Errorf("key '%s' expected type %v, got %v", key, expectedType, valueType)
	}

	return value, nil
}

// Stop stops the configuration.
func (c *ConfyImpl) Stop() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.started {
		return nil
	}

	if c.watchCancel != nil {
		c.watchCancel()
	}

	for _, source := range c.sources {
		if err := source.StopWatch(); err != nil {
			if c.logger != nil {
				c.logger.Error("failed to stop watching source",
					logger.String("source", source.Name()),
					logger.Error(err),
				)
			}
		}
	}

	c.started = false

	if c.logger != nil {
		c.logger.Info("configuration stopped")
	}

	if c.metrics != nil {
		c.metrics.Counter("config.watch_stopped").Inc()
	}

	return nil
}

// ConfigFileUsed returns the config file path.
func (c *ConfyImpl) ConfigFileUsed() string {
	sources := c.registry.GetSources()
	for _, source := range sources {
		if fileSource, ok := source.(interface {
			FilePath() string
		}); ok {
			return fileSource.FilePath()
		}
	}

	return ""
}

// =============================================================================
// INTERNAL HELPER METHODS
// =============================================================================

func (c *ConfyImpl) loadAllSources(ctx context.Context) error {
	mergedData := make(map[string]any)

	sources := c.registry.GetSources()

	// Sort sources by priority (lower number = lower priority, loaded first)
	// This ensures higher priority sources override lower priority ones
	type prioritySource struct {
		priority int
		source   ConfigSource
	}

	prioritySources := make([]prioritySource, 0, len(sources))
	for _, source := range sources {
		prioritySources = append(prioritySources, prioritySource{
			priority: source.Priority(),
			source:   source,
		})
	}

	// Sort by priority (ascending) using sort.Slice for O(n log n) performance
	sort.Slice(prioritySources, func(i, j int) bool {
		return prioritySources[i].priority < prioritySources[j].priority
	})

	// Load sources in priority order (lower priority first, so higher priority can override)
	for _, ps := range prioritySources {
		data, err := c.loader.LoadSource(ctx, ps.source)
		if err != nil {
			if c.errorHandler != nil {
				// nolint:gosec // G104: Error handler intentionally discards return value
				_ = c.errorHandler.HandleError(context.Background(), err)
			}

			return ErrConfigError("failed to load source "+ps.source.Name(), err)
		}

		c.mergeData(mergedData, data)
	}

	c.data = mergedData

	return nil
}

func (c *ConfyImpl) handleConfigChange(source string, data map[string]any) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.logger != nil {
		c.logger.Info("configuration change detected",
			logger.String("source", source),
			logger.Int("keys", len(data)),
		)
	}

	oldData := make(map[string]any)
	maps.Copy(oldData, c.data)

	c.mergeData(c.data, data)

	if err := c.validator.ValidateAll(c.data); err != nil {
		if c.logger != nil {
			c.logger.Error("configuration validation failed after change",
				logger.String("source", source),
				logger.Error(err),
			)
		}

		if c.validator.IsStrictMode() {
			c.data = oldData

			return
		}
	}

	change := ConfigChange{
		Source:    source,
		Type:      ChangeTypeUpdate,
		Timestamp: time.Now(),
	}
	c.notifyChangeCallbacks(change)
	c.notifyWatchCallbacks()

	if c.metrics != nil {
		c.metrics.Counter("config.changes_applied").Inc()
	}
}

func (c *ConfyImpl) getValue(key string) any {
	keys := strings.Split(key, ".")
	current := any(c.data)

	for _, k := range keys {
		if current == nil {
			return nil
		}

		switch v := current.(type) {
		case map[string]any:
			current = v[k]
		case map[any]any:
			current = v[k]
		default:
			return nil
		}
	}

	return current
}

func (c *ConfyImpl) setValue(key string, value any) {
	keys := strings.Split(key, ".")
	current := c.data

	for i, k := range keys {
		if i == len(keys)-1 {
			current[k] = value
		} else {
			if current[k] == nil {
				current[k] = make(map[string]any)
			}

			if next, ok := current[k].(map[string]any); ok {
				current = next
			} else {
				current[k] = make(map[string]any)
				current = current[k].(map[string]any)
			}
		}
	}
}

func (c *ConfyImpl) mergeData(target, source map[string]any) {
	c.merger.MergeInPlace(target, source)
}

// structToMap converts a struct to map[string]any using struct tags
// Supports yaml tags (preferred) and json tags as fallback, with optional custom tagName.
func (c *ConfyImpl) structToMap(v any, tagName string) (map[string]any, error) {
	val := reflect.ValueOf(v)

	// Handle pointer to struct
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return nil, errors.New("cannot convert nil pointer to map")
		}

		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return nil, fmt.Errorf("value must be a struct, got %s", val.Kind())
	}

	result := make(map[string]any)
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		fieldVal := val.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Get field name from tags (yaml takes precedence over json)
		fieldName := field.Name

		// Try yaml tag first
		if yamlTag := field.Tag.Get("yaml"); yamlTag != "" {
			if idx := strings.Index(yamlTag, ","); idx != -1 {
				fieldName = yamlTag[:idx]
			} else {
				fieldName = yamlTag
			}

			if fieldName == "-" {
				continue
			}
		} else if jsonTag := field.Tag.Get("json"); jsonTag != "" {
			// Fallback to json tag
			if idx := strings.Index(jsonTag, ","); idx != -1 {
				fieldName = jsonTag[:idx]
			} else {
				fieldName = jsonTag
			}

			if fieldName == "-" {
				continue
			}
		}

		// If using custom tagName from options (not yaml/json), respect it
		if tagName != "" && tagName != "yaml" && tagName != "json" {
			if customTag := field.Tag.Get(tagName); customTag != "" {
				if idx := strings.Index(customTag, ","); idx != -1 {
					fieldName = customTag[:idx]
				} else {
					fieldName = customTag
				}

				if fieldName == "-" {
					continue
				}
			}
		}

		// Handle nested structs recursively
		if fieldVal.Kind() == reflect.Struct {
			nested, err := c.structToMap(fieldVal.Interface(), tagName)
			if err == nil {
				result[fieldName] = nested

				continue
			}
		}

		// Set the value
		result[fieldName] = fieldVal.Interface()
	}

	return result, nil
}

func (c *ConfyImpl) bindValue(value any, target any) error {
	targetValue := reflect.ValueOf(target)
	if targetValue.Kind() != reflect.Ptr {
		return ErrConfigError("target must be a pointer", nil)
	}

	targetElem := targetValue.Elem()
	sourceValue := reflect.ValueOf(value)

	// Handle simple type binding (string, int, etc.)
	if targetElem.Kind() != reflect.Struct {
		if sourceValue.Type().AssignableTo(targetElem.Type()) {
			targetElem.Set(sourceValue)

			return nil
		}
		// Try to convert if possible
		if sourceValue.Type().ConvertibleTo(targetElem.Type()) {
			targetElem.Set(sourceValue.Convert(targetElem.Type()))

			return nil
		}

		return ErrConfigError(fmt.Sprintf("cannot convert %s to %s", sourceValue.Type(), targetElem.Type()), nil)
	}

	// Handle struct binding
	if sourceValue.Kind() == reflect.Map {
		return c.bindMapToStruct(sourceValue, targetElem)
	}

	return ErrConfigError("unsupported value type for binding", nil)
}

func (c *ConfyImpl) bindMapToStruct(mapValue reflect.Value, structValue reflect.Value) error {
	structType := structValue.Type()

	// First apply struct tag defaults
	if err := c.applyStructDefaults(structValue); err != nil {
		return err
	}

	for i := 0; i < structValue.NumField(); i++ {
		field := structValue.Field(i)
		fieldType := structType.Field(i)

		if !field.CanSet() {
			continue
		}

		fieldName := c.getFieldName(fieldType)
		if fieldName == "" {
			continue
		}

		mapKey := reflect.ValueOf(fieldName)
		mapVal := mapValue.MapIndex(mapKey)

		if !mapVal.IsValid() {
			// Check if field is required
			if fieldType.Tag.Get("required") == "true" {
				return ErrConfigError(fmt.Sprintf("required field '%s' is missing", fieldName), nil)
			}
			continue
		}

		if err := c.setFieldValue(field, mapVal); err != nil {
			return err
		}
	}

	return nil
}

func (c *ConfyImpl) getFieldName(field reflect.StructField) string {
	if tag := field.Tag.Get("yaml"); tag != "" && tag != "-" {
		return strings.Split(tag, ",")[0]
	}

	if tag := field.Tag.Get("json"); tag != "" && tag != "-" {
		return strings.Split(tag, ",")[0]
	}

	if tag := field.Tag.Get("config"); tag != "" && tag != "-" {
		return strings.Split(tag, ",")[0]
	}

	return field.Name
}

func (c *ConfyImpl) setFieldValue(field reflect.Value, value reflect.Value) error {
	if !value.IsValid() {
		return nil
	}

	valueInterface := value.Interface()

	switch field.Kind() {
	case reflect.String:
		field.SetString(c.converter.ToString(valueInterface))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		// Handle time.Duration specially since it's an int64
		if field.Type() == reflect.TypeOf(time.Duration(0)) {
			if durVal, err := c.converter.ToDuration(valueInterface); err == nil {
				field.SetInt(int64(durVal))
			}
		} else {
			if intVal, err := c.converter.ToInt64(valueInterface); err == nil {
				field.SetInt(intVal)
			}
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if uintVal, err := c.converter.ToUint64(valueInterface); err == nil {
			field.SetUint(uintVal)
		}
	case reflect.Float32, reflect.Float64:
		if floatVal, err := c.converter.ToFloat64(valueInterface); err == nil {
			field.SetFloat(floatVal)
		}
	case reflect.Bool:
		if boolVal, err := c.converter.ToBool(valueInterface); err == nil {
			field.SetBool(boolVal)
		}
	case reflect.Slice:
		if slice, ok := valueInterface.([]any); ok {
			return c.setSliceValue(field, slice)
		}
	case reflect.Map:
		if mapVal, ok := valueInterface.(map[string]any); ok {
			return c.setMapValue(field, mapVal)
		}
	case reflect.Struct:
		if mapVal, ok := valueInterface.(map[string]any); ok {
			return c.bindMapToStruct(reflect.ValueOf(mapVal), field)
		}
	case reflect.Ptr:
		if field.IsNil() {
			field.Set(reflect.New(field.Type().Elem()))
		}

		return c.setFieldValue(field.Elem(), value)
	}

	return nil
}

func (c *ConfyImpl) setSliceValue(field reflect.Value, slice []any) error {
	sliceValue := reflect.MakeSlice(field.Type(), len(slice), len(slice))
	for i, item := range slice {
		if err := c.setFieldValue(sliceValue.Index(i), reflect.ValueOf(item)); err != nil {
			return err
		}
	}

	field.Set(sliceValue)

	return nil
}

func (c *ConfyImpl) setMapValue(field reflect.Value, mapData map[string]any) error {
	mapValue := reflect.MakeMap(field.Type())
	mapValueType := field.Type().Elem()

	for key, value := range mapData {
		keyValue := reflect.ValueOf(key)

		// Convert value to the correct type for the map's value type
		var convertedValue reflect.Value

		// Check if the map value type is a struct
		if mapValueType.Kind() == reflect.Struct {
			// Create a new instance of the struct type
			structInstance := reflect.New(mapValueType).Elem()

			// If the value is a map[string]interface{}, bind it to the struct
			if valueMap, ok := value.(map[string]any); ok {
				if err := c.bindMapToStruct(reflect.ValueOf(valueMap), structInstance); err != nil {
					return fmt.Errorf("failed to bind map value for key '%s': %w", key, err)
				}

				convertedValue = structInstance
			} else {
				// Direct assignment if types match
				convertedValue = reflect.ValueOf(value)
			}
		} else if mapValueType.Kind() == reflect.Ptr && mapValueType.Elem().Kind() == reflect.Struct {
			// Handle pointer to struct
			structInstance := reflect.New(mapValueType.Elem())

			if valueMap, ok := value.(map[string]any); ok {
				if err := c.bindMapToStruct(reflect.ValueOf(valueMap), structInstance.Elem()); err != nil {
					return fmt.Errorf("failed to bind map value for key '%s': %w", key, err)
				}

				convertedValue = structInstance
			} else {
				convertedValue = reflect.ValueOf(value)
			}
		} else {
			// For primitive types or interfaces, try direct conversion
			convertedValue = reflect.ValueOf(value)

			// If types don't match, try to convert
			if convertedValue.Type() != mapValueType {
				// Try type conversion if possible
				if convertedValue.Type().ConvertibleTo(mapValueType) {
					convertedValue = convertedValue.Convert(mapValueType)
				} else {
					// If it's still a map[string]interface{} and we need a different type,
					// we need to recursively bind it
					if _, ok := value.(map[string]any); ok {
						newValue := reflect.New(mapValueType).Elem()
						if err := c.setFieldValue(newValue, convertedValue); err != nil {
							return fmt.Errorf("failed to convert map value for key '%s': %w", key, err)
						}

						convertedValue = newValue
					}
				}
			}
		}

		mapValue.SetMapIndex(keyValue, convertedValue)
	}

	field.Set(mapValue)

	return nil
}

func (c *ConfyImpl) bindValueWithOptions(value any, target any, options configcore.BindOptions) error {
	targetValue := reflect.ValueOf(target)
	if targetValue.Kind() != reflect.Ptr {
		return ErrConfigError("target must be a pointer", nil)
	}

	targetElem := targetValue.Elem()

	// Handle primitive target types
	if targetElem.Kind() != reflect.Struct {
		sourceValue := reflect.ValueOf(value)
		if sourceValue.Type().AssignableTo(targetElem.Type()) {
			targetElem.Set(sourceValue)

			return nil
		}

		if sourceValue.Type().ConvertibleTo(targetElem.Type()) {
			targetElem.Set(sourceValue.Convert(targetElem.Type()))

			return nil
		}

		return ErrConfigError(fmt.Sprintf("cannot convert %s to %s", sourceValue.Type(), targetElem.Type()), nil)
	}

	// Handle struct target (existing logic continues...)
	targetStruct := targetElem

	// Apply struct tag defaults (lowest precedence)
	if err := c.applyStructDefaults(targetStruct); err != nil {
		return err
	}

	// Apply passed default value (medium precedence)
	if options.DefaultValue != nil {
		if defaultMap, ok := options.DefaultValue.(map[string]any); ok {
			if options.DeepMerge {
				value = c.deepMergeValues(defaultMap, value)
			}
		}
	}

	// Apply config file values (highest precedence)
	sourceValue := reflect.ValueOf(value)
	if sourceValue.Kind() == reflect.Map {
		return c.bindMapToStructWithOptions(sourceValue, targetStruct, options)
	}

	return ErrConfigError("unsupported value type for binding", nil)
}

func (c *ConfyImpl) bindMapToStructWithOptions(mapValue reflect.Value, structValue reflect.Value, options configcore.BindOptions) error {
	structType := structValue.Type()

	// Track required fields
	requiredFields := make(map[string]bool)
	for _, field := range options.Required {
		requiredFields[field] = false
	}

	for i := 0; i < structValue.NumField(); i++ {
		field := structValue.Field(i)
		fieldType := structType.Field(i)

		if !field.CanSet() {
			continue
		}

		// Get field name from tags
		fieldName := c.getFieldNameWithOptions(fieldType, options)
		if fieldName == "" {
			continue
		}

		// Mark required field as potentially found
		if _, isRequired := requiredFields[fieldName]; isRequired {
			requiredFields[fieldName] = true
		}

		// Get value from config map
		var mapVal reflect.Value
		if options.IgnoreCase {
			mapVal = c.findMapValueIgnoreCase(mapValue, fieldName)
		} else {
			mapKey := reflect.ValueOf(fieldName)
			mapVal = mapValue.MapIndex(mapKey)
		}
		// Handle missing values with proper precedence
		if !mapVal.IsValid() {
			// Check required fields
			if _, isRequired := requiredFields[fieldName]; isRequired {
				if options.ErrorOnMissing {
					return ErrConfigError(fmt.Sprintf("required field '%s' not found", fieldName), nil)
				}
			}

			// Field not in config, keep existing value (could be from struct tag default or passed default)
			if options.UseDefaults {
				continue
			}

			continue
		}

		// Set field value with deep merge support
		if err := c.setFieldValueWithDeepMerge(field, mapVal, fieldType, options); err != nil {
			return err
		}
	}

	// Validate all required fields were found
	for fieldName, found := range requiredFields {
		if !found && options.ErrorOnMissing {
			return ErrConfigError(fmt.Sprintf("required field '%s' not found in configuration", fieldName), nil)
		}
	}

	return nil
}

func (c *ConfyImpl) getFieldNameWithOptions(field reflect.StructField, options configcore.BindOptions) string {
	tagName := options.TagName
	if tagName == "" {
		tagName = "yaml"
	}

	if tag := field.Tag.Get(tagName); tag != "" && tag != "-" {
		return strings.Split(tag, ",")[0]
	}

	if tagName != "yaml" {
		if tag := field.Tag.Get("yaml"); tag != "" && tag != "-" {
			return strings.Split(tag, ",")[0]
		}
	}

	if tagName != "json" {
		if tag := field.Tag.Get("json"); tag != "" && tag != "-" {
			return strings.Split(tag, ",")[0]
		}
	}

	return field.Name
}

func (c *ConfyImpl) findMapValueIgnoreCase(mapValue reflect.Value, fieldName string) reflect.Value {
	fieldNameLower := strings.ToLower(fieldName)

	for _, key := range mapValue.MapKeys() {
		if keyStr, ok := key.Interface().(string); ok {
			if strings.ToLower(keyStr) == fieldNameLower {
				return mapValue.MapIndex(key)
			}
		}
	}

	return reflect.Value{}
}

// deepMergeValues deeply merges two values with proper precedence
// configValue (from file) takes precedence over defaultValue.
func (c *ConfyImpl) deepMergeValues(defaultValue, configValue any) any {
	// If config value is nil, use default
	if configValue == nil {
		return defaultValue
	}

	// If default is nil, use config
	if defaultValue == nil {
		return configValue
	}

	// Both are maps - use merger
	defaultMap, defaultIsMap := defaultValue.(map[string]any)
	configMap, configIsMap := configValue.(map[string]any)

	if defaultIsMap && configIsMap {
		return c.merger.DeepMerge(defaultMap, configMap)
	}

	// For non-map values, config takes precedence
	return configValue
}

// applyStructDefaults applies default values from struct tags.
func (c *ConfyImpl) applyStructDefaults(structValue reflect.Value) error {
	structType := structValue.Type()

	for i := 0; i < structValue.NumField(); i++ {
		field := structValue.Field(i)
		fieldType := structType.Field(i)

		if !field.CanSet() {
			continue
		}

		// Check for default tag
		defaultTag := fieldType.Tag.Get("default")
		if defaultTag == "" || defaultTag == "-" {
			// Recursively apply defaults to nested structs
			if field.Kind() == reflect.Struct {
				if err := c.applyStructDefaults(field); err != nil {
					return err
				}
			}

			continue
		}

		// Only apply default if field is zero value
		if !field.IsZero() {
			continue
		}
		// Parse and set default value based on field type
		if err := c.setDefaultValue(field, defaultTag, fieldType); err != nil {
			return ErrConfigError(
				fmt.Sprintf("failed to set default for field '%s'", fieldType.Name),
				err,
			)
		}
	}

	return nil
}

// setDefaultValue sets a field value from a default tag string.
func (c *ConfyImpl) setDefaultValue(field reflect.Value, defaultTag string, fieldType reflect.StructField) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(defaultTag)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		// Check if it's a duration
		if field.Type() == reflect.TypeFor[time.Duration]() {
			if d, err := time.ParseDuration(defaultTag); err == nil {
				field.SetInt(int64(d))
			} else {
				return fmt.Errorf("invalid duration default: %s", defaultTag)
			}
		} else {
			if intVal, err := strconv.ParseInt(defaultTag, 10, 64); err == nil {
				field.SetInt(intVal)
			} else {
				return fmt.Errorf("invalid int default: %s", defaultTag)
			}
		}

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if uintVal, err := strconv.ParseUint(defaultTag, 10, 64); err == nil {
			field.SetUint(uintVal)
		} else {
			return fmt.Errorf("invalid uint default: %s", defaultTag)
		}

	case reflect.Float32, reflect.Float64:
		if floatVal, err := strconv.ParseFloat(defaultTag, 64); err == nil {
			field.SetFloat(floatVal)
		} else {
			return fmt.Errorf("invalid float default: %s", defaultTag)
		}

	case reflect.Bool:
		if boolVal, err := strconv.ParseBool(defaultTag); err == nil {
			field.SetBool(boolVal)
		} else {
			return fmt.Errorf("invalid bool default: %s", defaultTag)
		}

	case reflect.Slice:
		// Handle slice defaults (comma-separated)
		if field.Type().Elem().Kind() == reflect.String {
			values := strings.Split(defaultTag, ",")

			slice := reflect.MakeSlice(field.Type(), len(values), len(values))
			for i, val := range values {
				slice.Index(i).SetString(strings.TrimSpace(val))
			}

			field.Set(slice)
		} else {
			return errors.New("slice defaults only supported for []string")
		}

	case reflect.Struct:
		// Handle time.Time
		if field.Type() == reflect.TypeFor[time.Time]() {
			formats := []string{
				time.RFC3339,
				time.RFC3339Nano,
				"2006-01-02 15:04:05",
				"2006-01-02T15:04:05",
				"2006-01-02",
			}
			for _, format := range formats {
				if t, err := time.Parse(format, defaultTag); err == nil {
					field.Set(reflect.ValueOf(t))

					return nil
				}
			}

			return fmt.Errorf("invalid time default: %s", defaultTag)
		}

	default:
		return fmt.Errorf("unsupported default type: %v", field.Kind())
	}

	return nil
}

// setFieldValueWithDeepMerge sets field with deep merge support for nested structs.
func (c *ConfyImpl) setFieldValueWithDeepMerge(field reflect.Value, value reflect.Value, fieldType reflect.StructField, options configcore.BindOptions) error {
	if !value.IsValid() {
		return nil
	}

	valueInterface := value.Interface()

	switch field.Kind() {
	case reflect.String:
		field.SetString(c.converter.ToString(valueInterface))

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if intVal, err := c.converter.ToInt64(valueInterface); err == nil {
			field.SetInt(intVal)
		} else if !options.ErrorOnMissing {
			return nil
		} else {
			return err
		}

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if uintVal, err := c.converter.ToUint64(valueInterface); err == nil {
			field.SetUint(uintVal)
		} else if !options.ErrorOnMissing {
			return nil
		} else {
			return err
		}

	case reflect.Float32, reflect.Float64:
		if floatVal, err := c.converter.ToFloat64(valueInterface); err == nil {
			field.SetFloat(floatVal)
		} else if !options.ErrorOnMissing {
			return nil
		} else {
			return err
		}

	case reflect.Bool:
		if boolVal, err := c.converter.ToBool(valueInterface); err == nil {
			field.SetBool(boolVal)
		} else if !options.ErrorOnMissing {
			return nil
		} else {
			return err
		}

	case reflect.Slice:
		if slice, ok := valueInterface.([]any); ok {
			return c.setSliceValue(field, slice)
		}

	case reflect.Map:
		if mapVal, ok := valueInterface.(map[string]any); ok {
			if options.DeepMerge && !field.IsZero() {
				// Deep merge with existing map
				return c.mergeMapValue(field, mapVal, options)
			}

			return c.setMapValue(field, mapVal)
		}

	case reflect.Struct:
		if mapVal, ok := valueInterface.(map[string]any); ok {
			if options.DeepMerge && !field.IsZero() {
				// Deep merge with existing struct
				return c.mergeStructValue(field, mapVal, options)
			}

			return c.bindMapToStructWithOptions(reflect.ValueOf(mapVal), field, options)
		}

	case reflect.Ptr:
		if field.IsNil() {
			field.Set(reflect.New(field.Type().Elem()))
		}

		return c.setFieldValueWithDeepMerge(field.Elem(), value, fieldType, options)
	}

	return nil
}

// mergeMapValue deeply merges a map into an existing field.
func (c *ConfyImpl) mergeMapValue(field reflect.Value, newData map[string]any, options configcore.BindOptions) error {
	if field.IsNil() {
		return c.setMapValue(field, newData)
	}

	mapValueType := field.Type().Elem()

	// Create merged map
	merged := reflect.MakeMap(field.Type())

	// Copy existing values
	for _, key := range field.MapKeys() {
		merged.SetMapIndex(key, field.MapIndex(key))
	}

	// Merge new values
	for key, value := range newData {
		keyValue := reflect.ValueOf(key)

		// Convert value to correct type
		var convertedValue reflect.Value

		if mapValueType.Kind() == reflect.Struct {
			// Create new struct instance
			structInstance := reflect.New(mapValueType).Elem()

			if valueMap, ok := value.(map[string]any); ok {
				// Check if we should deep merge with existing
				if existingValue := field.MapIndex(keyValue); existingValue.IsValid() && options.DeepMerge {
					// Deep merge existing struct with new data
					if err := c.mergeStructValue(existingValue, valueMap, options); err != nil {
						return err
					}

					merged.SetMapIndex(keyValue, existingValue)

					continue
				} else {
					// Bind new struct
					if err := c.bindMapToStructWithOptions(reflect.ValueOf(valueMap), structInstance, options); err != nil {
						return fmt.Errorf("failed to bind map value for key '%v': %w", key, err)
					}

					convertedValue = structInstance
				}
			} else {
				convertedValue = reflect.ValueOf(value)
			}
		} else {
			convertedValue = reflect.ValueOf(value)

			// Convert if types don't match
			if convertedValue.Type() != mapValueType {
				if convertedValue.Type().ConvertibleTo(mapValueType) {
					convertedValue = convertedValue.Convert(mapValueType)
				}
			}
		}

		merged.SetMapIndex(keyValue, convertedValue)
	}

	field.Set(merged)

	return nil
}

func (c *ConfyImpl) mergeStructValue(structField reflect.Value, mapData map[string]any, options configcore.BindOptions) error {
	// Extract current struct values to map
	currentData := make(map[string]any)
	structType := structField.Type()

	for i := 0; i < structField.NumField(); i++ {
		field := structField.Field(i)
		fieldType := structType.Field(i)

		if !field.CanInterface() {
			continue
		}

		fieldName := c.getFieldNameWithOptions(fieldType, options)
		if fieldName != "" && !field.IsZero() {
			currentData[fieldName] = field.Interface()
		}
	}

	// Deep merge current with new data (new data takes precedence)
	mergedData := c.deepMergeValues(currentData, mapData)

	// Bind merged data back to struct
	if mergedMap, ok := mergedData.(map[string]any); ok {
		return c.bindMapToStructWithOptions(reflect.ValueOf(mergedMap), structField, options)
	}

	return nil
}

func (c *ConfyImpl) getAllKeys(data any, prefix string) []string {
	var keys []string

	if mapData, ok := data.(map[string]any); ok {
		for key, value := range mapData {
			fullKey := key
			if prefix != "" {
				fullKey = prefix + "." + key
			}

			keys = append(keys, fullKey)
			nestedKeys := c.getAllKeys(value, fullKey)
			keys = append(keys, nestedKeys...)
		}
	}

	return keys
}

func (c *ConfyImpl) expandEnvInMap(data map[string]any) {
	for key, value := range data {
		switch v := value.(type) {
		case string:
			data[key] = c.expandEnvInString(v)
		case map[string]any:
			c.expandEnvInMap(v)
		case []any:
			c.expandEnvInSlice(v)
		}
	}
}

func (c *ConfyImpl) expandEnvInSlice(slice []any) {
	for i, value := range slice {
		switch v := value.(type) {
		case string:
			slice[i] = c.expandEnvInString(v)
		case map[string]any:
			c.expandEnvInMap(v)
		case []any:
			c.expandEnvInSlice(v)
		}
	}
}

func (c *ConfyImpl) expandEnvInString(s string) string {
	return os.Expand(s, os.Getenv)
}

func (c *ConfyImpl) notifyWatchCallbacks() {
	for key, callbacks := range c.watchCallbacks {
		value := c.getValue(key)
		for _, callback := range callbacks {
			go callback(key, value)
		}
	}
}

func (c *ConfyImpl) notifyChangeCallbacks(change ConfigChange) {
	for _, callback := range c.changeCallbacks {
		go callback(change)
	}
}
