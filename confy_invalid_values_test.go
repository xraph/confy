package confy

import (
	"errors"
	"reflect"
	"testing"
	"time"

	configcore "github.com/xraph/confy/internal"
	logger "github.com/xraph/go-utils/log"
)

type invalidNested struct {
	N int    `yaml:"n"`
	S string `yaml:"s"`
}

type invalidTarget struct {
	Ratio    *float64        `yaml:"ratio"`
	Float    float64         `yaml:"float"`
	Int      int             `yaml:"int"`
	Uint     uint            `yaml:"uint"`
	Bool     bool            `yaml:"bool"`
	Duration time.Duration   `yaml:"duration"`
	Nested   invalidNested   `yaml:"nested"`
	NestedP  *invalidNested  `yaml:"nested_p"`
	List     []int           `yaml:"list"`
	Items    []invalidNested `yaml:"items"`
	Weights  map[string]int  `yaml:"weights"`
	Name     string          `yaml:"name"`
}

func priorTarget() invalidTarget {
	return invalidTarget{
		Float: 0.5, Int: 7, Uint: 7, Bool: true, Duration: time.Second,
		Nested: invalidNested{N: 7}, List: []int{9},
	}
}

func captureConfy(t *testing.T) (*ConfyImpl, *logger.TestLogger) {
	t.Helper()

	log, ok := logger.NewTestLogger().(*logger.TestLogger)
	if !ok {
		t.Fatal("NewTestLogger did not return a *TestLogger")
	}

	c, ok := New(WithLogger(log)).(*ConfyImpl)
	if !ok {
		t.Fatal("New did not return a *ConfyImpl")
	}

	return c, log
}

func warnKeys(log *logger.TestLogger) []string {
	var keys []string

	for _, e := range log.GetLogsByLevel("WARN") {
		if e.Message != "config: ignoring invalid value" {
			continue
		}

		k, _ := e.Field("key")
		keys = append(keys, k.(string))
	}

	return keys
}

func TestBindIgnoresInvalidValues(t *testing.T) {
	cases := []struct {
		name    string
		section map[string]any
		key     string
		check   func(invalidTarget) bool
	}{
		{"nil pointer stays nil", map[string]any{"ratio": "lots"}, "x.ratio",
			func(g invalidTarget) bool { return g.Ratio == nil }},
		{"float keeps prior", map[string]any{"float": "lots"}, "x.float",
			func(g invalidTarget) bool { return g.Float == 0.5 }},
		{"int keeps prior", map[string]any{"int": "ten"}, "x.int",
			func(g invalidTarget) bool { return g.Int == 7 }},
		{"uint rejects negative", map[string]any{"uint": -3}, "x.uint",
			func(g invalidTarget) bool { return g.Uint == 7 }},
		{"bool keeps prior", map[string]any{"bool": "maybe"}, "x.bool",
			func(g invalidTarget) bool { return g.Bool }},
		{"duration keeps prior", map[string]any{"duration": "forever"}, "x.duration",
			func(g invalidTarget) bool { return g.Duration == time.Second }},
		{"nested field keeps prior", map[string]any{"nested": map[string]any{"n": "x", "s": "ok"}}, "x.nested.n",
			func(g invalidTarget) bool { return g.Nested.N == 7 && g.Nested.S == "ok" }},
		{"pointer to struct keeps good fields", map[string]any{"nested_p": map[string]any{"n": "x", "s": "ok"}}, "x.nested_p.n",
			func(g invalidTarget) bool { return g.NestedP != nil && g.NestedP.S == "ok" && g.NestedP.N == 0 }},
		{"scalar list kept whole", map[string]any{"list": []any{1, "two"}}, "x.list[1]",
			func(g invalidTarget) bool { return reflect.DeepEqual(g.List, []int{9}) }},
		{"struct list keeps elements", map[string]any{"items": []any{map[string]any{"n": "x", "s": "a"}, map[string]any{"n": 2}}}, "x.items[0].n",
			func(g invalidTarget) bool { return len(g.Items) == 2 && g.Items[0].S == "a" && g.Items[1].N == 2 }},
		{"map drops bad entry", map[string]any{"weights": map[string]any{"a": "one", "b": 2}}, "x.weights.a",
			func(g invalidTarget) bool { return len(g.Weights) == 1 && g.Weights["b"] == 2 }},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c, log := captureConfy(t)
			c.Set("x", tc.section)

			got := priorTarget()
			if err := c.Bind("x", &got); err != nil {
				t.Fatalf("Bind returned %v, want nil", err)
			}

			if keys := warnKeys(log); !reflect.DeepEqual(keys, []string{tc.key}) {
				t.Fatalf("warned for %v, want [%s]", keys, tc.key)
			}

			if !tc.check(got) {
				t.Fatalf("bound %+v", got)
			}
		})
	}
}

func TestBindWarningNamesValueAndType(t *testing.T) {
	c, log := captureConfy(t)
	c.Set("tracing", map[string]any{"ratio": "lots"})

	var got invalidTarget
	_ = c.Bind("tracing", &got)

	warns := log.GetLogsByLevel("WARN")
	if len(warns) != 1 {
		t.Fatalf("got %d warnings, want 1", len(warns))
	}

	for field, want := range map[string]any{"key": "tracing.ratio", "value": "lots", "expected": "float64"} {
		if v, _ := warns[0].Field(field); v != want {
			t.Fatalf("%s = %#v, want %#v", field, v, want)
		}
	}
}

func TestBindValidValuesDoNotWarn(t *testing.T) {
	c, log := captureConfy(t)
	c.Set("x", map[string]any{
		"ratio": 0.25, "float": "1.5", "int": "12", "uint": 3, "bool": "yes",
		"duration": "30s", "nested": map[string]any{"n": 2}, "list": []any{1, "2"},
		"weights": map[string]any{"a": 1}, "name": "svc",
	})

	var got invalidTarget
	if err := c.Bind("x", &got); err != nil {
		t.Fatal(err)
	}

	if keys := warnKeys(log); len(keys) != 0 {
		t.Fatalf("valid config warned for %v", keys)
	}

	if *got.Ratio != 0.25 || got.Float != 1.5 || got.Int != 12 || got.Uint != 3 || !got.Bool ||
		got.Duration != 30*time.Second || got.Nested.N != 2 || !reflect.DeepEqual(got.List, []int{1, 2}) ||
		got.Weights["a"] != 1 || got.Name != "svc" {
		t.Fatalf("bound %+v", got)
	}

	c.Set("y", map[string]any{"duration": 5})

	var secs invalidTarget
	_ = c.Bind("y", &secs)

	if secs.Duration != 5*time.Second {
		t.Fatalf("int duration bound as %v, want 5s", secs.Duration)
	}
}

func TestBindStrictReportsEveryInvalidValue(t *testing.T) {
	c, log := captureConfy(t)
	c.Set("x", map[string]any{"ratio": "lots", "int": "ten", "name": "ok", "nested": map[string]any{"n": "x"}})

	got := priorTarget()
	err := c.BindStrict("x", &got)

	var be *BindError
	if !errors.As(err, &be) {
		t.Fatalf("BindStrict returned %v, want a *BindError", err)
	}

	var keys []string
	for _, f := range be.Fields {
		keys = append(keys, f.Key)
	}

	if want := []string{"x.ratio", "x.int", "x.nested.n"}; !reflect.DeepEqual(keys, want) {
		t.Fatalf("keys = %v, want %v", keys, want)
	}

	if got.Name != "ok" || got.Int != 7 || got.Ratio != nil {
		t.Fatalf("bound %+v", got)
	}

	if n := len(log.GetLogsByLevel("WARN")); n != 0 {
		t.Fatalf("BindStrict logged %d warnings, want none", n)
	}

	var _ StrictBinder = c
}

func TestBindStrictMissingKeyStillErrors(t *testing.T) {
	c, _ := captureConfy(t)

	var got invalidTarget

	err := c.BindStrict("nope", &got)

	var be *BindError
	if err == nil || errors.As(err, &be) {
		t.Fatalf("missing key returned %v, want the existing not-found error", err)
	}
}

func TestBindWithDefaultBindsDurationString(t *testing.T) {
	c, log := captureConfy(t)
	c.Set("x", map[string]any{"duration": "30s"})

	var got invalidTarget
	if err := c.BindWithDefault("x", &got, invalidTarget{}); err != nil {
		t.Fatal(err)
	}

	if got.Duration != 30*time.Second {
		t.Fatalf("Duration = %v, want 30s", got.Duration)
	}

	if keys := warnKeys(log); len(keys) != 0 {
		t.Fatalf("warned for %v", keys)
	}
}

func TestBindWithDefaultIgnoresInvalidValues(t *testing.T) {
	c, log := captureConfy(t)
	c.Set("x", map[string]any{"int": "ten", "name": "svc"})

	got := priorTarget()
	if err := c.BindWithDefault("x", &got, priorTarget()); err != nil {
		t.Fatalf("BindWithDefault returned %v, want nil", err)
	}

	if got.Int != 7 || got.Name != "svc" {
		t.Fatalf("bound %+v", got)
	}

	if keys := warnKeys(log); !reflect.DeepEqual(keys, []string{"x.int"}) {
		t.Fatalf("warned for %v, want [x.int]", keys)
	}
}

// Deep merge re-binds the struct's current values, which are typed
// (time.Duration, named ints). Those must not be reported.
func TestBindWithDefaultDeepMergeDoesNotWarnOnTypedValues(t *testing.T) {
	type withLevel struct {
		Level    level         `yaml:"level"`
		Duration time.Duration `yaml:"duration"`
		Inner    invalidNested `yaml:"inner"`
	}

	c, log := captureConfy(t)
	c.Set("x", map[string]any{"inner": map[string]any{"s": "a"}})

	got := withLevel{Level: 2, Duration: time.Minute, Inner: invalidNested{N: 4}}
	if err := c.BindWithDefault("x", &got, withLevel{Level: 2, Duration: time.Minute}); err != nil {
		t.Fatal(err)
	}

	if keys := warnKeys(log); len(keys) != 0 {
		t.Fatalf("warned for %v", keys)
	}

	if got.Level != 2 || got.Duration != time.Minute || got.Inner.S != "a" {
		t.Fatalf("bound %+v", got)
	}
}

func TestBindWithOptionsErrorOnMissingReturnsBindError(t *testing.T) {
	c, log := captureConfy(t)
	c.Set("x", map[string]any{"int": "ten", "float": "lots"})

	opts := configcore.DefaultBindOptions()
	opts.ErrorOnMissing = true

	got := priorTarget()
	err := c.BindWithOptions("x", &got, opts)

	var be *BindError
	if !errors.As(err, &be) || len(be.Fields) != 2 {
		t.Fatalf("err = %v, want a *BindError with 2 fields", err)
	}

	if n := len(log.GetLogsByLevel("WARN")); n != 0 {
		t.Fatalf("logged %d warnings with ErrorOnMissing, want none", n)
	}
}

func TestBindStructValuedMapEntryThatIsNotAMappingDoesNotPanic(t *testing.T) {
	type byValue struct {
		Servers map[string]invalidNested `yaml:"servers"`
	}

	type byPointer struct {
		Servers map[string]*invalidNested `yaml:"servers"`
	}

	section := map[string]any{"servers": map[string]any{
		"a": "oops",
		"b": map[string]any{"n": 3, "s": "ok"},
	}}

	t.Run("value", func(t *testing.T) {
		c, log := captureConfy(t)
		c.Set("x", section)

		var got byValue
		if err := c.Bind("x", &got); err != nil {
			t.Fatal(err)
		}

		if keys := warnKeys(log); !reflect.DeepEqual(keys, []string{"x.servers.a"}) {
			t.Fatalf("warned for %v, want [x.servers.a]", keys)
		}

		if _, ok := got.Servers["a"]; ok || got.Servers["b"].N != 3 {
			t.Fatalf("bound %+v", got.Servers)
		}
	})

	t.Run("pointer", func(t *testing.T) {
		c, log := captureConfy(t)
		c.Set("x", section)

		var got byPointer
		if err := c.Bind("x", &got); err != nil {
			t.Fatal(err)
		}

		if keys := warnKeys(log); !reflect.DeepEqual(keys, []string{"x.servers.a"}) {
			t.Fatalf("warned for %v, want [x.servers.a]", keys)
		}

		if _, ok := got.Servers["a"]; ok || got.Servers["b"] == nil || got.Servers["b"].N != 3 {
			t.Fatalf("bound %+v", got.Servers)
		}
	})
}

func TestBindWithDefaultStructValuedMapEntryDoesNotPanicOnMerge(t *testing.T) {
	type byValue struct {
		Servers map[string]invalidNested `yaml:"servers"`
	}

	type byPointer struct {
		Servers map[string]*invalidNested `yaml:"servers"`
	}

	section := map[string]any{"servers": map[string]any{
		"a": "oops",
		"b": map[string]any{"n": 3, "s": "ok"},
	}}

	t.Run("value", func(t *testing.T) {
		c, log := captureConfy(t)
		c.Set("x", section)

		got := byValue{Servers: map[string]invalidNested{"keep": {N: 1}}}
		if err := c.BindWithDefault("x", &got, byValue{Servers: map[string]invalidNested{"keep": {N: 1}}}); err != nil {
			t.Fatal(err)
		}

		if keys := warnKeys(log); !reflect.DeepEqual(keys, []string{"x.servers.a"}) {
			t.Fatalf("warned for %v, want [x.servers.a]", keys)
		}

		if _, ok := got.Servers["a"]; ok || got.Servers["b"].N != 3 || got.Servers["keep"].N != 1 {
			t.Fatalf("bound %+v", got.Servers)
		}
	})

	t.Run("pointer", func(t *testing.T) {
		c, log := captureConfy(t)
		c.Set("x", section)

		def := byPointer{Servers: map[string]*invalidNested{"keep": {N: 1}}}
		got := byPointer{Servers: map[string]*invalidNested{"keep": {N: 1}}}

		if err := c.BindWithDefault("x", &got, def); err != nil {
			t.Fatal(err)
		}

		if keys := warnKeys(log); !reflect.DeepEqual(keys, []string{"x.servers.a"}) {
			t.Fatalf("warned for %v, want [x.servers.a]", keys)
		}

		if _, ok := got.Servers["a"]; ok || got.Servers["b"] == nil || got.Servers["b"].N != 3 ||
			got.Servers["keep"] == nil || got.Servers["keep"].N != 1 {
			t.Fatalf("bound %+v", got.Servers)
		}
	})
}

// An existing struct entry merges with the new keys and reports a bad one, the
// same as a pointer entry does.
func TestBindWithDefaultMergesIntoExistingStructMapEntry(t *testing.T) {
	type byValue struct {
		Servers map[string]invalidNested `yaml:"servers"`
	}

	c, log := captureConfy(t)
	c.Set("x", map[string]any{"servers": map[string]any{
		"b": map[string]any{"n": "bad", "s": "new"},
	}})

	got := byValue{Servers: map[string]invalidNested{"b": {N: 1}}}
	if err := c.BindWithDefault("x", &got, byValue{Servers: map[string]invalidNested{"b": {N: 1}}}); err != nil {
		t.Fatal(err)
	}

	if want := (invalidNested{N: 1, S: "new"}); got.Servers["b"] != want {
		t.Fatalf("Servers[b] = %+v, want %+v", got.Servers["b"], want)
	}

	if keys := warnKeys(log); !reflect.DeepEqual(keys, []string{"x.servers.b.n"}) {
		t.Fatalf("warned for %v, want [x.servers.b.n]", keys)
	}
}

func TestBindReportsMapProblemsInKeyOrder(t *testing.T) {
	c, _ := captureConfy(t)
	c.Set("x", map[string]any{"weights": map[string]any{
		"d": "four", "a": "one", "c": 3, "b": "two",
	}})

	var want []string
	for _, k := range []string{"a", "b", "d"} {
		want = append(want, "x.weights."+k)
	}

	for range 20 {
		var got invalidTarget

		var be *BindError
		if err := c.BindStrict("x", &got); !errors.As(err, &be) {
			t.Fatalf("err = %v, want a *BindError", err)
		}

		var keys []string
		for _, f := range be.Fields {
			keys = append(keys, f.Key)
		}

		if !reflect.DeepEqual(keys, want) {
			t.Fatalf("keys = %v, want %v", keys, want)
		}
	}
}
