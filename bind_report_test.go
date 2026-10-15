package confy

import (
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/xraph/confy/internal"
)

type level int

func newBareConfy() *ConfyImpl {
	return &ConfyImpl{
		data:      make(map[string]any),
		converter: internal.NewTypeConverter(),
		merger:    internal.NewMergeUtil(),
	}
}

func TestConvertScalar(t *testing.T) {
	c := newBareConfy()

	cases := []struct {
		name    string
		target  any // pointer to the field type, holding the prior value
		value   any
		want    any
		wantErr bool
	}{
		{"float from string", new(float64), "0.25", 0.25, false},
		{"float rejects word", ptr(0.5), "lots", 0.5, true},
		{"int from float64", new(int), float64(3), 3, false},
		{"int rejects word", ptr(7), "ten", 7, true},
		{"uint rejects negative", ptr(uint(7)), -3, uint(7), true},
		{"bool from yes", new(bool), "yes", true, false},
		{"bool rejects maybe", ptr(true), "maybe", true, true},
		{"duration from string", new(time.Duration), "30s", 30 * time.Second, false},
		{"duration from int is seconds", new(time.Duration), 5, 5 * time.Second, false},
		{"duration from typed duration", new(time.Duration), 2 * time.Minute, 2 * time.Minute, false},
		{"duration from int64 is seconds", new(time.Duration), int64(5), 5 * time.Second, false},
		{"duration rejects word", ptr(time.Second), "forever", time.Second, true},
		{"named int from named int", new(level), level(3), level(3), false},
		{"named int from int", new(level), 4, level(4), false},
		{"string from int", new(string), 5, "5", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			field := reflect.ValueOf(tc.target).Elem()

			handled, err := c.convertScalar(field, tc.value)
			if !handled {
				t.Fatal("convertScalar did not handle a scalar kind")
			}

			if (err != nil) != tc.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tc.wantErr)
			}

			if got := field.Interface(); got != tc.want {
				t.Fatalf("field = %#v, want %#v", got, tc.want)
			}
		})
	}
}

func TestConvertScalarSkipsCompositeKinds(t *testing.T) {
	c := newBareConfy()

	for _, target := range []any{new([]int), new(map[string]int), new(struct{}), new(*int), new(any)} {
		handled, err := c.convertScalar(reflect.ValueOf(target).Elem(), "x")
		if handled || err != nil {
			t.Fatalf("%T: handled=%v err=%v, want false and nil", target, handled, err)
		}
	}
}

func TestBindErrorListsEveryField(t *testing.T) {
	st := &bindState{}
	st.add("x.ratio", "lots", reflect.TypeOf(0.0), errors.New("bad float"))
	st.add("x.list[1]", "two", reflect.TypeOf(0), errors.New("bad int"))

	err := st.err()

	var be *BindError
	if !errors.As(err, &be) || len(be.Fields) != 2 {
		t.Fatalf("err = %v, want a *BindError with 2 fields", err)
	}

	for _, want := range []string{"x.ratio", `"lots"`, "float64", "x.list[1]", `"two"`, "int"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("error %q does not mention %s", err.Error(), want)
		}
	}

	if got := len(be.Unwrap()); got != 2 {
		t.Fatalf("Unwrap returned %d errors, want 2", got)
	}

	if (&bindState{}).err() != nil {
		t.Fatal("an empty bindState must return a nil error")
	}
}

func TestUnwrapValue(t *testing.T) {
	floatField := reflect.ValueOf(new(float64)).Elem()

	if _, ok := unwrapValue(floatField, reflect.ValueOf((*float64)(nil))); ok {
		t.Fatal("a nil pointer should leave nothing to bind")
	}

	if v, ok := unwrapValue(floatField, reflect.ValueOf(ptr(0.3))); !ok || v.Interface() != 0.3 {
		t.Fatalf("got %v, %v; want 0.3, true", v, ok)
	}

	m := map[string]any{"k": ptr(0.3)}
	if v, ok := unwrapValue(floatField, reflect.ValueOf(m).MapIndex(reflect.ValueOf("k"))); !ok || v.Interface() != 0.3 {
		t.Fatalf("map entry: got %v, %v", v, ok)
	}

	anyField := reflect.ValueOf(new(any)).Elem()
	if v, _ := unwrapValue(anyField, reflect.ValueOf(ptr(0.3))); v.Kind() != reflect.Pointer {
		t.Fatal("an interface field should receive the pointer itself")
	}
}

func TestJoinKey(t *testing.T) {
	if got := joinKey("", "a"); got != "a" {
		t.Fatalf("joinKey(\"\", a) = %q", got)
	}

	if got := joinKey("x.y", "a"); got != "x.y.a" {
		t.Fatalf("joinKey(x.y, a) = %q", got)
	}
}

func TestIsScalar(t *testing.T) {
	for _, v := range []any{0, 0.0, "", true, time.Second, level(0), new(float64)} {
		if !isScalar(reflect.TypeOf(v)) {
			t.Fatalf("%T should be scalar", v)
		}
	}

	for _, v := range []any{[]int{}, map[string]int{}, struct{}{}, new(struct{})} {
		if isScalar(reflect.TypeOf(v)) {
			t.Fatalf("%T should not be scalar", v)
		}
	}
}

func ptr[T any](v T) *T { return &v }
