# Solace Native SDK Instrumentation

This package provides OpenTelemetry auto-instrumentation for the Solace PubSub+ Go SDK (`solace.dev/go/messaging`).

## Supported Operations

| Operation | Span Kind | Description |
|-----------|-----------|-------------|
| DirectMessagePublisher.Publish | Producer | Traces direct message publishing |
| PersistentMessagePublisher.Publish | Producer | Traces persistent/guaranteed message publishing |
| DirectMessageReceiver (callback) | Consumer | Traces direct message consumption |
| PersistentMessageReceiver (callback) | Consumer | Traces persistent message consumption |
| ReceiveMessage (sync) | Consumer | Traces synchronous message receive |

## Semantic Conventions

This instrumentation follows the [OpenTelemetry Messaging Semantic Conventions](https://opentelemetry.io/docs/specs/semconv/messaging/).

### Attributes

| Attribute | Type | Description |
|-----------|------|-------------|
| `messaging.system` | string | Always "solace" |
| `messaging.destination.name` | string | Topic or queue name |
| `messaging.operation.type` | string | "publish" or "receive" |
| `messaging.message.body.size` | int64 | Payload size in bytes |
| `messaging.message.id` | string | Application message ID |
| `messaging.message.conversation_id` | string | Correlation ID |
| `messaging.solace.destination_type` | string | "topic" or "queue" |
| `messaging.solace.redelivery_count` | int | Number of redeliveries |

## Trace Context Propagation

The Solace native SDK supports user properties on messages, which are used for W3C TraceContext propagation.

### Publisher Example

```go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/propagation"
    "solace.dev/go/messaging"
    "solace.dev/go/messaging/pkg/solace/message"
)

func publishWithTrace(ctx context.Context, publisher solace.DirectMessagePublisher, topic string, data []byte) {
    // Build message with trace context in user properties
    msgBuilder := messaging.NewOutboundMessageBuilder()
    
    // Inject trace context into user properties
    carrier := propagation.MapCarrier{}
    otel.GetTextMapPropagator().Inject(ctx, carrier)
    
    for key, value := range carrier {
        msgBuilder.WithProperty(key, value)
    }
    
    msg, _ := msgBuilder.
        WithPayload(data).
        Build()
    
    publisher.Publish(msg, resource.TopicOf(topic))
}
```

### Consumer Example

The instrumentation automatically extracts trace context from message user properties.

```go
func messageHandler(msg message.InboundMessage) {
    // Trace context is automatically extracted from user properties
    // The span is created with the extracted parent context
    
    // Process message...
}
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `OTEL_INSTRUMENTATION_SOLACE_ENABLED` | `true` | Enable/disable Solace instrumentation |

## Version Support

- Solace PubSub+ Go SDK: v1.0.0 and above

## Example Trace

```
[Trace: abc123...]
├── my/topic publish (Producer)
│   ├── messaging.system: solace
│   ├── messaging.destination.name: my/topic
│   ├── messaging.operation.type: publish
│   └── messaging.solace.destination_type: topic
│
└── my/topic receive (Consumer)
    ├── messaging.system: solace
    ├── messaging.destination.name: my/topic
    ├── messaging.operation.type: receive
    ├── messaging.solace.destination_type: topic
    └── messaging.solace.redelivery_count: 0
```

## Comparison with MQTT Instrumentation

| Feature | Solace Native SDK | Paho MQTT |
|---------|-------------------|-----------|
| User Properties | ✅ Native support | ❌ Requires payload wrapping |
| Trace Propagation | ✅ Automatic | ⚠️ Manual JSON envelope |
| Queue Support | ✅ Full support | ❌ Topics only |
| Guaranteed Delivery | ✅ Full support | ⚠️ QoS 1/2 only |

For applications using Solace PubSub+, the native SDK is recommended over MQTT for better trace propagation support.
