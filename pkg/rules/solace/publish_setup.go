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
	solace "solace.dev/go/messaging"
	"solace.dev/go/messaging/pkg/solace/message"
	"solace.dev/go/messaging/pkg/solace/resource"
)

// directPublisherPublishOnEnter intercepts DirectMessagePublisher.Publish calls
//
//go:linkname directPublisherPublishOnEnter solace.dev/go/messaging/pkg/solace.directPublisherPublishOnEnter
func directPublisherPublishOnEnter(call api.CallContext, publisher solace.DirectMessagePublisher, msg message.OutboundMessage, dest resource.Destination) {
	if !IsEnabled() {
		return
	}

	// Extract message properties
	var bodySize int64
	if payload, ok := msg.GetPayloadAsBytes(); ok {
		bodySize = int64(len(payload))
	}

	// Get destination name
	destName := ""
	if dest != nil {
		destName = dest.GetName()
	}

	// Get user properties for trace context
	headers := make(map[string]string)
	if props, ok := msg.GetProperties(); ok {
		for key, val := range props {
			if strVal, ok := val.(string); ok {
				headers[key] = strVal
			}
		}
	}

	request := SolacePublishRequest{
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

	// Start the span
	ctx := context.Background()
	var attributes []attribute.KeyValue
	attributes = append(attributes,
		attribute.String("messaging.solace.destination_type", request.DestinationType),
	)

	ctx = SolacePublishInstrumenter.Start(ctx, request, trace.WithAttributes(attributes...))

	// Store context and request for OnExit
	data := make(map[string]interface{})
	data["ctx"] = ctx
	data["solace_publish_request"] = request
	call.SetData(data)
}

//go:linkname directPublisherPublishOnExit solace.dev/go/messaging/pkg/solace.directPublisherPublishOnExit
func directPublisherPublishOnExit(call api.CallContext, err error) {
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

	request, ok := data["solace_publish_request"].(SolacePublishRequest)
	if !ok {
		return
	}

	SolacePublishInstrumenter.End(ctx, request, nil, err)
}

// persistentPublisherPublishOnEnter intercepts PersistentMessagePublisher.Publish calls
//
//go:linkname persistentPublisherPublishOnEnter solace.dev/go/messaging/pkg/solace.persistentPublisherPublishOnEnter
func persistentPublisherPublishOnEnter(call api.CallContext, publisher solace.PersistentMessagePublisher, msg message.OutboundMessage, dest resource.Destination) {
	if !IsEnabled() {
		return
	}

	// Extract message properties
	var bodySize int64
	if payload, ok := msg.GetPayloadAsBytes(); ok {
		bodySize = int64(len(payload))
	}

	// Get destination name and type
	destName := ""
	destType := "topic"
	if dest != nil {
		destName = dest.GetName()
		// Check if it's a queue
		if _, isQueue := dest.(*resource.Queue); isQueue {
			destType = "queue"
		}
	}

	// Get user properties for trace context
	headers := make(map[string]string)
	if props, ok := msg.GetProperties(); ok {
		for key, val := range props {
			if strVal, ok := val.(string); ok {
				headers[key] = strVal
			}
		}
	}

	request := SolacePublishRequest{
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

	// Start the span
	ctx := context.Background()
	var attributes []attribute.KeyValue
	attributes = append(attributes,
		attribute.String("messaging.solace.destination_type", request.DestinationType),
	)

	ctx = SolacePublishInstrumenter.Start(ctx, request, trace.WithAttributes(attributes...))

	// Store context and request for OnExit
	data := make(map[string]interface{})
	data["ctx"] = ctx
	data["solace_publish_request"] = request
	call.SetData(data)
}

//go:linkname persistentPublisherPublishOnExit solace.dev/go/messaging/pkg/solace.persistentPublisherPublishOnExit
func persistentPublisherPublishOnExit(call api.CallContext, publishReceipt solace.PublishReceipt, err error) {
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

	request, ok := data["solace_publish_request"].(SolacePublishRequest)
	if !ok {
		return
	}

	SolacePublishInstrumenter.End(ctx, request, nil, err)
}
