package confy

import (
	"github.com/xraph/confy/internal"
)

// Confy is the main interface for configuration management.
type Confy = internal.Confy

// =============================================================================
// ERROR CONSTRUCTORS
// =============================================================================

var (
	// ErrConfigError creates a configuration error.
	ErrConfigError = internal.ErrConfigError

	// ErrLifecycleError creates a lifecycle error (e.g., start/stop/watch operations).
	ErrLifecycleError = internal.ErrLifecycleError

	// ErrValidationError creates a validation error.
	ErrValidationError = internal.ErrValidationError

	// ErrSourceNotFound creates a source not found error.
	ErrSourceNotFound = internal.ErrSourceNotFound

	// ErrSourceAlreadyExists creates a source already exists error.
	ErrSourceAlreadyExists = internal.ErrSourceAlreadyExists

	// ErrSourceError creates a source-related error.
	ErrSourceError = internal.ErrSourceError

	// ErrLoaderError creates a loader error.
	ErrLoaderError = internal.ErrLoaderError

	// ErrTransformerError creates a transformer error.
	ErrTransformerError = internal.ErrTransformerError

	// ErrSecretsNotStarted creates an error for when secrets manager is not started.
	ErrSecretsNotStarted = internal.ErrSecretsNotStarted

	// ErrSecretsAlreadyStarted creates an error for when secrets manager is already started.
	ErrSecretsAlreadyStarted = internal.ErrSecretsAlreadyStarted

	// ErrSecretNotFound creates a secret not found error.
	ErrSecretNotFound = internal.ErrSecretNotFound

	// ErrSecretError creates a general secret operation error.
	ErrSecretError = internal.ErrSecretError

	// ErrProviderNotFound creates a provider not found error.
	ErrProviderNotFound = internal.ErrProviderNotFound

	// ErrProviderError creates a provider operation error.
	ErrProviderError = internal.ErrProviderError

	// ErrUnknownProviderType creates an unknown provider type error.
	ErrUnknownProviderType = internal.ErrUnknownProviderType

	// ErrEncryptionError creates an encryption/decryption error.
	ErrEncryptionError = internal.ErrEncryptionError

	// ErrFormatError creates a format-related error.
	ErrFormatError = internal.ErrFormatError

	// ErrKeyNotFound creates a key not found error.
	ErrKeyNotFound = internal.ErrKeyNotFound

	// ErrKeyEmpty creates an empty key error.
	ErrKeyEmpty = internal.ErrKeyEmpty

	// ErrRequiredKeyMissing creates a required key missing error.
	ErrRequiredKeyMissing = internal.ErrRequiredKeyMissing

	// ErrKeyTypeMismatch creates a type mismatch error.
	ErrKeyTypeMismatch = internal.ErrKeyTypeMismatch

	// ErrConversionFailed creates a type conversion error.
	ErrConversionFailed = internal.ErrConversionFailed

	// ErrBindingFailed creates a binding error.
	ErrBindingFailed = internal.ErrBindingFailed

	// ErrInvalidDefault creates an invalid default value error.
	ErrInvalidDefault = internal.ErrInvalidDefault

	// ErrUnsupportedType creates an unsupported type error.
	ErrUnsupportedType = internal.ErrUnsupportedType

	// ErrMergeNotSupported creates a merge not supported error.
	ErrMergeNotSupported = internal.ErrMergeNotSupported

	// ErrWatchAlreadyActive creates a watch already active error.
	ErrWatchAlreadyActive = internal.ErrWatchAlreadyActive

	// ErrConfigFileNotFound creates a config file not found error.
	ErrConfigFileNotFound = internal.ErrConfigFileNotFound

	// ErrConfigFileRequired creates a required config file not found error.
	ErrConfigFileRequired = internal.ErrConfigFileRequired

	// ErrAutodiscoveryFailed creates an autodiscovery error.
	ErrAutodiscoveryFailed = internal.ErrAutodiscoveryFailed

	// ErrAppConfigNotFound creates an app-scoped config not found error.
	ErrAppConfigNotFound = internal.ErrAppConfigNotFound

	// ErrNotImplemented creates a not implemented error.
	ErrNotImplemented = internal.ErrNotImplemented

	// ErrHealthCheckFailed creates a health check failed error.
	ErrHealthCheckFailed = internal.ErrHealthCheckFailed

	// ErrFileOperation creates a file operation error.
	ErrFileOperation = internal.ErrFileOperation

	// ErrEnvironmentVariable creates an environment variable not found error.
	ErrEnvironmentVariable = internal.ErrEnvironmentVariable

	// ErrInvalidStructType creates an invalid struct type error.
	ErrInvalidStructType = internal.ErrInvalidStructType

	// ErrNilPointer creates a nil pointer error.
	ErrNilPointer = internal.ErrNilPointer

	// ErrValidationFailed creates a validation failed error with specific details.
	ErrValidationFailed = internal.ErrValidationFailed

	// ErrFormatValidation creates a format validation error.
	ErrFormatValidation = internal.ErrFormatValidation

	// ErrPortRange creates a port range validation error.
	ErrPortRange = internal.ErrPortRange
)

// =============================================================================
// CONSTANTS
// =============================================================================

// ConfigKey is the service key for configuration.
const ConfigKey = "confy:service"
