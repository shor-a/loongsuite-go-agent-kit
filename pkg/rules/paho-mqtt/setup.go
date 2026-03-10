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

import "os"

// MqttPublishInstrumenter is the global instrumenter for MQTT publish operations
var MqttPublishInstrumenter = BuildMqttPublishInstrumenter()

// MqttSubscribeInstrumenter is the global instrumenter for MQTT subscribe/receive operations
var MqttSubscribeInstrumenter = BuildMqttSubscribeInstrumenter()

// mqttEnabler controls whether MQTT instrumentation is enabled
type mqttEnabler struct {
	enabled bool
}

func (e mqttEnabler) Enable() bool {
	return e.enabled
}

// MqttInstrumentationEnabled checks if MQTT instrumentation is enabled via environment variable
var mqttEnabled = mqttEnabler{os.Getenv("OTEL_INSTRUMENTATION_MQTT_ENABLED") != "false"}

// IsEnabled returns whether MQTT instrumentation is enabled
func IsEnabled() bool {
	return mqttEnabled.Enable()
}
