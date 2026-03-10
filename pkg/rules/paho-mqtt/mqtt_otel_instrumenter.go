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
	"fmt"

	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api-semconv/instrumenter/message"
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api/instrumenter"
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api/utils"
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api/version"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/instrumentation"
)

// MqttPublishAttrsGetter implements MessageAttrsGetter for MQTT publish operations
type MqttPublishAttrsGetter struct{}

var _ message.MessageAttrsGetter[MqttPublishRequest, any] = MqttPublishAttrsGetter{}

func (MqttPublishAttrsGetter) GetSystem(request MqttPublishRequest) string {
	return "mqtt"
}

func (MqttPublishAttrsGetter) GetDestination(request MqttPublishRequest) string {
	return request.Topic
}

func (MqttPublishAttrsGetter) GetDestinationTemplate(request MqttPublishRequest) string {
	return ""
}

func (MqttPublishAttrsGetter) IsTemporaryDestination(request MqttPublishRequest) bool {
	return false
}

func (MqttPublishAttrsGetter) IsAnonymousDestination(request MqttPublishRequest) bool {
	return false
}

func (MqttPublishAttrsGetter) GetConversationId(request MqttPublishRequest) string {
	return fmt.Sprintf("%d", request.MessageID)
}

func (MqttPublishAttrsGetter) GetMessageBodySize(request MqttPublishRequest) int64 {
	return request.BodySize
}

func (MqttPublishAttrsGetter) GetMessageEnvelopSize(request MqttPublishRequest) int64 {
	return 0
}

func (MqttPublishAttrsGetter) GetMessageId(request MqttPublishRequest, response any) string {
	return fmt.Sprintf("%d", request.MessageID)
}

func (MqttPublishAttrsGetter) GetClientId(request MqttPublishRequest) string {
	return request.ClientID
}

func (MqttPublishAttrsGetter) GetBatchMessageCount(request MqttPublishRequest, response any) int64 {
	return 1
}

func (MqttPublishAttrsGetter) GetMessageHeader(request MqttPublishRequest, name string) []string {
	if request.Headers == nil {
		return []string{}
	}
	if val, ok := request.Headers[name]; ok {
		return []string{val}
	}
	return []string{}
}

func (MqttPublishAttrsGetter) GetDestinationPartitionId(request MqttPublishRequest) string {
	return ""
}

// MqttSubscribeAttrsGetter implements MessageAttrsGetter for MQTT subscribe/receive operations
type MqttSubscribeAttrsGetter struct{}

var _ message.MessageAttrsGetter[MqttSubscribeRequest, any] = MqttSubscribeAttrsGetter{}

func (MqttSubscribeAttrsGetter) GetSystem(request MqttSubscribeRequest) string {
	return "mqtt"
}

func (MqttSubscribeAttrsGetter) GetDestination(request MqttSubscribeRequest) string {
	return request.Topic
}

func (MqttSubscribeAttrsGetter) GetDestinationTemplate(request MqttSubscribeRequest) string {
	return ""
}

func (MqttSubscribeAttrsGetter) IsTemporaryDestination(request MqttSubscribeRequest) bool {
	return false
}

func (MqttSubscribeAttrsGetter) IsAnonymousDestination(request MqttSubscribeRequest) bool {
	return false
}

func (MqttSubscribeAttrsGetter) GetConversationId(request MqttSubscribeRequest) string {
	return fmt.Sprintf("%d", request.MessageID)
}

func (MqttSubscribeAttrsGetter) GetMessageBodySize(request MqttSubscribeRequest) int64 {
	return request.BodySize
}

func (MqttSubscribeAttrsGetter) GetMessageEnvelopSize(request MqttSubscribeRequest) int64 {
	return 0
}

func (MqttSubscribeAttrsGetter) GetMessageId(request MqttSubscribeRequest, response any) string {
	return fmt.Sprintf("%d", request.MessageID)
}

func (MqttSubscribeAttrsGetter) GetClientId(request MqttSubscribeRequest) string {
	return request.ClientID
}

func (MqttSubscribeAttrsGetter) GetBatchMessageCount(request MqttSubscribeRequest, response any) int64 {
	return 1
}

func (MqttSubscribeAttrsGetter) GetMessageHeader(request MqttSubscribeRequest, name string) []string {
	if request.Headers == nil {
		return []string{}
	}
	if val, ok := request.Headers[name]; ok {
		return []string{val}
	}
	return []string{}
}

func (MqttSubscribeAttrsGetter) GetDestinationPartitionId(request MqttSubscribeRequest) string {
	return ""
}

// publishCarrier implements propagation.TextMapCarrier for MQTT publish
type publishCarrier struct {
	req *MqttPublishRequest
}

func (c *publishCarrier) Get(key string) string {
	if c.req.Headers == nil {
		return ""
	}
	return c.req.Headers[key]
}

func (c *publishCarrier) Set(key, value string) {
	if c.req.Headers == nil {
		c.req.Headers = make(map[string]string)
	}
	c.req.Headers[key] = value
}

func (c *publishCarrier) Keys() []string {
	if c.req.Headers == nil {
		return []string{}
	}
	keys := make([]string, 0, len(c.req.Headers))
	for k := range c.req.Headers {
		keys = append(keys, k)
	}
	return keys
}

// subscribeCarrier implements propagation.TextMapCarrier for MQTT subscribe
type subscribeCarrier struct {
	req MqttSubscribeRequest
}

func (c *subscribeCarrier) Get(key string) string {
	if c.req.Headers == nil {
		return ""
	}
	return c.req.Headers[key]
}

func (c *subscribeCarrier) Set(key, value string) {
	// Subscribe carrier is read-only for extraction
}

func (c *subscribeCarrier) Keys() []string {
	if c.req.Headers == nil {
		return []string{}
	}
	keys := make([]string, 0, len(c.req.Headers))
	for k := range c.req.Headers {
		keys = append(keys, k)
	}
	return keys
}

// BuildMqttPublishInstrumenter creates the instrumenter for MQTT publish operations
func BuildMqttPublishInstrumenter() *instrumenter.PropagatingToDownstreamInstrumenter[MqttPublishRequest, any] {
	builder := instrumenter.Builder[MqttPublishRequest, any]{}
	return builder.Init().
		SetSpanNameExtractor(&message.MessageSpanNameExtractor[MqttPublishRequest, any]{
			Getter:        MqttPublishAttrsGetter{},
			OperationName: message.PUBLISH,
		}).
		SetSpanKindExtractor(&instrumenter.AlwaysProducerExtractor[MqttPublishRequest]{}).
		AddAttributesExtractor(&message.MessageAttrsExtractor[MqttPublishRequest, any, MqttPublishAttrsGetter]{
			Operation: message.PUBLISH,
		}).
		SetInstrumentationScope(instrumentation.Scope{
			Name:    utils.PAHO_MQTT_SCOPE_NAME,
			Version: version.Tag,
		}).
		BuildPropagatingToDownstreamInstrumenter(func(req MqttPublishRequest) propagation.TextMapCarrier {
			return &publishCarrier{req: &req}
		}, otel.GetTextMapPropagator())
}

// BuildMqttSubscribeInstrumenter creates the instrumenter for MQTT subscribe/receive operations
func BuildMqttSubscribeInstrumenter() *instrumenter.PropagatingFromUpstreamInstrumenter[MqttSubscribeRequest, any] {
	builder := instrumenter.Builder[MqttSubscribeRequest, any]{}
	return builder.Init().
		SetSpanNameExtractor(&message.MessageSpanNameExtractor[MqttSubscribeRequest, any]{
			Getter:        MqttSubscribeAttrsGetter{},
			OperationName: message.RECEIVE,
		}).
		SetSpanKindExtractor(&instrumenter.AlwaysConsumerExtractor[MqttSubscribeRequest]{}).
		AddAttributesExtractor(&message.MessageAttrsExtractor[MqttSubscribeRequest, any, MqttSubscribeAttrsGetter]{
			Operation: message.RECEIVE,
		}).
		SetInstrumentationScope(instrumentation.Scope{
			Name:    utils.PAHO_MQTT_SCOPE_NAME,
			Version: version.Tag,
		}).
		BuildPropagatingFromUpstreamInstrumenter(func(req MqttSubscribeRequest) propagation.TextMapCarrier {
			return &subscribeCarrier{req: req}
		}, otel.GetTextMapPropagator())
}
