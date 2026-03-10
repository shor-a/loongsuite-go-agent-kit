// Copyright (c) 2025 Alibaba Group Holding Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pahomqtt

import (
	"context"
	_ "unsafe"

	"github.com/alibaba/loongsuite-go-agent/pkg/api"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// messageHandlerOnEnter intercepts when a message handler callback is invoked
// This is called when a subscribed message is received and the callback is executed
//
//go:linkname messageHandlerOnEnter github.com/eclipse/paho.mqtt.golang.messageHandlerOnEnter
func messageHandlerOnEnter(call api.CallContext, client mqtt.Client, msg mqtt.Message) {
	if !IsEnabled() {
		return
	}

	if msg == nil {
		return
	}

	request := MqttSubscribeRequest{
		Topic:     msg.Topic(),
		QoS:       msg.Qos(),
		MessageID: msg.MessageID(),
		BodySize:  int64(len(msg.Payload())),
		Headers:   make(map[string]string),
	}

	// Try to extract trace context from message payload if it's JSON with trace headers
	// This is a best-effort approach since MQTT doesn't have native header support
	extractTraceContextFromPayload(msg.Payload(), request.Headers)

	// Start consumer span
	ctx := context.Background()
	var attributes []attribute.KeyValue
	attributes = append(attributes,
		attribute.Int("messaging.mqtt.qos", int(msg.Qos())),
		attribute.Int("messaging.mqtt.message_id", int(msg.MessageID())),
		attribute.Bool("messaging.mqtt.duplicate", msg.Duplicate()),
		attribute.Bool("messaging.mqtt.retained", msg.Retained()),
	)

	ctx = MqttSubscribeInstrumenter.Start(ctx, request, trace.WithAttributes(attributes...))

	// Store context and request for OnExit
	data := make(map[string]interface{})
	data["ctx"] = ctx
	data["mqtt_subscribe_request"] = request
	call.SetData(data)
}

//go:linkname messageHandlerOnExit github.com/eclipse/paho.mqtt.golang.messageHandlerOnExit
func messageHandlerOnExit(call api.CallContext) {
	if !IsEnabled() {
		return
	}

	data, ok := call.GetData().(map[string]interface{})
	if !ok {
		return
	}

	ctx, ok := data["ctx"].(context.Context)
	if !ok {
		return
	}

	request, ok := data["mqtt_subscribe_request"].(MqttSubscribeRequest)
	if !ok {
		return
	}

	MqttSubscribeInstrumenter.End(ctx, request, nil, nil)
}

// subscribeOnEnter intercepts Subscribe calls
//
//go:linkname subscribeOnEnter github.com/eclipse/paho.mqtt.golang.subscribeOnEnter
func subscribeOnEnter(call api.CallContext, client mqtt.Client, topic string, qos byte, callback mqtt.MessageHandler) {
	if !IsEnabled() {
		return
	}

	// We don't create a span for the subscribe call itself
	// The actual message processing spans are created in messageHandlerOnEnter
	// This is just for potential future enhancements like tracking subscriptions
}

//go:linkname subscribeOnExit github.com/eclipse/paho.mqtt.golang.subscribeOnExit
func subscribeOnExit(call api.CallContext, token mqtt.Token) {
	// No-op for now, subscription tracking could be added here
}

// subscribeMultipleOnEnter intercepts SubscribeMultiple calls
//
//go:linkname subscribeMultipleOnEnter github.com/eclipse/paho.mqtt.golang.subscribeMultipleOnEnter
func subscribeMultipleOnEnter(call api.CallContext, client mqtt.Client, filters map[string]byte, callback mqtt.MessageHandler) {
	if !IsEnabled() {
		return
	}
	// Similar to subscribeOnEnter, we track at message handler level
}

//go:linkname subscribeMultipleOnExit github.com/eclipse/paho.mqtt.golang.subscribeMultipleOnExit
func subscribeMultipleOnExit(call api.CallContext, token mqtt.Token) {
	// No-op for now
}

// extractTraceContextFromPayload attempts to extract W3C trace context from JSON payload
// This is a best-effort approach for trace propagation in MQTT
// Expected format: {"traceparent": "...", "tracestate": "...", "payload": ...}
func extractTraceContextFromPayload(payload []byte, headers map[string]string) {
	if len(payload) == 0 {
		return
	}

	// Simple JSON parsing for trace headers
	// Looking for "traceparent" and "tracestate" fields
	// This is intentionally simple to avoid heavy JSON parsing overhead

	payloadStr := string(payload)

	// Try to find traceparent
	if idx := findJSONField(payloadStr, "traceparent"); idx >= 0 {
		if value := extractJSONStringValue(payloadStr[idx:]); value != "" {
			headers["traceparent"] = value
		}
	}

	// Try to find tracestate
	if idx := findJSONField(payloadStr, "tracestate"); idx >= 0 {
		if value := extractJSONStringValue(payloadStr[idx:]); value != "" {
			headers["tracestate"] = value
		}
	}
}

// findJSONField finds the position of a JSON field in a string
func findJSONField(s, field string) int {
	// Look for "field": pattern
	pattern := `"` + field + `"`
	for i := 0; i <= len(s)-len(pattern); i++ {
		if s[i:i+len(pattern)] == pattern {
			return i + len(pattern)
		}
	}
	return -1
}

// extractJSONStringValue extracts a string value after a colon in JSON
func extractJSONStringValue(s string) string {
	// Skip whitespace and colon
	i := 0
	for i < len(s) && (s[i] == ' ' || s[i] == '\t' || s[i] == ':') {
		i++
	}

	if i >= len(s) || s[i] != '"' {
		return ""
	}

	// Find the closing quote
	i++ // skip opening quote
	start := i
	for i < len(s) && s[i] != '"' {
		if s[i] == '\\' && i+1 < len(s) {
			i++ // skip escaped character
		}
		i++
	}

	if i >= len(s) {
		return ""
	}

	return s[start:i]
}
