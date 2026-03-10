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

// clientInfo stores client metadata for instrumentation
type clientInfo struct {
	brokerURL string
	clientID  string
}

// clientRegistry stores client information by client pointer
var clientRegistry = make(map[uintptr]clientInfo)

//go:linkname afterNewClient github.com/eclipse/paho.mqtt.golang.afterNewClient
func afterNewClient(call api.CallContext, client mqtt.Client) {
	if !IsEnabled() {
		return
	}
	// Store client info for later use in publish/subscribe
	// Note: We can't easily get broker URL from client after creation
	// This would need to be enhanced based on actual paho.mqtt.golang internals
}

//go:linkname publishOnEnter github.com/eclipse/paho.mqtt.golang.publishOnEnter
func publishOnEnter(call api.CallContext, client mqtt.Client, topic string, qos byte, retained bool, payload interface{}) {
	if !IsEnabled() {
		return
	}

	// Calculate payload size
	var bodySize int64
	switch p := payload.(type) {
	case []byte:
		bodySize = int64(len(p))
	case string:
		bodySize = int64(len(p))
	default:
		bodySize = 0
	}

	request := MqttPublishRequest{
		Topic:    topic,
		QoS:      qos,
		Retained: retained,
		BodySize: bodySize,
		Headers:  make(map[string]string),
	}

	// Start the span with producer kind
	ctx := context.Background()
	var attributes []attribute.KeyValue
	attributes = append(attributes,
		attribute.Int("messaging.mqtt.qos", int(qos)),
		attribute.Bool("messaging.mqtt.retained", retained),
	)

	ctx = MqttPublishInstrumenter.Start(ctx, request, trace.WithAttributes(attributes...))

	// Store context and request for OnExit
	data := make(map[string]interface{})
	data["ctx"] = ctx
	data["mqtt_publish_request"] = request
	call.SetData(data)
}

//go:linkname publishOnExit github.com/eclipse/paho.mqtt.golang.publishOnExit
func publishOnExit(call api.CallContext, token mqtt.Token) {
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

	request, ok := data["mqtt_publish_request"].(MqttPublishRequest)
	if !ok {
		return
	}

	// Check for errors from the token
	var err error
	if token != nil {
		token.Wait()
		err = token.Error()
	}

	MqttPublishInstrumenter.End(ctx, request, nil, err)
}

// publishWithContextOnEnter handles Publish calls that include context
//
//go:linkname publishWithContextOnEnter github.com/eclipse/paho.mqtt.golang.publishWithContextOnEnter
func publishWithContextOnEnter(call api.CallContext, client mqtt.Client, ctx context.Context, topic string, qos byte, retained bool, payload interface{}) {
	if !IsEnabled() {
		return
	}

	// Calculate payload size
	var bodySize int64
	switch p := payload.(type) {
	case []byte:
		bodySize = int64(len(p))
	case string:
		bodySize = int64(len(p))
	default:
		bodySize = 0
	}

	request := MqttPublishRequest{
		Topic:    topic,
		QoS:      qos,
		Retained: retained,
		BodySize: bodySize,
		Headers:  make(map[string]string),
	}

	// Use provided context as parent
	var attributes []attribute.KeyValue
	attributes = append(attributes,
		attribute.Int("messaging.mqtt.qos", int(qos)),
		attribute.Bool("messaging.mqtt.retained", retained),
	)

	ctx = MqttPublishInstrumenter.Start(ctx, request, trace.WithAttributes(attributes...))

	// Store context and request for OnExit
	data := make(map[string]interface{})
	data["ctx"] = ctx
	data["mqtt_publish_request"] = request
	call.SetData(data)
}

//go:linkname publishWithContextOnExit github.com/eclipse/paho.mqtt.golang.publishWithContextOnExit
func publishWithContextOnExit(call api.CallContext, token mqtt.Token) {
	// Same as publishOnExit
	publishOnExit(call, token)
}
