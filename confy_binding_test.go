package confy

import (
	"reflect"
	"testing"
	"time"

	"github.com/xraph/confy/internal"
)

// Characterization tests for binding logic
// These tests document current struct binding behavior before refactoring

type BindingTestConfig struct {
	Host     string         `config:"host"`
	Port     int            `config:"port"`
	Enabled  bool           `config:"enabled"`
	Timeout  time.Duration  `config:"timeout"`
	Tags     []string       `config:"tags"`
	Metadata map[string]any `config:"metadata"`
}

type NestedConfig struct {
	Database struct {
		Host string `config:"host"`
		Port int    `config:"port"`
	} `config:"database"`
	Server struct {
		Host string `config:"host"`
		Port int    `config:"port"`
	} `config:"server"`
}

type DefaultConfig struct {
	Host    string `config:"host" default:"localhost"`
	Port    int    `config:"port" default:"8080"`
	Enabled bool   `config:"enabled" default:"true"`
}

type RequiredConfig struct {
	Host string `config:"host" required:"true"`
	Port int    `config:"port" required:"true"`
}

func TestBindStruct_BasicTypes(t *testing.T) {
	m := &ConfyImpl{
		converter: internal.NewTypeConverter(),
		merger:    internal.NewMergeUtil(),
		data: map[string]any{
			"host":    "example.com",
			"port":    3000,
			"enabled": true,
			"timeout": "5s",
			"tags":    []any{"api", "web"},
			"metadata": map[string]any{
				"version": "1.0",
			},
		},
	}

	var cfg BindingTestConfig
	err := m.Bind("", &cfg)
	if err != nil {
		t.Fatalf("BindStruct() error = %v", err)
	}

	if cfg.Host != "example.com" {
		t.Errorf("Host = %v, want example.com", cfg.Host)
	}
	if cfg.Port != 3000 {
		t.Errorf("Port = %v, want 3000", cfg.Port)
	}
	if !cfg.Enabled {
		t.Error("Enabled = false, want true")
	}
	if cfg.Timeout != 5*time.Second {
		t.Errorf("Timeout = %v, want 5s", cfg.Timeout)
	}
	if !reflect.DeepEqual(cfg.Tags, []string{"api", "web"}) {
		t.Errorf("Tags = %v, want [api web]", cfg.Tags)
	}
}

func TestBindStruct_Nested(t *testing.T) {
	m := &ConfyImpl{
		converter: internal.NewTypeConverter(),
		merger:    internal.NewMergeUtil(),
		data: map[string]any{
			"database": map[string]any{
				"host": "db.example.com",
				"port": 5432,
			},
			"server": map[string]any{
				"host": "api.example.com",
				"port": 8080,
			},
		},
	}

	var cfg NestedConfig
	err := m.Bind("", &cfg)
	if err != nil {
		t.Fatalf("BindStruct() error = %v", err)
	}

	if cfg.Database.Host != "db.example.com" {
		t.Errorf("Database.Host = %v, want db.example.com", cfg.Database.Host)
	}
	if cfg.Database.Port != 5432 {
		t.Errorf("Database.Port = %v, want 5432", cfg.Database.Port)
	}
	if cfg.Server.Host != "api.example.com" {
		t.Errorf("Server.Host = %v, want api.example.com", cfg.Server.Host)
	}
	if cfg.Server.Port != 8080 {
		t.Errorf("Server.Port = %v, want 8080", cfg.Server.Port)
	}
}

func TestBindStruct_Defaults(t *testing.T) {
	m := &ConfyImpl{
		data:      map[string]any{},
		converter: internal.NewTypeConverter(),
		merger:    internal.NewMergeUtil(),
	}

	var cfg DefaultConfig
	err := m.Bind("", &cfg)
	if err != nil {
		t.Fatalf("BindStruct() error = %v", err)
	}

	if cfg.Host != "localhost" {
		t.Errorf("Host = %v, want localhost (default)", cfg.Host)
	}
	if cfg.Port != 8080 {
		t.Errorf("Port = %v, want 8080 (default)", cfg.Port)
	}
	if !cfg.Enabled {
		t.Error("Enabled = false, want true (default)")
	}
}

func TestBindStruct_DefaultOverride(t *testing.T) {
	m := &ConfyImpl{
		data: map[string]any{
			"host": "custom.com",
			"port": 9000,
		},
		converter: internal.NewTypeConverter(),
		merger:    internal.NewMergeUtil(),
	}

	var cfg DefaultConfig
	err := m.Bind("", &cfg)
	if err != nil {
		t.Fatalf("BindStruct() error = %v", err)
	}

	// Actual values should override defaults
	if cfg.Host != "custom.com" {
		t.Errorf("Host = %v, want custom.com (override)", cfg.Host)
	}
	if cfg.Port != 9000 {
		t.Errorf("Port = %v, want 9000 (override)", cfg.Port)
	}
	// Unset value should use default
	if !cfg.Enabled {
		t.Error("Enabled = false, want true (default)")
	}
}

func TestBindStruct_Required(t *testing.T) {
	tests := []struct {
		name    string
		data    map[string]any
		wantErr bool
	}{
		{
			name: "all required present",
			data: map[string]any{
				"host": "example.com",
				"port": 8080,
			},
			wantErr: false,
		},
		{
			name: "missing required host",
			data: map[string]any{
				"port": 8080,
			},
			wantErr: true,
		},
		{
			name: "missing required port",
			data: map[string]any{
				"host": "example.com",
			},
			wantErr: true,
		},
		{
			name:    "all required missing",
			data:    map[string]any{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &ConfyImpl{data: tt.data, converter: internal.NewTypeConverter(), merger: internal.NewMergeUtil()}
			var cfg RequiredConfig
			err := m.Bind("", &cfg)

			if tt.wantErr && err == nil {
				t.Error("BindStruct() expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("BindStruct() unexpected error = %v", err)
			}
		})
	}
}

func TestBindStruct_TypeConversion(t *testing.T) {
	tests := []struct {
		name  string
		data  map[string]any
		check func(*BindingTestConfig) bool
	}{
		{
			name: "string to int",
			data: map[string]any{"port": "3000"},
			check: func(cfg *BindingTestConfig) bool {
				return cfg.Port == 3000
			},
		},
		{
			name: "int to string",
			data: map[string]any{"host": 123},
			check: func(cfg *BindingTestConfig) bool {
				return cfg.Host == "123"
			},
		},
		{
			name: "string to bool",
			data: map[string]any{"enabled": "true"},
			check: func(cfg *BindingTestConfig) bool {
				return cfg.Enabled == true
			},
		},
		{
			name: "int to bool",
			data: map[string]any{"enabled": 1},
			check: func(cfg *BindingTestConfig) bool {
				return cfg.Enabled == true
			},
		},
		{
			name: "string to duration",
			data: map[string]any{"timeout": "10s"},
			check: func(cfg *BindingTestConfig) bool {
				return cfg.Timeout == 10*time.Second
			},
		},
		{
			name: "int to duration",
			data: map[string]any{"timeout": 30},
			check: func(cfg *BindingTestConfig) bool {
				return cfg.Timeout == 30*time.Second
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &ConfyImpl{data: tt.data, converter: internal.NewTypeConverter(), merger: internal.NewMergeUtil()}
			var cfg BindingTestConfig
			if err := m.Bind("", &cfg); err != nil {
				t.Fatalf("BindStruct() error = %v", err)
			}
			if !tt.check(&cfg) {
				t.Errorf("Type conversion check failed for %v", tt.data)
			}
		})
	}
}

func TestBindStruct_PointerValidation(t *testing.T) {
	m := &ConfyImpl{data: make(map[string]any), converter: internal.NewTypeConverter(), merger: internal.NewMergeUtil()}

	// Non-pointer should error
	var cfg BindingTestConfig
	err := m.Bind("", cfg)
	if err == nil {
		t.Error("BindStruct() with non-pointer should error")
	}

	// Nil pointer should error
	err = m.Bind("", nil)
	if err == nil {
		t.Error("BindStruct() with nil should error")
	}

	// Pointer to non-struct should error
	var str string
	err = m.Bind("", &str)
	if err == nil {
		t.Error("BindStruct() with non-struct should error")
	}
}

func TestBindStruct_EmptyStruct(t *testing.T) {
	m := &ConfyImpl{
		data: map[string]any{
			"unused": "value",
		},
		converter: internal.NewTypeConverter(),
		merger:    internal.NewMergeUtil(),
	}

	type EmptyConfig struct{}
	var cfg EmptyConfig

	// Should not error even with no fields to bind
	err := m.Bind("", &cfg)
	if err != nil {
		t.Errorf("BindStruct() with empty struct error = %v", err)
	}
}
