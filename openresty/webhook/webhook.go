package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

type TraceContext struct {
	Version    string
	TraceID    string
	ParentID   string
	TraceFlags string
}

func parseTraceparent(tp string) (*TraceContext, error) {
	parts := strings.Split(tp, "-")
	if len(parts) != 4 {
		return nil, fmt.Errorf("invalid traceparent format")
	}

	return &TraceContext{
		Version:    parts[0],
		TraceID:    parts[1],
		ParentID:   parts[2],
		TraceFlags: parts[3],
	}, nil
}

func handler(w http.ResponseWriter, r *http.Request) {
	tp := r.Header.Get("Traceparent")
	traceCtx, err := parseTraceparent(tp)
	
	if err != nil {
		log.Printf("Error parsing traceparent: %v", err)
	} else {
		log.Printf("Trace Context - Version: %s, TraceID: %s, ParentID: %s, Flags: %s",
			traceCtx.Version, traceCtx.TraceID, traceCtx.ParentID, traceCtx.TraceFlags)
	}

	// 設置響應 header
	w.Header().Set("Content-Type", "application/json")
	if traceCtx != nil {
		w.Header().Set("Traceparent", tp)
	}

	fmt.Fprintf(w, `{"status": "success", "trace_id": "%s"}`, 
		func() string {
			if traceCtx != nil {
				return traceCtx.TraceID
			}
			return ""
		}())
}

func main() {
	http.HandleFunc("/", handler)
	log.Println("Starting webhook on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
