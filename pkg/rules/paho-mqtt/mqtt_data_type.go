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

// MqttPublishRequest represents an MQTT publish operation for tracing
type MqttPublishRequest struct {
	Topic      string
	QoS        byte
	Retained   bool
	MessageID  uint16
	BodySize   int64
	BrokerURL  string
	ClientID   string
	Headers    map[string]string // For trace context propagation (embedded in payload)
}

// MqttSubscribeRequest represents an MQTT message receive operation for tracing
type MqttSubscribeRequest struct {
	Topic      string
	QoS        byte
	MessageID  uint16
	BodySize   int64
	BrokerURL  string
	ClientID   string
	Headers    map[string]string // Extracted from payload for trace context
}

// MqttConnectRequest represents an MQTT connection operation
type MqttConnectRequest struct {
	BrokerURL string
	ClientID  string
	Username  string
}
