package internal

import (
	"fmt"

	errors "github.com/xraph/go-utils/errs"
)

// =============================================================================
// ERROR CODES
// =============================================================================

const (
	// Configuration error codes
	CodeConfig         = "CONFIG_ERROR"
	CodeLifecycle      = "LIFECYCLE_ERROR"
	CodeSource         = "SOURCE_ERROR"
	CodeLoader         = "LOADER_ERROR"
	CodeTransformer    = "TRANSFORMER_ERROR"
	CodeSecrets        = "SECRETS_ERROR"
	CodeProvider       = "PROVIDER_ERROR"
	CodeEncryption     = "ENCRYPTION_ERROR"
	CodeFormat         = "FORMAT_ERROR"
	CodeBinding        = "BINDING_ERROR"
	CodeConversion     = "CONVERSION_ERROR"
	CodeMerge          = "MERGE_ERROR"
	CodeWatch          = "WATCH_ERROR"
	CodeAutodiscovery  = "AUTODISCOVERY_ERROR"
	CodeUnsupported    = "UNSUPPORTED_ERROR"
	CodeNotImplemented = "NOT_IMPLEMENTED"
)

// =============================================================================
// ERROR CONSTRUCTORS
// =============================================================================

// ErrConfigError creates a configuration error.
func ErrConfigError(message string, cause error) error {
	return errors.NewError(CodeConfig, message, cause)
}

// ErrLifecycleError creates a lifecycle error (e.g., start/stop/watch operations).
func ErrLifecycleError(operation string, cause error) error {
	msg := fmt.Sprintf("lifecycle error during %s operation", operation)
	if cause != nil {
		msg = fmt.Sprintf("%s: %s", msg, cause.Error())
	}
	return errors.NewError(CodeLifecycle, msg, cause).
		WithContext("operation", operation)
}

// ErrValidationError creates a validation error.
func ErrValidationError(field string, cause error) error {
	msg := fmt.Sprintf("validation failed for field '%s'", field)
	if cause != nil {
		msg = fmt.Sprintf("%s: %s", msg, cause.Error())
	}
	return errors.NewError(errors.CodeValidation, msg, cause).
		WithContext("field", field)
}

// ErrSourceNotFound creates a source not found error.
func ErrSourceNotFound(sourceName string) error {
	return errors.NewError(errors.CodeNotFound, fmt.Sprintf("source not found: %s", sourceName), nil).
		WithContext("source", sourceName)
}

// ErrSourceAlreadyExists creates a source already exists error.
func ErrSourceAlreadyExists(sourceName string) error {
	return errors.NewError(errors.CodeAlreadyExists, fmt.Sprintf("source already exists: %s", sourceName), nil).
		WithContext("source", sourceName)
}

// ErrSourceError creates a source-related error.
func ErrSourceError(sourceName string, operation string, cause error) error {
	msg := fmt.Sprintf("source '%s' failed during %s", sourceName, operation)
	return errors.NewError(CodeSource, msg, cause).
		WithContext("source", sourceName).
		WithContext("operation", operation)
}

// ErrLoaderError creates a loader error.
func ErrLoaderError(operation string, cause error) error {
	msg := fmt.Sprintf("loader failed during %s", operation)
	return errors.NewError(CodeLoader, msg, cause).
		WithContext("operation", operation)
}

// ErrTransformerError creates a transformer error.
func ErrTransformerError(transformerName string, cause error) error {
	msg := fmt.Sprintf("transformer '%s' failed", transformerName)
	return errors.NewError(CodeTransformer, msg, cause).
		WithContext("transformer", transformerName)
}

// ErrSecretsNotStarted creates an error for when secrets manager is not started.
func ErrSecretsNotStarted(operation string) error {
	return errors.NewError(CodeSecrets, fmt.Sprintf("secrets manager not started for operation: %s", operation), nil).
		WithContext("operation", operation)
}

// ErrSecretsAlreadyStarted creates an error for when secrets manager is already started.
func ErrSecretsAlreadyStarted() error {
	return errors.NewError(CodeLifecycle, "secrets manager already started", nil)
}

// ErrSecretNotFound creates a secret not found error.
func ErrSecretNotFound(key string, cause error) error {
	return errors.NewError(errors.CodeNotFound, fmt.Sprintf("secret '%s' not found", key), cause).
		WithContext("key", key)
}

// ErrSecretError creates a general secret operation error.
func ErrSecretError(operation string, key string, cause error) error {
	msg := fmt.Sprintf("secret operation '%s' failed for key '%s'", operation, key)
	return errors.NewError(CodeSecrets, msg, cause).
		WithContext("operation", operation).
		WithContext("key", key)
}

// ErrProviderNotFound creates a provider not found error.
func ErrProviderNotFound(providerName string) error {
	return errors.NewError(errors.CodeNotFound, fmt.Sprintf("provider '%s' not found", providerName), nil).
		WithContext("provider", providerName)
}

// ErrProviderError creates a provider operation error.
func ErrProviderError(providerName string, operation string, cause error) error {
	msg := fmt.Sprintf("provider '%s' failed during %s", providerName, operation)
	return errors.NewError(CodeProvider, msg, cause).
		WithContext("provider", providerName).
		WithContext("operation", operation)
}

// ErrUnknownProviderType creates an unknown provider type error.
func ErrUnknownProviderType(providerType string) error {
	return errors.NewError(errors.CodeInvalidInput, fmt.Sprintf("unknown provider type: %s", providerType), nil).
		WithContext("type", providerType)
}

// ErrEncryptionError creates an encryption/decryption error.
func ErrEncryptionError(operation string, cause error) error {
	msg := fmt.Sprintf("encryption operation '%s' failed", operation)
	return errors.NewError(CodeEncryption, msg, cause).
		WithContext("operation", operation)
}

// ErrFormatError creates a format-related error.
func ErrFormatError(format string, cause error) error {
	msg := fmt.Sprintf("unsupported or invalid format: %s", format)
	return errors.NewError(CodeFormat, msg, cause).
		WithContext("format", format)
}

// ErrKeyNotFound creates a key not found error.
func ErrKeyNotFound(key string) error {
	return errors.NewError(errors.CodeNotFound, fmt.Sprintf("key '%s' not found", key), nil).
		WithContext("key", key)
}

// ErrKeyEmpty creates an empty key error.
func ErrKeyEmpty(key string) error {
	return errors.NewError(errors.CodeValidation, fmt.Sprintf("key '%s' is empty", key), nil).
		WithContext("key", key)
}

// ErrRequiredKeyMissing creates a required key missing error.
func ErrRequiredKeyMissing(key string) error {
	return errors.NewError(errors.CodeValidation, fmt.Sprintf("required key '%s' not found", key), nil).
		WithContext("key", key).
		WithContext("required", true)
}

// ErrKeyTypeMismatch creates a type mismatch error.
func ErrKeyTypeMismatch(key string, expectedType, actualType string) error {
	msg := fmt.Sprintf("key '%s' expected type %s, got %s", key, expectedType, actualType)
	return errors.NewError(errors.CodeValidation, msg, nil).
		WithContext("key", key).
		WithContext("expected_type", expectedType).
		WithContext("actual_type", actualType)
}

// ErrConversionFailed creates a type conversion error.
func ErrConversionFailed(key string, targetType string, cause error) error {
	msg := fmt.Sprintf("failed to convert key '%s' to %s", key, targetType)
	return errors.NewError(CodeConversion, msg, cause).
		WithContext("key", key).
		WithContext("target_type", targetType)
}

// ErrBindingFailed creates a binding error.
func ErrBindingFailed(key string, cause error) error {
	msg := fmt.Sprintf("failed to bind key '%s'", key)
	return errors.NewError(CodeBinding, msg, cause).
		WithContext("key", key)
}

// ErrInvalidDefault creates an invalid default value error.
func ErrInvalidDefault(fieldType string, defaultValue string, cause error) error {
	msg := fmt.Sprintf("invalid %s default: %s", fieldType, defaultValue)
	return errors.NewError(errors.CodeInvalidInput, msg, cause).
		WithContext("field_type", fieldType).
		WithContext("default_value", defaultValue)
}

// ErrUnsupportedType creates an unsupported type error.
func ErrUnsupportedType(typeName string, context string) error {
	msg := fmt.Sprintf("unsupported type: %s in context: %s", typeName, context)
	return errors.NewError(CodeUnsupported, msg, nil).
		WithContext("type", typeName).
		WithContext("context", context)
}

// ErrMergeNotSupported creates a merge not supported error.
func ErrMergeNotSupported() error {
	return errors.NewError(CodeMerge, "merge not supported for this ConfigManager implementation", nil)
}

// ErrWatchAlreadyActive creates a watch already active error.
func ErrWatchAlreadyActive() error {
	return errors.NewError(CodeWatch, "configuration manager already watching", nil)
}

// ErrConfigFileNotFound creates a config file not found error.
func ErrConfigFileNotFound(context string) error {
	msg := fmt.Sprintf("config file not found: %s", context)
	return errors.NewError(errors.CodeNotFound, msg, nil).
		WithContext("context", context)
}

// ErrConfigFileRequired creates a required config file not found error.
func ErrConfigFileRequired(fileType string) error {
	msg := fmt.Sprintf("%s config file required but not found", fileType)
	return errors.NewError(errors.CodeNotFound, msg, nil).
		WithContext("file_type", fileType).
		WithContext("required", true)
}

// ErrAutodiscoveryFailed creates an autodiscovery error.
func ErrAutodiscoveryFailed(operation string, cause error) error {
	msg := fmt.Sprintf("autodiscovery failed during %s", operation)
	return errors.NewError(CodeAutodiscovery, msg, cause).
		WithContext("operation", operation)
}

// ErrAppConfigNotFound creates an app-scoped config not found error.
func ErrAppConfigNotFound(appName string) error {
	return errors.NewError(errors.CodeNotFound, fmt.Sprintf("app-scoped config not found for app: %s", appName), nil).
		WithContext("app", appName)
}

// ErrNotImplemented creates a not implemented error.
func ErrNotImplemented(feature string) error {
	return errors.NewError(CodeNotImplemented, fmt.Sprintf("%s not implemented", feature), nil).
		WithContext("feature", feature)
}

// ErrHealthCheckFailed creates a health check failed error.
func ErrHealthCheckFailed(component string, cause error) error {
	msg := fmt.Sprintf("%s health check failed", component)
	return errors.NewError(errors.CodeUnavailable, msg, cause).
		WithContext("component", component)
}

// ErrFileOperation creates a file operation error.
func ErrFileOperation(operation string, filePath string, cause error) error {
	msg := fmt.Sprintf("failed to %s file %s", operation, filePath)
	return errors.NewError(errors.CodeInternal, msg, cause).
		WithContext("operation", operation).
		WithContext("file_path", filePath)
}

// ErrEnvironmentVariable creates an environment variable not found error.
func ErrEnvironmentVariable(envKey string) error {
	return errors.NewError(errors.CodeNotFound, fmt.Sprintf("environment variable %s not found", envKey), nil).
		WithContext("env_key", envKey)
}

// ErrInvalidStructType creates an invalid struct type error.
func ErrInvalidStructType(expectedType string, actualType string) error {
	msg := fmt.Sprintf("value must be a %s, got %s", expectedType, actualType)
	return errors.NewError(errors.CodeInvalidInput, msg, nil).
		WithContext("expected", expectedType).
		WithContext("actual", actualType)
}

// ErrNilPointer creates a nil pointer error.
func ErrNilPointer(context string) error {
	return errors.NewError(errors.CodeInvalidInput, fmt.Sprintf("cannot convert nil pointer in context: %s", context), nil).
		WithContext("context", context)
}

// ErrValidationFailed creates a validation failed error with specific details.
func ErrValidationFailed(key string, reason string) error {
	msg := fmt.Sprintf("validation failed for key '%s': %s", key, reason)
	return errors.NewError(errors.CodeValidation, msg, nil).
		WithContext("key", key).
		WithContext("reason", reason)
}

// ErrFormatValidation creates a format validation error.
func ErrFormatValidation(format string, value string) error {
	msg := fmt.Sprintf("invalid %s format", format)
	return errors.NewError(errors.CodeValidation, msg, nil).
		WithContext("format", format).
		WithContext("value", value)
}

// ErrPortRange creates a port range validation error.
func ErrPortRange() error {
	return errors.NewError(errors.CodeValidation, "port must be between 1 and 65535", nil)
}
