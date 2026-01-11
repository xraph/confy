package internal

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// TypeConverter provides generic type conversion utilities.
// This replaces the 22 duplicate Get* methods in Manager with a single, testable implementation.
type TypeConverter struct{}

// NewTypeConverter creates a new type converter.
func NewTypeConverter() *TypeConverter {
	return &TypeConverter{}
}

// =============================================================================
// NUMERIC CONVERSIONS
// =============================================================================

// ToInt converts any value to int.
func (tc *TypeConverter) ToInt(value any) (int, error) {
	return convertToInt[int](value)
}

// ToInt8 converts any value to int8.
func (tc *TypeConverter) ToInt8(value any) (int8, error) {
	return convertToInt[int8](value)
}

// ToInt16 converts any value to int16.
func (tc *TypeConverter) ToInt16(value any) (int16, error) {
	return convertToInt[int16](value)
}

// ToInt32 converts any value to int32.
func (tc *TypeConverter) ToInt32(value any) (int32, error) {
	return convertToInt[int32](value)
}

// ToInt64 converts any value to int64.
func (tc *TypeConverter) ToInt64(value any) (int64, error) {
	return convertToInt[int64](value)
}

// ToUint converts any value to uint.
func (tc *TypeConverter) ToUint(value any) (uint, error) {
	return convertToUint[uint](value)
}

// ToUint8 converts any value to uint8.
func (tc *TypeConverter) ToUint8(value any) (uint8, error) {
	return convertToUint[uint8](value)
}

// ToUint16 converts any value to uint16.
func (tc *TypeConverter) ToUint16(value any) (uint16, error) {
	return convertToUint[uint16](value)
}

// ToUint32 converts any value to uint32.
func (tc *TypeConverter) ToUint32(value any) (uint32, error) {
	return convertToUint[uint32](value)
}

// ToUint64 converts any value to uint64.
func (tc *TypeConverter) ToUint64(value any) (uint64, error) {
	return convertToUint[uint64](value)
}

// ToFloat32 converts any value to float32.
func (tc *TypeConverter) ToFloat32(value any) (float32, error) {
	return convertToFloat[float32](value)
}

// ToFloat64 converts any value to float64.
func (tc *TypeConverter) ToFloat64(value any) (float64, error) {
	return convertToFloat[float64](value)
}

// =============================================================================
// INTEGER TYPE CONVERSION - GENERIC
// =============================================================================

type signedInt interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

func convertToInt[T signedInt](value any) (T, error) {
	if value == nil {
		return 0, fmt.Errorf("cannot convert nil to %T", *new(T))
	}

	switch v := value.(type) {
	case int:
		return T(v), nil
	case int8:
		return T(v), nil
	case int16:
		return T(v), nil
	case int32:
		return T(v), nil
	case int64:
		return T(v), nil
	case uint:
		return T(v), nil
	case uint8:
		return T(v), nil
	case uint16:
		return T(v), nil
	case uint32:
		return T(v), nil
	case uint64:
		return T(v), nil
	case float32:
		return T(v), nil
	case float64:
		return T(v), nil
	case bool:
		if v {
			return 1, nil
		}
		return 0, nil
	case string:
		i, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("cannot convert string %q to int: %w", v, err)
		}
		return T(i), nil
	default:
		return 0, fmt.Errorf("cannot convert %T to %T", value, *new(T))
	}
}

// =============================================================================
// UNSIGNED INTEGER TYPE CONVERSION - GENERIC
// =============================================================================

type unsignedInt interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
}

func convertToUint[T unsignedInt](value any) (T, error) {
	if value == nil {
		return 0, fmt.Errorf("cannot convert nil to %T", *new(T))
	}

	switch v := value.(type) {
	case int:
		if v < 0 {
			return 0, fmt.Errorf("cannot convert negative int %d to unsigned", v)
		}
		return T(v), nil
	case int8:
		if v < 0 {
			return 0, fmt.Errorf("cannot convert negative int8 %d to unsigned", v)
		}
		return T(v), nil
	case int16:
		if v < 0 {
			return 0, fmt.Errorf("cannot convert negative int16 %d to unsigned", v)
		}
		return T(v), nil
	case int32:
		if v < 0 {
			return 0, fmt.Errorf("cannot convert negative int32 %d to unsigned", v)
		}
		return T(v), nil
	case int64:
		if v < 0 {
			return 0, fmt.Errorf("cannot convert negative int64 %d to unsigned", v)
		}
		return T(v), nil
	case uint:
		return T(v), nil
	case uint8:
		return T(v), nil
	case uint16:
		return T(v), nil
	case uint32:
		return T(v), nil
	case uint64:
		return T(v), nil
	case float32:
		if v < 0 {
			return 0, fmt.Errorf("cannot convert negative float32 %f to unsigned", v)
		}
		return T(v), nil
	case float64:
		if v < 0 {
			return 0, fmt.Errorf("cannot convert negative float64 %f to unsigned", v)
		}
		return T(v), nil
	case bool:
		if v {
			return 1, nil
		}
		return 0, nil
	case string:
		u, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("cannot convert string %q to uint: %w", v, err)
		}
		return T(u), nil
	default:
		return 0, fmt.Errorf("cannot convert %T to %T", value, *new(T))
	}
}

// =============================================================================
// FLOAT TYPE CONVERSION - GENERIC
// =============================================================================

type floatType interface {
	~float32 | ~float64
}

func convertToFloat[T floatType](value any) (T, error) {
	if value == nil {
		return 0, fmt.Errorf("cannot convert nil to %T", *new(T))
	}

	switch v := value.(type) {
	case float32:
		return T(v), nil
	case float64:
		return T(v), nil
	case int:
		return T(v), nil
	case int8:
		return T(v), nil
	case int16:
		return T(v), nil
	case int32:
		return T(v), nil
	case int64:
		return T(v), nil
	case uint:
		return T(v), nil
	case uint8:
		return T(v), nil
	case uint16:
		return T(v), nil
	case uint32:
		return T(v), nil
	case uint64:
		return T(v), nil
	case bool:
		if v {
			return 1.0, nil
		}
		return 0.0, nil
	case string:
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return 0, fmt.Errorf("cannot convert string %q to float: %w", v, err)
		}
		return T(f), nil
	default:
		return 0, fmt.Errorf("cannot convert %T to %T", value, *new(T))
	}
}

// =============================================================================
// OTHER TYPE CONVERSIONS
// =============================================================================

// ToString converts any value to string.
func (tc *TypeConverter) ToString(value any) string {
	if value == nil {
		return ""
	}

	switch v := value.(type) {
	case string:
		return v
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%d", v)
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", v)
	case float32, float64:
		return fmt.Sprintf("%v", v)
	case bool:
		return strconv.FormatBool(v)
	case time.Duration:
		return v.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}

// ToBool converts any value to bool.
func (tc *TypeConverter) ToBool(value any) (bool, error) {
	if value == nil {
		return false, fmt.Errorf("cannot convert nil to bool")
	}

	switch v := value.(type) {
	case bool:
		return v, nil
	case int, int8, int16, int32, int64:
		// Any integer value - use reflect to handle all int types
		return fmt.Sprintf("%d", v) != "0", nil
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", v) != "0", nil
	case float32:
		return v != 0, nil
	case float64:
		return v != 0, nil
	case string:
		// Try parsing as bool first
		if b, err := strconv.ParseBool(v); err == nil {
			return b, nil
		}
		// Check common string representations
		lower := v
		switch lower {
		case "yes", "y", "on", "1":
			return true, nil
		case "no", "n", "off", "0", "":
			return false, nil
		}
		return false, fmt.Errorf("cannot convert string %q to bool", v)
	default:
		return false, fmt.Errorf("cannot convert %T to bool", value)
	}
}

// ToDuration converts any value to time.Duration.
func (tc *TypeConverter) ToDuration(value any) (time.Duration, error) {
	if value == nil {
		return 0, fmt.Errorf("cannot convert nil to duration")
	}

	switch v := value.(type) {
	case time.Duration:
		return v, nil
	case int, int8, int16, int32, int64:
		// Interpret integers as seconds
		seconds := fmt.Sprintf("%d", v)
		i, _ := strconv.ParseInt(seconds, 10, 64)
		return time.Duration(i) * time.Second, nil
	case uint, uint8, uint16, uint32, uint64:
		seconds := fmt.Sprintf("%d", v)
		i, _ := strconv.ParseUint(seconds, 10, 64)
		return time.Duration(i) * time.Second, nil
	case float32:
		return time.Duration(float64(v) * float64(time.Second)), nil
	case float64:
		return time.Duration(v * float64(time.Second)), nil
	case string:
		d, err := time.ParseDuration(v)
		if err != nil {
			return 0, fmt.Errorf("cannot convert string %q to duration: %w", v, err)
		}
		return d, nil
	default:
		return 0, fmt.Errorf("cannot convert %T to duration", value)
	}
}

// =============================================================================
// SLICE CONVERSIONS
// =============================================================================

// ToStringSlice converts any value to []string.
func (tc *TypeConverter) ToStringSlice(value any) ([]string, error) {
	if value == nil {
		return nil, fmt.Errorf("cannot convert nil to []string")
	}

	switch v := value.(type) {
	case []string:
		return v, nil
	case []any:
		result := make([]string, len(v))
		for i, item := range v {
			result[i] = tc.ToString(item)
		}
		return result, nil
	case string:
		// Single string becomes single-element slice
		return []string{v}, nil
	default:
		return nil, fmt.Errorf("cannot convert %T to []string", value)
	}
}

// ToIntSlice converts any value to []int.
func (tc *TypeConverter) ToIntSlice(value any) ([]int, error) {
	if value == nil {
		return nil, fmt.Errorf("cannot convert nil to []int")
	}

	switch v := value.(type) {
	case []int:
		return v, nil
	case []any:
		result := make([]int, len(v))
		for i, item := range v {
			val, err := tc.ToInt(item)
			if err != nil {
				return nil, fmt.Errorf("cannot convert element %d: %w", i, err)
			}
			result[i] = val
		}
		return result, nil
	default:
		return nil, fmt.Errorf("cannot convert %T to []int", value)
	}
}

// ToInt64Slice converts any value to []int64.
func (tc *TypeConverter) ToInt64Slice(value any) ([]int64, error) {
	if value == nil {
		return nil, fmt.Errorf("cannot convert nil to []int64")
	}

	switch v := value.(type) {
	case []int64:
		return v, nil
	case []any:
		result := make([]int64, len(v))
		for i, item := range v {
			val, err := tc.ToInt64(item)
			if err != nil {
				return nil, fmt.Errorf("cannot convert element %d: %w", i, err)
			}
			result[i] = val
		}
		return result, nil
	default:
		return nil, fmt.Errorf("cannot convert %T to []int64", value)
	}
}

// ToFloat64Slice converts any value to []float64.
func (tc *TypeConverter) ToFloat64Slice(value any) ([]float64, error) {
	if value == nil {
		return nil, fmt.Errorf("cannot convert nil to []float64")
	}

	switch v := value.(type) {
	case []float64:
		return v, nil
	case []any:
		result := make([]float64, len(v))
		for i, item := range v {
			val, err := tc.ToFloat64(item)
			if err != nil {
				return nil, fmt.Errorf("cannot convert element %d: %w", i, err)
			}
			result[i] = val
		}
		return result, nil
	default:
		return nil, fmt.Errorf("cannot convert %T to []float64", value)
	}
}

// ToBoolSlice converts any value to []bool.
func (tc *TypeConverter) ToBoolSlice(value any) ([]bool, error) {
	if value == nil {
		return nil, fmt.Errorf("cannot convert nil to []bool")
	}

	switch v := value.(type) {
	case []bool:
		return v, nil
	case []any:
		result := make([]bool, len(v))
		for i, item := range v {
			val, err := tc.ToBool(item)
			if err != nil {
				return nil, fmt.Errorf("cannot convert element %d: %w", i, err)
			}
			result[i] = val
		}
		return result, nil
	default:
		return nil, fmt.Errorf("cannot convert %T to []bool", value)
	}
}

// ToTime converts any value to time.Time.
func (tc *TypeConverter) ToTime(value any) (time.Time, error) {
	if value == nil {
		return time.Time{}, fmt.Errorf("cannot convert nil to time.Time")
	}

	switch v := value.(type) {
	case time.Time:
		return v, nil
	case string:
		formats := []string{
			time.RFC3339,
			time.RFC3339Nano,
			"2006-01-02 15:04:05",
			"2006-01-02T15:04:05",
			"2006-01-02",
		}
		for _, format := range formats {
			if t, err := time.Parse(format, v); err == nil {
				return t, nil
			}
		}
		return time.Time{}, fmt.Errorf("cannot parse time string %q", v)
	case int64:
		return time.Unix(v, 0), nil
	case float64:
		sec := int64(v)
		nsec := int64((v - float64(sec)) * 1e9)
		return time.Unix(sec, nsec), nil
	default:
		return time.Time{}, fmt.Errorf("cannot convert %T to time.Time", value)
	}
}

// ToSizeInBytes parses a size string and returns the value in bytes.
// Supports units: B, KB, MB, GB, TB, PB (binary: 1024) and K, M, G, T, P (decimal: 1000).
func (tc *TypeConverter) ToSizeInBytes(value any) (uint64, error) {
	// Handle numeric types directly
	switch v := value.(type) {
	case uint64:
		return v, nil
	case uint:
		return uint64(v), nil
	case int:
		if v < 0 {
			return 0, fmt.Errorf("size cannot be negative: %d", v)
		}
		return uint64(v), nil
	case int64:
		if v < 0 {
			return 0, fmt.Errorf("size cannot be negative: %d", v)
		}
		return uint64(v), nil
	case string:
		return tc.parseSizeString(v)
	default:
		return 0, fmt.Errorf("cannot convert %T to size in bytes", value)
	}
}

// parseSizeString parses a size string like "10MB" or "1.5GB".
func (tc *TypeConverter) parseSizeString(s string) (uint64, error) {
	s = strings.TrimSpace(strings.ToUpper(s))
	if s == "" {
		return 0, fmt.Errorf("empty size string")
	}

	units := map[string]uint64{
		"B":  1,
		"KB": 1024,
		"MB": 1024 * 1024,
		"GB": 1024 * 1024 * 1024,
		"TB": 1024 * 1024 * 1024 * 1024,
		"PB": 1024 * 1024 * 1024 * 1024 * 1024,
		"K":  1000,
		"M":  1000 * 1000,
		"G":  1000 * 1000 * 1000,
		"T":  1000 * 1000 * 1000 * 1000,
		"P":  1000 * 1000 * 1000 * 1000 * 1000,
	}

	for unit, multiplier := range units {
		if strings.HasSuffix(s, unit) {
			numberStr := strings.TrimSuffix(s, unit)
			numberStr = strings.TrimSpace(numberStr)
			number, err := strconv.ParseFloat(numberStr, 64)
			if err != nil {
				return 0, fmt.Errorf("invalid size format %q: %w", s, err)
			}
			return uint64(number * float64(multiplier)), nil
		}
	}

	// No unit specified, try to parse as plain number
	number, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid size format %q: %w", s, err)
	}
	return number, nil
}
