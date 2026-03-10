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

import "os"

// SolacePublishInstrumenter is the global instrumenter for Solace publish operations
var SolacePublishInstrumenter = BuildSolacePublishInstrumenter()

// SolaceConsumeInstrumenter is the global instrumenter for Solace consume operations
var SolaceConsumeInstrumenter = BuildSolaceConsumeInstrumenter()

// SolaceAckInstrumenter is the global instrumenter for Solace ack operations
var SolaceAckInstrumenter = BuildSolaceAckInstrumenter()

// solaceEnabler controls whether Solace instrumentation is enabled
type solaceEnabler struct {
	enabled bool
}

func (e solaceEnabler) Enable() bool {
	return e.enabled
}

// SolaceInstrumentationEnabled checks if Solace instrumentation is enabled via environment variable
var solaceEnabled = solaceEnabler{os.Getenv("OTEL_INSTRUMENTATION_SOLACE_ENABLED") != "false"}

// IsEnabled returns whether Solace instrumentation is enabled
func IsEnabled() bool {
	return solaceEnabled.Enable()
}
