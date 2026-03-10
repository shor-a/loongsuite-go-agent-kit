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
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api-semconv/instrumenter/message"
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api/instrumenter"
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api/utils"
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api/version"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/instrumentation"
)

// SolacePublishAttrsGetter implements MessageAttrsGetter for Solace publish operations
type SolacePublishAttrsGetter struct{}

var _ message.MessageAttrsGetter[SolacePublishRequest, any] = SolacePublishAttrsGetter{}

func (SolacePublishAttrsGetter) GetSystem(request SolacePublishRequest) string {
	return "solace"
}

func (SolacePublishAttrsGetter) GetDestination(request SolacePublishRequest) string {
	return request.Topic
}

func (SolacePublishAttrsGetter) GetDestinationTemplate(request SolacePublishRequest) string {
	return ""
}

func (SolacePublishAttrsGetter) IsTemporaryDestination(request SolacePublishRequest) bool {
	return false
}

func (SolacePublishAttrsGetter) IsAnonymousDestination(request SolacePublishRequest) bool {
	return false
}

func (SolacePublishAttrsGetter) GetConversationId(request SolacePublishRequest) string {
	return request.CorrelationID
}

func (SolacePublishAttrsGetter) GetMessageBodySize(request SolacePublishRequest) int64 {
	return request.BodySize
}

func (SolacePublishAttrsGetter) GetMessageEnvelopSize(request SolacePublishRequest) int64 {
	return 0
}

func (SolacePublishAttrsGetter) GetMessageId(request SolacePublishRequest, response any) string {
	return request.MessageID
}

func (SolacePublishAttrsGetter) GetClientId(request SolacePublishRequest) string {
	return ""
}

func (SolacePublishAttrsGetter) GetBatchMessageCount(request SolacePublishRequest, response any) int64 {
	return 1
}

func (SolacePublishAttrsGetter) GetMessageHeader(request SolacePublishRequest, name string) []string {
	if request.Headers == nil {
		return []string{}
	}
	if val, ok := request.Headers[name]; ok {
		return []string{val}
	}
	return []string{}
}

func (SolacePublishAttrsGetter) GetDestinationPartitionId(request SolacePublishRequest) string {
	return ""
}

// SolaceConsumeAttrsGetter implements MessageAttrsGetter for Solace consume operations
type SolaceConsumeAttrsGetter struct{}

var _ message.MessageAttrsGetter[SolaceConsumeRequest, any] = SolaceConsumeAttrsGetter{}

func (SolaceConsumeAttrsGetter) GetSystem(request SolaceConsumeRequest) string {
	return "solace"
}

func (SolaceConsumeAttrsGetter) GetDestination(request SolaceConsumeRequest) string {
	return request.Topic
}

func (SolaceConsumeAttrsGetter) GetDestinationTemplate(request SolaceConsumeRequest) string {
	return ""
}

func (SolaceConsumeAttrsGetter) IsTemporaryDestination(request SolaceConsumeRequest) bool {
	return false
}

func (SolaceConsumeAttrsGetter) IsAnonymousDestination(request SolaceConsumeRequest) bool {
	return false
}

func (SolaceConsumeAttrsGetter) GetConversationId(request SolaceConsumeRequest) string {
	return request.CorrelationID
}

func (SolaceConsumeAttrsGetter) GetMessageBodySize(request SolaceConsumeRequest) int64 {
	return request.BodySize
}

func (SolaceConsumeAttrsGetter) GetMessageEnvelopSize(request SolaceConsumeRequest) int64 {
	return 0
}

func (SolaceConsumeAttrsGetter) GetMessageId(request SolaceConsumeRequest, response any) string {
	return request.MessageID
}

func (SolaceConsumeAttrsGetter) GetClientId(request SolaceConsumeRequest) string {
	return ""
}

func (SolaceConsumeAttrsGetter) GetBatchMessageCount(request SolaceConsumeRequest, response any) int64 {
	return 1
}

func (SolaceConsumeAttrsGetter) GetMessageHeader(request SolaceConsumeRequest, name string) []string {
	if request.Headers == nil {
		return []string{}
	}
	if val, ok := request.Headers[name]; ok {
		return []string{val}
	}
	return []string{}
}

func (SolaceConsumeAttrsGetter) GetDestinationPartitionId(request SolaceConsumeRequest) string {
	return ""
}

// publishCarrier implements propagation.TextMapCarrier for Solace publish
type publishCarrier struct {
	req *SolacePublishRequest
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

// consumeCarrier implements propagation.TextMapCarrier for Solace consume
type consumeCarrier struct {
	req SolaceConsumeRequest
}

func (c *consumeCarrier) Get(key string) string {
	if c.req.Headers == nil {
		return ""
	}
	return c.req.Headers[key]
}

func (c *consumeCarrier) Set(key, value string) {
	// Consume carrier is read-only for extraction
}

func (c *consumeCarrier) Keys() []string {
	if c.req.Headers == nil {
		return []string{}
	}
	keys := make([]string, 0, len(c.req.Headers))
	for k := range c.req.Headers {
		keys = append(keys, k)
	}
	return keys
}

// BuildSolacePublishInstrumenter creates the instrumenter for Solace publish operations
func BuildSolacePublishInstrumenter() *instrumenter.PropagatingToDownstreamInstrumenter[SolacePublishRequest, any] {
	builder := instrumenter.Builder[SolacePublishRequest, any]{}
	return builder.Init().
		SetSpanNameExtractor(&message.MessageSpanNameExtractor[SolacePublishRequest, any]{
			Getter:        SolacePublishAttrsGetter{},
			OperationName: message.PUBLISH,
		}).
		SetSpanKindExtractor(&instrumenter.AlwaysProducerExtractor[SolacePublishRequest]{}).
		AddAttributesExtractor(&message.MessageAttrsExtractor[SolacePublishRequest, any, SolacePublishAttrsGetter]{
			Operation: message.PUBLISH,
		}).
		SetInstrumentationScope(instrumentation.Scope{
			Name:    utils.SOLACE_SCOPE_NAME,
			Version: version.Tag,
		}).
		BuildPropagatingToDownstreamInstrumenter(func(req SolacePublishRequest) propagation.TextMapCarrier {
			return &publishCarrier{req: &req}
		}, otel.GetTextMapPropagator())
}

// BuildSolaceConsumeInstrumenter creates the instrumenter for Solace consume operations
func BuildSolaceConsumeInstrumenter() *instrumenter.PropagatingFromUpstreamInstrumenter[SolaceConsumeRequest, any] {
	builder := instrumenter.Builder[SolaceConsumeRequest, any]{}
	return builder.Init().
		SetSpanNameExtractor(&message.MessageSpanNameExtractor[SolaceConsumeRequest, any]{
			Getter:        SolaceConsumeAttrsGetter{},
			OperationName: message.RECEIVE,
		}).
		SetSpanKindExtractor(&instrumenter.AlwaysConsumerExtractor[SolaceConsumeRequest]{}).
		AddAttributesExtractor(&message.MessageAttrsExtractor[SolaceConsumeRequest, any, SolaceConsumeAttrsGetter]{
			Operation: message.RECEIVE,
		}).
		SetInstrumentationScope(instrumentation.Scope{
			Name:    utils.SOLACE_SCOPE_NAME,
			Version: version.Tag,
		}).
		BuildPropagatingFromUpstreamInstrumenter(func(req SolaceConsumeRequest) propagation.TextMapCarrier {
			return &consumeCarrier{req: req}
		}, otel.GetTextMapPropagator())
}

// SolaceAckAttrsGetter implements MessageAttrsGetter for Solace ack operations
type SolaceAckAttrsGetter struct{}

var _ message.MessageAttrsGetter[SolaceConsumeRequest, any] = SolaceAckAttrsGetter{}

func (SolaceAckAttrsGetter) GetSystem(request SolaceConsumeRequest) string {
	return "solace"
}

func (SolaceAckAttrsGetter) GetDestination(request SolaceConsumeRequest) string {
	return request.Topic
}

func (SolaceAckAttrsGetter) GetDestinationTemplate(request SolaceConsumeRequest) string {
	return ""
}

func (SolaceAckAttrsGetter) IsTemporaryDestination(request SolaceConsumeRequest) bool {
	return false
}

func (SolaceAckAttrsGetter) IsAnonymousDestination(request SolaceConsumeRequest) bool {
	return false
}

func (SolaceAckAttrsGetter) GetConversationId(request SolaceConsumeRequest) string {
	return request.CorrelationID
}

func (SolaceAckAttrsGetter) GetMessageBodySize(request SolaceConsumeRequest) int64 {
	return 0
}

func (SolaceAckAttrsGetter) GetMessageEnvelopSize(request SolaceConsumeRequest) int64 {
	return 0
}

func (SolaceAckAttrsGetter) GetMessageId(request SolaceConsumeRequest, response any) string {
	return request.MessageID
}

func (SolaceAckAttrsGetter) GetClientId(request SolaceConsumeRequest) string {
	return ""
}

func (SolaceAckAttrsGetter) GetBatchMessageCount(request SolaceConsumeRequest, response any) int64 {
	return 1
}

func (SolaceAckAttrsGetter) GetMessageHeader(request SolaceConsumeRequest, name string) []string {
	return []string{}
}

func (SolaceAckAttrsGetter) GetDestinationPartitionId(request SolaceConsumeRequest) string {
	return ""
}

// BuildSolaceAckInstrumenter creates the instrumenter for Solace ack operations
func BuildSolaceAckInstrumenter() *instrumenter.Instrumenter[SolaceConsumeRequest, any] {
	builder := instrumenter.Builder[SolaceConsumeRequest, any]{}
	return builder.Init().
		SetSpanNameExtractor(&message.MessageSpanNameExtractor[SolaceConsumeRequest, any]{
			Getter:        SolaceAckAttrsGetter{},
			OperationName: message.SETTLE,
		}).
		SetSpanKindExtractor(&instrumenter.AlwaysClientExtractor[SolaceConsumeRequest]{}).
		AddAttributesExtractor(&message.MessageAttrsExtractor[SolaceConsumeRequest, any, SolaceAckAttrsGetter]{
			Operation: message.SETTLE,
		}).
		SetInstrumentationScope(instrumentation.Scope{
			Name:    utils.SOLACE_SCOPE_NAME,
			Version: version.Tag,
		}).
		BuildInstrumenter()
}
