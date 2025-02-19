package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

const (
	serviceName = "webhook"
)

var (
	tracer = otel.GetTracerProvider().Tracer(serviceName)
)

func initTracer() (*sdktrace.TracerProvider, error) {
	jaegerEndpoint := os.Getenv("JAEGER_ENDPOINT")
	if jaegerEndpoint == "" {
		jaegerEndpoint = "http://jaeger:14268/api/traces"
	}
	log.Printf("Initializing Jaeger exporter with endpoint: %s", jaegerEndpoint)

	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(jaegerEndpoint)))
	if err != nil {
		return nil, err
	}

	// 創建自定義的 SpanProcessor
	customProcessor := sdktrace.NewSimpleSpanProcessor(
		&loggingExporter{
			wrapped: exp,
		},
	)

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSpanProcessor(customProcessor),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(serviceName),
		)),
		sdktrace.WithSampler(sdktrace.AlwaysSample()), // 強制採樣所有追蹤
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return tp, nil
}

// loggingExporter 包裝 Jaeger exporter 並添加日誌
type loggingExporter struct {
	wrapped sdktrace.SpanExporter
}

func (e *loggingExporter) ExportSpans(ctx context.Context, spans []sdktrace.ReadOnlySpan) error {
	for _, span := range spans {
		log.Printf("[Webhook] Exporting span to Jaeger: TraceID=%s SpanID=%s Name=%s",
			span.SpanContext().TraceID(),
			span.SpanContext().SpanID(),
			span.Name())
	}
	return e.wrapped.ExportSpans(ctx, spans)
}

func (e *loggingExporter) Shutdown(ctx context.Context) error {
	return e.wrapped.Shutdown(ctx)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"service": serviceName,
	})
}

func webhookHandler(w http.ResponseWriter, r *http.Request) {
	// 記錄收到的請求頭
	log.Printf("Received headers: %v", r.Header)

	ctx := r.Context()
	propagator := otel.GetTextMapPropagator()
	ctx = propagator.Extract(ctx, propagation.HeaderCarrier(r.Header))

	ctx, span := tracer.Start(ctx, "webhook.handle")
	defer span.End()

	spanContext := span.SpanContext()
	log.Printf("Webhook Span: TraceID=%s SpanID=%s IsSampled=%t",
		spanContext.TraceID(),
		spanContext.SpanID(),
		spanContext.IsSampled())

	span.SetAttributes(
		attribute.String("http.method", r.Method),
		attribute.String("http.url", r.URL.String()),
		attribute.String("http.host", r.Host),
	)

	// 處理請求
	response := map[string]interface{}{
		"status":   "success",
		"message":  "Webhook received",
		"service":  serviceName,
		"trace_id": spanContext.TraceID().String(),
		"span_id":  spanContext.SpanID().String(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	log.Printf("Webhook response sent with trace_id=%s span_id=%s",
		spanContext.TraceID(),
		spanContext.SpanID())
}

func main() {
	tp, err := initTracer()
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}
	}()

	tracer = tp.Tracer(serviceName)

	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/", webhookHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	log.Printf("Webhook server starting on port %s...", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
