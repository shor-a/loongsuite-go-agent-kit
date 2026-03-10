# Paho MQTT Instrumentation

This package provides OpenTelemetry auto-instrumentation for the Eclipse Paho MQTT Go client (`github.com/eclipse/paho.mqtt.golang`).

## Supported Operations

| Operation | Span Kind | Description |
|-----------|-----------|-------------|
| Publish | Producer | Traces MQTT message publishing |
| Subscribe/Message Handler | Consumer | Traces MQTT message consumption |

## Semantic Conventions

This instrumentation follows the [OpenTelemetry Messaging Semantic Conventions](https://opentelemetry.io/docs/specs/semconv/messaging/).

### Attributes

| Attribute | Type | Description |
|-----------|------|-------------|
| `messaging.system` | string | Always "mqtt" |
| `messaging.destination.name` | string | MQTT topic |
| `messaging.operation.type` | string | "publish" or "receive" |
| `messaging.message.body.size` | int64 | Payload size in bytes |
| `messaging.mqtt.qos` | int | QoS level (0, 1, or 2) |
| `messaging.mqtt.retained` | bool | Whether message is retained |
| `messaging.mqtt.message_id` | int | MQTT message ID |
| `messaging.mqtt.duplicate` | bool | Whether message is a duplicate |

## Trace Context Propagation

Since MQTT protocol doesn't natively support message headers, trace context propagation requires embedding trace information in the message payload.

### Recommended Payload Format

For trace context propagation to work, wrap your payload in a JSON envelope:

```json
{
  "traceparent": "00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01",
  "tracestate": "congo=t61rcWkgMzE",
  "payload": {
    "your": "actual data"
  }
}
```

### Publisher Example

```go
import (
    "encoding/json"
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/propagation"
)

func publishWithTrace(ctx context.Context, client mqtt.Client, topic string, data interface{}) {
    // Create carrier for trace context
    carrier := propagation.MapCarrier{}
    otel.GetTextMapPropagator().Inject(ctx, carrier)
    
    // Wrap payload with trace context
    envelope := map[string]interface{}{
        "traceparent": carrier.Get("traceparent"),
        "tracestate":  carrier.Get("tracestate"),
        "payload":     data,
    }
    
    payload, _ := json.Marshal(envelope)
    client.Publish(topic, 1, false, payload)
}
```

### Consumer Example

The instrumentation automatically extracts trace context from JSON payloads that contain `traceparent` and `tracestate` fields.

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `OTEL_INSTRUMENTATION_MQTT_ENABLED` | `true` | Enable/disable MQTT instrumentation |

## Limitations

1. **No Native Header Support**: MQTT protocol doesn't support message headers, so trace context must be embedded in the payload.

2. **JSON Payload Assumption**: Automatic trace extraction assumes JSON-formatted payloads with trace fields at the root level.

3. **Callback-based Consumption**: Message consumption is traced at the callback handler level, not at the network receive level.

## Version Support

- Paho MQTT Go Client: v1.3.0 and above

## Example Trace

```
[Trace: abc123...]
├── my/topic publish (Producer)
│   ├── messaging.system: mqtt
│   ├── messaging.destination.name: my/topic
│   ├── messaging.operation.type: publish
│   └── messaging.mqtt.qos: 1
│
└── my/topic receive (Consumer)
    ├── messaging.system: mqtt
    ├── messaging.destination.name: my/topic
    ├── messaging.operation.type: receive
    └── messaging.mqtt.qos: 1
```

## Integration with Solace

This instrumentation works with Solace PubSub+ when using the MQTT protocol. For Solace-specific features, consider using the Solace native SDK instrumentation (if available) or the SMF protocol which supports user properties for header propagation.
