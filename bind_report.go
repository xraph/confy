package confy

import (
	"fmt"
	"reflect"
	"slices"
	"strings"
	"time"

	logger "github.com/xraph/go-utils/log"
)

var durationType = reflect.TypeFor[time.Duration]()

// FieldError describes one config value that could not be converted to its
// field's type. The field kept the value it had before the bind.
type FieldError struct {
	Key   string // full dotted path, e.g. "tracing.sample_ratio"
	Value any    // the value as it appeared in the config
	Type  string // the Go type expected, e.g. "float64", "time.Duration"
	Err   error  // the converter's error
}

// String formats the field as key=value (want type).
func (e FieldError) String() string {
	return fmt.Sprintf("%s=%#v (want %s)", e.Key, e.Value, e.Type)
}

// BindError lists every value a bind could not convert. Every other field was
// bound normally.
type BindError struct {
	Fields []FieldError
}

func (e *BindError) Error() string {
	parts := make([]string, len(e.Fields))
	for i, f := range e.Fields {
		parts[i] = f.String()
	}

	return fmt.Sprintf("config: %d invalid value(s): %s", len(e.Fields), strings.Join(parts, "; "))
}

// Unwrap returns the converter error behind each field.
func (e *BindError) Unwrap() []error {
	errs := make([]error, len(e.Fields))
	for i, f := range e.Fields {
		errs[i] = f.Err
	}

	return errs
}

// StrictBinder is implemented by binders that can report values they could not
// convert instead of logging them. *ConfyImpl implements it.
type StrictBinder interface {
	BindStrict(key string, target any) error
}

// bindState collects the values one bind call could not convert.
type bindState struct {
	problems []FieldError
}

// add records one value that could not be converted. Maps and lists are copied,
// because the caller reads Value after the bind has released the lock and Set
// can change the loaded config in place.
func (s *bindState) add(key string, value any, t reflect.Type, err error) {
	s.problems = append(s.problems, FieldError{Key: key, Value: copyConfigValue(value), Type: t.String(), Err: err})
}

// copyConfigValue deep-copies the maps and lists a loaded config is made of.
// Every other value is returned as is.
func copyConfigValue(value any) any {
	switch v := value.(type) {
	case map[string]any:
		out := make(map[string]any, len(v))
		for k, item := range v {
			out[k] = copyConfigValue(item)
		}

		return out
	case []any:
		out := make([]any, len(v))
		for i, item := range v {
			out[i] = copyConfigValue(item)
		}

		return out
	default:
		return value
	}
}

func (s *bindState) count() int { return len(s.problems) }

// err returns a *BindError when anything failed, and nil otherwise.
func (s *bindState) err() error {
	if s == nil || len(s.problems) == 0 {
		return nil
	}

	return &BindError{Fields: slices.Clone(s.problems)}
}

// joinKey appends child to a dotted key path.
func joinKey(parent, child string) string {
	if parent == "" {
		return child
	}

	return parent + "." + child
}

// isScalar reports whether t (or what t points to) is a kind convertScalar
// handles.
func isScalar(t reflect.Type) bool {
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	switch t.Kind() {
	case reflect.String, reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return true
	default:
		return false
	}
}

// unwrapValue strips interfaces and, unless field is itself an interface,
// pointers from value. ok is false when there is nothing to bind. null tells
// the two reasons apart: it is true for an explicit null in the config (`name:`
// or `name: null`) and false for a typed nil pointer, which BindWithDefault
// produces from a nil pointer field in the defaults struct.
func unwrapValue(field, value reflect.Value) (v reflect.Value, ok, null bool) {
	for value.IsValid() && (value.Kind() == reflect.Interface ||
		(value.Kind() == reflect.Pointer && field.Kind() != reflect.Interface)) {
		if value.IsNil() {
			return reflect.Value{}, false, value.Kind() == reflect.Interface
		}

		value = value.Elem()
	}

	if !value.IsValid() {
		return reflect.Value{}, false, true
	}

	return value, true, false
}

// bindNull applies an explicit null to field. A string field becomes "", which
// is what ToString(nil) always bound. Every other field keeps its value, and
// nothing is reported: a null is a choice, not a typo.
func bindNull(field reflect.Value) {
	if field.Kind() == reflect.String {
		field.SetString("")
	}
}

// mapKey converts a config key to keyType. Any string kind works, named ones
// such as `type env string` included. A config key can't become anything else,
// so other key types return an error for the caller to report.
func mapKey(key string, keyType reflect.Type) (reflect.Value, error) {
	if keyType.Kind() != reflect.String {
		return reflect.Value{}, fmt.Errorf("unsupported map key type %s", keyType)
	}

	return reflect.ValueOf(key).Convert(keyType), nil
}

// convertScalar sets field from value when field is a scalar kind. handled is
// false for every other kind. On error field is left untouched.
func (c *ConfyImpl) convertScalar(field reflect.Value, value any) (bool, error) {
	if !isScalar(field.Type()) || field.Kind() == reflect.Pointer {
		return false, nil
	}

	// A value of the same kind (a typed time.Duration, a named int) converts
	// directly. The converters only know the built-in types. A plain integer
	// headed for a time.Duration still goes through ToDuration, which reads it
	// as seconds the way Bind always has.
	rv := reflect.ValueOf(value)
	if rv.IsValid() && rv.Kind() == field.Kind() &&
		(field.Type() != durationType || rv.Type() == durationType) {
		field.Set(rv.Convert(field.Type()))

		return true, nil
	}

	switch field.Kind() {
	case reflect.String:
		field.SetString(c.converter.ToString(value))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if field.Type() == durationType {
			d, err := c.converter.ToDuration(value)
			if err != nil {
				return true, err
			}

			field.SetInt(int64(d))

			return true, nil
		}

		i, err := c.converter.ToInt64(value)
		if err != nil {
			return true, err
		}

		field.SetInt(i)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		u, err := c.converter.ToUint64(value)
		if err != nil {
			return true, err
		}

		field.SetUint(u)
	case reflect.Float32, reflect.Float64:
		f, err := c.converter.ToFloat64(value)
		if err != nil {
			return true, err
		}

		field.SetFloat(f)
	case reflect.Bool:
		b, err := c.converter.ToBool(value)
		if err != nil {
			return true, err
		}

		field.SetBool(b)
	}

	return true, nil
}

// setPointer binds into a fresh copy of what field points to and stores it
// with set's result. A pointer to a scalar is only replaced when the value
// converted, so a bad value never leaves a non-nil pointer to zero behind. A
// pointer to anything else is always replaced, keeping its good fields.
//
// The field gets a new pointer. It never writes through the old one, so a
// variable the old pointer shared with other code is not changed by a bind.
// That is on purpose: a bind must not reach past the target it was given.
func (c *ConfyImpl) setPointer(field reflect.Value, st *bindState, set func(elem reflect.Value) error) error {
	elem := reflect.New(field.Type().Elem())
	if !field.IsNil() {
		elem.Elem().Set(field.Elem())
	}

	before := st.count()
	if err := set(elem.Elem()); err != nil {
		return err
	}

	if isScalar(field.Type()) && st.count() > before {
		return nil
	}

	field.Set(elem)

	return nil
}

// logBindProblems logs one warning per value a bind could not convert.
func (c *ConfyImpl) logBindProblems(st *bindState) {
	if c.logger == nil || st == nil {
		return
	}

	for _, p := range st.problems {
		c.logger.Warn("config: ignoring invalid value",
			logger.String("key", p.Key),
			logger.Any("value", p.Value),
			logger.String("expected", p.Type),
			logger.Error(p.Err),
		)
	}
}
