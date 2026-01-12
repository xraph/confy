package confy

import (
	"testing"
	"time"

	"github.com/xraph/confy/internal"
)

// Characterization tests for type conversion logic
// These tests document current behavior before extracting to generic converters

func TestTypeConverters_IntegerTypes(t *testing.T) {
	m := &ConfyImpl{
		data:      make(map[string]any),
		converter: internal.NewTypeConverter(),
		merger:    internal.NewMergeUtil(),
	}

	tests := []struct {
		name      string
		key       string
		value     any
		getInt8   int8
		getInt16  int16
		getInt32  int32
		getInt64  int64
		getUint8  uint8
		getUint16 uint16
		getUint32 uint32
		getUint64 uint64
	}{
		{
			name:      "int value",
			key:       "test.int",
			value:     42,
			getInt8:   42,
			getInt16:  42,
			getInt32:  42,
			getInt64:  42,
			getUint8:  42,
			getUint16: 42,
			getUint32: 42,
			getUint64: 42,
		},
		{
			name:      "int8 value",
			key:       "test.int8",
			value:     int8(10),
			getInt8:   10,
			getInt16:  10,
			getInt32:  10,
			getInt64:  10,
			getUint8:  10,
			getUint16: 10,
			getUint32: 10,
			getUint64: 10,
		},
		{
			name:      "int16 value",
			key:       "test.int16",
			value:     int16(1000),
			getInt8:   0, // overflow
			getInt16:  1000,
			getInt32:  1000,
			getInt64:  1000,
			getUint8:  0, // overflow
			getUint16: 1000,
			getUint32: 1000,
			getUint64: 1000,
		},
		{
			name:      "uint value",
			key:       "test.uint",
			value:     uint(100),
			getInt8:   100,
			getInt16:  100,
			getInt32:  100,
			getInt64:  100,
			getUint8:  100,
			getUint16: 100,
			getUint32: 100,
			getUint64: 100,
		},
		{
			name:      "float64 value",
			key:       "test.float",
			value:     42.7,
			getInt8:   42,
			getInt16:  42,
			getInt32:  42,
			getInt64:  42,
			getUint8:  42,
			getUint16: 42,
			getUint32: 42,
			getUint64: 42,
		},
		{
			name:      "string number",
			key:       "test.string",
			value:     "123",
			getInt8:   123,
			getInt16:  123,
			getInt32:  123,
			getInt64:  123,
			getUint8:  123,
			getUint16: 123,
			getUint32: 123,
			getUint64: 123,
		},
		{
			name:      "bool true",
			key:       "test.bool",
			value:     true,
			getInt8:   1,
			getInt16:  1,
			getInt32:  1,
			getInt64:  1,
			getUint8:  1,
			getUint16: 1,
			getUint32: 1,
			getUint64: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m.Set(tt.key, tt.value)

			if got := m.GetInt8(tt.key); got != tt.getInt8 {
				t.Errorf("GetInt8() = %v, want %v", got, tt.getInt8)
			}
			if got := m.GetInt16(tt.key); got != tt.getInt16 {
				t.Errorf("GetInt16() = %v, want %v", got, tt.getInt16)
			}
			if got := m.GetInt32(tt.key); got != tt.getInt32 {
				t.Errorf("GetInt32() = %v, want %v", got, tt.getInt32)
			}
			if got := m.GetInt64(tt.key); got != tt.getInt64 {
				t.Errorf("GetInt64() = %v, want %v", got, tt.getInt64)
			}
			if got := m.GetUint8(tt.key); got != tt.getUint8 {
				t.Errorf("GetUint8() = %v, want %v", got, tt.getUint8)
			}
			if got := m.GetUint16(tt.key); got != tt.getUint16 {
				t.Errorf("GetUint16() = %v, want %v", got, tt.getUint16)
			}
			if got := m.GetUint32(tt.key); got != tt.getUint32 {
				t.Errorf("GetUint32() = %v, want %v", got, tt.getUint32)
			}
			if got := m.GetUint64(tt.key); got != tt.getUint64 {
				t.Errorf("GetUint64() = %v, want %v", got, tt.getUint64)
			}
		})
	}
}

func TestTypeConverters_FloatTypes(t *testing.T) {
	m := &ConfyImpl{
		data:      make(map[string]any),
		converter: internal.NewTypeConverter(),
		merger:    internal.NewMergeUtil(),
	}

	tests := []struct {
		name       string
		key        string
		value      any
		getFloat32 float32
		getFloat64 float64
	}{
		{
			name:       "int value",
			key:        "test.int",
			value:      42,
			getFloat32: 42.0,
			getFloat64: 42.0,
		},
		{
			name:       "float32 value",
			key:        "test.float32",
			value:      float32(3.14),
			getFloat32: 3.14,
			getFloat64: 3.14,
		},
		{
			name:       "float64 value",
			key:        "test.float64",
			value:      float64(2.71828),
			getFloat32: 2.71828,
			getFloat64: 2.71828,
		},
		{
			name:       "string number",
			key:        "test.string",
			value:      "1.5",
			getFloat32: 1.5,
			getFloat64: 1.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m.Set(tt.key, tt.value)

			if got := m.GetFloat32(tt.key); got != tt.getFloat32 {
				t.Errorf("GetFloat32() = %v, want %v", got, tt.getFloat32)
			}
			if got := m.GetFloat64(tt.key); got != tt.getFloat64 {
				t.Errorf("GetFloat64() = %v, want %v", got, tt.getFloat64)
			}
		})
	}
}

func TestTypeConverters_Defaults(t *testing.T) {
	m := &ConfyImpl{
		data:      make(map[string]any),
		converter: internal.NewTypeConverter(),
		merger:    internal.NewMergeUtil(),
	}

	// Test missing keys with defaults
	if got := m.GetInt("missing", 99); got != 99 {
		t.Errorf("GetInt with default = %v, want 99", got)
	}
	if got := m.GetInt8("missing", 88); got != 88 {
		t.Errorf("GetInt8 with default = %v, want 88", got)
	}
	if got := m.GetUint("missing", 77); got != 77 {
		t.Errorf("GetUint with default = %v, want 77", got)
	}
	if got := m.GetFloat32("missing", 1.5); got != 1.5 {
		t.Errorf("GetFloat32 with default = %v, want 1.5", got)
	}
	if got := m.GetBool("missing", true); got != true {
		t.Errorf("GetBool with default = %v, want true", got)
	}
	if got := m.GetString("missing", "default"); got != "default" {
		t.Errorf("GetString with default = %v, want 'default'", got)
	}
	if got := m.GetDuration("missing", 5*time.Second); got != 5*time.Second {
		t.Errorf("GetDuration with default = %v, want 5s", got)
	}
}

func TestTypeConverters_StringConversions(t *testing.T) {
	m := &ConfyImpl{
		converter: internal.NewTypeConverter(),
		merger:    internal.NewMergeUtil(),
		data: map[string]any{
			"int":      42,
			"float":    3.14,
			"bool":     true,
			"string":   "hello",
			"duration": 5 * time.Second,
		},
	}

	tests := []struct {
		name string
		key  string
		want string
	}{
		{"int to string", "int", "42"},
		{"float to string", "float", "3.14"},
		{"bool to string", "bool", "true"},
		{"string to string", "string", "hello"},
		{"duration to string", "duration", "5s"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := m.GetString(tt.key); got != tt.want {
				t.Errorf("GetString(%q) = %v, want %v", tt.key, got, tt.want)
			}
		})
	}
}

func TestTypeConverters_BoolConversions(t *testing.T) {
	m := &ConfyImpl{
		data:      make(map[string]any),
		converter: internal.NewTypeConverter(),
		merger:    internal.NewMergeUtil(),
	}

	tests := []struct {
		name  string
		value any
		want  bool
	}{
		{"bool true", true, true},
		{"bool false", false, false},
		{"int 1", 1, true},
		{"int 0", 0, false},
		{"int non-zero", 42, true},
		{"string true", "true", true},
		{"string false", "false", false},
		{"string 1", "1", true},
		{"string 0", "0", false},
		{"string yes", "yes", true},
		{"string no", "no", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m.data["test"] = tt.value
			if got := m.GetBool("test"); got != tt.want {
				t.Errorf("GetBool() with %v = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

func TestTypeConverters_DurationConversions(t *testing.T) {
	m := &ConfyImpl{
		data:      make(map[string]any),
		converter: internal.NewTypeConverter(),
		merger:    internal.NewMergeUtil(),
	}

	tests := []struct {
		name  string
		value any
		want  time.Duration
	}{
		{"duration value", 5 * time.Second, 5 * time.Second},
		{"int seconds", 30, 30 * time.Second},
		{"int64 seconds", int64(60), 60 * time.Second},
		{"string duration", "2m", 2 * time.Minute},
		{"string seconds", "45s", 45 * time.Second},
		{"float64 seconds", 1.5, time.Duration(1.5 * float64(time.Second))},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m.data["test"] = tt.value
			if got := m.GetDuration("test"); got != tt.want {
				t.Errorf("GetDuration() with %v = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

func TestTypeConverters_InvalidConversions(t *testing.T) {
	m := &ConfyImpl{
		converter: internal.NewTypeConverter(),
		merger:    internal.NewMergeUtil(),
		data: map[string]any{
			"invalid.int":      "not-a-number",
			"invalid.float":    "not-a-float",
			"invalid.duration": "invalid-duration",
			"invalid.bool":     "not-a-bool",
		},
	}

	// Should return zero values for invalid conversions
	if got := m.GetInt("invalid.int"); got != 0 {
		t.Errorf("GetInt(invalid) = %v, want 0", got)
	}
	if got := m.GetFloat64("invalid.float"); got != 0 {
		t.Errorf("GetFloat64(invalid) = %v, want 0", got)
	}
	if got := m.GetDuration("invalid.duration"); got != 0 {
		t.Errorf("GetDuration(invalid) = %v, want 0", got)
	}
	if got := m.GetBool("invalid.bool"); got != false {
		t.Errorf("GetBool(invalid) = %v, want false", got)
	}
}
