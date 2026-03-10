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

package solace

import (
	"context"
	_ "unsafe"

	"github.com/alibaba/loongsuite-go-agent/pkg/api"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"solace.dev/go/messaging/pkg/solace/message"
)

// directReceiverMessageHandlerOnEnter intercepts when a message handler callback is invoked
// for DirectMessageReceiver
//
//go:linkname directReceiverMessageHandlerOnEnter solace.dev/go/messaging/pkg/solace.directReceiverMessageHandlerOnEnter
func directReceiverMessageHandlerOnEnter(call api.CallContext, msg message.InboundMessage) {
	if !IsEnabled() {
		return
	}

	if msg == nil {
		return
	}

	// Extract message properties
	var bodySize int64
	if payload, ok := msg.GetPayloadAsBytes(); ok {
		bodySize = int64(len(payload))
	}

	// Get destination topic
	destName := ""
	if dest := msg.GetDestinationName(); dest != "" {
		destName = dest
	}

	// Get user properties for trace context extraction
	headers := make(map[string]string)
	if props, ok := msg.GetProperties(); ok {
		for key, val := range props {
			if strVal, ok := val.(string); ok {
				headers[key] = strVal
			}
		}
	}

	request := SolaceConsumeRequest{
		Topic:           destName,
		DestinationType: "topic",
		BodySize:        bodySize,
		Headers:         headers,
	}

	// Get message ID and correlation ID if available
	if msgID, ok := msg.GetApplicationMessageId(); ok {
		request.MessageID = msgID
	}
	if corrID, ok := msg.GetCorrelationId(); ok {
		request.CorrelationID = corrID
	}
	if redeliveryCount, ok := msg.GetRedeliveryCount(); ok {
		request.RedeliveryCount = int(redeliveryCount)
	}

	// Start consumer span
	ctx := context.Background()
	var attributes []attribute.KeyValue
	attributes = append(attributes,
		attribute.String("messaging.solace.destination_type", request.DestinationType),
		attribute.Int("messaging.solace.redelivery_count", request.RedeliveryCount),
	)

	ctx = SolaceConsumeInstrumenter.Start(ctx, request, trace.WithAttributes(attributes...))

	// Store context and request for OnExit
	data := make(map[string]interface{})
	data["ctx"] = ctx
	data["solace_consume_request"] = request
	call.SetData(data)
}

//go:linkname directReceiverMessageHandlerOnExit solace.dev/go/messaging/pkg/solace.directReceiverMessageHandlerOnExit
func directReceiverMessageHandlerOnExit(call api.CallContext) {
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

	request, ok := data["solace_consume_request"].(SolaceConsumeRequest)
	if !ok {
		return
	}

	SolaceConsumeInstrumenter.End(ctx, request, nil, nil)
}

// persistentReceiverMessageHandlerOnEnter intercepts when a message handler callback is invoked
// for PersistentMessageReceiver
//
//go:linkname persistentReceiverMessageHandlerOnEnter solace.dev/go/messaging/pkg/solace.persistentReceiverMessageHandlerOnEnter
func persistentReceiverMessageHandlerOnEnter(call api.CallContext, msg message.InboundMessage) {
	if !IsEnabled() {
		return
	}

	if msg == nil {
		return
	}

	// Extract message properties
	var bodySize int64
	if payload, ok := msg.GetPayloadAsBytes(); ok {
		bodySize = int64(len(payload))
	}

	// Get destination
	destName := ""
	destType := "queue" // Persistent receivers typically use queues
	if dest := msg.GetDestinationName(); dest != "" {
		destName = dest
	}

	// Get user properties for trace context extraction
	headers := make(map[string]string)
	if props, ok := msg.GetProperties(); ok {
		for key, val := range props {
			if strVal, ok := val.(string); ok {
				headers[key] = strVal
			}
		}
	}

	request := SolaceConsumeRequest{
		Topic:           destName,
		DestinationType: destType,
		BodySize:        bodySize,
		Headers:         headers,
	}

	// Get message ID and correlation ID if available
	if msgID, ok := msg.GetApplicationMessageId(); ok {
		request.MessageID = msgID
	}
	if corrID, ok := msg.GetCorrelationId(); ok {
		request.CorrelationID = corrID
	}
	if redeliveryCount, ok := msg.GetRedeliveryCount(); ok {
		request.RedeliveryCount = int(redeliveryCount)
	}

	// Start consumer span
	ctx := context.Background()
	var attributes []attribute.KeyValue
	attributes = append(attributes,
		attribute.String("messaging.solace.destination_type", request.DestinationType),
		attribute.Int("messaging.solace.redelivery_count", request.RedeliveryCount),
	)

	ctx = SolaceConsumeInstrumenter.Start(ctx, request, trace.WithAttributes(attributes...))

	// Store context and request for OnExit
	data := make(map[string]interface{})
	data["ctx"] = ctx
	data["solace_consume_request"] = request
	call.SetData(data)
}

//go:linkname persistentReceiverMessageHandlerOnExit solace.dev/go/messaging/pkg/solace.persistentReceiverMessageHandlerOnExit
func persistentReceiverMessageHandlerOnExit(call api.CallContext) {
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

	request, ok := data["solace_consume_request"].(SolaceConsumeRequest)
	if !ok {
		return
	}

	SolaceConsumeInstrumenter.End(ctx, request, nil, nil)
}

// receiveMessageOnEnter intercepts synchronous ReceiveMessage calls
//
//go:linkname receiveMessageOnEnter solace.dev/go/messaging/pkg/solace.receiveMessageOnEnter
func receiveMessageOnEnter(call api.CallContext) {
	// No-op for enter, we trace on exit when we have the message
}

//go:linkname receiveMessageOnExit solace.dev/go/messaging/pkg/solace.receiveMessageOnExit
func receiveMessageOnExit(call api.CallContext, msg message.InboundMessage, err error) {
	if !IsEnabled() {
		return
	}

	if msg == nil || err != nil {
		return
	}

	// Extract message properties
	var bodySize int64
	if payload, ok := msg.GetPayloadAsBytes(); ok {
		bodySize = int64(len(payload))
	}

	// Get destination
	destName := ""
	if dest := msg.GetDestinationName(); dest != "" {
		destName = dest
	}

	// Get user properties for trace context extraction
	headers := make(map[string]string)
	if props, ok := msg.GetProperties(); ok {
		for key, val := range props {
			if strVal, ok := val.(string); ok {
				headers[key] = strVal
			}
		}
	}

	request := SolaceConsumeRequest{
		Topic:           destName,
		DestinationType: "queue",
		BodySize:        bodySize,
		Headers:         headers,
	}

	// Get message ID and correlation ID if available
	if msgID, ok := msg.GetApplicationMessageId(); ok {
		request.MessageID = msgID
	}
	if corrID, ok := msg.GetCorrelationId(); ok {
		request.CorrelationID = corrID
	}

	// Start and immediately end the span for synchronous receive
	ctx := context.Background()
	var attributes []attribute.KeyValue
	attributes = append(attributes,
		attribute.String("messaging.solace.destination_type", request.DestinationType),
	)

	ctx = SolaceConsumeInstrumenter.Start(ctx, request, trace.WithAttributes(attributes...))
	SolaceConsumeInstrumenter.End(ctx, request, nil, nil)
}

// MessageHandler is the callback function type for ReceiveAsync
type MessageHandler = func(inboundMessage message.InboundMessage)

// receiveAsyncOnEnter intercepts ReceiveAsync calls and wraps the callback with tracing
//
//go:linkname receiveAsyncOnEnter solace.dev/go/messaging/pkg/solace.receiveAsyncOnEnter
func receiveAsyncOnEnter(call api.CallContext, handler MessageHandler) {
	if !IsEnabled() {
		return
	}

	if handler == nil {
		return
	}

	// Wrap the original handler with tracing
	wrappedHandler := func(msg message.InboundMessage) {
		if msg == nil {
			handler(msg)
			return
		}

		// Extract message properties
		var bodySize int64
		if payload, ok := msg.GetPayloadAsBytes(); ok {
			bodySize = int64(len(payload))
		}

		// Get destination
		destName := ""
		destType := "queue"
		if dest := msg.GetDestinationName(); dest != "" {
			destName = dest
		}

		// Get user properties for trace context extraction
		headers := make(map[string]string)
		if props, ok := msg.GetProperties(); ok {
			for key, val := range props {
				if strVal, ok := val.(string); ok {
					headers[key] = strVal
				}
			}
		}

		request := SolaceConsumeRequest{
			Topic:           destName,
			DestinationType: destType,
			BodySize:        bodySize,
			Headers:         headers,
		}

		// Get message ID and correlation ID if available
		if msgID, ok := msg.GetApplicationMessageId(); ok {
			request.MessageID = msgID
		}
		if corrID, ok := msg.GetCorrelationId(); ok {
			request.CorrelationID = corrID
		}
		if redeliveryCount, ok := msg.GetRedeliveryCount(); ok {
			request.RedeliveryCount = int(redeliveryCount)
		}

		// Start consumer span
		ctx := context.Background()
		var attributes []attribute.KeyValue
		attributes = append(attributes,
			attribute.String("messaging.solace.destination_type", request.DestinationType),
			attribute.Int("messaging.solace.redelivery_count", request.RedeliveryCount),
		)

		ctx = SolaceConsumeInstrumenter.Start(ctx, request, trace.WithAttributes(attributes...))

		// Call original handler
		handler(msg)

		// End span
		SolaceConsumeInstrumenter.End(ctx, request, nil, nil)
	}

	// Replace the handler parameter with the wrapped version
	call.SetParam(0, wrappedHandler)
}


// ackOnEnter intercepts PersistentMessageReceiver.Ack calls
//
//go:linkname ackOnEnter solace.dev/go/messaging/pkg/solace.ackOnEnter
func ackOnEnter(call api.CallContext, msg message.InboundMessage) {
	if !IsEnabled() {
		return
	}

	if msg == nil {
		return
	}

	// Get destination for span name
	destName := ""
	if dest := msg.GetDestinationName(); dest != "" {
		destName = dest
	}

	// Get message ID if available
	msgID := ""
	if id, ok := msg.GetApplicationMessageId(); ok {
		msgID = id
	}

	request := SolaceConsumeRequest{
		Topic:           destName,
		DestinationType: "queue",
		MessageID:       msgID,
	}

	// Start ack span
	ctx := context.Background()
	var attributes []attribute.KeyValue
	attributes = append(attributes,
		attribute.String("messaging.operation.type", "settle"),
		attribute.String("messaging.solace.settle_outcome", "accepted"),
	)

	ctx = SolaceAckInstrumenter.Start(ctx, request, trace.WithAttributes(attributes...))

	// Store context and request for OnExit
	data := make(map[string]interface{})
	data["ctx"] = ctx
	data["solace_ack_request"] = request
	call.SetData(data)
}

//go:linkname ackOnExit solace.dev/go/messaging/pkg/solace.ackOnExit
func ackOnExit(call api.CallContext, err error) {
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

	request, ok := data["solace_ack_request"].(SolaceConsumeRequest)
	if !ok {
		return
	}

	SolaceAckInstrumenter.End(ctx, request, nil, err)
}
