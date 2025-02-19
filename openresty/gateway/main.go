package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

const (
	serviceName = "gateway"
)

var (
	tracer = otel.GetTracerProvider().Tracer(serviceName)
)

// ServiceRoute 定義服務路由
type ServiceRoute struct {
	Path        string
	ServiceHost string
}

// 服務路由配置
var serviceRoutes = []ServiceRoute{
	{
		Path:        "/webhook",
		ServiceHost: "http://webhook:8081",
	},
	{
		Path:        "/",
		ServiceHost: "http://webhook:8081",
	},
}

func initTracer() (*sdktrace.TracerProvider, error) {
	jaegerEndpoint := os.Getenv("JAEGER_ENDPOINT")
	if jaegerEndpoint == "" {
		jaegerEndpoint = "http://jaeger:14268/api/traces"
	}
	log.Printf("Initializing Jaeger exporter with endpoint: %s", jaegerEndpoint)

	exp, err := jaeger.New(
		jaeger.WithCollectorEndpoint(
			jaeger.WithEndpoint(jaegerEndpoint),
			jaeger.WithUsername(""),
			jaeger.WithPassword(""),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create jaeger exporter: %v", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(serviceName),
		)),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
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
		log.Printf("[Gateway] Exporting span to Jaeger: TraceID=%s SpanID=%s ParentSpanID=%s Name=%s",
			span.SpanContext().TraceID(),
			span.SpanContext().SpanID(),
			span.Parent().SpanID(),
			span.Name())
	}
	return e.wrapped.ExportSpans(ctx, spans)
}

func (e *loggingExporter) Shutdown(ctx context.Context) error {
	return e.wrapped.Shutdown(ctx)
}

// parseW3CTraceContext 解析 W3C Trace Context
func parseW3CTraceContext(traceparent string) (trace.TraceID, trace.SpanID, trace.TraceFlags, error) {
	parts := strings.Split(traceparent, "-")
	if len(parts) != 4 {
		return trace.TraceID{}, trace.SpanID{}, 0, fmt.Errorf("invalid traceparent format")
	}

	traceID, err := hex.DecodeString(parts[1])
	if err != nil {
		return trace.TraceID{}, trace.SpanID{}, 0, fmt.Errorf("invalid trace ID: %w", err)
	}
	var tid trace.TraceID
	copy(tid[:], traceID)

	parentID, err := hex.DecodeString(parts[2])
	if err != nil {
		return trace.TraceID{}, trace.SpanID{}, 0, fmt.Errorf("invalid parent ID: %w", err)
	}
	var sid trace.SpanID
	copy(sid[:], parentID)

	flags := trace.TraceFlags(0)
	if parts[3] == "01" {
		flags = trace.FlagsSampled
	}

	return tid, sid, flags, nil
}

// createReverseProxy 創建反向代理
func createReverseProxy(target *url.URL) *httputil.ReverseProxy {
	proxy := httputil.NewSingleHostReverseProxy(target)

	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.Host = target.Host
	}

	return proxy
}

// healthHandler 處理健康檢查請求
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"service": serviceName,
	})
}

// findRoute 查找匹配的路由
func findRoute(path string) *ServiceRoute {
	log.Printf("Finding route for path: %s", path)

	// 首先嘗試精確匹配
	for _, route := range serviceRoutes {
		if path == route.Path {
			log.Printf("Found exact match route: %s -> %s", path, route.ServiceHost)
			return &route
		}
	}

	// 然後嘗試前綴匹配
	for _, route := range serviceRoutes {
		if strings.HasPrefix(path, route.Path) && route.Path != "/" {
			log.Printf("Found prefix match route: %s -> %s", path, route.ServiceHost)
			return &route
		}
	}

	// 最後嘗試根路徑
	for _, route := range serviceRoutes {
		if route.Path == "/" {
			log.Printf("Using root route: %s -> %s", path, route.ServiceHost)
			return &route
		}
	}

	log.Printf("No route found for path: %s", path)
	return nil
}

// proxyHandler 處理代理請求
func proxyHandler(w http.ResponseWriter, r *http.Request) {
	// 檢查是否是健康檢查請求
	if r.URL.Path == "/health" {
		healthHandler(w, r)
		return
	}

	// 從請求中獲取 traceparent
	ctx := r.Context()
	propagator := otel.GetTextMapPropagator()
	ctx = propagator.Extract(ctx, propagation.HeaderCarrier(r.Header))

	ctx, span := tracer.Start(ctx, "gateway.proxy")
	defer span.End()

	spanContext := span.SpanContext()
	log.Printf("Gateway Span: TraceID=%s SpanID=%s IsSampled=%t",
		spanContext.TraceID(),
		spanContext.SpanID(),
		spanContext.IsSampled())

	span.SetAttributes(
		attribute.String("http.method", r.Method),
		attribute.String("http.url", r.URL.String()),
		attribute.String("http.host", r.Host),
	)

	// 查找匹配的路由
	route := findRoute(r.URL.Path)
	if route == nil {
		http.Error(w, "Service not found", http.StatusNotFound)
		return
	}

	targetURL, err := url.Parse(route.ServiceHost)
	if err != nil {
		http.Error(w, "Service route configuration error", http.StatusInternalServerError)
		return
	}

	log.Printf("Routing request %s to %s", r.URL.Path, targetURL.String())

	// 創建反向代理
	proxy := createReverseProxy(targetURL)

	// 注入 trace context 到請求頭
	r = r.WithContext(ctx)
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(r.Header))

	// 記錄注入的 trace context
	log.Printf("Injected trace context headers: %v", r.Header)

	// 執行代理請求
	proxy.ServeHTTP(w, r)
}

func main() {
	// 初始化 tracer
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

	http.HandleFunc("/", proxyHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Gateway server starting on port %s...", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
