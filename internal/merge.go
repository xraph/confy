package internal

import (
	"reflect"
)

// MergeUtil provides utilities for merging configuration data.
// This consolidates the three duplicate merge implementations in the codebase.
type MergeUtil struct{}

// NewMergeUtil creates a new merge utility.
func NewMergeUtil() *MergeUtil {
	return &MergeUtil{}
}

// DeepMerge performs a deep merge of two maps.
// Values from 'new' override values in 'existing'.
// For nested maps, merging continues recursively.
// Slices and other types are replaced entirely (not merged).
func (mu *MergeUtil) DeepMerge(existing, new map[string]any) map[string]any {
	if existing == nil {
		return mu.DeepCopy(new)
	}
	if new == nil {
		return mu.DeepCopy(existing)
	}

	result := mu.DeepCopy(existing)

	for key, newValue := range new {
		if existingValue, exists := result[key]; exists {
			result[key] = mu.mergeValues(existingValue, newValue)
		} else {
			result[key] = mu.DeepCopyValue(newValue)
		}
	}

	return result
}

// MergeInPlace merges 'new' into 'existing' without creating a copy.
// This modifies the existing map in place.
func (mu *MergeUtil) MergeInPlace(existing, new map[string]any) {
	if existing == nil || new == nil {
		return
	}

	for key, newValue := range new {
		if existingValue, exists := existing[key]; exists {
			existing[key] = mu.mergeValues(existingValue, newValue)
		} else {
			existing[key] = mu.DeepCopyValue(newValue)
		}
	}
}

// mergeValues determines how to merge two values based on their types.
func (mu *MergeUtil) mergeValues(existing, new any) any {
	// If new value is nil, use it (explicit null/unset)
	if new == nil {
		return nil
	}

	// If existing is nil, use new value
	if existing == nil {
		return mu.DeepCopyValue(new)
	}

	// Both values are maps - deep merge them
	existingMap, existingIsMap := existing.(map[string]any)
	newMap, newIsMap := new.(map[string]any)

	if existingIsMap && newIsMap {
		return mu.DeepMerge(existingMap, newMap)
	}

	// For all other types (including slices), replace with new value
	return mu.DeepCopyValue(new)
}

// DeepCopy creates a deep copy of a map.
func (mu *MergeUtil) DeepCopy(src map[string]any) map[string]any {
	if src == nil {
		return nil
	}

	result := make(map[string]any, len(src))
	for key, value := range src {
		result[key] = mu.DeepCopyValue(value)
	}

	return result
}

// deepCopyValue creates a deep copy of any value.
func (mu *MergeUtil) DeepCopyValue(value any) any {
	if value == nil {
		return nil
	}

	switch v := value.(type) {
	case map[string]any:
		return mu.DeepCopy(v)
	case []any:
		return mu.deepCopySlice(v)
	case []string:
		// Copy string slice
		result := make([]string, len(v))
		copy(result, v)
		return result
	case []int:
		result := make([]int, len(v))
		copy(result, v)
		return result
	case []float64:
		result := make([]float64, len(v))
		copy(result, v)
		return result
	default:
		// For primitive types and unknown types, use reflection for safety
		return mu.deepCopyReflect(value)
	}
}

// deepCopySlice creates a deep copy of a slice.
func (mu *MergeUtil) deepCopySlice(src []any) []any {
	if src == nil {
		return nil
	}

	result := make([]any, len(src))
	for i, item := range src {
		result[i] = mu.DeepCopyValue(item)
	}

	return result
}

// deepCopyReflect uses reflection to deep copy complex types.
// This is a fallback for types we don't handle explicitly.
func (mu *MergeUtil) deepCopyReflect(value any) any {
	if value == nil {
		return nil
	}

	val := reflect.ValueOf(value)

	switch val.Kind() {
	case reflect.Ptr:
		if val.IsNil() {
			return nil
		}
		// For pointers, we dereference and copy the value
		return mu.deepCopyReflect(val.Elem().Interface())

	case reflect.Slice, reflect.Array:
		length := val.Len()
		result := make([]any, length)
		for i := 0; i < length; i++ {
			result[i] = mu.deepCopyReflect(val.Index(i).Interface())
		}
		return result

	case reflect.Map:
		result := make(map[string]any)
		iter := val.MapRange()
		for iter.Next() {
			key := iter.Key().Interface()
			keyStr, ok := key.(string)
			if !ok {
				keyStr = toString(key)
			}
			result[keyStr] = mu.deepCopyReflect(iter.Value().Interface())
		}
		return result

	case reflect.Struct:
		// For structs, we can't easily deep copy without knowing the type
		// Return the value as-is (shallow copy)
		// In practice, config values are usually maps, slices, or primitives
		return value

	default:
		// Primitive types (int, string, bool, etc.) can be copied by value
		return value
	}
}

// toString converts any value to a string for use as a map key.
func toString(value any) string {
	if value == nil {
		return ""
	}

	if str, ok := value.(string); ok {
		return str
	}

	// Use %v format for other types
	return reflect.ValueOf(value).String()
}

// ShallowMerge performs a shallow merge (only top-level keys).
// This is faster than DeepMerge when you don't need recursive merging.
func (mu *MergeUtil) ShallowMerge(existing, new map[string]any) map[string]any {
	if existing == nil {
		existing = make(map[string]any)
	}

	result := make(map[string]any, len(existing)+len(new))

	// Copy existing
	for key, value := range existing {
		result[key] = value
	}

	// Override with new
	for key, value := range new {
		result[key] = value
	}

	return result
}

// MergeMaps merges multiple maps from left to right.
// Later maps override earlier ones.
func (mu *MergeUtil) MergeMaps(maps ...map[string]any) map[string]any {
	if len(maps) == 0 {
		return make(map[string]any)
	}

	result := make(map[string]any)

	for _, m := range maps {
		if m != nil {
			mu.MergeInPlace(result, m)
		}
	}

	return result
}
