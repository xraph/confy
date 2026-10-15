# confy

A configuration library for Go that doesn't get in your way.

```go
cfg := confy.New(confy.Config{})
cfg.LoadFrom(sources.NewFileSource("config.yaml", sources.FileSourceOptions{}))

port := cfg.GetInt("server.port", 8080)
debug := cfg.GetBool("debug", false)
timeout := cfg.GetDuration("timeout", 30*time.Second)
```

## Why confy?

Most config libraries either do too little or too much. confy sits in the middle:

- **Multiple sources** - files, environment variables, Consul, Kubernetes ConfigMaps
- **Type-safe getters** with sensible defaults
- **Struct binding** when you want it
- **Hot reload** with file watching
- **Auto-discovery** that finds your config files

No magic globals. No init() surprises. Just a struct you control.

## Install

```bash
go get github.com/xraph/confy
```

## Quick Start

### Load from a file

```go
package main

import (
    "github.com/xraph/confy"
    "github.com/xraph/confy/sources"
)

func main() {
    cfg := confy.New(confy.Config{})
    
    source, _ := sources.NewFileSource("config.yaml", sources.FileSourceOptions{
        WatchEnabled: true,
    })
    cfg.LoadFrom(source)
    
    // Get values with defaults
    host := cfg.GetString("database.host", "localhost")
    port := cfg.GetInt("database.port", 5432)
    maxConns := cfg.GetInt("database.max_connections", 10)
}
```

### Auto-discover config files

confy can find your config files automatically. It searches the current directory and parent directories for `config.yaml` and `config.local.yaml`.

```go
cfg, err := confy.AutoLoadConfy("myapp", nil)
if err != nil {
    log.Fatal(err)
}
```

This is useful for monorepos where your app might be nested several directories deep.

### Bind to a struct

```go
type DatabaseConfig struct {
    Host        string        `yaml:"host"`
    Port        int           `yaml:"port"`
    MaxConns    int           `yaml:"max_connections" default:"10"`
    IdleTimeout time.Duration `yaml:"idle_timeout" default:"5m"`
}

var dbCfg DatabaseConfig
cfg.Bind("database", &dbCfg)
```

The `default` tag works when the key is missing from your config.

### Environment variables

```go
envSource := sources.NewEnvSource(sources.EnvSourceOptions{
    Prefix:    "MYAPP_",
    Separator: "_",
})
cfg.LoadFrom(envSource)

// MYAPP_DATABASE_HOST becomes database.host
host := cfg.GetString("database.host")
```

### Watch for changes

```go
cfg.WatchChanges(func(change confy.ConfigChange) {
    log.Printf("config changed: %s = %v", change.Key, change.NewValue)
})

cfg.Watch(context.Background())
```

## Invalid values

`Bind` doesn't fail when a value can't be converted. Say your config has this:

```yaml
tracing:
  sample_ratio: "lots"
```

and you bind it into a struct with a `float64` field. `Bind` returns nil and the field keeps whatever it held before the call, so a value you set in the struct before binding still applies. confy logs one warning for each key it skipped:

```
config: ignoring invalid value  key=tracing.sample_ratio value=lots expected=float64 error=...
```

Every other field binds as usual. Here's what happens to the rest of a struct:

- A pointer to a scalar (`*float64`, `*int`, `*string`) stays nil when its value is bad, and one that already points somewhere keeps its target. A pointer to a struct isn't like that: if a field inside its block is bad, it gets a new pointer with the good fields applied and the bad field at its prior value, and a nil one comes back pointing at a struct.
- A list of scalars with one bad element keeps its old value whole. A list of structs keeps every element, and the bad field inside it is reported.
- A bad map entry is left out of the map. A struct-valued entry that isn't a mapping is reported too.
- Map entries are reported in key order, so the log reads the same on every run.
- An explicit null (`name:` or `name: null`) sets a string field to `""`. Any other field keeps its value, and you get no warning.

confy reports scalars, pointers to scalars, list elements and map entries. It doesn't yet report a list, map or struct given a value of the wrong shape, like `hosts: "a,b"` for a `[]string` or `tls: true` for a `*TLSConfig`. That field keeps its value without a warning, except that a nil pointer to a struct can come back pointing at an empty one.

`BindWithDefault` and `BindWithOptions` log the same warnings. Only `BindWithOptions` takes `ErrorOnMissing`; `BindWithDefault` always has it off. Set `ErrorOnMissing: true` on the options you pass to `BindWithOptions` and it logs nothing, binds the whole struct, then returns the `*BindError` described below.

### Upgrading from v1.0.x

Some values bind differently now:

- `BindWithDefault` binds duration strings such as `"30s"`, which it used to drop. It reads an int bound into a `time.Duration` as seconds, the way `Bind` always has.
- Map values go through the same converter as struct fields. An int into a `map[string]time.Duration` is seconds, and an int into a `map[string]string` becomes `"5"`.
- Under `BindWithDefault`, a struct entry that's already in a map takes the new keys and keeps the rest. It used to keep the old entry whole.
- A pointer field gets a new pointer when its value binds. confy used to write through the old one, so a variable you shared with that pointer changed too. Now it doesn't.
- A map keyed by a named string type (`map[envName]int`) binds. It used to panic. A map keyed by anything else, such as `map[int]int`, can't hold a config key, so each entry is reported and skipped.

### Fail on invalid values

If you'd rather stop at startup than read a log, use `BindStrict`. It binds exactly like `Bind`, logs nothing, and returns a `*confy.BindError` listing every value it skipped. The error comes back after the whole struct is bound, so you see all the problems at once.

`BindStrict` isn't on the `Confy` interface, so reach it through `StrictBinder`. `New` returns a value that satisfies it, and so does `NewTestConfyImpl`.

```go
var s ServerConfig

if sb, ok := cfg.(confy.StrictBinder); ok {
    var be *confy.BindError
    if err := sb.BindStrict("server", &s); errors.As(err, &be) {
        for _, f := range be.Fields {
            fmt.Println(f.Key, f.Value, f.Type)
        }
    }
}
```

`Key` is the full dotted path (`server.port`, or `server.hosts[1]` for a list element), `Value` is what the config held, and `Type` is the Go type confy wanted. `Err` has the converter's own error. Any error that isn't a `*BindError`, such as a missing key, is a real failure.

## Sources

confy supports multiple configuration sources out of the box:

| Source | Description |
|--------|-------------|
| `sources.FileSource` | YAML, JSON, TOML files |
| `sources.EnvSource` | Environment variables |
| `sources.ConsulSource` | HashiCorp Consul KV |
| `sources.K8sConfigMapSource` | Kubernetes ConfigMaps |

Sources have priorities. Higher priority sources override lower ones. By default:

- Base config file: 100
- Local config file: 200  
- Environment variables: 300

### File source

```go
source, err := sources.NewFileSource("config.yaml", sources.FileSourceOptions{
    Priority:      100,
    WatchEnabled:  true,
    ExpandEnvVars: true,  // expands ${VAR} in values
})
```

## Variable Resolution

confy supports environment variable expansion in YAML files with bash-style default value syntax.

### Basic expansion

```yaml
# config.yaml
database:
  host: ${DB_HOST}
  port: ${DB_PORT}
```

```bash
export DB_HOST=localhost
export DB_PORT=5432
```

### Default values

Use bash-style syntax for default values when environment variables aren't set:

```yaml
database:
  # Use default if DB_HOST is unset or empty
  host: ${DB_HOST:-localhost}
  
  # Use default only if DB_PORT is unset (not if empty)
  port: ${DB_PORT-5432}
  
  # Full connection string with multiple defaults
  dsn: postgres://${DB_USER:-postgres}:${DB_PASS:-postgres}@${DB_HOST:-localhost}:${DB_PORT:-5432}/${DB_NAME:-mydb}
```

**Syntax variants:**

| Syntax | Behavior |
|--------|----------|
| `${VAR}` or `$VAR` | Standard expansion, empty string if not set |
| `${VAR:-default}` | Use `default` if VAR is unset **or empty** |
| `${VAR-default}` | Use `default` only if VAR is **unset** (not if empty) |
| `${VAR:=default}` | Assign and use `default` if VAR is unset or empty |
| `${VAR=default}` | Assign and use `default` only if VAR is unset |

**Example with some variables set:**

```yaml
server:
  address: ${SERVER_HOST:-0.0.0.0}:${SERVER_PORT:-8080}
  timeout: ${TIMEOUT:-30s}
```

```bash
export SERVER_PORT=3000
# SERVER_HOST and TIMEOUT will use defaults
# Result: address = "0.0.0.0:3000", timeout = "30s"
```

### Complex examples

**Database connection string:**

```yaml
database:
  dsn: ${DATABASE_DSN:-postgres://postgres:postgres@localhost:5432/testdb?sslmode=disable}
```

This allows you to either:
- Set `DATABASE_DSN` to override the entire connection string, or
- Use individual environment variables with defaults for each component

**Service URLs:**

```yaml
services:
  auth: ${AUTH_URL:-http://localhost:8080}
  api: ${API_URL:-http://localhost:3000}
  cache: ${REDIS_URL:-redis://localhost:6379/0}
```

### Secret references

confy can resolve secret references from external secret managers (AWS Secrets Manager, HashiCorp Vault, etc.):

```yaml
database:
  password: ${secret:db/production/password}
  api_key: ${secret:services/stripe/api_key}
```

Enable secret expansion in your file source:

```go
source, err := sources.NewFileSource("config.yaml", sources.FileSourceOptions{
    ExpandSecrets: true,
})

// Provide a secrets manager implementation
cfg.SetSecretsManager(mySecretsManager)
```

Secret references use the format `${secret:key}` where `key` is passed to your `SecretsManager` implementation.

### Consul source

```go
source, err := sources.NewConsulSource(sources.ConsulSourceOptions{
    Address: "localhost:8500",
    Path:    "myapp/config",
    Token:   os.Getenv("CONSUL_TOKEN"),
})
```

### Kubernetes ConfigMap

```go
source, err := sources.NewK8sConfigMapSource(sources.K8sConfigMapSourceOptions{
    Namespace:     "default",
    ConfigMapName: "myapp-config",
    Key:           "config.yaml",
})
```

## Type-safe getters

Every getter has an optional default value:

```go
cfg.GetString("key")                    // returns "" if missing
cfg.GetString("key", "default")         // returns "default" if missing

cfg.GetInt("port", 8080)
cfg.GetBool("debug", false)
cfg.GetDuration("timeout", 10*time.Second)
cfg.GetFloat64("rate", 1.5)
cfg.GetStringSlice("hosts", []string{"localhost"})
cfg.GetSizeInBytes("max_size", 1024*1024)  // supports "10MB", "1GB" strings
```

## App-scoped config

For monorepos, you can scope config per application:

```yaml
# config.yaml
database:
  host: shared-db.internal

apps:
  api:
    port: 8080
    database:
      host: api-db.internal
  
  worker:
    concurrency: 10
```

```go
cfg, _ := confy.LoadConfigWithAppScope("api", logger, nil)

// Gets "api-db.internal" - app config overrides global
host := cfg.GetString("database.host")
```

## Validation

```go
validator := confy.NewValidator(confy.ValidatorConfig{
    Mode: confy.ValidationModeStrict,
})

validator.AddRule(confy.ValidationRule{
    Key:      "server.port",
    Required: true,
    Min:      1,
    Max:      65535,
})

if err := cfg.Validate(); err != nil {
    log.Fatal(err)
}
```

## Testing

confy includes a test implementation for unit tests:

```go
func TestMyHandler(t *testing.T) {
    cfg := confy.NewTestConfyImpl()
    cfg.Set("feature.enabled", true)
    cfg.Set("timeout", "5s")
    
    handler := NewHandler(cfg)
    // ...
}
```

`TestConfyImpl` implements `StrictBinder` too, so you can test how your code reacts to a bad value:

```go
func TestServerRejectsBadPort(t *testing.T) {
    cfg := confy.NewTestConfyImpl()
    cfg.Set("server", map[string]any{"port": "eighty"})

    var s ServerConfig
    err := cfg.(confy.StrictBinder).BindStrict("server", &s)

    var be *confy.BindError
    if !errors.As(err, &be) || be.Fields[0].Key != "server.port" {
        t.Fatalf("got %v, want a BindError for server.port", err)
    }
}
```

`TestConfyImpl.Bind` still uses the double's own simple binder. Only `BindStrict` goes through the real one.

Or use the builder:

```go
cfg := confy.NewTestConfigBuilder().
    WithString("api.url", "http://test.local").
    WithInt("retry.count", 3).
    WithBool("debug", true).
    Build()
```

## License

MIT

