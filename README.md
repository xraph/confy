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

