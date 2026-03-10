//go:build ignore

// Copyright (c) 2026 Alibaba Group Holding Ltd.
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

package trace_context

import (
	"context"
	"os"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const (
	goroutineScopeName = "loongsuite.instrumentation.goroutine"
)

var (
	// goroutineAutoSpanEnabled controls automatic span creation for goroutines
	// Default: false (disabled) - set OTEL_INSTRUMENTATION_GOROUTINE_AUTO_SPAN=true to enable
	goroutineAutoSpanEnabled = false

	// goroutineAutoSpanCount tracks the number of auto-created goroutine spans
	goroutineAutoSpanCount uint64

	// goroutineTracer is the tracer for goroutine spans
	goroutineTracer trace.Tracer

	// initOnce ensures tracer is initialized only once
	initOnce sync.Once
)

func init() {
	// Check environment variable to enable automatic goroutine span creation
	if val := os.Getenv("OTEL_INSTRUMENTATION_GOROUTINE_AUTO_SPAN"); val != "" {
		goroutineAutoSpanEnabled = strings.ToLower(val) == "true"
	}
}

// initGoroutineTracer initializes the goroutine tracer lazily
func initGoroutineTracer() {
	initOnce.Do(func() {
		goroutineTracer = otel.Tracer(goroutineScopeName)
	})
}

// asyncTraceContext is an enhanced trace context that supports automatic span creation
// for goroutines. When a goroutine is spawned, this context is copied to the new goroutine.
// On first access to the trace context in the new goroutine, a child span is automatically
// created if goroutineAutoSpanEnabled is true.
type asyncTraceContext struct {
	// parentSpanCtx is the span context from the parent goroutine
	parentSpanCtx trace.SpanContext

	// callerInfo contains information about where the goroutine was spawned
	callerInfo string

	// autoSpanCreated indicates if an automatic span has been created
	autoSpanCreated bool

	// autoSpan is the automatically created span (if any)
	autoSpan trace.Span

	// mu protects the auto span creation
	mu sync.Mutex
}

// TakeSnapShot implements ContextSnapshoter interface
// This is called by runtime.newproc1 when a goroutine is spawned
func (atc *asyncTraceContext) TakeSnapShot() interface{} {
	if !atc.parentSpanCtx.IsValid() {
		return &asyncTraceContext{}
	}

	// Get caller info for the new goroutine
	callerInfo := getGoroutineCallerInfo(4) // Skip: TakeSnapShot -> contextPropagate -> newproc1 -> go statement

	return &asyncTraceContext{
		parentSpanCtx:   atc.parentSpanCtx,
		callerInfo:      callerInfo,
		autoSpanCreated: false,
		autoSpan:        nil,
	}
}

// ensureAutoSpan creates an automatic span for this goroutine if enabled
func (atc *asyncTraceContext) ensureAutoSpan() trace.Span {
	if !goroutineAutoSpanEnabled {
		return nil
	}

	atc.mu.Lock()
	defer atc.mu.Unlock()

	if atc.autoSpanCreated {
		return atc.autoSpan
	}

	if !atc.parentSpanCtx.IsValid() {
		return nil
	}

	initGoroutineTracer()

	// Create span name from caller info
	spanName := "go:" + extractFunctionName(atc.callerInfo)

	// Create a context with the parent span context
	parentCtx := trace.ContextWithSpanContext(context.Background(), atc.parentSpanCtx)

	// Start a new child span
	_, span := goroutineTracer.Start(parentCtx, spanName,
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(
			attribute.String("goroutine.caller", atc.callerInfo),
			attribute.Bool("goroutine.auto", true),
		),
	)

	atc.autoSpan = span
	atc.autoSpanCreated = true
	atomic.AddUint64(&goroutineAutoSpanCount, 1)

	return span
}

// getSpanContext returns the current span context
// If auto span was created, returns its context; otherwise returns parent context
func (atc *asyncTraceContext) getSpanContext() trace.SpanContext {
	if atc.autoSpanCreated && atc.autoSpan != nil {
		return atc.autoSpan.SpanContext()
	}
	return atc.parentSpanCtx
}

// endAutoSpan ends the automatically created span
func (atc *asyncTraceContext) endAutoSpan() {
	atc.mu.Lock()
	defer atc.mu.Unlock()

	if atc.autoSpanCreated && atc.autoSpan != nil {
		atc.autoSpan.End()
		atc.autoSpan = nil
	}
}

// getGoroutineCallerInfo returns the caller function name for goroutine creation
func getGoroutineCallerInfo(skip int) string {
	pc, _, _, ok := runtime.Caller(skip)
	if !ok {
		return "unknown"
	}
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return "unknown"
	}
	return fn.Name()
}

// extractFunctionName extracts a clean function name from full path
func extractFunctionName(fullName string) string {
	if fullName == "" || fullName == "unknown" {
		return "goroutine"
	}

	// Remove file path prefix, keep only package.function
	parts := strings.Split(fullName, "/")
	if len(parts) > 0 {
		lastPart := parts[len(parts)-1]
		// Handle method receivers: (*Type).method -> Type.method
		lastPart = strings.ReplaceAll(lastPart, "(*", "")
		lastPart = strings.ReplaceAll(lastPart, ")", "")
		if lastPart != "" {
			return lastPart
		}
	}

	return "goroutine"
}

// GetGoroutineAutoSpanCount returns the number of auto-created goroutine spans
func GetGoroutineAutoSpanCount() uint64 {
	return atomic.LoadUint64(&goroutineAutoSpanCount)
}

// IsGoroutineAutoSpanEnabled returns whether automatic goroutine span creation is enabled
func IsGoroutineAutoSpanEnabled() bool {
	return goroutineAutoSpanEnabled
}
