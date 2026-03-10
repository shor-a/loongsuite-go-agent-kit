//go:build ignore

// Copyright (c) 2024 Alibaba Group Holding Ltd.
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
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	trace "go.opentelemetry.io/otel/trace"
)

const maxSpans = 300

// ============================================================================
// Goroutine Auto-Span Feature
// ============================================================================

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

// TakeSnapShot implements ContextSnapshoter interface for asyncTraceContext
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

// ============================================================================
// Core Trace Context
// ============================================================================

type traceContext struct {
	sw  *spanWrapper
	n   int
	lcs trace.Span
}

type spanWrapper struct {
	span trace.Span
	prev *spanWrapper
}

func (tc *traceContext) size() int {
	return tc.n
}

func (tc *traceContext) add(span trace.Span) bool {
	if tc.n > 0 {
		if tc.n >= maxSpans {
			return false
		}
	}
	wrapper := &spanWrapper{span, tc.sw}
	// local root span
	if tc.n == 0 {
		tc.lcs = span
	}
	tc.sw = wrapper
	tc.n++
	return true
}

//go:norace
func (tc *traceContext) tail() trace.Span {
	if tc.n == 0 {
		return nil
	} else {
		return tc.sw.span
	}
}

func (tc *traceContext) localRootSpan() trace.Span {
	if tc.n == 0 {
		return nil
	} else {
		return tc.lcs
	}
}

func (tc *traceContext) del(span trace.Span) {
	if tc.n == 0 {
		return
	}
	addr := &tc.sw
	cur := tc.sw
	for cur != nil {
		sc1 := cur.span.SpanContext()
		sc2 := span.SpanContext()
		if sc1.TraceID() == sc2.TraceID() && sc1.SpanID() == sc2.SpanID() {
			*addr = cur.prev
			tc.n--
			break
		}
		addr = &cur.prev
		cur = cur.prev
	}
}

func (tc *traceContext) clear() {
	tc.sw = nil
	tc.n = 0
	SetBaggageContainerToGLS(nil)
}

//go:norace
func (tc *traceContext) TakeSnapShot() interface{} {
	// take a deep copy to avoid reading & writing the same map at the same time
	if tc.n == 0 {
		return &traceContext{nil, 0, nil}
	}
	last := tc.tail()

	// If automatic goroutine span creation is enabled, return an asyncTraceContext
	// that will create a child span when the goroutine first accesses the trace context
	if goroutineAutoSpanEnabled && last != nil {
		callerInfo := getGoroutineCallerInfo(3) // Skip: TakeSnapShot -> contextPropagate -> newproc1
		return &asyncTraceContext{
			parentSpanCtx:   last.SpanContext(),
			callerInfo:      callerInfo,
			autoSpanCreated: false,
			autoSpan:        nil,
		}
	}

	sw := &spanWrapper{last, nil}
	return &traceContext{sw, 1, nil}
}

func GetGLocalData(key string) interface{} {
	return nil
}

func SetGLocalData(key string, value interface{}) {
	t := getOrInitTraceContext()
	setTraceContext(t)
}

func getOrInitTraceContext() *traceContext {
	tc := GetTraceContextFromGLS()
	if tc == nil {
		newTc := &traceContext{nil, 0, nil}
		setTraceContext(newTc)
		return newTc
	} else {
		return tc.(*traceContext)
	}
}

func setTraceContext(tc *traceContext) {
	SetTraceContextToGLS(tc)
}

func traceContextAddSpan(span trace.Span) {
	tc := getOrInitTraceContext()
	if !tc.add(span) {
		fmt.Println("Failed to add span to TraceContext")
	}
}

func GetTraceAndSpanId() (string, string) {
	tc := GetTraceContextFromGLS()
	if tc == nil {
		return "", ""
	}

	// Handle asyncTraceContext
	if atc, ok := tc.(*asyncTraceContext); ok {
		spanCtx := atc.getSpanContext()
		if spanCtx.IsValid() {
			return spanCtx.TraceID().String(), spanCtx.SpanID().String()
		}
		return "", ""
	}

	if tc.(*traceContext).tail() == nil {
		return "", ""
	}
	ctx := tc.(*traceContext).tail().SpanContext()
	return ctx.TraceID().String(), ctx.SpanID().String()
}

func traceContextDelSpan(span trace.Span) {
	ctx := getOrInitTraceContext()
	ctx.del(span)
}

func clearTraceContext() {
	getOrInitTraceContext().clear()
}

func SpanFromGLS() trace.Span {
	gls := GetTraceContextFromGLS()
	if gls == nil {
		return nil
	}

	// Handle asyncTraceContext - create auto span if enabled
	if atc, ok := gls.(*asyncTraceContext); ok {
		if goroutineAutoSpanEnabled {
			return atc.ensureAutoSpan()
		}
		// If auto span not enabled, return nil (no span in this goroutine yet)
		return nil
	}

	return gls.(*traceContext).tail()
}

func LocalRootSpanFromGLS() trace.Span {
	gls := GetTraceContextFromGLS()
	if gls == nil {
		return nil
	}

	// Handle asyncTraceContext
	if atc, ok := gls.(*asyncTraceContext); ok {
		if goroutineAutoSpanEnabled && atc.autoSpanCreated {
			return atc.autoSpan
		}
		return nil
	}

	return gls.(*traceContext).lcs
}
