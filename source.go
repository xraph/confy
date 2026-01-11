package confy

import (
	"github.com/xraph/confy/internal"
)

// ConfigSource represents a source of configuration data.
type ConfigSource = internal.ConfigSource

// ConfigSourceOptions contains options for creating a configuration source.
type ConfigSourceOptions = internal.ConfigSourceOptions

// ValidationOptions contains validation configuration.
type ValidationOptions = internal.ValidationOptions

// ValidationRule represents a custom validation rule.
type ValidationRule = internal.ValidationRule

// SourceMetadata contains metadata about a configuration source.
type SourceMetadata = internal.SourceMetadata

// ChangeType represents the type of configuration change.
type ChangeType = internal.ChangeType

const (
	ChangeTypeSet    ChangeType = internal.ChangeTypeSet
	ChangeTypeUpdate ChangeType = internal.ChangeTypeUpdate
	ChangeTypeDelete ChangeType = internal.ChangeTypeDelete
	ChangeTypeReload ChangeType = internal.ChangeTypeReload
)

// ConfigChange represents a configuration change event.
type ConfigChange = internal.ConfigChange

// ConfigSourceFactory creates configuration sources.
type ConfigSourceFactory = internal.ConfigSourceFactory

// SourceConfig contains common configuration for all sources.
type SourceConfig = internal.SourceConfig

// ValidationConfig contains validation configuration for sources.
type ValidationConfig = internal.ValidationConfig

// SourceRegistry manages registered configuration sources.
type SourceRegistry = internal.SourceRegistry

// SourceEvent represents an event from a configuration source.
type SourceEvent = internal.SourceEvent

// SourceEventHandler handles events from configuration sources.
type SourceEventHandler = internal.SourceEventHandler

// WatchContext contains context for watching configuration changes.
type WatchContext = internal.WatchContext
