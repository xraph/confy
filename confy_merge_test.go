package confy

import (
	"reflect"
	"testing"

	"github.com/xraph/confy/internal"
)

// Characterization tests for merge functions
// These tests document current merge behavior before consolidation

func TestMergeData_BasicTypes(t *testing.T) {
	m := &ConfyImpl{
		data:      make(map[string]any),
		converter: internal.NewTypeConverter(),
		merger:    internal.NewMergeUtil(),
	}

	tests := []struct {
		name     string
		existing map[string]any
		new      map[string]any
		want     map[string]any
	}{
		{
			name:     "merge strings",
			existing: map[string]any{"key": "old"},
			new:      map[string]any{"key": "new"},
			want:     map[string]any{"key": "new"},
		},
		{
			name:     "merge different keys",
			existing: map[string]any{"key1": "value1"},
			new:      map[string]any{"key2": "value2"},
			want:     map[string]any{"key1": "value1", "key2": "value2"},
		},
		{
			name:     "merge integers",
			existing: map[string]any{"count": 1},
			new:      map[string]any{"count": 2},
			want:     map[string]any{"count": 2},
		},
		{
			name:     "nil in new keeps existing",
			existing: map[string]any{"key": "value"},
			new:      map[string]any{"key": nil},
			want:     map[string]any{"key": nil},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m.data = tt.existing
			m.mergeData(m.data, tt.new)
			if !reflect.DeepEqual(m.data, tt.want) {
				t.Errorf("mergeData() = %v, want %v", m.data, tt.want)
			}
		})
	}
}

func TestMergeData_NestedMaps(t *testing.T) {
	m := &ConfyImpl{
		data:      make(map[string]any),
		converter: internal.NewTypeConverter(),
		merger:    internal.NewMergeUtil(),
	}

	tests := []struct {
		name     string
		existing map[string]any
		new      map[string]any
		want     map[string]any
	}{
		{
			name: "merge nested maps",
			existing: map[string]any{
				"database": map[string]any{
					"host": "localhost",
					"port": 5432,
				},
			},
			new: map[string]any{
				"database": map[string]any{
					"port":     3306,
					"username": "admin",
				},
			},
			want: map[string]any{
				"database": map[string]any{
					"host":     "localhost",
					"port":     3306,
					"username": "admin",
				},
			},
		},
		{
			name: "deep nested maps",
			existing: map[string]any{
				"app": map[string]any{
					"server": map[string]any{
						"host": "0.0.0.0",
					},
				},
			},
			new: map[string]any{
				"app": map[string]any{
					"server": map[string]any{
						"port": 8080,
					},
				},
			},
			want: map[string]any{
				"app": map[string]any{
					"server": map[string]any{
						"host": "0.0.0.0",
						"port": 8080,
					},
				},
			},
		},
		{
			name: "replace non-map with map",
			existing: map[string]any{
				"database": "simple-string",
			},
			new: map[string]any{
				"database": map[string]any{
					"host": "localhost",
				},
			},
			want: map[string]any{
				"database": map[string]any{
					"host": "localhost",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m.data = tt.existing
			m.mergeData(m.data, tt.new)
			if !reflect.DeepEqual(m.data, tt.want) {
				t.Errorf("mergeData() = %#v, want %#v", m.data, tt.want)
			}
		})
	}
}

func TestMergeData_Slices(t *testing.T) {
	m := &ConfyImpl{
		data:      make(map[string]any),
		converter: internal.NewTypeConverter(),
		merger:    internal.NewMergeUtil(),
	}

	tests := []struct {
		name     string
		existing map[string]any
		new      map[string]any
		want     map[string]any
	}{
		{
			name:     "replace slice",
			existing: map[string]any{"items": []any{1, 2, 3}},
			new:      map[string]any{"items": []any{4, 5}},
			want:     map[string]any{"items": []any{4, 5}},
		},
		{
			name: "slice in nested map",
			existing: map[string]any{
				"config": map[string]any{
					"tags": []any{"old"},
				},
			},
			new: map[string]any{
				"config": map[string]any{
					"tags": []any{"new1", "new2"},
				},
			},
			want: map[string]any{
				"config": map[string]any{
					"tags": []any{"new1", "new2"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m.data = tt.existing
			m.mergeData(m.data, tt.new)
			if !reflect.DeepEqual(m.data, tt.want) {
				t.Errorf("mergeData() = %v, want %v", m.data, tt.want)
			}
		})
	}
}

func TestDeepCopyMap_Immutability(t *testing.T) {
	m := &ConfyImpl{
		converter: internal.NewTypeConverter(),
		merger:    internal.NewMergeUtil(),
	}

	original := map[string]any{
		"simple": "value",
		"nested": map[string]any{
			"inner": "data",
		},
		"slice": []any{1, 2, 3},
	}

	copied := m.merger.DeepCopy(original)

	// Modify the copy
	copied["simple"] = "modified"
	if nestedMap, ok := copied["nested"].(map[string]any); ok {
		nestedMap["inner"] = "changed"
	}
	if sliceVal, ok := copied["slice"].([]any); ok {
		sliceVal[0] = 999
	}

	// Original should be unchanged
	if original["simple"] != "value" {
		t.Error("Original simple value was modified")
	}
	if nestedMap, ok := original["nested"].(map[string]any); ok {
		if nestedMap["inner"] != "data" {
			t.Error("Original nested value was modified")
		}
	}
	if sliceVal, ok := original["slice"].([]any); ok {
		if sliceVal[0] != 1 {
			t.Error("Original slice was modified")
		}
	}
}

func TestDeepMergeValues_ComplexScenarios(t *testing.T) {
	m := &ConfyImpl{
		converter: internal.NewTypeConverter(),
		merger:    internal.NewMergeUtil(),
	}

	tests := []struct {
		name     string
		existing any
		new      any
		want     any
	}{
		{
			name:     "both maps",
			existing: map[string]any{"a": 1},
			new:      map[string]any{"b": 2},
			want:     map[string]any{"a": 1, "b": 2},
		},
		{
			name:     "new overwrites non-map",
			existing: "old",
			new:      map[string]any{"key": "value"},
			want:     map[string]any{"key": "value"},
		},
		{
			name:     "existing not map, new is map",
			existing: 42,
			new:      map[string]any{"new": "data"},
			want:     map[string]any{"new": "data"},
		},
		{
			name:     "both primitives",
			existing: "old",
			new:      "new",
			want:     "new",
		},
		{
			name:     "nil new value",
			existing: "value",
			new:      nil,
			want:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := m.deepMergeValues(tt.existing, tt.new)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("deepMergeValues() = %v, want %v", got, tt.want)
			}
		})
	}
}
