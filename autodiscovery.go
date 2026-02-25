package confy

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	configcore "github.com/xraph/confy/internal"
	"github.com/xraph/confy/sources"
	errors "github.com/xraph/go-utils/errs"
	logger "github.com/xraph/go-utils/log"
)

// F is a helper function to create logger fields.
func F(key string, value any) logger.Field {
	return logger.Any(key, value)
}

// Priority constants for configuration sources.
const (
	PriorityEnvLow      = 50  // Environment variables with lower priority (files override env)
	PriorityBaseConfig  = 100 // Base configuration file priority
	PriorityLocalConfig = 200 // Local configuration file priority (overrides base)
	PriorityEnvHigh     = 300 // Environment variables with higher priority (env overrides files)
)

// AutoDiscoveryConfig configures automatic config file discovery.
type AutoDiscoveryConfig struct {
	// AppName is the application name to look for in app-scoped configs
	// If provided, will look for "apps.{AppName}" section in config
	AppName string

	// SearchPaths are directories to search for config files
	// Defaults to current directory and parent directories
	SearchPaths []string

	// ConfigNames are the config file names to search for
	// Defaults to ["config.yaml", "config.yml"]
	ConfigNames []string

	// LocalConfigNames are the local override config file names
	// Defaults to ["config.local.yaml", "config.local.yml"]
	LocalConfigNames []string

	// MaxDepth is the maximum number of parent directories to search
	// Defaults to 5
	MaxDepth int

	// RequireBase determines if base config file is required
	// Defaults to false
	RequireBase bool

	// RequireLocal determines if local config file is required
	// Defaults to false
	RequireLocal bool

	// EnableAppScoping enables app-scoped config extraction
	// If true and AppName is set, will extract "apps.{AppName}" section
	// Defaults to true
	EnableAppScoping bool

	// Environment Variable Source Configuration
	// EnableEnvSource enables loading config from environment variables
	// Defaults to true
	EnableEnvSource bool

	// EnvPrefix is the prefix for environment variables
	// If empty, defaults to AppName uppercase with trailing underscore
	EnvPrefix string

	// EnvSeparator is the separator for nested keys in env vars
	// Defaults to "_"
	EnvSeparator string

	// EnvOverridesFile controls whether env vars override file config values
	// When true, env source gets higher priority than file sources
	// Defaults to true
	EnvOverridesFile bool

	// Logger for discovery operations
	Logger logger.Logger

	// ErrorHandler for error handling
	ErrorHandler errors.ErrorHandler
}

// AutoDiscoveryResult contains the result of config discovery.
type AutoDiscoveryResult struct {
	// BaseConfigPaths contains all discovered base config file paths (ordered by search path).
	BaseConfigPaths []string

	// LocalConfigPaths contains all discovered local config file paths (ordered by search path).
	LocalConfigPaths []string

	// WorkingDirectory is the directory where the first config was found.
	WorkingDirectory string

	// IsMonorepo indicates if this is a monorepo layout.
	IsMonorepo bool

	// AppName is the app name for app-scoped configs.
	AppName string
}

// BaseConfigPath returns the first base config path, or "" if none were found.
func (r *AutoDiscoveryResult) BaseConfigPath() string {
	if len(r.BaseConfigPaths) > 0 {
		return r.BaseConfigPaths[0]
	}
	return ""
}

// LocalConfigPath returns the first local config path, or "" if none were found.
func (r *AutoDiscoveryResult) LocalConfigPath() string {
	if len(r.LocalConfigPaths) > 0 {
		return r.LocalConfigPaths[0]
	}
	return ""
}

// DefaultAutoDiscoveryConfig returns default auto-discovery configuration.
func DefaultAutoDiscoveryConfig() AutoDiscoveryConfig {
	return AutoDiscoveryConfig{
		ConfigNames:      []string{"config.yaml", "config.yml"},
		LocalConfigNames: []string{"config.local.yaml", "config.local.yml"},
		MaxDepth:         5,
		RequireBase:      false,
		RequireLocal:     false,
		EnableAppScoping: true,
		// Environment variable source defaults
		EnableEnvSource:  true,
		EnvSeparator:     "_",
		EnvOverridesFile: true,
	}
}

// DiscoverAndLoadConfigs automatically discovers and loads config files.
func DiscoverAndLoadConfigs(cfg AutoDiscoveryConfig) (Confy, *AutoDiscoveryResult, error) {
	// Apply defaults
	if len(cfg.ConfigNames) == 0 {
		cfg.ConfigNames = []string{"config.yaml", "config.yml"}
	}

	if len(cfg.LocalConfigNames) == 0 {
		cfg.LocalConfigNames = []string{"config.local.yaml", "config.local.yml"}
	}

	if cfg.MaxDepth == 0 {
		cfg.MaxDepth = 5
	}

	if len(cfg.SearchPaths) == 0 {
		// Default to current directory
		cwd, err := os.Getwd()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get current directory: %w", err)
		}

		cfg.SearchPaths = []string{cwd}
	}

	// Default env separator
	if cfg.EnvSeparator == "" {
		cfg.EnvSeparator = "_"
	}

	// Discover config files
	result, err := discoverConfigFiles(cfg)
	if err != nil {
		return nil, nil, err
	}

	// Create confy
	confy := NewFromConfig(Config{
		Logger:       cfg.Logger,
		ErrorHandler: cfg.ErrorHandler,
	})

	// Priority scheme:
	// - Base config: 100
	// - Local config: 200
	// - Environment (if EnvOverridesFile=true): 300
	// - Environment (if EnvOverridesFile=false): 50

	// Load all base configs (priority 100, 101, 102, …)
	if len(result.BaseConfigPaths) == 0 && cfg.RequireBase {
		return nil, nil, errors.New("base config file required but not found")
	}
	for i, basePath := range result.BaseConfigPaths {
		source, err := sources.NewFileSource(basePath, sources.FileSourceOptions{
			Name:          fmt.Sprintf("config.base.%d", i),
			Priority:      PriorityBaseConfig + i,
			WatchEnabled:  true,
			ExpandEnvVars: true,
			RequireFile:   cfg.RequireBase,
			Logger:        cfg.Logger,
			ErrorHandler:  cfg.ErrorHandler,
		})
		if err != nil {
			if cfg.RequireBase {
				return nil, nil, fmt.Errorf("failed to create base config source %q: %w", basePath, err)
			}
			continue
		}
		if err := confy.LoadFrom(source); err != nil {
			if cfg.RequireBase {
				return nil, nil, fmt.Errorf("failed to load base config %q: %w", basePath, err)
			}
		}
	}

	// Load all local configs (priority 200, 201, 202, …)
	if len(result.LocalConfigPaths) == 0 && cfg.RequireLocal {
		return nil, nil, errors.New("local config file required but not found")
	}
	for i, localPath := range result.LocalConfigPaths {
		source, err := sources.NewFileSource(localPath, sources.FileSourceOptions{
			Name:          fmt.Sprintf("config.local.%d", i),
			Priority:      PriorityLocalConfig + i,
			WatchEnabled:  true,
			ExpandEnvVars: true,
			RequireFile:   cfg.RequireLocal,
			Logger:        cfg.Logger,
			ErrorHandler:  cfg.ErrorHandler,
		})
		if err != nil {
			if cfg.RequireLocal {
				return nil, nil, fmt.Errorf("failed to create local config source %q: %w", localPath, err)
			}
			continue
		}
		if err := confy.LoadFrom(source); err != nil {
			if cfg.RequireLocal {
				return nil, nil, fmt.Errorf("failed to load local config %q: %w", localPath, err)
			}
		}
	}

	// Load environment variable source if enabled
	if cfg.EnableEnvSource {
		// Determine env prefix - default to AppName uppercase with trailing underscore
		envPrefix := cfg.EnvPrefix
		if envPrefix == "" && cfg.AppName != "" {
			envPrefix = strings.ToUpper(cfg.AppName) + cfg.EnvSeparator
		}

		// Determine priority based on EnvOverridesFile setting
		envPriority := PriorityEnvHigh // Higher than file sources (default: env overrides files)
		if !cfg.EnvOverridesFile {
			envPriority = PriorityEnvLow // Lower than file sources (files override env)
		}

		envSource, err := sources.NewEnvSource(envPrefix, sources.EnvSourceOptions{
			Name:           "config.env",
			Prefix:         envPrefix,
			Priority:       envPriority,
			Separator:      cfg.EnvSeparator,
			WatchEnabled:   false, // Env watching is expensive, disabled by default
			CaseSensitive:  false,
			IgnoreEmpty:    true,
			TypeConversion: true,
			Logger:         cfg.Logger,
			ErrorHandler:   cfg.ErrorHandler,
		})
		if err != nil {
			// Log but continue - environment variables are optional and shouldn't block config loading
			if cfg.Logger != nil {
				cfg.Logger.Warn("failed to create env config source",
					F("error", err.Error()),
				)
			}
		} else {
			if err := confy.LoadFrom(envSource); err != nil {
				// Log but continue - env loading failures shouldn't block the entire config load
				if cfg.Logger != nil {
					cfg.Logger.Warn("failed to load env config source",
						F("error", err.Error()),
					)
				}
			} else {
				if cfg.Logger != nil {
					cfg.Logger.Debug("loaded environment variable config source",
						F("prefix", envPrefix),
						F("priority", envPriority),
						F("overrides_files", cfg.EnvOverridesFile),
					)
				}
			}
		}
	}

	// Extract app-scoped config if enabled and AppName is provided
	// We need to do this AFTER loading all sources to maintain proper priority
	if cfg.EnableAppScoping && cfg.AppName != "" {
		// Get the source data before merging
		if mgr, ok := confy.(*ConfyImpl); ok {
			if err := extractAppScopedWithPriority(mgr, cfg.AppName); err != nil {
				if cfg.Logger != nil {
					cfg.Logger.Debug("app-scoped config not found, using global config",
						F("app", cfg.AppName),
					)
				}
			}
		}
	}

	return confy, result, nil
}

// discoverConfigFiles searches for config files in all specified paths.
func discoverConfigFiles(cfg AutoDiscoveryConfig) (*AutoDiscoveryResult, error) {
	result := &AutoDiscoveryResult{
		AppName: cfg.AppName,
	}

	// Search in ALL paths — do not stop on first match.
	for _, searchPath := range cfg.SearchPaths {
		searchPath = filepath.Clean(searchPath)
		// Errors in individual paths are non-fatal; continue searching.
		_, _ = searchInPathHierarchy(searchPath, cfg, result)
	}

	// Deduplicate paths in case overlapping hierarchies found the same file.
	result.BaseConfigPaths = dedup(result.BaseConfigPaths)
	result.LocalConfigPaths = dedup(result.LocalConfigPaths)

	// If nothing was found and configs are required, return an error.
	if len(result.BaseConfigPaths) == 0 && len(result.LocalConfigPaths) == 0 {
		if cfg.RequireBase || cfg.RequireLocal {
			return nil, errors.New("config files not found in search paths")
		}
	}

	return result, nil
}

// searchInPathHierarchy searches for config files in a path and its parents.
// It walks up from startPath until it finds at least one config or hits MaxDepth.
// Found paths are appended to result so that multiple search paths accumulate.
func searchInPathHierarchy(startPath string, cfg AutoDiscoveryConfig, result *AutoDiscoveryResult) (bool, error) {
	currentPath := startPath
	depth := 0
	found := false

	for depth < cfg.MaxDepth {
		// Look for base config files
		for _, configName := range cfg.ConfigNames {
			configPath := filepath.Join(currentPath, configName)
			if fileExists(configPath) {
				result.BaseConfigPaths = append(result.BaseConfigPaths, configPath)
				if result.WorkingDirectory == "" {
					result.WorkingDirectory = currentPath
				}
				found = true
				break
			}
		}

		// Look for local config files
		for _, localName := range cfg.LocalConfigNames {
			localPath := filepath.Join(currentPath, localName)
			if fileExists(localPath) {
				result.LocalConfigPaths = append(result.LocalConfigPaths, localPath)
				if result.WorkingDirectory == "" {
					result.WorkingDirectory = currentPath
				}
				found = true
			}
		}

		// Check if this looks like a monorepo (has apps/ directory)
		if dirExists(filepath.Join(currentPath, "apps")) {
			result.IsMonorepo = true
		}

		// Stop walking up once we found something at this directory level.
		if found {
			return true, nil
		}

		// Move to parent directory
		parentPath := filepath.Dir(currentPath)
		if parentPath == currentPath {
			break
		}

		currentPath = parentPath
		depth++
	}

	return false, nil
}

// extractAppScopedConfig extracts app-scoped configuration from the Confy instance
// Looks for config under "apps.{appName}" and promotes it to root level.
func extractAppScopedConfig(c Confy, appName string) error {
	// Try to get app-scoped config
	appConfigKey := "apps." + appName
	appConfig := c.GetSection(appConfigKey)

	if len(appConfig) == 0 {
		return fmt.Errorf("app-scoped config not found for app: %s", appName)
	}

	// Get all current settings
	allSettings := c.GetAllSettings()

	// Merge app config with global config using deep merge
	// Global settings are base, app-specific settings override
	merger := configcore.NewMergeUtil()
	mergedConfig := make(map[string]any)

	// Start with global settings (excluding apps section)
	for key, value := range allSettings {
		if key != "apps" {
			mergedConfig[key] = merger.DeepCopyValue(value)
		}
	}

	// Deep merge app-specific settings over global
	merger.MergeInPlace(mergedConfig, appConfig)

	// Clear and reload with merged config
	if mgr, ok := c.(*ConfyImpl); ok {
		mgr.mu.Lock()
		mgr.data = mergedConfig
		mgr.mu.Unlock()
	}

	return nil
}

// extractAppScopedWithPriority extracts app-scoped config respecting source priorities
// This ensures that local global overrides take precedence over base app-scoped settings.
func extractAppScopedWithPriority(mgr *ConfyImpl, appName string) error {
	mgr.mu.Lock()
	defer mgr.mu.Unlock()

	// Build a priority-aware merged config
	// We need to extract app-scoped from each source separately, then merge by priority

	type sourceData struct {
		priority int
		data     map[string]any
		appData  map[string]any
	}

	sources := mgr.registry.GetSources()
	sourceDataList := make([]sourceData, 0, len(sources))

	// Load each source and extract its app-scoped section
	for _, source := range sources {
		data, err := mgr.loader.LoadSource(context.Background(), source)
		if err != nil {
			continue
		}

		// Extract app-scoped section if it exists
		var appData map[string]any

		if appsSection, ok := data["apps"].(map[string]any); ok {
			if appSection, ok := appsSection[appName].(map[string]any); ok {
				appData = appSection
			}
		}

		// Remove apps section from global data
		merger := configcore.NewMergeUtil()
		globalData := make(map[string]any)

		for k, v := range data {
			if k != "apps" {
				globalData[k] = merger.DeepCopyValue(v)
			}
		}

		sourceDataList = append(sourceDataList, sourceData{
			priority: source.Priority(),
			data:     globalData,
			appData:  appData,
		})
	}

	// Sort by priority (ascending) using sort.Slice for O(n log n) performance
	sort.Slice(sourceDataList, func(i, j int) bool {
		return sourceDataList[i].priority < sourceDataList[j].priority
	})

	// Merge in priority order:
	// 1. Start with lowest priority global
	// 2. Merge same-priority app-scoped over global
	// 3. Merge next priority global
	// 4. Merge next priority app-scoped
	// etc.

	merger := configcore.NewMergeUtil()
	mergedConfig := make(map[string]any)

	for _, sd := range sourceDataList {
		// First merge global from this source
		merger.MergeInPlace(mergedConfig, sd.data)

		// Then merge app-scoped from this source (app overrides global at same priority)
		if sd.appData != nil {
			merger.MergeInPlace(mergedConfig, sd.appData)
		}
	}

	mgr.data = mergedConfig

	return nil
}

// AutoLoadConfy automatically discovers and loads config files
// This is a convenience function that uses default settings.
func AutoLoadConfy(appName string, logger logger.Logger) (Confy, error) {
	cfg := DefaultAutoDiscoveryConfig()
	cfg.AppName = appName
	cfg.Logger = logger

	c, _, err := DiscoverAndLoadConfigs(cfg)

	return c, err
}

// LoadConfigWithAppScope loads config with app-scoped extraction
// This is the recommended way to load configs in a monorepo environment.
func LoadConfigWithAppScope(appName string, logger logger.Logger, errorHandler errors.ErrorHandler) (Confy, error) {
	cfg := DefaultAutoDiscoveryConfig()
	cfg.AppName = appName
	cfg.EnableAppScoping = true
	cfg.Logger = logger
	cfg.ErrorHandler = errorHandler

	c, result, err := DiscoverAndLoadConfigs(cfg)
	if err != nil {
		return nil, err
	}

	// Log discovery results
	if logger != nil {
		for _, p := range result.BaseConfigPaths {
			logger.Info("discovered base config",
				F("path", p),
			)
		}

		for _, p := range result.LocalConfigPaths {
			logger.Info("discovered local config",
				F("path", p),
			)
		}

		if result.IsMonorepo {
			logger.Info("detected monorepo layout",
				F("app", appName),
			)
		}
	}

	return c, nil
}

// Helper functions

func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	return !info.IsDir()
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	return info.IsDir()
}

// dedup removes duplicate strings from a slice while preserving order.
func dedup(paths []string) []string {
	if len(paths) <= 1 {
		return paths
	}
	seen := make(map[string]struct{}, len(paths))
	out := make([]string, 0, len(paths))
	for _, p := range paths {
		if _, ok := seen[p]; !ok {
			seen[p] = struct{}{}
			out = append(out, p)
		}
	}
	return out
}

// LoadConfigFromPaths is a helper that loads config from explicit paths
// Useful when you know exactly where your config files are.
func LoadConfigFromPaths(basePath, localPath, appName string, logger logger.Logger) (Confy, error) {
	confy := NewFromConfig(Config{
		Logger: logger,
	})

	// Load base config if provided
	if basePath != "" && fileExists(basePath) {
		source, err := sources.NewFileSource(basePath, sources.FileSourceOptions{
			Name:          "config.base",
			Priority:      PriorityBaseConfig,
			WatchEnabled:  true,
			ExpandEnvVars: true,
			Logger:        logger,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create base config source: %w", err)
		}

		if err := confy.LoadFrom(source); err != nil {
			return nil, fmt.Errorf("failed to load base config: %w", err)
		}
	}

	// Load local config if provided (overrides base)
	if localPath != "" && fileExists(localPath) {
		source, err := sources.NewFileSource(localPath, sources.FileSourceOptions{
			Name:          "config.local",
			Priority:      PriorityLocalConfig,
			WatchEnabled:  true,
			ExpandEnvVars: true,
			Logger:        logger,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create local config source: %w", err)
		}

		if err := confy.LoadFrom(source); err != nil {
			return nil, fmt.Errorf("failed to load local config: %w", err)
		}
	}

	// Extract app-scoped config if app name provided
	if appName != "" {
		if err := extractAppScopedConfig(confy, appName); err != nil {
			// Log but don't fail - app scoping is optional
			if logger != nil {
				logger.Debug("app-scoped config not found, using global config",
					F("app", appName),
				)
			}
		}
	}

	return confy, nil
}

// GetConfigSearchInfo returns information about where configs would be searched
// Useful for debugging config loading issues.
func GetConfigSearchInfo(appName string) string {
	cwd, _ := os.Getwd()
	cfg := DefaultAutoDiscoveryConfig()
	cfg.AppName = appName

	var info strings.Builder
	fmt.Fprintf(&info, "Config Search Information for app '%s':\n", appName)
	fmt.Fprintf(&info, "  Working Directory: %s\n", cwd)
	fmt.Fprintf(&info, "  Base Config Names: %v\n", cfg.ConfigNames)
	fmt.Fprintf(&info, "  Local Config Names: %v\n", cfg.LocalConfigNames)
	fmt.Fprintf(&info, "  Max Search Depth: %d parent directories\n", cfg.MaxDepth)
	fmt.Fprintf(&info, "  App Scoping: %v\n", cfg.EnableAppScoping)

	if cfg.EnableAppScoping && appName != "" {
		fmt.Fprintf(&info, "  App-Scoped Key: apps.%s\n", appName)
	}

	return info.String()
}
