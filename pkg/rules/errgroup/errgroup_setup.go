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

package errgroup

import (
	"context"
	"os"
	"runtime"
	"strings"
	_ "unsafe"

	"github.com/alibaba/loongsuite-go-agent/pkg/api"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const (
	errgroupScopeName = "loongsuite.instrumentation.errgroup"
)

var (
	errgroupTracer  = otel.Tracer(errgroupScopeName)
	errgroupEnabled = true
)

func init() {
	if val := os.Getenv("OTEL_INSTRUMENTATION_ERRGROUP_ENABLED"); val != "" {
		errgroupEnabled = strings.ToLower(val) != "false"
	}
}

// errgroupGoOnEnter is called when errgroup.Group.Go() is invoked
// It wraps the function to automatically create a child span
//
//go:linkname errgroupGoOnEnter golang.org/x/sync/errgroup.errgroupGoOnEnter
func errgroupGoOnEnter(call api.CallContext, g interface{}, f func() error) {
	if !errgroupEnabled {
		return
	}

	// Get the context from the errgroup (if it was created with WithContext)
	ctx := getErrgroupContext(g)
	if ctx == nil {
		ctx = context.Background()
	}

	// Get caller information for span name
	callerInfo := getCallerInfo(3) // Skip: errgroupGoOnEnter -> Go -> caller

	// Wrap the function to create a span
	wrappedFunc := func() error {
		// Start a child span for this goroutine task
		_, span := errgroupTracer.Start(ctx, "errgroup.task:"+callerInfo,
			trace.WithSpanKind(trace.SpanKindInternal),
			trace.WithAttributes(
				attribute.String("errgroup.task.caller", callerInfo),
				attribute.Bool("errgroup.task.async", true),
			),
		)
		defer span.End()

		// Execute the original function
		err := f()
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		return err
	}

	// Replace the function parameter with the wrapped version
	call.SetParam(1, wrappedFunc)
}

// errgroupTryGoOnEnter is called when errgroup.Group.TryGo() is invoked
//
//go:linkname errgroupTryGoOnEnter golang.org/x/sync/errgroup.errgroupTryGoOnEnter
func errgroupTryGoOnEnter(call api.CallContext, g interface{}, f func() error) {
	if !errgroupEnabled {
		return
	}

	ctx := getErrgroupContext(g)
	if ctx == nil {
		ctx = context.Background()
	}

	callerInfo := getCallerInfo(3)

	wrappedFunc := func() error {
		_, span := errgroupTracer.Start(ctx, "errgroup.trygo:"+callerInfo,
			trace.WithSpanKind(trace.SpanKindInternal),
			trace.WithAttributes(
				attribute.String("errgroup.task.caller", callerInfo),
				attribute.Bool("errgroup.task.async", true),
				attribute.Bool("errgroup.task.trygo", true),
			),
		)
		defer span.End()

		err := f()
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		return err
	}

	call.SetParam(1, wrappedFunc)
}

// errgroupWaitOnEnter is called when errgroup.Group.Wait() is invoked
//
//go:linkname errgroupWaitOnEnter golang.org/x/sync/errgroup.errgroupWaitOnEnter
func errgroupWaitOnEnter(call api.CallContext, g interface{}) {
	if !errgroupEnabled {
		return
	}

	ctx := getErrgroupContext(g)
	if ctx == nil {
		ctx = context.Background()
	}

	_, span := errgroupTracer.Start(ctx, "errgroup.wait",
		trace.WithSpanKind(trace.SpanKindInternal),
	)

	call.SetData(span)
}

// errgroupWaitOnExit is called after errgroup.Group.Wait() returns
//
//go:linkname errgroupWaitOnExit golang.org/x/sync/errgroup.errgroupWaitOnExit
func errgroupWaitOnExit(call api.CallContext, err error) {
	span, ok := call.GetData().(trace.Span)
	if !ok || span == nil {
		return
	}

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	span.End()
}

// getErrgroupContext extracts the context from an errgroup.Group
// The errgroup stores context internally when created with WithContext
func getErrgroupContext(g interface{}) context.Context {
	// Use reflection to access the internal ctx field
	// errgroup.Group has: ctx context.Context
	// This is a simplified approach - in production, use proper reflection
	if ctxProvider, ok := g.(interface{ Context() context.Context }); ok {
		return ctxProvider.Context()
	}
	return nil
}

// getCallerInfo returns the caller function name
func getCallerInfo(skip int) string {
	pc, _, _, ok := runtime.Caller(skip)
	if !ok {
		return "unknown"
	}
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return "unknown"
	}
	name := fn.Name()
	// Extract just the function name without full path
	if idx := strings.LastIndex(name, "/"); idx >= 0 {
		name = name[idx+1:]
	}
	return name
}
