# Errgroup Instrumentation

This package provides automatic OpenTelemetry instrumentation for `golang.org/x/sync/errgroup`.

## Features

- Automatic child span creation for each `Go()` and `TryGo()` task
- Span for `Wait()` operation
- Error recording and status propagation
- Caller information in span attributes

## Supported Operations

| Operation | Span Name | Description |
|-----------|-----------|-------------|
| `Group.Go(f)` | `errgroup.task:{caller}` | Creates child span for async task |
| `Group.TryGo(f)` | `errgroup.trygo:{caller}` | Creates child span for try-go task |
| `Group.Wait()` | `errgroup.wait` | Creates span for wait operation |

## Attributes

| Attribute | Type | Description |
|-----------|------|-------------|
| `errgroup.task.caller` | string | Function name that spawned the task |
| `errgroup.task.async` | bool | Always `true` |
| `errgroup.task.trygo` | bool | `true` for TryGo tasks |

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `OTEL_INSTRUMENTATION_ERRGROUP_ENABLED` | `true` | Enable/disable errgroup instrumentation |

## Example Trace

```
[Trace: abc123]
└── handle-request (50ms)
    ├── errgroup.task:processItem (10ms)
    ├── errgroup.task:processItem (12ms)
    ├── errgroup.task:processItem (8ms)
    └── errgroup.wait (15ms)
```

## Usage

No code changes required. When building with `otel go build`, errgroup operations are automatically instrumented.

```go
// This code is automatically instrumented
g, ctx := errgroup.WithContext(ctx)

g.Go(func() error {
    // This runs in a child span: "errgroup.task:main.processItem"
    return processItem(ctx, item1)
})

g.Go(func() error {
    // This runs in a child span: "errgroup.task:main.processItem"
    return processItem(ctx, item2)
})

// Wait span is created here
err := g.Wait()
```
