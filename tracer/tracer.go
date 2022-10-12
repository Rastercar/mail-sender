package tracer

import (
	"context"
	"log"
	"mailer-ms/config"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
	"go.opentelemetry.io/otel/trace"
)

var tp *tracesdk.TracerProvider

func Start(c *config.TracerConfig) error {
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(c.Url)))
	if err != nil {
		return err
	}

	tp = tracesdk.NewTracerProvider(
		tracesdk.WithBatcher(exp),
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(c.ServiceName),
		)),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}))

	return nil
}

func Stop(ctx context.Context) {
	if tp == nil {
		return
	}

	// Do not make the application hang when it is shutdown.
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	if err := tp.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}
}

// NewSpan returns a new span from the global tracer. Each resulting
// span should call `span.End()`, ideally with `defer span.End()`.
func NewSpan(ctx context.Context, tracerName, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	if opts == nil {
		return otel.Tracer(tracerName).Start(ctx, spanName)
	}

	return otel.Tracer(tracerName).Start(ctx, spanName, opts...)
}

// SpanFromContext returns the current span from a context. If you wish to avoid
// creating child spans for each operation and just rely on the parent span, use
// this function throughout the application. With such practise you will get
// flatter span tree as opposed to deeper version. You can always mix and match
// both functions.
func SpanFromContext(ctx context.Context) trace.Span {
	return trace.SpanFromContext(ctx)
}

// AddSpanTags adds a new tags to the span. It will appear under "Tags" section
// of the selected span. Use this if you think the tag and its value could be
// useful while debugging.
func AddSpanTags(span trace.Span, tags map[string]string) {
	list := make([]attribute.KeyValue, len(tags))

	var i int
	for k, v := range tags {
		list[i] = attribute.Key(k).String(v)
		i++
	}

	span.SetAttributes(list...)
}

// AddSpanEvents adds a new events to the span. It will appear under the "Logs"
// section of the selected span. Use this if the event could mean anything
// valuable while debugging.
func AddSpanEvent(span trace.Span, name string, attributes map[string]string) {
	list := make([]trace.EventOption, len(attributes))

	var i int
	for k, v := range attributes {
		list[i] = trace.WithAttributes(attribute.Key(k).String(v))
		i++
	}

	span.AddEvent(name, list...)
}

// AddSpanError adds a new event to the span. It will appear under the "Logs"
// section of the selected span. This is not going to flag the span as "failed".
// Use this if you think you should log any exceptions such as critical, error,
// warning, caution etc. Avoid logging sensitive data!
func AddSpanError(span trace.Span, err error) {
	span.RecordError(err)
}

// FailSpan flags the span as "failed" and adds "error" label on listed trace.
// Use this after calling the `AddSpanError` function so that there is some sort
// of relevant exception logged against it.
func FailSpan(span trace.Span, msg string) {
	span.SetStatus(codes.Error, msg)
}

// AddSpanErrorAndFail adds a new event to the span. It will appear under the "Logs"
// section of the selected span. This is going to flag the span as "failed".
func AddSpanErrorAndFail(span trace.Span, err error, msg string) {
	span.RecordError(err)
	span.SetStatus(codes.Error, msg)
}
